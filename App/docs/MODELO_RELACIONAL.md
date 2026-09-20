# Modelo relacional — Concordia

## Entidades e relações

```text
users 1 ───── N articles
article_sources 1 ───── N articles
articles N ───── N authors        via article_authors
articles 1 ───── N term_occurrences
terms    1 ───── N term_occurrences
articles 1 ───── 1 article_index_stats
```

## Tabelas

### users

Identidade, autenticação e perfil (`PESQUISADOR` / `CURADOR`).

### article_sources

Origem lógica do artigo, como `local`, `arxiv`, `pubmed` e `openalex`.

### articles

Metadados do documento e referência ao PDF no filesystem. Também guarda `index_status`, `indexed_at` e SHA-256.

### authors

Autores normalizados.

### article_authors

Relacionamento N:N entre artigo e autor, preservando `author_order`.

### terms

Vocabulário global do índice invertido. `document_frequency` é usado no cálculo do IDF do BM25.

### term_occurrences

Posting do índice invertido por termo/documento, com `term_frequency` e `positions`.

### article_index_stats

Comprimento do documento e quantidade de termos indexados, usados na normalização BM25.
