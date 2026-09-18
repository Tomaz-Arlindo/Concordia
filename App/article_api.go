package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"desktop/internal/models"
	"desktop/internal/services"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      uint64 `json:"id,omitempty"`
}

type FileSelectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Name    string `json:"name,omitempty"`
	Size    int64  `json:"size,omitempty"`
}


type BatchImportResponse struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	Selected   int      `json:"selected"`
	Imported   int      `json:"imported"`
	Duplicates int      `json:"duplicates"`
	Failed     int      `json:"failed"`
	Errors     []string `json:"errors"`
}

type ArticleDetailResponse struct {
	Success bool                  `json:"success"`
	Message string                `json:"message"`
	Article *models.ArticleDetail `json:"article,omitempty"`
}

func (a *App) ListSources() []models.ArticleSource {
	if _, err := a.authenticatedUser(); err != nil || a.articleService == nil {
		return []models.ArticleSource{}
	}
	items, err := a.articleService.ListSources(a.ctx)
	if err != nil {
		return []models.ArticleSource{}
	}
	return items
}

func (a *App) ListArticles() []models.ArticleSummary {
	if _, err := a.authenticatedUser(); err != nil || a.articleService == nil {
		return []models.ArticleSummary{}
	}
	items, err := a.articleService.ListArticles(a.ctx)
	if err != nil {
		return []models.ArticleSummary{}
	}
	return items
}

func (a *App) SearchArticles(query string) []models.ArticleSummary {
	if _, err := a.authenticatedUser(); err != nil || a.articleService == nil {
		return []models.ArticleSummary{}
	}
	items, err := a.articleService.SearchArticles(a.ctx, query)
	if err != nil {
		return []models.ArticleSummary{}
	}
	return items
}

func (a *App) GetArticle(id uint64) ArticleDetailResponse {
	if _, err := a.authenticatedUser(); err != nil {
		return ArticleDetailResponse{Success: false, Message: "Usuário não autenticado."}
	}
	if a.articleService == nil {
		return ArticleDetailResponse{Success: false, Message: "Serviço de artigos indisponível."}
	}

	article, err := a.articleService.GetArticle(a.ctx, id)
	if err != nil {
		return ArticleDetailResponse{Success: false, Message: "Artigo não encontrado."}
	}

	return ArticleDetailResponse{
		Success: true,
		Message: "Artigo carregado.",
		Article: article,
	}
}

func (a *App) SelectPDF() FileSelectionResponse {
	if _, err := a.requireCurator(); err != nil {
		return FileSelectionResponse{Success: false, Message: "Somente Curador pode importar artigos."}
	}

	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Selecionar artigo em PDF",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "Documento PDF (*.pdf)",
				Pattern:     "*.pdf",
			},
		},
	})
	if err != nil {
		return FileSelectionResponse{Success: false, Message: "Não foi possível abrir o seletor de arquivos."}
	}
	if path == "" {
		return FileSelectionResponse{Success: false, Message: "Nenhum arquivo selecionado."}
	}

	info, err := os.Stat(path)
	if err != nil {
		return FileSelectionResponse{Success: false, Message: "Não foi possível ler o arquivo selecionado."}
	}

	a.mu.Lock()
	a.selectedPDF = path
	a.mu.Unlock()

	return FileSelectionResponse{
		Success: true,
		Message: "PDF selecionado.",
		Name:    filepath.Base(path),
		Size:    info.Size(),
	}
}


func (a *App) ImportPDFBatch(sourceID uint64, languageCode string) BatchImportResponse {
	user, err := a.requireCurator()
	if err != nil {
		return BatchImportResponse{
			Success: false,
			Message: "Somente Curador pode importar artigos.",
			Errors:  []string{},
		}
	}
	if a.articleService == nil {
		return BatchImportResponse{
			Success: false,
			Message: "Serviço de artigos indisponível.",
			Errors:  []string{},
		}
	}
	if sourceID == 0 {
		return BatchImportResponse{
			Success: false,
			Message: "Fonte inválida para importação em lote.",
			Errors:  []string{},
		}
	}

	paths, err := wailsruntime.OpenMultipleFilesDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Selecionar PDFs para benchmark",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "Documentos PDF (*.pdf)",
				Pattern:     "*.pdf",
			},
		},
	})
	if err != nil {
		return BatchImportResponse{
			Success: false,
			Message: "Não foi possível abrir o seletor de arquivos.",
			Errors:  []string{err.Error()},
		}
	}
	if len(paths) == 0 {
		return BatchImportResponse{
			Success: false,
			Message: "Nenhum PDF selecionado.",
			Errors:  []string{},
		}
	}

	sort.Strings(paths)

	report := BatchImportResponse{
		Selected: len(paths),
		Errors:   []string{},
	}

	lang := strings.TrimSpace(languageCode)
	if lang == "" {
		lang = "pt-BR"
	}

	for _, path := range paths {
		base := filepath.Base(path)
		title := strings.TrimSpace(strings.TrimSuffix(base, filepath.Ext(base)))
		if title == "" {
			title = "Documento importado"
		}

		_, importErr := a.articleService.ImportPDF(a.ctx, services.ImportArticleInput{
			Title:        title,
			SourceID:     sourceID,
			LanguageCode: lang,
			OriginalPath: path,
			UserID:       user.ID,
		})
		if importErr == nil {
			report.Imported++
			continue
		}

		if errors.Is(importErr, services.ErrDuplicateFile) {
			report.Duplicates++
			continue
		}

		report.Failed++
		report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", base, importErr))
	}

	report.Success = report.Imported > 0 && report.Failed == 0
	if report.Imported == 0 && report.Duplicates > 0 && report.Failed == 0 {
		report.Message = "Os PDFs selecionados já estavam no acervo."
		return report
	}

	report.Message = fmt.Sprintf(
		"Importação em lote concluída: %d importados, %d duplicados e %d falhas.",
		report.Imported,
		report.Duplicates,
		report.Failed,
	)
	return report
}

