# Sprint 3 — Concordia

## 1. Identificação da equipe

- **Integrantes:**
    - Maria Karolina Rodrigues de Andrade
    - Pedro Felipe Oliveira Tavares
    - Pedro Oliveira Barros Batista de Lima
    - Tomaz Arlindo Silva Ribeiro
    - Vladison Lucas Costa Dos Santos
- **Turma:** CC-8-MB
- **Nome do projeto:** Concordia

## 2. Banco de dados conectado

A conexão com o MySQL está implementada em `internal/config/database.go` (`ConnectDatabase`), lendo as credenciais de variáveis de ambiente:

| Variável | Padrão |
|---|---|
| `DB_HOST` | `localhost` |
| `DB_PORT` | `3306` |
| `DB_USER` | `root` |
| `DB_PASSWORD` | *(sem padrão)* |
| `DB_NAME` | `concordia` |

A conexão usa `database/sql` com o driver `go-sql-driver/mysql`, pool configurado (`SetConnMaxLifetime(5min)`, `SetMaxOpenConns(10)`, `SetMaxIdleConns(5)`) e é validada com `Ping()` no startup da aplicação (`app.go`).

O schema (8 tabelas, índices e chaves estrangeiras) está versionado em `App/database/schema.sql` e pode ser aplicado com `mysql -u root -p < App/database/schema.sql` antes de subir a aplicação.

## 3. Login funcional

Implementado em `AuthService.Login` (`services/auth_service.go`), exposto ao frontend via `App.Login` (`app.go`):

1. Normaliza o email (trim + lowercase).
2. Busca o usuário por email (`UserRepository.FindByEmail`).
3. Compara a senha informada com o hash armazenado via `bcrypt.CompareHashAndPassword`.
4. Em caso de sucesso, define o usuário atual na sessão da aplicação (`setCurrentUser`) e retorna os dados do usuário; em caso de falha, retorna `ErrInvalidCredentials` sem revelar se o erro foi de email ou senha.

## 4. Cadastro de usuários

Implementado em `AuthService.Register` (`services/auth_service.go`), exposto via `App.Register`:

1. Valida nome e email não vazios e senha com no mínimo 8 caracteres (`ErrInvalidInput` caso contrário).
2. Verifica se o email já existe (`ErrEmailAlreadyExists` caso positivo).
3. Gera o hash da senha com `bcrypt.GenerateFromPassword`.
4. Persiste o usuário (`UserRepository.Create`) com papel padrão `PESQUISADOR`.
5. Define o usuário recém-criado como usuário atual da sessão.

## 5. Controle de perfis

O sistema define dois perfis (`models.RolePesquisador = "PESQUISADOR"`, `models.RoleCurador = "CURADOR"`):

- **PESQUISADOR:** login, pesquisa, filtros, visualização de detalhes.
- **CURADOR:** funções do Pesquisador + importação, edição/exclusão de artigos, indexação e benchmark.

O controle é feito no backend, por `App.requireCurator()` (`app.go`), chamado em todas as rotas restritas: `ImportArticle`, `ImportPDFBatch`, `SelectPDF`, `UpdateArticle`, `DeleteArticle`, `StartIndexing` e `RunIndexingBenchmark`. As demais rotas de leitura (`ListArticles`, `SearchArticles`, `GetArticle`, `GetCollectionStats`, `GetIndexingOverview`, `SearchBM25`, `SearchBM25Filtered`) exigem apenas `authenticatedUser()` (usuário logado, de qualquer perfil).

> **PENDENTE :** hoje não existe nenhum caminho, pelo próprio sistema, para criar ou promover um usuário a `CURADOR` — `Register` sempre atribui `PESQUISADOR`, e não há endpoint de promoção nem usuário `CURADOR` inicial (seed). Na prática, isso só é possível hoje editando a linha diretamente no banco (`UPDATE users SET role = 'CURADOR' WHERE id = ...`).  a equipe ainda irá decidir a melhor forma de criar e/ou promover um usuário a `CURADOR`.

## 6. CRUD principal funcionando

CRUD da entidade `Article`, implementado em `ArticleService` (`services/article_service.go`) e `ArticleRepository` (`repositories/article_repository.go`), exposto via `ArticleAPI` (`article_api.go`):

- **Create:** `ImportPDF`/`ImportArticle` — grava o artigo em transação (`createArticleTransaction`), vinculando autores (`replaceAuthors`) e evitando duplicidade por `checksum_sha256`.
- **Read:** `ListArticles`, `SearchArticles` (busca simples por título/descrição/fonte/autores) e `GetArticle` (detalhe por id).
- **Update:** `UpdateArticle` — atualiza título, descrição, fonte, identificadores externos, idioma e data de publicação.
- **Delete:** `DeleteArticle` — remove o registro e devolve o caminho do arquivo para exclusão no filesystem.

Todas as operações de escrita (Create/Update/Delete) exigem perfil `CURADOR`; leitura exige apenas usuário autenticado (ver item 5).

## 7. Primeiro deploy local

**Requisitos:** Go, Node.js + npm, MySQL 8, Wails v2, WebView2 (no Windows).

**Configuração:** na pasta `App`, criar `.env` a partir de `.env.example` e preencher as credenciais locais do MySQL.

**Execução:**
```powershell
mysql -u root -p < App/database/schema.sql
cd App
go mod tidy
go test ./...
wails dev
```

> O arquivo `.env`, PDFs, `node_modules` e binários gerados não devem ser versionados.
