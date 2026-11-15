package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"desktop/internal/models"
)

type fakeIndexRepository struct {
	targets        []models.IndexTarget
	listErr        error
	statusErr      error
	persistErr     error
	recalculateErr error
	overview       models.IndexingOverview
	statuses       []string
	recalculates   int
}

func (r *fakeIndexRepository) ListTargets(context.Context, bool) ([]models.IndexTarget, error) {
	return r.targets, r.listErr
}

func (r *fakeIndexRepository) MarkStatus(_ context.Context, _ uint64, status string) error {
	r.statuses = append(r.statuses, status)
	return r.statusErr
}

func (r *fakeIndexRepository) PersistDocument(context.Context, models.IndexedDocument) error {
	return r.persistErr
}

func (r *fakeIndexRepository) RecalculateDocumentFrequency(context.Context) error {
	r.recalculates++
	return r.recalculateErr
}

func (r *fakeIndexRepository) Overview(context.Context) (models.IndexingOverview, error) {
	return r.overview, nil
}

func TestIndexingServiceRunHandlesEmptyAndInvalidTargets(t *testing.T) {
	ctx := context.Background()
	emptyRepo := &fakeIndexRepository{}
	empty, err := newIndexingService(emptyRepo).Run(ctx, 0, false)
	if err != nil || empty.Total != 0 || empty.Workers != 1 || emptyRepo.recalculates != 0 {
		t.Fatalf("Run vazio inesperado: report=%+v recalculates=%d erro=%v", empty, emptyRepo.recalculates, err)
	}

	targetRepo := &fakeIndexRepository{
		targets: []models.IndexTarget{{
			ID:       42,
			Title:    "TESTE artigo",
			FilePath: filepath.Join(t.TempDir(), "TESTE_inexistente.pdf"),
		}},
	}
	report, err := newIndexingService(targetRepo).Run(ctx, 2, false)
	if err != nil {
		t.Fatalf("Run deveria reportar falha do alvo sem erro global: %v", err)
	}
	if report.Total != 1 || report.Failed != 1 || report.Indexed != 0 || len(report.Errors) != 1 {
		t.Fatalf("Run não reportou falha do alvo: %+v", report)
	}
	if len(targetRepo.statuses) != 2 || targetRepo.statuses[0] != "PROCESSING" || targetRepo.statuses[1] != "ERROR" {
		t.Fatalf("status do artigo inválido: %v", targetRepo.statuses)
	}
	if targetRepo.recalculates != 1 {
		t.Fatalf("Run deveria recalcular DF após processar alvos; chamadas=%d", targetRepo.recalculates)
	}
}

func TestIndexingServicePropagatesRepositoryErrors(t *testing.T) {
	want := errors.New("TESTE erro de banco")
	service := newIndexingService(&fakeIndexRepository{listErr: want})
	if _, err := service.Run(context.Background(), 1, false); !errors.Is(err, want) {
		t.Fatalf("Run deveria propagar erro do repositório: %v", err)
	}
	if _, err := service.Benchmark(context.Background()); err == nil {
		t.Fatal("Benchmark deveria propagar erro do repositório")
	}

	overviewErrRepo := &fakeOverviewErrorRepository{err: want}
	if _, err := newIndexingService(overviewErrRepo).Overview(context.Background()); !errors.Is(err, want) {
		t.Fatalf("Overview deveria propagar erro do repositório: %v", err)
	}
	wantOverview := models.IndexingOverview{Pending: 2, Indexed: 3}
	overview, err := newIndexingService(&fakeIndexRepository{overview: wantOverview}).Overview(context.Background())
	if err != nil || overview != wantOverview {
		t.Fatalf("Overview não retornou os indicadores do repositório: overview=%+v erro=%v", overview, err)
	}
}

type fakeOverviewErrorRepository struct {
	fakeIndexRepository
	err error
}

func (r *fakeOverviewErrorRepository) Overview(context.Context) (models.IndexingOverview, error) {
	return models.IndexingOverview{}, r.err
}

func TestIndexingServiceBenchmarkReportsFourWorkerCounts(t *testing.T) {
	report, err := newIndexingService(&fakeIndexRepository{}).Benchmark(context.Background())
	if err != nil {
		t.Fatalf("Benchmark sem alvos falhou: %v", err)
	}
	if len(report.Results) != 4 {
		t.Fatalf("esperados quatro resultados, obtidos %d", len(report.Results))
	}
	for i, workers := range []int{1, 2, 4, 8} {
		if report.Results[i].Workers != workers {
			t.Errorf("resultado %d usou %d workers; esperado %d", i, report.Results[i].Workers, workers)
		}
	}
	if report.Note == "" {
		t.Fatal("Benchmark deveria explicar as métricas")
	}
}

func TestPrepareArticleContentReportsMissingAndInvalidPDF(t *testing.T) {
	service := newIndexingService(&fakeIndexRepository{})
	missing := service.prepareArticleContent(context.Background(), models.IndexTarget{
		ID: 1, Title: "TESTE ausente", FilePath: filepath.Join(t.TempDir(), "TESTE_ausente.pdf"),
	})
	if missing.Err == nil {
		t.Fatal("arquivo ausente deveria ser reportado")
	}

	invalidPath := filepath.Join(t.TempDir(), "TESTE_invalido.pdf")
	if err := os.WriteFile(invalidPath, []byte("%PDF-1.4 TESTE inválido"), 0o600); err != nil {
		t.Fatal(err)
	}
	invalid := service.prepareArticleContent(context.Background(), models.IndexTarget{
		ID: 2, Title: "TESTE inválido", FilePath: invalidPath,
	})
	if invalid.Err == nil {
		t.Fatal("PDF inválido deveria ser reportado")
	}
}
