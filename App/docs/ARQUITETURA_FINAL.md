# Arquitetura final — Concordia

## Visão geral

```text
┌──────────────────────────────────────────────────────────┐
│                    CONCORDIA DESKTOP                     │
│                                                          │
│  React + TypeScript                                      │
│           │                                              │
│           ▼                                              │
│      Wails Bindings                                      │
│           │                                              │
│           ▼                                              │
│          Go                                              │
│   ┌───────┼───────────────┐                              │
│   │       │               │                              │
│ Auth   Curadoria     Busca/Indexação                     │
│           │               │                              │
│           │        Worker Pool / Goroutines              │
│           │               │                              │
│           └───────┬───────┘                              │
│                   ▼                                      │
│                 MySQL              filesystem/PDFs       │
└──────────────────────────────────────────────────────────┘
```

## Camadas Go

```text
App / Wails bindings
        ↓
Services
        ↓
Repositories
        ↓
MySQL
```

Principais responsabilidades:

- `config`: conexão MySQL;
- `models`: estruturas usadas pelo backend e bindings;
- `repositories`: SQL e persistência;
- `services/auth_service.go`: autenticação;
- `services/article_service.go`: curadoria e PDFs;
- `services/indexing_service.go`: worker pool e orquestração paralela;
- `services/pdf_extractor.go`: extração de texto;
- `services/tokenizer.go`: normalização e tokenização;
- `services/search_service.go`: BM25, filtros e busca aproximada.

## Persistência

Tabelas principais:

```text
users
articles
article_sources
authors
article_authors
terms
term_occurrences
article_index_stats
```

Os PDFs não são armazenados como BLOB. O arquivo fica em:

```text
storage/articles/
```

O banco mantém `file_path`, metadados e `checksum_sha256`.

## Indexação paralela

```text
Artigos pendentes
       ↓
canal de jobs
       ↓
┌─────────┬─────────┬─────────┐
│Worker 1 │Worker 2 │Worker N │  ← goroutines
└────┬────┴────┬────┴────┬────┘
     │         │         │
     └─────────┼─────────┘
               ↓
         índice invertido
               ↓
   terms + term_occurrences
               ↓
      article_index_stats
```

A quantidade de workers é configurável pelo Curador.

## BM25

A busca usa os postings do índice invertido e os comprimentos persistidos em `article_index_stats`.

Parâmetros usados:

```text
k1 = 1.2
b  = 0.75
```

Os resultados são ordenados por score decrescente.

## Perfis

### PESQUISADOR

- login;
- pesquisa;
- filtros;
- visualização de resultados e detalhes.

### CURADOR

Além das funções do Pesquisador:

- importação de PDFs;
- edição/exclusão de artigos;
- indexação;
- métricas;
- benchmark.

As permissões de curadoria são verificadas no backend Go, e não apenas escondidas no frontend.
