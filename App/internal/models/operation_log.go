package models

import "time"

// Operações registradas no histórico do sistema.
const (
	OperationLogin       = "LOGIN"
	OperationLogout      = "LOGOUT"
	OperationRegister    = "CADASTRO_USUARIO"
	OperationImport      = "IMPORTACAO_ARTIGO"
	OperationBatchImport = "IMPORTACAO_LOTE"
	OperationUpdate      = "EDICAO_ARTIGO"
	OperationDelete      = "EXCLUSAO_ARTIGO"
	OperationIndexing    = "INDEXACAO"
	OperationBenchmark   = "BENCHMARK"
	OperationSearch      = "BUSCA"
)

// é um registro já persistido, pronto para exibição.
type OperationLog struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Operation   string    `json:"operation"`
	Entity      string    `json:"entity"`
	EntityID    uint64    `json:"entity_id"`
	Details     string    `json:"details"`
	ResultCount int       `json:"result_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// LogEntry descreve uma operação a ser registrada.
type LogEntry struct {
	UserID      uint64
	Operation   string
	Entity      string
	EntityID    uint64
	Details     string
	ResultCount int
}

// LogFilters controla a consulta ao histórico.
type LogFilters struct {
	Operation string `json:"operation"`
	UserID    uint64 `json:"user_id"`
	DateFrom  string `json:"date_from"`
	DateTo    string `json:"date_to"`
	Limit     int    `json:"limit"`
}

type OperationCount struct {
	Operation string `json:"operation"`
	Total     uint64 `json:"total"`
}

type TermCount struct {
	Term  string `json:"term"`
	Total uint64 `json:"total"`
}

// LogSummary alimenta os indicadores do dashboard.
type LogSummary struct {
	Total       uint64           `json:"total"`
	Today       uint64           `json:"today"`
	ByOperation []OperationCount `json:"by_operation"`
	TopSearches []TermCount      `json:"top_searches"`
}