func (a *App) ImportArticle(
	title string,
	description string,
	sourceID uint64,
	authorsCSV string,
	languageCode string,
	publicationDate string,
	doi string,
	externalID string,
) ActionResponse {
	user, err := a.requireCurator()
	if err != nil {
		return ActionResponse{Success: false, Message: "Somente Curador pode importar artigos."}
	}
	if a.articleService == nil {
		return ActionResponse{Success: false, Message: "Serviço de artigos indisponível."}
	}

	a.mu.RLock()
	selectedPath := a.selectedPDF
	a.mu.RUnlock()

	if selectedPath == "" {
		return ActionResponse{Success: false, Message: "Selecione um arquivo PDF antes de importar."}
	}

	id, err := a.articleService.ImportPDF(a.ctx, services.ImportArticleInput{
		Title:           title,
		Description:     description,
		SourceID:        sourceID,
		AuthorsCSV:      authorsCSV,
		LanguageCode:    languageCode,
		PublicationDate: publicationDate,
		DOI:             doi,
		ExternalID:      externalID,
		OriginalPath:    selectedPath,
		UserID:          user.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidArticle):
			return ActionResponse{Success: false, Message: "Título, fonte e PDF válido são obrigatórios."}
		case errors.Is(err, services.ErrDuplicateFile):
			return ActionResponse{Success: false, Message: "Este PDF já foi importado anteriormente."}
		default:
			return ActionResponse{Success: false, Message: fmt.Sprintf("Falha ao importar artigo: %v", err)}
		}
	}

	a.mu.Lock()
	a.selectedPDF = ""
	a.mu.Unlock()

	return ActionResponse{
		Success: true,
		Message: "Artigo importado e persistido com sucesso.",
		ID:      id,
	}
}

func (a *App) UpdateArticle(
	id uint64,
	title string,
	description string,
	sourceID uint64,
	authorsCSV string,
	languageCode string,
	publicationDate string,
	doi string,
	externalID string,
) ActionResponse {
	if _, err := a.requireCurator(); err != nil {
		return ActionResponse{Success: false, Message: "Somente Curador pode editar artigos."}
	}
	if a.articleService == nil {
		return ActionResponse{Success: false, Message: "Serviço de artigos indisponível."}
	}

	err := a.articleService.UpdateArticle(a.ctx, services.UpdateArticleInput{
		ID:              id,
		Title:           title,
		Description:     description,
		SourceID:        sourceID,
		AuthorsCSV:      authorsCSV,
		LanguageCode:    languageCode,
		PublicationDate: publicationDate,
		DOI:             doi,
		ExternalID:      externalID,
	})
	if err != nil {
		return ActionResponse{Success: false, Message: fmt.Sprintf("Falha ao atualizar artigo: %v", err)}
	}

	return ActionResponse{Success: true, Message: "Metadados atualizados com sucesso.", ID: id}
}

func (a *App) DeleteArticle(id uint64) ActionResponse {
	if _, err := a.requireCurator(); err != nil {
		return ActionResponse{Success: false, Message: "Somente Curador pode excluir artigos."}
	}
	if a.articleService == nil {
		return ActionResponse{Success: false, Message: "Serviço de artigos indisponível."}
	}

	if err := a.articleService.DeleteArticle(a.ctx, id); err != nil {
		return ActionResponse{Success: false, Message: fmt.Sprintf("Falha ao excluir artigo: %v", err)}
	}

	return ActionResponse{Success: true, Message: "Artigo removido com sucesso.", ID: id}
}

func (a *App) GetCollectionStats() models.CollectionStats {
	if _, err := a.authenticatedUser(); err != nil || a.articleService == nil {
		return models.CollectionStats{}
	}
	stats, err := a.articleService.GetStats(a.ctx)
	if err != nil {
		return models.CollectionStats{}
	}
	return stats
}
