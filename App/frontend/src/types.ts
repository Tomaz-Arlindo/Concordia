export type User = {
  id: number
  name: string
  email: string
  role: 'PESQUISADOR' | 'CURADOR' | string
  created_at: string
}

export type HealthStatus = {
  status: string
  database: string
}

export type AuthResponse = {
  success: boolean
  message: string
  user?: User
}

export type ArticleSource = {
  id: number
  name: string
  description: string
}

export type ArticleSummary = {
  id: number
  title: string
  description: string
  filename: string
  source_name: string
  authors: string
  language_code: string
  publication_date: string
  index_status: string
  created_at: string
}

export type ArticleDetail = ArticleSummary & {
  file_path: string
  content_type: string
  size: number
  user_id: number
  source_id: number
  abstract_text: string
  external_id: string
  doi: string
  checksum_sha256: string
  updated_at: string
}

export type CollectionStats = {
  total_articles: number
  pending_articles: number
  indexed_articles: number
  total_authors: number
  total_sources: number
}

export type IndexingOverview = {
  pending: number
  processing: number
  indexed: number
  errors: number
  terms: number
  occurrences: number
  documents: number
}

export type IndexingReport = {
  workers: number
  total: number
  indexed: number
  failed: number
  duration_ms: number
  reindexed_all: boolean
  errors: string[]
}

export type SearchResult = {
  id: number
  title: string
  description: string
  filename: string
  source_name: string
  authors: string
  language_code: string
  publication_date: string
  index_status: string
  score: number
  matched_terms: string
}

export type BenchmarkResult = {
  workers: number
  processing_ms: number
  database_ms: number
  duration_ms: number
  indexed: number
  failed: number
  processing_speedup: number
  processing_efficiency: number
  total_speedup: number
  database_share: number
  errors: string[]
}

export type BenchmarkReport = {
  results: BenchmarkResult[]
  note: string
}
