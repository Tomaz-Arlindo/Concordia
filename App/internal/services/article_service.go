package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"desktop/internal/models"
	"desktop/internal/repositories"
)

var (
	ErrInvalidArticle = errors.New("dados do artigo inválidos")
	ErrDuplicateFile  = errors.New("arquivo já importado")
)

type ArticleService struct {
	db       *sql.DB
	articles *repositories.ArticleRepository
	sources  *repositories.SourceRepository
}

type ImportArticleInput struct {
	Title           string
	Description     string
	SourceID        uint64
	AuthorsCSV      string
	LanguageCode    string
	PublicationDate string
	DOI             string
	ExternalID      string
	OriginalPath    string
	UserID          uint64
}

type UpdateArticleInput struct {
	ID              uint64
	Title           string
	Description     string
	SourceID        uint64
	AuthorsCSV      string
	LanguageCode    string
	PublicationDate string
	DOI             string
	ExternalID      string
}

func NewArticleService(
	db *sql.DB,
	articles *repositories.ArticleRepository,
	sources *repositories.SourceRepository,
) *ArticleService {
	return &ArticleService{
		db:       db,
		articles: articles,
		sources:  sources,
	}
}

func (s *ArticleService) ListSources(ctx context.Context) ([]models.ArticleSource, error) {
	return s.sources.List(ctx)
}

func (s *ArticleService) ListArticles(ctx context.Context) ([]models.ArticleSummary, error) {
	return s.articles.List(ctx)
}

func (s *ArticleService) SearchArticles(ctx context.Context, query string) ([]models.ArticleSummary, error) {
	return s.articles.Search(ctx, query)
}

func (s *ArticleService) GetArticle(ctx context.Context, id uint64) (*models.ArticleDetail, error) {
	return s.articles.GetByID(ctx, id)
}

func (s *ArticleService) GetStats(ctx context.Context) (models.CollectionStats, error) {
	return s.articles.GetStats(ctx)
}

func (s *ArticleService) ImportPDF(ctx context.Context, input ImportArticleInput) (uint64, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.LanguageCode = strings.TrimSpace(input.LanguageCode)
	input.PublicationDate = strings.TrimSpace(input.PublicationDate)
	input.DOI = strings.TrimSpace(input.DOI)
	input.ExternalID = strings.TrimSpace(input.ExternalID)

	if input.Title == "" || input.SourceID == 0 || input.OriginalPath == "" || input.UserID == 0 {
		return 0, ErrInvalidArticle
	}

	if strings.ToLower(filepath.Ext(input.OriginalPath)) != ".pdf" {
		return 0, ErrInvalidArticle
	}

	checksum, size, err := inspectPDF(input.OriginalPath)
	if err != nil {
		return 0, ErrInvalidArticle
	}

	exists, err := s.articles.ChecksumExists(ctx, checksum)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, ErrDuplicateFile
	}

	storageDir := filepath.Join("storage", "articles")
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return 0, err
	}

	destinationRelative := filepath.ToSlash(filepath.Join(storageDir, checksum+".pdf"))
	destinationOS := filepath.FromSlash(destinationRelative)

	if err := copyFile(input.OriginalPath, destinationOS); err != nil {
		return 0, err
	}

	articleID, err := s.createArticleTransaction(ctx, input, checksum, uint64(size), destinationRelative)
	if err != nil {
		_ = os.Remove(destinationOS)
		return 0, err
	}

	return articleID, nil
}

