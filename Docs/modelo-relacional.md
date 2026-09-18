# Modelo Relacional

## Relações principais

- `users 1:N articles`
- `article_sources 1:N articles`
- `articles N:N authors` por `article_authors`
- `articles N:N terms` por `term_occurrences`
- `articles 1:1 article_index_stats`

## Tabelas

### users
`id`, `name`, `email`, `password_hash`, `role`, `created_at`

### articles
`id`, `title`, `description`, `filename`, `file_path`, `content_type`, `size`, `user_id`, `source_id`, `abstract_text`, `external_id`, `doi`, `language_code`, `publication_date`, `checksum_sha256`, `index_status`, `indexed_at`, `created_at`, `updated_at`

### article_sources
`id`, `name`, `description`, `created_at`

### authors
`id`, `name`, `orcid`, `created_at`

### article_authors
`article_id`, `author_id`, `author_order`

### terms
`id`, `term`, `document_frequency`, `created_at`

### term_occurrences
`term_id`, `article_id`, `term_frequency`, `positions`

### article_index_stats
`article_id`, `document_length`, `indexed_terms`, `updated_at`
