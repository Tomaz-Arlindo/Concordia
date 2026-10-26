package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"desktop/internal/models"
	"desktop/internal/repositories"
)

type LogService struct {
	logs *repositories.LogRepository
}

func NewLogService(logs *repositories.LogRepository) *LogService {
	return &LogService{logs: logs}
}


// A falha ao registrar nunca interrompe a operação principal: o histórico é
// apenas um recurso auxiliar, e não deve impedir o funcionamento do sistema.
func (s *LogService) Registrar(ctx context.Context, entry models.LogEntry) {
	if s == nil || s.logs == nil {
		return
	}
	if strings.TrimSpace(entry.Operation) == "" {
		return
	}

	if err := s.logs.Create(ctx, entry); err != nil {
		fmt.Printf("Concordia: falha ao registrar operação %s: %v\n", entry.Operation, err)
	}
}

func (s *LogService) List(ctx context.Context, filters models.LogFilters) ([]models.OperationLog, error) {
	return s.logs.List(ctx, filters)
}

func (s *LogService) Summary(ctx context.Context) (models.LogSummary, error) {
	return s.logs.Summary(ctx)
}

// BuildCSV converte o histórico em CSV com separador ponto e vírgula, que é o
// esperado pelo Excel em português.
func BuildCSV(logs []models.OperationLog) (string, error) {
	var builder strings.Builder

	// BOM UTF-8: sem ele o Excel exibe os acentos incorretamente.
	builder.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(&builder)
	writer.Comma = ';'

	header := []string{
		"id",
		"data_hora",
		"usuario",
		"operacao",
		"entidade",
		"entidade_id",
		"detalhes",
		"resultados",
	}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	for _, item := range logs {
		entityID := ""
		if item.EntityID > 0 {
			entityID = strconv.FormatUint(item.EntityID, 10)
		}

		resultCount := ""
		if item.Operation == models.OperationSearch {
			resultCount = strconv.Itoa(item.ResultCount)
		}

		record := []string{
			strconv.FormatUint(item.ID, 10),
			item.CreatedAt.Format("2006-01-02 15:04:05"),
			item.UserName,
			item.Operation,
			item.Entity,
			entityID,
			item.Details,
			resultCount,
		}

		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return builder.String(), nil
}
