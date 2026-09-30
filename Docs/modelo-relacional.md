# Modelo Relacional

*(gerado a partir do `App/database/schema.sql`, schema oficial do MySQL 8, adicionado em 30/09)*

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
| name | VARCHAR(150) | | |
| email | VARCHAR(255) | UNIQUE | usado no login |
| password_hash | VARCHAR(255) | | hash bcrypt, nunca exposto pela API |
| role | VARCHAR(30) | | `PESQUISADOR` (padrão) ou `CURADOR` |
| created_at | DATETIME | | |

### articles
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | |
| title | VARCHAR(500) | | |
| description | TEXT | NULL | |
| filename | VARCHAR(255) | NULL | |
| file_path | VARCHAR(1000) | NULL | caminho no filesystem (`storage/articles/`) |
| content_type | VARCHAR(100) | NULL | ex.: `application/pdf` |
| size | BIGINT UNSIGNED | NULL | tamanho do arquivo em bytes |
| user_id | BIGINT UNSIGNED | FK → users.id | `ON DELETE SET NULL`; curador que importou |
| source_id | BIGINT UNSIGNED | FK → article_sources.id | `ON DELETE SET NULL` |
| abstract_text | TEXT | NULL | |
| external_id | VARCHAR(255) | UNIQUE, NULL | |
| doi | VARCHAR(255) | UNIQUE, NULL | |
| language_code | VARCHAR(10) | NULL | |
| publication_date | DATE | NULL | |
| checksum_sha256 | CHAR(64) | UNIQUE, NULL | evita reimportação do mesmo PDF |
| index_status | VARCHAR(30) | | padrão `PENDING`; também `PROCESSING`, `INDEXED` |
| indexed_at | DATETIME | NULL | preenchido quando `index_status = INDEXED` |
| created_at | DATETIME | | |
| updated_at | DATETIME | | atualizado automaticamente (`ON UPDATE CURRENT_TIMESTAMP`) |

### article_sources
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | |
| name | VARCHAR(100) | UNIQUE | `local`, `arxiv`, `pubmed`, `openalex` (seed inicial) |
| description | VARCHAR(500) | NULL | |
| created_at | DATETIME | | |

### authors
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | |
| name | VARCHAR(255) | | deduplicado por nome na aplicação (case-insensitive) |
| orcid | VARCHAR(50) | UNIQUE, NULL | |
| created_at | DATETIME | | |

### article_authors
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| article_id | BIGINT UNSIGNED | PK (composta), FK → articles.id | `ON DELETE CASCADE` |
| author_id | BIGINT UNSIGNED | PK (composta), FK → authors.id | `ON DELETE CASCADE` |
| author_order | INT UNSIGNED | | ordem de exibição dos autores (padrão 1) |

### terms
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | |
| term | VARCHAR(255) | UNIQUE | token normalizado |
| document_frequency | BIGINT UNSIGNED | | usado no IDF do BM25 |
| created_at | DATETIME | | |

### term_occurrences
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| term_id | BIGINT UNSIGNED | PK (composta), FK → terms.id | `ON DELETE CASCADE` |
| article_id | BIGINT UNSIGNED | PK (composta), FK → articles.id | `ON DELETE CASCADE` |
| term_frequency | BIGINT UNSIGNED | | ocorrências do termo no documento |
| positions | JSON | NULL | posições do termo no texto |

### article_index_stats
| Coluna | Tipo | Chave | Observação |
|---|---|---|---|
| article_id | BIGINT UNSIGNED | PK, FK → articles.id | `ON DELETE CASCADE` |
| document_length | BIGINT UNSIGNED | | usado na normalização do BM25 |
| indexed_terms | BIGINT UNSIGNED | | |
| updated_at | DATETIME | | atualizado automaticamente |

> **Nota:** este documento descreve o esquema físico (tabelas do MySQL, agora versionado em `App/database/schema.sql`). Ele é diferente do modelo em [`Modelos/modelo-relacional-concordia.md`](./Modelos/modelo-relacional-concordia.md), que documenta as *structs* Go da camada de aplicação (DTOs/views como `ArticleSummary`, `ArticleDetail`) — útil como referência de código, mas não corresponde 1:1 às tabelas do banco.
