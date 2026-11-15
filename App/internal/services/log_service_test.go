package services

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"desktop/internal/models"
	"desktop/internal/repositories"
)

func TestLogServiceMySQLIntegration(t *testing.T) {
	db := servicesIntegrationDB(t)
	ctx := context.Background()
	operation := fmt.Sprintf("TESTE_Log_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, `DELETE FROM operation_logs WHERE operation = ?`, operation); err != nil {
			t.Errorf("limpeza do log TESTE_ falhou: %v", err)
		}
		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM operation_logs WHERE operation = ?`, operation).Scan(&count); err != nil {
			t.Errorf("verificação da limpeza de logs TESTE_ falhou: %v", err)
		} else if count != 0 {
			t.Errorf("limpeza deixou %d logs TESTE_", count)
		}
	})

	service := NewLogService(repositories.NewLogRepository(db))
	service.Registrar(ctx, models.LogEntry{
		Operation: operation,
		Entity:    "TESTE_",
		Details:   `TESTE_ "detalhe"; CSV`,
	})
	service.Registrar(ctx, models.LogEntry{Operation: " ", Entity: "TESTE_"})

	logs, err := service.List(ctx, models.LogFilters{Operation: operation})
	if err != nil || len(logs) != 1 {
		t.Fatalf("List de logs inesperado: quantidade=%d erro=%v", len(logs), err)
	}
	if logs[0].Details != `TESTE_ "detalhe"; CSV` || logs[0].UserName != "usuário removido" {
		t.Fatalf("registro de log inesperado: %+v", logs[0])
	}
	summary, err := service.Summary(ctx)
	if err != nil {
		t.Fatalf("Summary falhou: %v", err)
	}
	found := false
	for _, item := range summary.ByOperation {
		if item.Operation == operation && item.Total == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("Summary não incluiu operação TESTE_: %+v", summary.ByOperation)
	}

	csv, err := BuildCSV(logs)
	if err != nil {
		t.Fatalf("BuildCSV falhou: %v", err)
	}
	if !strings.HasPrefix(csv, "\uFEFFid;data_hora;usuario;operacao;") || !strings.Contains(csv, `"TESTE_ ""detalhe""; CSV"`) {
		t.Fatalf("CSV não preservou cabeçalho BOM/separador ou escapamento: %q", csv)
	}
}

func TestLogServiceRegistrarWithoutServiceOrOperationIsNoOp(t *testing.T) {
	ctx := context.Background()
	var nilService *LogService
	nilService.Registrar(ctx, models.LogEntry{Operation: "TESTE_"})
	NewLogService(nil).Registrar(ctx, models.LogEntry{Operation: "TESTE_"})
	NewLogService(nil).Registrar(ctx, models.LogEntry{Operation: " "})
}
