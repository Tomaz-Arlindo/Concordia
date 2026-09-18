package repositories

import (
	"context"
	"database/sql"

	"desktop/internal/models"
)

type SourceRepository struct {
	DB *sql.DB
}

func NewSourceRepository(db *sql.DB) *SourceRepository {
	return &SourceRepository{DB: db}
}

func (r *SourceRepository) EnsureDefaults(ctx context.Context) error {
	defaults := []struct {
		Name        string
		Description string
	}{
		{"local", "Artigos importados manualmente pelo Curador"},
		{"arxiv", "Artigos provenientes do arXiv"},
		{"pubmed", "Artigos provenientes do PubMed"},
		{"openalex", "Artigos provenientes do OpenAlex"},
	}

	for _, item := range defaults {
		_, err := r.DB.ExecContext(
			ctx,
			`INSERT INTO article_sources (name, description)
			 VALUES (?, ?)
			 ON DUPLICATE KEY UPDATE description = COALESCE(description, VALUES(description))`,
			item.Name,
			item.Description,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SourceRepository) List(ctx context.Context) ([]models.ArticleSource, error) {
	rows, err := r.DB.QueryContext(
		ctx,
		`SELECT id, name, COALESCE(description, '')
		 FROM article_sources
		 ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.ArticleSource, 0)
	for rows.Next() {
		var item models.ArticleSource
		if err := rows.Scan(&item.ID, &item.Name, &item.Description); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
