# Modelo Relacional

*(colunas conferidas contra o código de acesso a dados — `internal/repositories/*.go` e `internal/services/article_service.go` — em 21/09)*

## Relações principais

- `users 1:N articles`
- `article_sources 1:N articles`
- `articles N:N authors` por `article_authors`
- `articles N:N terms` por `term_occurrences`
- `articles 1:1 article_index_stats`

## Tabelas

### users
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | auto-incremento |
| name | VARCHAR | | |
| email | VARCHAR | UNIQUE | usado no login |
| password_hash | VARCHAR | | hash bcrypt, nunca exposto pela API |
| role | VARCHAR/ENUM | | `PESQUISADOR` ou `CURADOR` |
| created_at | DATETIME | | |

### articles
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | |
| title | VARCHAR | | |
| description | TEXT | NULL | |
| filename | VARCHAR | | |
| file_path | VARCHAR | | caminho no filesystem (`storage/articles/`) |
| content_type | VARCHAR | | ex.: `application/pdf` |
| size | BIGINT UNSIGNED | | tamanho do arquivo em bytes |
| user_id | BIGINT UNSIGNED | FK → users.id | curador que importou |
| source_id | BIGINT UNSIGNED | FK → article_sources.id | |
| abstract_text | TEXT | NULL | |
| external_id | VARCHAR | NULL | |
| doi | VARCHAR | NULL | |
| language_code | VARCHAR | NULL | |
| publication_date | DATE | NULL | |
| checksum_sha256 | VARCHAR | UNIQUE | evita reimportação do mesmo PDF |
| index_status | VARCHAR/ENUM | | `PENDING`, `PROCESSING`, `INDEXED` (entre outros) |
| indexed_at | DATETIME | NULL | preenchido quando `index_status = INDEXED` |
| created_at | DATETIME | | |
| updated_at | DATETIME | | |

### article_sources
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | |
| name | VARCHAR | UNIQUE | ex.: `local`, `arxiv`, `pubmed` |
| description | VARCHAR | NULL | |
| created_at | DATETIME | | |

### authors
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | |
| name | VARCHAR | | deduplicado por nome (case-insensitive) na importação |
| orcid | VARCHAR | NULL | |
| created_at | DATETIME | | |

### article_authors
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| article_id | BIGINT UNSIGNED | PK (composta), FK → articles.id | |
| author_id | BIGINT UNSIGNED | PK (composta), FK → authors.id | |
| author_order | INT | | ordem de exibição dos autores |

### terms
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | |
| term | VARCHAR | UNIQUE | token normalizado |
| document_frequency | BIGINT UNSIGNED | | usado no IDF do BM25 |
| created_at | DATETIME | | |

### term_occurrences
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| term_id | BIGINT UNSIGNED | PK (composta), FK → terms.id | |
| article_id | BIGINT UNSIGNED | PK (composta), FK → articles.id | |
| term_frequency | BIGINT UNSIGNED | | ocorrências do termo no documento |
| positions | JSON/TEXT | | posições do termo no texto |

### article_index_stats
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| article_id | BIGINT UNSIGNED | PK, FK → articles.id | |
| document_length | BIGINT UNSIGNED | | usado na normalização do BM25 |
| indexed_terms | BIGINT UNSIGNED | | |
| updated_at | DATETIME | | |

> **Nota:** este documento descreve o esquema físico (tabelas do MySQL). Ele é diferente do modelo em [`Modelos/modelo-relacional-concordia.md`](./Modelos/modelo-relacional-concordia.md), que documenta as *structs* Go da camada de aplicação (DTOs/views como `ArticleSummary`, `ArticleDetail`) — útil como referência de código, mas não corresponde 1:1 às tabelas do banco.
