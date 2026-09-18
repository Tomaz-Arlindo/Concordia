package models

type IndexTarget struct {
	ID       uint64 `json:"id"`
	Title    string `json:"title"`
	FilePath string `json:"file_path"`
}

type TermStat struct {
	Term      string `json:"term"`
	Frequency uint64 `json:"frequency"`
	Positions []int  `json:"positions"`
}

type IndexedDocument struct {
	ArticleID      uint64     `json:"article_id"`
	DocumentLength uint64     `json:"document_length"`
	Terms          []TermStat `json:"terms"`
}

type IndexingReport struct {
	Workers      int      `json:"workers"`
	Total        int      `json:"total"`
	Indexed      int      `json:"indexed"`
	Failed       int      `json:"failed"`
	DurationMS   int64    `json:"duration_ms"`
	ReindexedAll bool     `json:"reindexed_all"`
	Errors       []string `json:"errors"`
}

type IndexingOverview struct {
	Pending     uint64 `json:"pending"`
	Processing  uint64 `json:"processing"`
	Indexed     uint64 `json:"indexed"`
	Errors      uint64 `json:"errors"`
	Terms       uint64 `json:"terms"`
	Occurrences uint64 `json:"occurrences"`
	Documents   uint64 `json:"documents"`
}

type SearchResult struct {
	ID              uint64  `json:"id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Filename        string  `json:"filename"`
	SourceName      string  `json:"source_name"`
	Authors         string  `json:"authors"`
	LanguageCode    string  `json:"language_code"`
	PublicationDate string  `json:"publication_date"`
	IndexStatus     string  `json:"index_status"`
	Score           float64 `json:"score"`
	MatchedTerms    string  `json:"matched_terms"`
}

type SearchFilters struct {
	Source   string `json:"source"`
	Author   string `json:"author"`
	Language string `json:"language"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
	Fuzzy    bool   `json:"fuzzy"`
	Limit    int    `json:"limit"`
}

type BenchmarkResult struct {
	Workers              int      `json:"workers"`
	ProcessingMS         int64    `json:"processing_ms"`
	DatabaseMS           int64    `json:"database_ms"`
	DurationMS           int64    `json:"duration_ms"`
	Indexed              int      `json:"indexed"`
	Failed               int      `json:"failed"`
	ProcessingSpeedup    float64  `json:"processing_speedup"`
	ProcessingEfficiency float64  `json:"processing_efficiency"`
	TotalSpeedup         float64  `json:"total_speedup"`
	DatabaseShare        float64  `json:"database_share"`
	Errors               []string `json:"errors"`
}

type BenchmarkReport struct {
	Results []BenchmarkResult `json:"results"`
	Note    string            `json:"note"`
}
