package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"desktop/internal/models"
	"desktop/internal/services"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const maxLogTitleRunes = 200

func (a *App) ExportArticlePDF(id uint64) ExportResponse {
	if _, err := a.authenticatedUser(); err != nil {
		return ExportResponse{Success: false, Message: "Usuário não autenticado."}
	}
	if a.articleService == nil {
		return ExportResponse{Success: false, Message: "Serviço de artigos indisponível."}
	}

	article, err := a.articleService.GetArticle(a.ctx, id)
	if err != nil || article == nil {
		return ExportResponse{Success: false, Message: "Artigo não encontrado."}
	}

	source := filepath.FromSlash(article.FilePath)
	if _, err := os.Stat(source); err != nil {
		return ExportResponse{
			Success: false,
			Message: "O arquivo PDF deste artigo não foi encontrado no armazenamento.",
		}
	}

	path, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "Exportar artigo em PDF",
		DefaultFilename: services.ExportPDFFileName(article.Title, article.ID),
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "Documento PDF (*.pdf)",
				Pattern:     "*.pdf",
			},
		},
	})
	if err != nil {
		return ExportResponse{Success: false, Message: "Não foi possível abrir o seletor de arquivos."}
	}
	if path == "" {
		return ExportResponse{Success: false, Message: "Exportação cancelada."}
	}
	if !strings.EqualFold(filepath.Ext(path), ".pdf") {
		path += ".pdf"
	}

	if err := exportCopyFile(source, path); err != nil {
		fmt.Printf("Concordia: falha ao exportar PDF do artigo %d: %v\n", article.ID, err)
		return ExportResponse{Success: false, Message: "Não foi possível gravar o arquivo."}
	}

	title := article.Title
	if runes := []rune(title); len(runes) > maxLogTitleRunes {
		title = string(runes[:maxLogTitleRunes]) + "..."
	}
	a.registrarOperacao(models.LogEntry{
		Operation: models.OperationExportPDF,
		Entity:    "articles",
		EntityID:  article.ID,
		Details:   fmt.Sprintf("PDF exportado: %s", title),
	})

	return ExportResponse{
		Success: true,
		Message: fmt.Sprintf("PDF exportado como \"%s\".", filepath.Base(path)),
		Path:    path,
	}
}

func exportCopyFile(source, destination string) error {
	if srcInfo, err := os.Stat(source); err == nil {
		if dstInfo, err := os.Stat(destination); err == nil && os.SameFile(srcInfo, dstInfo) {
			return errors.New("origem e destino são o mesmo arquivo")
		}
	}

	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		_ = os.Remove(destination)
		return err
	}

	return out.Close()
}