package repositories

import (
	"context"
	"database/sql"
	"strings"

	"desktop/internal/models"
)

type LogRepository struct {
	DB *sql.DB
}

func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{DB: db}
}

// Create grava uma operação no histórico.
func (r *LogRepository) Create(ctx context.Context, entry models.LogEntry) error {
	var userID any
	if entry.UserID > 0 {
		userID = entry.UserID
	}

	var entityID any
	if entry.EntityID > 0 {
		entityID = entry.EntityID
	}

	details := entry.Details
	if len(details) > 500 {
		details = details[:500]
	}

	_, err := r.DB.ExecContext(
		ctx,
		`INSERT INTO operation_logs
		     (user_id, operation, entity, entity_id, details, result_count)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID,
		entry.Operation,
		entry.Entity,
		entityID,
		details,
		entry.ResultCount,
	)

	return err
}

// List devolve o histórico aplicando os filtros informados.
func (r *LogRepository) List(ctx context.Context, filters models.LogFilters) ([]models.OperationLog, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT l.id,
		       l.user_id,
		       u.name,
		       l.operation,
		       l.entity,
		       l.entity_id,
		       l.details,
		       l.result_count,
		       l.created_at
		FROM operation_logs l
		LEFT JOIN users u ON u.id = l.user_id
		WHERE 1 = 1`)

	args := make([]any, 0, 5)

	if operation := strings.TrimSpace(filters.Operation); operation != "" {
		query.WriteString(" AND l.operation = ?")
		args = append(args, operation)
	}
	if filters.UserID > 0 {
		query.WriteString(" AND l.user_id = ?")
		args = append(args, filters.UserID)
	}
	if dateFrom := strings.TrimSpace(filters.DateFrom); dateFrom != "" {
		query.WriteString(" AND l.created_at >= ?")
		args = append(args, dateFrom)
	}
	if dateTo := strings.TrimSpace(filters.DateTo); dateTo != "" {
		query.WriteString(" AND l.created_at < DATE_ADD(?, INTERVAL 1 DAY)")
		args = append(args, dateTo)
	}

	limit := filters.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query.WriteString(" ORDER BY l.created_at DESC, l.id DESC LIMIT ?")
	args = append(args, limit)

	rows, err := r.DB.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]models.OperationLog, 0)

	for rows.Next() {
		var (
			item     models.OperationLog
			userID   sql.NullInt64
			userName sql.NullString
			entityID sql.NullInt64
		)

		if err := rows.Scan(
			&item.ID,
			&userID,
			&userName,
			&item.Operation,
			&item.Entity,
			&entityID,
			&item.Details,
			&item.ResultCount,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		if userID.Valid {
			item.UserID = uint64(userID.Int64)
		}
		if userName.Valid {
			item.UserName = userName.String
		} else {
			item.UserName = "usuário removido"
		}
		if entityID.Valid {
			item.EntityID = uint64(entityID.Int64)
		}

		logs = append(logs, item)
	}

	return logs, rows.Err()
}

// Summary consolida os indicadores do histórico usados no dashboard.
func (r *LogRepository) Summary(ctx context.Context) (models.LogSummary, error) {
	summary := models.LogSummary{
		ByOperation: make([]models.OperationCount, 0),
		TopSearches: make([]models.TermCount, 0),
	}

	if err := r.DB.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM operation_logs`,
	).Scan(&summary.Total); err != nil {
		return summary, err
	}

	if err := r.DB.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM operation_logs WHERE DATE(created_at) = CURDATE()`,
	).Scan(&summary.Today); err != nil {
		return summary, err
	}

	rows, err := r.DB.QueryContext(
		ctx,
		`SELECT operation, COUNT(*) AS total
		 FROM operation_logs
		 GROUP BY operation
		 ORDER BY total DESC`,
	)
	if err != nil {
		return summary, err
	}
	for rows.Next() {
		var item models.OperationCount
		if err := rows.Scan(&item.Operation, &item.Total); err != nil {
			rows.Close()
			return summary, err
		}
		summary.ByOperation = append(summary.ByOperation, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return summary, err
	}
	rows.Close()

	searchRows, err := r.DB.QueryContext(
		ctx,
		`SELECT details, COUNT(*) AS total
		 FROM operation_logs
		 WHERE operation = ? AND details <> ''
		 GROUP BY details
		 ORDER BY total DESC, details ASC
		 LIMIT 5`,
		models.OperationSearch,
	)
	if err != nil {
		return summary, err
	}
	defer searchRows.Close()

	for searchRows.Next() {
		var item models.TermCount
		if err := searchRows.Scan(&item.Term, &item.Total); err != nil {
			return summary, err
		}
		summary.TopSearches = append(summary.TopSearches, item)
	}

	return summary, searchRows.Err()
}
