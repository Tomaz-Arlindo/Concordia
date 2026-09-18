package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"desktop/internal/models"
)

type IndexRepository struct {
	DB *sql.DB
}

func NewIndexRepository(db *sql.DB) *IndexRepository {
	return &IndexRepository{DB: db}
}

func (r *IndexRepository) ListTargets(ctx context.Context, reindexAll bool) ([]models.IndexTarget, error) {
	query := `SELECT id, title, COALESCE(file_path, '') FROM articles WHERE COALESCE(file_path, '') <> ''`
	if !reindexAll {
		query += ` AND UPPER(index_status) IN ('PENDING', 'ERROR')`
	}
	query += ` ORDER BY id`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.IndexTarget, 0)
	for rows.Next() {
		var item models.IndexTarget
		if err := rows.Scan(&item.ID, &item.Title, &item.FilePath); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *IndexRepository) MarkStatus(ctx context.Context, articleID uint64, status string) error {
	_, err := r.DB.ExecContext(
		ctx,
		`UPDATE articles
		 SET index_status = ?,
		     indexed_at = CASE WHEN UPPER(?) = 'INDEXED' THEN CURRENT_TIMESTAMP ELSE indexed_at END
		 WHERE id = ?`,
		status,
		status,
		articleID,
	)
	return err
}

