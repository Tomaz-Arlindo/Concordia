package main

import (
	"desktop/internal/models"
)

type IndexingResponse struct {
	Success bool                  `json:"success"`
	Message string                `json:"message"`
	Report  models.IndexingReport `json:"report"`
}

type BenchmarkResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Report  models.BenchmarkReport `json:"report"`
}

func (a *App) StartIndexing(workers int, reindexAll bool) IndexingResponse {
	if _, err := a.requireCurator(); err != nil {
		return IndexingResponse{Success: false, Message: "Somente Curador pode iniciar a indexação."}
	}
	if a.indexingService == nil {
		return IndexingResponse{Success: false, Message: "Serviço de indexação indisponível."}
	}

	report, err := a.indexingService.Run(a.ctx, workers, reindexAll)
	if err != nil {
		return IndexingResponse{Success: false, Message: "Falha durante a indexação: " + err.Error(), Report: report}
	}

	message := "Indexação concluída."
	if report.Total == 0 {
		message = "Nenhum artigo pendente para indexar."
	} else if report.Failed > 0 {
		message = "Indexação concluída com alguns erros."
	}

	return IndexingResponse{Success: true, Message: message, Report: report}
}

func (a *App) GetIndexingOverview() models.IndexingOverview {
	if _, err := a.authenticatedUser(); err != nil || a.indexingService == nil {
		return models.IndexingOverview{}
	}
	overview, err := a.indexingService.Overview(a.ctx)
	if err != nil {
		return models.IndexingOverview{}
	}
	return overview
}

func (a *App) SearchBM25(query string, limit int) []models.SearchResult {
	if _, err := a.authenticatedUser(); err != nil || a.searchService == nil {
		return []models.SearchResult{}
	}
	results, err := a.searchService.BM25(a.ctx, query, limit)
	if err != nil {
		return []models.SearchResult{}
	}
	return results
}

func (a *App) SearchBM25Filtered(
	query string,
	source string,
	author string,
	language string,
	dateFrom string,
	dateTo string,
	fuzzy bool,
	limit int,
) []models.SearchResult {
	if _, err := a.authenticatedUser(); err != nil || a.searchService == nil {
		return []models.SearchResult{}
	}
	results, err := a.searchService.BM25Filtered(a.ctx, query, models.SearchFilters{
		Source:   source,
		Author:   author,
		Language: language,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Fuzzy:    fuzzy,
		Limit:    limit,
	})
	if err != nil {
		return []models.SearchResult{}
	}
	return results
}

func (a *App) RunIndexingBenchmark() BenchmarkResponse {
	if _, err := a.requireCurator(); err != nil {
		return BenchmarkResponse{Success: false, Message: "Somente Curador pode executar benchmark."}
	}
	if a.indexingService == nil {
		return BenchmarkResponse{Success: false, Message: "Serviço de indexação indisponível."}
	}

	report, err := a.indexingService.Benchmark(a.ctx)
	if err != nil {
		return BenchmarkResponse{Success: false, Message: "Falha no benchmark: " + err.Error()}
	}
	return BenchmarkResponse{Success: true, Message: "Benchmark concluído.", Report: report}
}
