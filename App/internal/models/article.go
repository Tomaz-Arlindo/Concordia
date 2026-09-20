package models

import "time"

type ArticleSource struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ArticleSummary struct {
	ID              uint64    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Filename        string    `json:"filename"`
	SourceName      string    `json:"source_name"`
	Authors         string    `json:"authors"`
	LanguageCode    string    `json:"language_code"`
	PublicationDate string    `json:"publication_date"`
	IndexStatus     string    `json:"index_status"`
	CreatedAt       time.Time `json:"created_at"`
}

type ArticleDetail struct {
	ID              uint64    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Filename        string    `json:"filename"`
	FilePath        string    `json:"file_path"`
	ContentType     string    `json:"content_type"`
	Size            uint64    `json:"size"`
	UserID          uint64    `json:"user_id"`
	SourceID        uint64    `json:"source_id"`
	SourceName      string    `json:"source_name"`
	Authors         string    `json:"authors"`
	AbstractText    string    `json:"abstract_text"`
	ExternalID      string    `json:"external_id"`
	DOI             string    `json:"doi"`
	LanguageCode    string    `json:"language_code"`
	PublicationDate string    `json:"publication_date"`
	ChecksumSHA256  string    `json:"checksum_sha256"`
	IndexStatus     string    `json:"index_status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CollectionStats struct {
	TotalArticles   uint64 `json:"total_articles"`
	PendingArticles uint64 `json:"pending_articles"`
	IndexedArticles uint64 `json:"indexed_articles"`
	TotalAuthors    uint64 `json:"total_authors"`
	TotalSources    uint64 `json:"total_sources"`
}
