package services

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"desktop/internal/models"
)

// benchmarkOnce separa explicitamente o benchmark em duas fases:
//
//  1. processamento paralelizável: PDF -> texto -> tokens -> índice local;
//  2. persistência serial: escrita do índice no MySQL + recálculo de DF.
//
// A separação é intencional. O pipeline normal continua usando workers e um
// writer único para evitar deadlocks. Para fins acadêmicos, o benchmark mede
// isoladamente onde as goroutines conseguem acelerar e quanto do tempo total
// permanece preso ao gargalo serial de persistência.
func (s *IndexingService) benchmarkOnce(ctx context.Context, workers int) (models.BenchmarkResult, error) {
	workers = normalizeWorkerCount(workers)

	targets, err := s.index.ListTargets(ctx, true)
	if err != nil {
		return models.BenchmarkResult{}, err
	}

	result := models.BenchmarkResult{
		Workers: workers,
		Errors:  []string{},
	}
	if len(targets) == 0 {
		return result, nil
	}

	// Fase 1: processamento puramente paralelo, sem escrita no MySQL.
	processingStart := time.Now()
	jobs := make(chan models.IndexTarget)
	results := make(chan workerResult, len(targets))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				results <- s.prepareArticleContent(ctx, target)
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

	prepared := make([]workerResult, 0, len(targets))
	for item := range results {
		prepared = append(prepared, item)
	}
	result.ProcessingMS = time.Since(processingStart).Milliseconds()

	// Mantém a ordem de persistência estável entre execuções para reduzir ruído
	// do benchmark causado apenas pela ordem em que os workers terminaram.
	sort.Slice(prepared, func(i, j int) bool {
		return prepared[i].ArticleID < prepared[j].ArticleID
	})

	// Fase 2: persistência serial. Ela é propositalmente separada porque foi a
	// solução arquitetural usada para eliminar deadlocks entre transações que
	// compartilhavam muitos termos.
	databaseStart := time.Now()
	for _, item := range prepared {
		if item.Err != nil {
			_ = s.index.MarkStatus(ctx, item.ArticleID, "ERROR")
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("#%d %s: %v", item.ArticleID, item.Title, item.Err))
			continue
		}

		if err := s.index.MarkStatus(ctx, item.ArticleID, "PROCESSING"); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("#%d %s: atualizar status: %v", item.ArticleID, item.Title, err))
			continue
		}

		if err := s.index.PersistDocument(ctx, item.Document); err != nil {
			_ = s.index.MarkStatus(ctx, item.ArticleID, "ERROR")
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("#%d %s: %v", item.ArticleID, item.Title, err))
			continue
		}
		result.Indexed++
	}

	if err := s.index.RecalculateDocumentFrequency(ctx); err != nil {
		return result, err
	}
	result.DatabaseMS = time.Since(databaseStart).Milliseconds()

	result.DurationMS = result.ProcessingMS + result.DatabaseMS
	if result.DurationMS > 0 {
		result.DatabaseShare = float64(result.DatabaseMS) / float64(result.DurationMS)
	}

	return result, nil
}

func (s *IndexingService) Benchmark(ctx context.Context) (models.BenchmarkReport, error) {
	workersList := []int{1, 2, 4, 8}
	results := make([]models.BenchmarkResult, 0, len(workersList))

	var processingBaseline int64
	var totalBaseline int64

	for _, workers := range workersList {
		result, err := s.benchmarkOnce(ctx, workers)
		if err != nil {
			return models.BenchmarkReport{}, fmt.Errorf("benchmark com %d workers: %w", workers, err)
		}

		if workers == 1 {
			processingBaseline = result.ProcessingMS
			totalBaseline = result.DurationMS
		}

		if processingBaseline > 0 && result.ProcessingMS > 0 {
			result.ProcessingSpeedup = float64(processingBaseline) / float64(result.ProcessingMS)
			result.ProcessingEfficiency = result.ProcessingSpeedup / float64(result.Workers)
		}
		if totalBaseline > 0 && result.DurationMS > 0 {
			result.TotalSpeedup = float64(totalBaseline) / float64(result.DurationMS)
		}

		results = append(results, result)
	}

	return models.BenchmarkReport{
		Results: results,
		Note:    "O benchmark separa a etapa paralelizável (extração + tokenização + índice local) da persistência serial no MySQL. O speedup de processamento mostra o efeito das goroutines; o speedup total inclui o gargalo do banco. A porcentagem MySQL mostra quanto do tempo fim-a-fim ficou na etapa serial.",
	}, nil
}