func (r *IndexRepository) PersistDocument(ctx context.Context, doc models.IndexedDocument) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM term_occurrences WHERE article_id = ?`, doc.ArticleID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM article_index_stats WHERE article_id = ?`, doc.ArticleID); err != nil {
		return err
	}

	for _, stat := range doc.Terms {
		result, err := tx.ExecContext(
			ctx,
			`INSERT INTO terms (term, document_frequency)
			 VALUES (?, 0)
			 ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)`,
			stat.Term,
		)
		if err != nil {
			return fmt.Errorf("salvar termo %q: %w", stat.Term, err)
		}

		termID, err := result.LastInsertId()
		if err != nil || termID == 0 {
			if scanErr := tx.QueryRowContext(ctx, `SELECT id FROM terms WHERE term = ?`, stat.Term).Scan(&termID); scanErr != nil {
				return scanErr
			}
		}

		positionsJSON, err := json.Marshal(stat.Positions)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO term_occurrences (term_id, article_id, term_frequency, positions)
			 VALUES (?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE
			   term_frequency = VALUES(term_frequency),
			   positions = VALUES(positions)`,
			termID,
			doc.ArticleID,
			stat.Frequency,
			string(positionsJSON),
		)
		if err != nil {
			return fmt.Errorf("salvar ocorrência do termo %q: %w", stat.Term, err)
		}
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO article_index_stats (article_id, document_length, indexed_terms)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   document_length = VALUES(document_length),
		   indexed_terms = VALUES(indexed_terms),
		   updated_at = CURRENT_TIMESTAMP`,
		doc.ArticleID,
		doc.DocumentLength,
		len(doc.Terms),
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE articles
		 SET index_status = 'INDEXED', indexed_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		doc.ArticleID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *IndexRepository) RecalculateDocumentFrequency(ctx context.Context) error {
	_, err := r.DB.ExecContext(
		ctx,
		`UPDATE terms t
		 LEFT JOIN (
		   SELECT o.term_id, COUNT(*) AS df
		   FROM term_occurrences o
		   JOIN articles a ON a.id = o.article_id
		   WHERE UPPER(a.index_status) = 'INDEXED'
		   GROUP BY o.term_id
		 ) x ON x.term_id = t.id
		 SET t.document_frequency = COALESCE(x.df, 0)`,
	)
	if err != nil {
		return err
	}

	_, err = r.DB.ExecContext(ctx, `DELETE FROM terms WHERE document_frequency = 0`)
	return err
}

func (r *IndexRepository) Overview(ctx context.Context) (models.IndexingOverview, error) {
	var out models.IndexingOverview

	queries := []struct {
		SQL  string
		Dest *uint64
	}{
		{`SELECT COUNT(*) FROM articles WHERE UPPER(index_status) = 'PENDING'`, &out.Pending},
		{`SELECT COUNT(*) FROM articles WHERE UPPER(index_status) = 'PROCESSING'`, &out.Processing},
		{`SELECT COUNT(*) FROM articles WHERE UPPER(index_status) = 'INDEXED'`, &out.Indexed},
		{`SELECT COUNT(*) FROM articles WHERE UPPER(index_status) = 'ERROR'`, &out.Errors},
		{`SELECT COUNT(*) FROM terms`, &out.Terms},
		{`SELECT COUNT(*) FROM term_occurrences`, &out.Occurrences},
		{`SELECT COUNT(*) FROM article_index_stats`, &out.Documents},
	}

	for _, query := range queries {
		if err := r.DB.QueryRowContext(ctx, query.SQL).Scan(query.Dest); err != nil {
			return out, err
		}
	}
	return out, nil
}

func (r *IndexRepository) CorpusStats(ctx context.Context) (documents uint64, avgLength float64, err error) {
	err = r.DB.QueryRowContext(
		ctx,
		`SELECT COUNT(*), COALESCE(AVG(document_length), 0)
		 FROM article_index_stats`,
	).Scan(&documents, &avgLength)
	return
}

func (r *IndexRepository) TermInfo(ctx context.Context, term string) (termID uint64, df uint64, found bool, err error) {
	err = r.DB.QueryRowContext(
		ctx,
		`SELECT id, document_frequency FROM terms WHERE term = ? LIMIT 1`,
		term,
	).Scan(&termID, &df)
	if err == sql.ErrNoRows {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}
	return termID, df, true, nil
}

type Posting struct {
	ArticleID      uint64
	TermFrequency  uint64
	DocumentLength uint64
}

func (r *IndexRepository) Postings(ctx context.Context, termID uint64) ([]Posting, error) {
	rows, err := r.DB.QueryContext(
		ctx,
		`SELECT o.article_id, o.term_frequency, s.document_length
		 FROM term_occurrences o
		 JOIN article_index_stats s ON s.article_id = o.article_id
		 JOIN articles a ON a.id = o.article_id
		 WHERE o.term_id = ? AND UPPER(a.index_status) = 'INDEXED'`,
		termID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Posting, 0)
	for rows.Next() {
		var p Posting
		if err := rows.Scan(&p.ArticleID, &p.TermFrequency, &p.DocumentLength); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *IndexRepository) SearchResultByArticleID(ctx context.Context, articleID uint64) (models.SearchResult, error) {
	var item models.SearchResult
	var authors sql.NullString

	err := r.DB.QueryRowContext(
		ctx,
		`SELECT
			a.id,
			a.title,
			COALESCE(a.description, ''),
			COALESCE(a.filename, ''),
			COALESCE(s.name, ''),
			auth.authors,
			COALESCE(a.language_code, ''),
			COALESCE(DATE_FORMAT(a.publication_date, '%Y-%m-%d'), ''),
			a.index_status
		 FROM articles a
		 LEFT JOIN article_sources s ON s.id = a.source_id
		 LEFT JOIN (
			SELECT aa.article_id,
			       GROUP_CONCAT(au.name ORDER BY aa.author_order SEPARATOR ', ') AS authors
			FROM article_authors aa
			JOIN authors au ON au.id = aa.author_id
			GROUP BY aa.article_id
		 ) auth ON auth.article_id = a.id
		 WHERE a.id = ?
		 LIMIT 1`,
		articleID,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Filename,
		&item.SourceName,
		&authors,
		&item.LanguageCode,
		&item.PublicationDate,
		&item.IndexStatus,
	)
	if err != nil {
		return item, err
	}
	item.Authors = strings.TrimSpace(authors.String)
	return item, nil
}

type TermCandidate struct {
	Term              string
	DocumentFrequency uint64
}

func (r *IndexRepository) CandidateTerms(ctx context.Context, prefix string, minLen int, maxLen int, limit int) ([]TermCandidate, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if minLen < 1 {
		minLen = 1
	}
	if maxLen < minLen {
		maxLen = minLen
	}

	rows, err := r.DB.QueryContext(
		ctx,
		`SELECT term, document_frequency
		 FROM terms
		 WHERE term LIKE CONCAT(?, '%')
		   AND CHAR_LENGTH(term) BETWEEN ? AND ?
		 ORDER BY document_frequency DESC, term ASC
		 LIMIT ?`,
		prefix,
		minLen,
		maxLen,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]TermCandidate, 0)
	for rows.Next() {
		var item TermCandidate
		if err := rows.Scan(&item.Term, &item.DocumentFrequency); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