func (s *ArticleService) createArticleTransaction(
	ctx context.Context,
	input ImportArticleInput,
	checksum string,
	size uint64,
	filePath string,
) (uint64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO articles (
			title,
			description,
			filename,
			file_path,
			content_type,
			size,
			user_id,
			source_id,
			external_id,
			doi,
			language_code,
			publication_date,
			checksum_sha256,
			index_status
		) VALUES (
			?,
			NULLIF(?, ''),
			?,
			?,
			'application/pdf',
			?,
			?,
			?,
			NULLIF(?, ''),
			NULLIF(?, ''),
			NULLIF(?, ''),
			NULLIF(?, ''),
			?,
			'PENDING'
		)`,
		input.Title,
		input.Description,
		filepath.Base(input.OriginalPath),
		filePath,
		size,
		input.UserID,
		input.SourceID,
		input.ExternalID,
		input.DOI,
		input.LanguageCode,
		input.PublicationDate,
		checksum,
	)
	if err != nil {
		return 0, err
	}

	rawID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	articleID := uint64(rawID)

	if err := replaceAuthors(ctx, tx, articleID, parseAuthors(input.AuthorsCSV)); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return articleID, nil
}

func (s *ArticleService) UpdateArticle(ctx context.Context, input UpdateArticleInput) error {
	input.Title = strings.TrimSpace(input.Title)
	if input.ID == 0 || input.Title == "" || input.SourceID == 0 {
		return ErrInvalidArticle
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`UPDATE articles
		 SET
			title = ?,
			description = NULLIF(?, ''),
			source_id = ?,
			external_id = NULLIF(?, ''),
			doi = NULLIF(?, ''),
			language_code = NULLIF(?, ''),
			publication_date = NULLIF(?, '')
		 WHERE id = ?`,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Description),
		input.SourceID,
		strings.TrimSpace(input.ExternalID),
		strings.TrimSpace(input.DOI),
		strings.TrimSpace(input.LanguageCode),
		strings.TrimSpace(input.PublicationDate),
		input.ID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		if _, err := s.articles.GetByID(ctx, input.ID); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM article_authors WHERE article_id = ?`, input.ID); err != nil {
		return err
	}
	if err := replaceAuthors(ctx, tx, input.ID, parseAuthors(input.AuthorsCSV)); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *ArticleService) DeleteArticle(ctx context.Context, id uint64) error {
	path, err := s.articles.Delete(ctx, id)
	if err != nil {
		return err
	}

	if path != "" {
		_ = os.Remove(filepath.FromSlash(path))
	}
	return nil
}

func inspectPDF(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	header := make([]byte, 5)
	if _, err := io.ReadFull(file, header); err != nil {
		return "", 0, err
	}
	if string(header) != "%PDF-" {
		return "", 0, errors.New("arquivo não é um PDF válido")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", 0, err
	}

	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func copyFile(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()

	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func parseAuthors(csv string) []string {
	raw := strings.Split(csv, ",")
	seen := make(map[string]string)

	for _, value := range raw {
		name := strings.TrimSpace(value)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; !exists {
			seen[key] = name
		}
	}

	names := make([]string, 0, len(seen))
	for _, name := range seen {
		names = append(names, name)
	}
	sort.SliceStable(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})
	return names
}

func replaceAuthors(ctx context.Context, tx *sql.Tx, articleID uint64, authors []string) error {
	for index, name := range authors {
		var authorID uint64

		err := tx.QueryRowContext(
			ctx,
			`SELECT id FROM authors WHERE LOWER(name) = LOWER(?) ORDER BY id LIMIT 1`,
			name,
		).Scan(&authorID)

		if errors.Is(err, sql.ErrNoRows) {
			result, insertErr := tx.ExecContext(ctx, `INSERT INTO authors (name) VALUES (?)`, name)
			if insertErr != nil {
				return insertErr
			}
			rawID, insertErr := result.LastInsertId()
			if insertErr != nil {
				return insertErr
			}
			authorID = uint64(rawID)
		} else if err != nil {
			return err
		}

		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO article_authors (article_id, author_id, author_order)
			 VALUES (?, ?, ?)`,
			articleID,
			authorID,
			index+1,
		)
		if err != nil {
			return fmt.Errorf("vincular autor %q: %w", name, err)
		}
	}

	return nil
}
