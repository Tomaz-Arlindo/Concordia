package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"desktop/internal/models"
)

var ErrArticleNotFound = errors.New("artigo não encontrado")

type ArticleRepository struct {
	DB *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{DB: db}
}

const articleSummarySelect = `
SELECT
	a.id,
	a.title,
	COALESCE(a.description, ''),
	COALESCE(a.filename, ''),
	COALESCE(s.name, ''),
	COALESCE(auth.authors, ''),
	COALESCE(a.language_code, ''),
	COALESCE(DATE_FORMAT(a.publication_date, '%Y-%m-%d'), ''),
	a.index_status,
	a.created_at
FROM articles a
LEFT JOIN article_sources s ON s.id = a.source_id
LEFT JOIN (
	SELECT
		aa.article_id,
		GROUP_CONCAT(au.name ORDER BY aa.author_order SEPARATOR ', ') AS authors
	FROM article_authors aa
	JOIN authors au ON au.id = aa.author_id
	GROUP BY aa.article_id
) auth ON auth.article_id = a.id
`

func (r *ArticleRepository) List(ctx context.Context) ([]models.ArticleSummary, error) {
	rows, err := r.DB.QueryContext(ctx, articleSummarySelect+` ORDER BY a.created_at DESC, a.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticleSummaries(rows)
}

func (r *ArticleRepository) Search(ctx context.Context, query string) ([]models.ArticleSummary, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return r.List(ctx)
	}

	pattern := "%" + query + "%"
	rows, err := r.DB.QueryContext(
		ctx,
		articleSummarySelect+`
		WHERE
			a.title LIKE ?
			OR COALESCE(a.description, '') LIKE ?
			OR COALESCE(s.name, '') LIKE ?
			OR COALESCE(auth.authors, '') LIKE ?
		ORDER BY a.created_at DESC, a.id DESC`,
		pattern, pattern, pattern, pattern,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticleSummaries(rows)
}

func scanArticleSummaries(rows *sql.Rows) ([]models.ArticleSummary, error) {
	items := make([]models.ArticleSummary, 0)
	for rows.Next() {
		var item models.ArticleSummary
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Filename,
			&item.SourceName,
			&item.Authors,
			&item.LanguageCode,
			&item.PublicationDate,
			&item.IndexStatus,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ArticleRepository) GetByID(ctx context.Context, id uint64) (*models.ArticleDetail, error) {
	var item models.ArticleDetail

	err := r.DB.QueryRowContext(
		ctx,
		`SELECT
			a.id,
			a.title,
			COALESCE(a.description, ''),
			COALESCE(a.filename, ''),
			COALESCE(a.file_path, ''),
			COALESCE(a.content_type, ''),
			COALESCE(a.size, 0),
			COALESCE(a.user_id, 0),
			COALESCE(a.source_id, 0),
			COALESCE(s.name, ''),
			COALESCE(auth.authors, ''),
			COALESCE(a.abstract_text, ''),
			COALESCE(a.external_id, ''),
			COALESCE(a.doi, ''),
			COALESCE(a.language_code, ''),
			COALESCE(DATE_FORMAT(a.publication_date, '%Y-%m-%d'), ''),
			COALESCE(a.checksum_sha256, ''),
			a.index_status,
			a.created_at,
			a.updated_at
		FROM articles a
		LEFT JOIN article_sources s ON s.id = a.source_id
		LEFT JOIN (
			SELECT
				aa.article_id,
				GROUP_CONCAT(au.name ORDER BY aa.author_order SEPARATOR ', ') AS authors
			FROM article_authors aa
			JOIN authors au ON au.id = aa.author_id
			GROUP BY aa.article_id
		) auth ON auth.article_id = a.id
		WHERE a.id = ?
		LIMIT 1`,
		id,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Filename,
		&item.FilePath,
		&item.ContentType,
		&item.Size,
		&item.UserID,
		&item.SourceID,
		&item.SourceName,
		&item.Authors,
		&item.AbstractText,
		&item.ExternalID,
		&item.DOI,
		&item.LanguageCode,
		&item.PublicationDate,
		&item.ChecksumSHA256,
		&item.IndexStatus,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrArticleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ArticleRepository) ChecksumExists(ctx context.Context, checksum string) (bool, error) {
	var count uint64
	err := r.DB.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM articles WHERE checksum_sha256 = ?`,
		checksum,
	).Scan(&count)
	return count > 0, err
}

func (r *ArticleRepository) Delete(ctx context.Context, id uint64) (string, error) {
	var path string
	err := r.DB.QueryRowContext(
		ctx,
		`SELECT COALESCE(file_path, '') FROM articles WHERE id = ?`,
		id,
	).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrArticleNotFound
	}
	if err != nil {
		return "", err
	}

	result, err := r.DB.ExecContext(ctx, `DELETE FROM articles WHERE id = ?`, id)
	if err != nil {
		return "", err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return "", err
	}
	if affected == 0 {
		return "", ErrArticleNotFound
	}
	return path, nil
}

func (r *ArticleRepository) GetStats(ctx context.Context) (models.CollectionStats, error) {
	var stats models.CollectionStats

	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM articles`).Scan(&stats.TotalArticles); err != nil {
		return stats, err
	}
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM articles WHERE UPPER(index_status) = 'PENDING'`).Scan(&stats.PendingArticles); err != nil {
		return stats, err
	}
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM articles WHERE UPPER(index_status) = 'INDEXED'`).Scan(&stats.IndexedArticles); err != nil {
		return stats, err
	}
	if err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT aa.author_id)
		FROM article_authors aa
		JOIN articles a ON a.id = aa.article_id
	`).Scan(&stats.TotalAuthors); err != nil {
		return stats, err
	}
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM article_sources`).Scan(&stats.TotalSources); err != nil {
		return stats, err
	}

	return stats, nil
}
