# Concordia

Motor de busca paralelo para artigos científicos desenvolvido como projeto integrado das disciplinas de **Fábrica de Software** e **Tópicos Avançados em Ciência da Computação**.

## Sobre o projeto

O Concordia é uma aplicação desktop para importação, indexação e busca de artigos científicos. O núcleo é desenvolvido em **Go**, com interface em **React + TypeScript** integrada por **Wails**, persistência em **MySQL 8** e armazenamento dos PDFs no filesystem local.

A indexação utiliza um **worker pool com goroutines** para paralelizar leitura, extração, normalização, tokenização e construção do índice local. A busca usa **índice invertido + BM25**, além de filtros e busca aproximada.

## Equipe

- Maria Karolina Rodrigues de Andrade
- Pedro Felipe Oliveira Tavares
- Pedro Oliveira Barros Batista de Lima
- Tomaz Arlindo Silva Ribeiro
- Vladison Lucas Costa Dos Santos

**Turma:** CC-8-MB

## Tecnologias

- **Backend / núcleo:** Go
- **Desktop:** Wails v2
- **Frontend:** React + TypeScript
- **Banco de dados:** MySQL 8
- **Armazenamento de documentos:** filesystem local
- **Paralelismo:** goroutines + worker pool
- **Busca:** índice invertido + BM25
- **Segurança:** hash de senha e controle de acesso por perfil

## Funcionalidades implementadas

### Pesquisador
- cadastro e autenticação;
- pesquisa no conteúdo dos artigos;
- ranking por BM25;
- busca aproximada;
- filtros por fonte, autor, idioma e data;
- visualização de detalhes.

### Curador
Além das funcionalidades do Pesquisador:
- importação individual de PDFs;
- importação em lote para dataset;
- gerenciamento de artigos;
- indexação do acervo;
- escolha do número de workers;
- métricas técnicas;
- benchmark de paralelismo.

## Pipeline de indexação

```text
PDF
→ extração de texto
→ normalização
→ tokenização
→ worker pool / goroutines
→ índice local
→ writer controlado
→ MySQL
```

A persistência usa um writer controlado para evitar contenção e deadlocks em escritas concorrentes.

## Benchmark da Sprint 2

Dataset utilizado: **21 documentos**.

| Workers | Processamento | Speedup proc. | Eficiência proc. | MySQL | Total | Speedup total |
|---:|---:|---:|---:|---:|---:|---:|
| 1 | 895 ms | 1,00x | 100,0% | 34.046 ms | 34.941 ms | 1,00x |
| 2 | 707 ms | 1,27x | 63,3% | 32.639 ms | 33.346 ms | 1,05x |
| 4 | 440 ms | 2,03x | 50,9% | 30.655 ms | 31.095 ms | 1,12x |
| 8 | 284 ms | **3,15x** | 39,4% | 29.614 ms | 29.898 ms | **1,17x** |

A fase em que as goroutines atuam apresentou speedup de **3,15x** com 8 workers. O ganho fim-a-fim é menor porque a persistência no MySQL representa a maior fração do tempo total, ilustrando na prática a **Lei de Amdahl**.

## Banco de dados

Principais tabelas:

- `users`
- `articles`
- `article_sources`
- `authors`
- `article_authors`
- `terms`
- `term_occurrences`
- `article_index_stats`

Os PDFs não são salvos como BLOB no MySQL. O banco guarda os metadados e o caminho do arquivo.

## Estrutura principal

```text
Concordia/
├── App/
├── Assets/
├── Docs/
│   ├── Sprints/
│   ├── arquitetura.md
│   ├── modelo-relacional.md
│   ├── diagramas.md
│   └── benchmark-sprint2.md
├── LICENSE
└── README.md
```

## Executando o projeto

### Requisitos

- Go
- Node.js + npm
- MySQL 8
- Wails v2
- WebView2 no Windows

### Configuração

Na pasta `App`, crie `.env` a partir de `.env.example` e preencha as credenciais locais do MySQL.

Depois:

```powershell
cd App
go mod tidy
go test ./...
wails dev
```

> O arquivo `.env`, PDFs, `node_modules` e binários gerados não devem ser versionados.

## Status

### Sprint 2 — versão funcional

Nesta Sprint foram concluídos os artefatos de arquitetura e banco previstos e, adicionalmente, já existe uma versão funcional integrada com autenticação, importação, indexação paralela, busca BM25, filtros e benchmark.

Consulte a documentação em [`Docs/`](Docs/).
