package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"desktop/internal/models"
	"desktop/internal/repositories"
)

type IndexingService struct {
	index *repositories.IndexRepository
}

func NewIndexingService(index *repositories.IndexRepository) *IndexingService {
	return &IndexingService{index: index}
}

type workerResult struct {
	ArticleID uint64
	Title     string
	Document  models.IndexedDocument
	Err       error
}

func normalizeWorkerCount(workers int) int {
	if workers < 1 {
		workers = 1
	}
	maxWorkers := runtime.NumCPU() * 2
	if maxWorkers < 4 {
		maxWorkers = 4
	}
	if maxWorkers > 32 {
		maxWorkers = 32
	}
	if workers > maxWorkers {
		workers = maxWorkers
	}
	return workers
}

func (s *IndexingService) Run(ctx context.Context, workers int, reindexAll bool) (models.IndexingReport, error) {
	workers = normalizeWorkerCount(workers)

	targets, err := s.index.ListTargets(ctx, reindexAll)
	if err != nil {
		return models.IndexingReport{}, err
	}

	report := models.IndexingReport{
		Workers:      workers,
		Total:        len(targets),
		ReindexedAll: reindexAll,
		Errors:       []string{},
	}
	if len(targets) == 0 {
		return report, nil
	}

	start := time.Now()
	jobs := make(chan models.IndexTarget)
	results := make(chan workerResult, len(targets))

	// Os workers fazem a parte paralela e custosa: leitura do PDF,
	// extração de texto, normalização, tokenização e montagem do índice local.
	//
	// A persistência MySQL é feita por um único consumidor abaixo. Isso evita
	// que várias transações concorrentes tentem inserir/atualizar os mesmos
	// termos da tabela `terms`, cenário que causava deadlocks (MySQL 1213)
	// quando vários artigos compartilhavam palavras.
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				results <- s.prepareArticle(ctx, target)
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, target := range targets {
			select {
			case <-ctx.Done():
				return
			case jobs <- target:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	// Um único writer persiste os documentos. Enquanto ele grava um resultado,
	// os workers continuam extraindo/tokenizando os demais PDFs em paralelo.
	for result := range results {
		if result.Err != nil {
			_ = s.index.MarkStatus(ctx, result.ArticleID, "ERROR")
			report.Failed++
			report.Errors = append(report.Errors, fmt.Sprintf("#%d %s: %v", result.ArticleID, result.Title, result.Err))
			continue
		}

		if err := s.index.PersistDocument(ctx, result.Document); err != nil {
			_ = s.index.MarkStatus(ctx, result.ArticleID, "ERROR")
			report.Failed++
			report.Errors = append(report.Errors, fmt.Sprintf("#%d %s: %v", result.ArticleID, result.Title, err))
			continue
		}

		report.Indexed++
	}

	if err := s.index.RecalculateDocumentFrequency(ctx); err != nil {
		return report, err
	}

	report.DurationMS = time.Since(start).Milliseconds()
	return report, nil
}

func (s *IndexingService) prepareArticle(ctx context.Context, target models.IndexTarget) workerResult {
	if err := s.index.MarkStatus(ctx, target.ID, "PROCESSING"); err != nil {
		return workerResult{ArticleID: target.ID, Title: target.Title, Err: err}
	}

	return s.prepareArticleContent(ctx, target)
}

// prepareArticleContent executa somente a parte paralelizável da indexação:
// leitura do arquivo, extração do texto, normalização, tokenização e montagem
// do índice local. Não escreve no MySQL, por isso também é usada pelo
// benchmark para medir o ganho das goroutines sem misturar o gargalo de I/O
// do banco de dados.
func (s *IndexingService) prepareArticleContent(ctx context.Context, target models.IndexTarget) workerResult {
	result := workerResult{ArticleID: target.ID, Title: target.Title}

	path := filepath.FromSlash(target.FilePath)
	if _, err := os.Stat(path); err != nil {
		result.Err = fmt.Errorf("arquivo não encontrado em %s", target.FilePath)
		return result
	}

	text, err := ExtractPDFText(ctx, path)
	if err != nil {
		result.Err = fmt.Errorf("extração de texto: %w", err)
		return result
	}

	doc := BuildIndexedDocument(target.ID, text)
	if doc.DocumentLength == 0 || len(doc.Terms) == 0 {
		result.Err = ErrNoExtractableText
		return result
	}

	result.Document = doc
	return result
}

func (s *IndexingService) Overview(ctx context.Context) (models.IndexingOverview, error) {
	return s.index.Overview(ctx)
}
