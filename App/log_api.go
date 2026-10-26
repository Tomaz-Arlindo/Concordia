package main

import (
	"fmt"
	"os"
	"time"

	"desktop/internal/models"
	"desktop/internal/services"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type LogSummaryResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Summary models.LogSummary `json:"summary"`
}

type ExportResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
	Rows    int    `json:"rows"`
}

// registrarOperacao é o atalho usado pelas demais rotas para alimentar o
// histórico com o usuário autenticado no momento da ação.
func (a *App) registrarOperacao(entry models.LogEntry) {
	if a.logService == nil {
		return
	}
	if entry.UserID == 0 {
		if user, err := a.authenticatedUser(); err == nil {
			entry.UserID = user.ID
		}
	}
	a.logService.Registrar(a.ctx, entry)
}

// ListOperationLogs devolve o histórico de operações com filtros opcionais.
func (a *App) ListOperationLogs(operation, dateFrom, dateTo string, limit int) []models.OperationLog {
	if _, err := a.authenticatedUser(); err != nil || a.logService == nil {
		return []models.OperationLog{}
	}

	logs, err := a.logService.List(a.ctx, models.LogFilters{
		Operation: operation,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Limit:     limit,
	})
	if err != nil {
		fmt.Printf("Concordia: falha ao consultar histórico: %v\n", err)
		return []models.OperationLog{}
	}

	return logs
}

// GetLogSummary alimenta os indicadores de histórico do dashboard.
func (a *App) GetLogSummary() LogSummaryResponse {
	if _, err := a.authenticatedUser(); err != nil {
		return LogSummaryResponse{Success: false, Message: "Usuário não autenticado."}
	}
	if a.logService == nil {
		return LogSummaryResponse{Success: false, Message: "Serviço de histórico indisponível."}
	}

	summary, err := a.logService.Summary(a.ctx)
	if err != nil {
		return LogSummaryResponse{Success: false, Message: "Falha ao consolidar o histórico."}
	}

	return LogSummaryResponse{Success: true, Message: "Histórico consolidado.", Summary: summary}
}

// ExportLogsCSV grava o histórico filtrado em um arquivo CSV escolhido pelo
// usuário.
func (a *App) ExportLogsCSV(operation, dateFrom, dateTo string, limit int) ExportResponse {
	if _, err := a.requireCurator(); err != nil {
		return ExportResponse{Success: false, Message: "Somente Curador pode exportar o histórico."}
	}
	if a.logService == nil {
		return ExportResponse{Success: false, Message: "Serviço de histórico indisponível."}
	}

	logs, err := a.logService.List(a.ctx, models.LogFilters{
		Operation: operation,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Limit:     limit,
	})
	if err != nil {
		return ExportResponse{Success: false, Message: "Falha ao consultar o histórico."}
	}
	if len(logs) == 0 {
		return ExportResponse{Success: false, Message: "Nenhum registro para exportar com os filtros atuais."}
	}

	content, err := services.BuildCSV(logs)
	if err != nil {
		return ExportResponse{Success: false, Message: "Falha ao gerar o arquivo CSV."}
	}

	suggested := fmt.Sprintf("historico-concordia-%s.csv", time.Now().Format("2006-01-02"))

	path, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "Exportar histórico de operações",
		DefaultFilename: suggested,
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "Arquivo CSV (*.csv)",
				Pattern:     "*.csv",
			},
		},
	})
	if err != nil {
		return ExportResponse{Success: false, Message: "Não foi possível abrir o seletor de arquivos."}
	}
	if path == "" {
		return ExportResponse{Success: false, Message: "Exportação cancelada."}
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return ExportResponse{Success: false, Message: "Não foi possível gravar o arquivo."}
	}

	a.registrarOperacao(models.LogEntry{
		Operation: "EXPORTACAO_CSV",
		Entity:    "operation_logs",
		Details:   fmt.Sprintf("%d registros exportados", len(logs)),
	})

	return ExportResponse{
		Success: true,
		Message: fmt.Sprintf("%d registros exportados.", len(logs)),
		Path:    path,
		Rows:    len(logs),
	}
}
