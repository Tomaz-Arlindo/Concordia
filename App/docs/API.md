# Documentação das APIs — Concordia

Este documento descreve os **30 métodos** que o backend em Go expõe à interface do Concordia. Eles são o equivalente aos *endpoints* do sistema.

## Sumário

1. [Visão geral e justificativa](#1-visão-geral-e-justificativa)
2. [Como chamar](#2-como-chamar)
3. [Autenticação e autorização](#3-autenticação-e-autorização)
4. [Convenções de resposta](#4-convenções-de-resposta)
5. [Tipos de dados](#5-tipos-de-dados)
6. [Sistema e sessão](#6-sistema-e-sessão)
7. [Artigos](#7-artigos)
8. [Indexação e busca](#8-indexação-e-busca)
9. [Histórico de operações](#9-histórico-de-operações)
10. [Gestão de usuários](#10-gestão-de-usuários)
11. [Operações registradas no histórico](#11-operações-registradas-no-histórico)
12. [Resumo dos métodos](#12-resumo-dos-métodos)

---

## 1. Visão geral e justificativa

O Concordia **não possui API HTTP**: não há servidor web, porta de rede, rotas, verbos GET/POST nem tokens. Ele é um **aplicativo desktop** construído com o [Wails](https://wails.io), que executa o backend em Go e a interface em React dentro do mesmo processo.

Nesse modelo, o mecanismo de integração é a **ponte do Wails (*bindings*)**: cada método público da struct `App` (arquivos `*_api.go` na raiz do projeto) é exposto à interface como uma função TypeScript assíncrona, e os dados trafegam serializados em JSON dentro do próprio processo. Essa escolha elimina a superfície de ataque de um servidor HTTP local e dispensa autenticação por token, já que a sessão vive no processo do aplicativo.

| Conceito em uma API REST | Equivalente no Concordia |
| --- | --- |
| Endpoint (`POST /login`) | Método Go exposto, chamado como `Login(email, password)` |
| Corpo da requisição (JSON) | Parâmetros posicionais do método |
| Corpo da resposta (JSON) | Valor retornado, serializado em JSON |
| Código HTTP 2xx | `success: true` |
| Código 400, 401, 403, 404 ou 409 | `success: false` com `message` descrevendo o motivo |
| Token / cookie de sessão | Usuário da sessão mantido em memória pelo backend |

Os arquivos de ligação gerados pelo Wails ficam em `frontend/wailsjs/go/main/App.js` (funções) e `App.d.ts` (tipos).

## 2. Como chamar

Cada método é importado de `frontend/wailsjs/go/main/App` e chamado com os parâmetros **na ordem** em que aparecem neste documento. Toda chamada devolve uma `Promise`.

```ts
import { Login, SearchBM25Filtered } from '../wailsjs/go/main/App'

const auth = await Login('maria@exemplo.com', 'SenhaForte123')
if (!auth.success) {
  console.log(auth.message) // "Email ou senha inválidos."
}

const results = await SearchBM25Filtered('indexação paralela', '', '', '', '', '', true, 20)
```

Tipos numéricos de identificadores (`uint64` no Go) chegam como `number`. Datas de entrada seguem o formato `AAAA-MM-DD`; datas de saída (`created_at`) são texto no padrão RFC 3339, por exemplo `"2026-10-22T19:15:03-03:00"`.

## 3. Autenticação e autorização

**Sessão.** O backend mantém **uma sessão por instância do aplicativo**, em memória. `Login` (ou `Register` e `CreateInitialAdmin`) abre a sessão e `Logout` a encerra. O perfil do usuário é lido do banco no momento do login. As senhas são armazenadas com hash **bcrypt** e nunca são devolvidas pela API.

**Níveis de acesso.** Cada método exige um dos níveis abaixo, verificado no backend antes de qualquer ação:

| Nível | Quem acessa | Verificação |
| --- | --- | --- |
| Público | Qualquer um, sem login | — |
| Autenticado | Pesquisador, Curador e Administrador | Há usuário na sessão |
| Curador | Curador e Administrador | Perfil `CURADOR` ou `ADMIN` |
| Administrador | Somente o Administrador | Perfil `ADMIN` |

Quando o acesso é recusado, os métodos que devolvem um envelope respondem `success: false` com uma mensagem; os que devolvem listas ou indicadores devolvem uma lista vazia ou um objeto zerado.

**Perfis.** `PESQUISADOR` (padrão de todo cadastro público), `CURADOR` (atribuído pelo Administrador) e `ADMIN` (criado na primeira execução).

## 4. Convenções de resposta

- **Envelope:** a maioria dos métodos que alteram dados devolve um objeto com `success` (booleano) e `message` (texto em português, pronto para exibir ao usuário). Alguns acrescentam campos próprios (`id`, `report`, `user`).
- **Consultas:** métodos de listagem devolvem diretamente um *array* (vazio quando não há dados, quando o usuário não tem acesso ou quando ocorre um erro).
- **Erros:** não há exceções nem códigos numéricos. O motivo da falha está em `message`.
- **Diálogos nativos:** `SelectPDF`, `ImportPDFBatch`, `ExportArticlePDF` e `ExportLogsCSV` abrem o seletor de arquivos do sistema operacional; o caminho do arquivo é escolhido pelo usuário e nunca trafega pela interface.

## 5. Tipos de dados

Campos marcados com `?` são opcionais (aparecem apenas quando preenchidos).

**`AuthResponse`**

| Campo | Tipo | Descrição |
| --- | --- | --- |
| `success` | boolean | Resultado da operação |
| `message` | string | Mensagem para o usuário |
| `user?` | `User` | Usuário autenticado |

**`User`**

| Campo | Tipo | Descrição |
| --- | --- | --- |
| `id` | number | Identificador |
| `name` | string | Nome |
| `email` | string | E-mail (único) |
| `role` | string | `PESQUISADOR`, `CURADOR` ou `ADMIN` |
| `created_at` | string | Data de criação |

**`ActionResponse`**: `success` (boolean), `message` (string), `id?` (number, identificador do registro afetado).

**`HealthResponse`**: `status` (`"ok"` ou `"error"`) e `database` (`"connected"` ou `"disconnected"`).

**`FileSelectionResponse`**: `success`, `message`, `name?` (nome do arquivo escolhido) e `size?` (tamanho em bytes).

**`BatchImportResponse`**: `success`, `message`, `selected`, `imported`, `duplicates`, `failed` (contadores) e `errors` (lista de textos, um por arquivo com falha).

**`ExportResponse`**: `success`, `message`, `path?` (caminho do arquivo gravado) e `rows` (quantidade de registros exportados, nas exportações de histórico).

**`ArticleSource`**: `id`, `name`, `description`.

**`ArticleSummary`**

| Campo | Tipo | Descrição |
| --- | --- | --- |
| `id` | number | Identificador |
| `title` | string | Título |
| `description` | string | Descrição |
| `filename` | string | Nome original do arquivo |
| `source_name` | string | Fonte (`local`, `arxiv`, `pubmed`, `openalex`) |
| `authors` | string | Autores, separados por vírgula |
| `language_code` | string | Idioma (ex.: `pt-BR`) |
| `publication_date` | string | Data de publicação (`AAAA-MM-DD`) |
| `index_status` | string | `PENDING`, `PROCESSING`, `INDEXED` ou `ERROR` |
| `created_at` | string | Data de importação |

**`ArticleDetail`**: todos os campos de `ArticleSummary`, mais `file_path`, `content_type`, `size`, `user_id`, `source_id`, `abstract_text`, `external_id`, `doi`, `checksum_sha256` e `updated_at`.

**`ArticleDetailResponse`**: `success`, `message` e `article?` (`ArticleDetail`).

**`CollectionStats`**: `total_articles`, `pending_articles`, `indexed_articles`, `total_authors`, `total_sources`.

**`IndexingOverview`**: `pending`, `processing`, `indexed`, `errors`, `terms`, `occurrences`, `documents`.

**`IndexingReport`**: `workers`, `total`, `indexed`, `failed`, `duration_ms`, `reindexed_all` (boolean) e `errors` (lista de textos).

**`IndexingResponse`**: `success`, `message` e `report` (`IndexingReport`).

**`SearchResult`**: `id`, `title`, `description`, `filename`, `source_name`, `authors`, `language_code`, `publication_date`, `index_status`, `score` (pontuação BM25, número) e `matched_terms` (termos que casaram com a consulta, em texto).

**`BenchmarkResult`**: `workers`, `processing_ms`, `database_ms`, `duration_ms`, `indexed`, `failed`, `processing_speedup`, `processing_efficiency`, `total_speedup`, `database_share` e `errors`.

**`BenchmarkResponse`**: `success`, `message` e `report` com `results` (lista de `BenchmarkResult`) e `note` (texto explicativo).

**`OperationLog`**: `id`, `user_id`, `user_name`, `operation`, `entity`, `entity_id`, `details`, `result_count` e `created_at`.

**`LogSummaryResponse`**: `success`, `message` e `summary` com `total`, `today`, `by_operation` (lista de `{ operation, total }`) e `top_searches` (lista de `{ term, total }`).

---

## 6. Sistema e sessão

### `Health()`

Informa se o backend está conectado ao banco de dados.

- **Acesso:** público
- **Parâmetros:** nenhum
- **Retorno:** `HealthResponse`
- **Respostas:** `{ "status": "ok", "database": "connected" }` ou `{ "status": "error", "database": "disconnected" }`

```ts
const health = await Health()
```

### `ReconnectDatabase()`

Tenta conectar ao MySQL quando a conexão não foi estabelecida na abertura. Releitura do `.env`, abertura da conexão e criação dos serviços, sem reiniciar o aplicativo. Se já estiver conectado, apenas devolve o estado atual.

- **Acesso:** público
- **Parâmetros:** nenhum
- **Retorno:** `HealthResponse`

### `NeedsAdminSetup()`

Informa se o sistema ainda precisa de um Administrador.

- **Acesso:** público
- **Parâmetros:** nenhum
- **Retorno:** `boolean`: `true` quando o banco está conectado e não existe nenhum usuário `ADMIN`. Devolve `false` se o banco estiver indisponível.

### `CreateInitialAdmin(name, email, password)`

Cria o primeiro Administrador. Só funciona enquanto não existir nenhum.

- **Acesso:** público (recusado se já existir um Administrador)
- **Parâmetros:** `name` (string), `email` (string), `password` (string, mínimo de 8 caracteres)
- **Retorno:** `AuthResponse`. Em caso de sucesso, a sessão do Administrador é aberta.
- **Respostas de erro:**
  - `"Banco de dados não conectado."`
  - `"Preencha nome e email e use uma senha com pelo menos 8 caracteres."`
  - `"Informe um e-mail válido (ex.: nome@dominio.com)."`
  - `"Este email já está cadastrado. Use outro email."`
  - `"Já existe um administrador cadastrado."`
- **Sucesso:** `"Administrador criado com sucesso."`

```ts
const r = await CreateInitialAdmin('Admin', 'admin@exemplo.com', 'SenhaForte123')
```

### `Register(name, email, password)`

Cadastro público de usuário. O perfil é sempre `PESQUISADOR`.

- **Acesso:** público
- **Parâmetros:** `name`, `email`, `password` (mínimo de 8 caracteres)
- **Retorno:** `AuthResponse`. Em caso de sucesso, a sessão do novo usuário é aberta.
- **Respostas de erro:** `"Banco de dados não conectado."`, `"Preencha nome e email e use uma senha com pelo menos 8 caracteres."`, `"Informe um e-mail válido (ex.: nome@dominio.com)."`, `"Este email já está cadastrado."`
- **Sucesso:** `"Cadastro realizado com sucesso."`
- **Histórico:** registra `CADASTRO_USUARIO`

### `Login(email, password)`

Autentica o usuário e abre a sessão.

- **Acesso:** público
- **Parâmetros:** `email`, `password`
- **Retorno:** `AuthResponse`
- **Respostas:** `"Login realizado com sucesso."` (com `user`), `"Email ou senha inválidos."` (a mesma mensagem para e-mail inexistente ou senha incorreta) e `"Banco de dados não conectado."`
- **Histórico:** registra `LOGIN`

Exemplo de resposta de sucesso:

```json
{
  "success": true,
  "message": "Login realizado com sucesso.",
  "user": {
    "id": 3,
    "name": "Maria",
    "email": "maria@exemplo.com",
    "role": "CURADOR",
    "created_at": "2026-10-22T20:44:10-03:00"
  }
}
```

### `Logout()`

Encerra a sessão e limpa o PDF selecionado.

- **Acesso:** público (não faz nada sem sessão)
- **Parâmetros:** nenhum
- **Retorno:** nenhum
- **Histórico:** registra `LOGOUT`

### `GetCurrentUser()`

Devolve o usuário da sessão atual.

- **Acesso:** público
- **Parâmetros:** nenhum
- **Retorno:** `User`, ou `null` se não houver sessão

---

## 7. Artigos

### `ListSources()`

Lista as fontes cadastradas.

- **Acesso:** autenticado
- **Retorno:** `ArticleSource[]`

### `ListArticles()`

Lista todos os artigos do acervo, do mais recente para o mais antigo. Não há paginação nem limite.

- **Acesso:** autenticado
- **Retorno:** `ArticleSummary[]`

### `SearchArticles(query)`

Busca textual simples por `LIKE` em título, descrição, fonte e autores. A busca por relevância está em [`SearchBM25Filtered`](#searchbm25filteredquery-source-author-language-datefrom-dateto-fuzzy-limit).

- **Acesso:** autenticado
- **Parâmetros:** `query` (string)
- **Retorno:** `ArticleSummary[]`

### `GetArticle(id)`

Devolve os metadados completos de um artigo.

- **Acesso:** autenticado
- **Parâmetros:** `id` (number)
- **Retorno:** `ArticleDetailResponse`
- **Respostas:** `"Artigo carregado."`, `"Artigo não encontrado."`, `"Usuário não autenticado."`

### `GetCollectionStats()`

Indicadores do acervo.

- **Acesso:** autenticado
- **Retorno:** `CollectionStats` (zerado se não houver acesso)

```json
{ "total_articles": 12, "pending_articles": 3, "indexed_articles": 8, "total_authors": 17, "total_sources": 4 }
```

### `SelectPDF()`

Abre o seletor de arquivos e guarda, na sessão, o PDF escolhido para a próxima importação individual.

- **Acesso:** Curador
- **Retorno:** `FileSelectionResponse`
- **Respostas:** `"PDF selecionado."` (com `name` e `size`), `"Nenhum arquivo selecionado."`, `"Não foi possível ler o arquivo selecionado."`, `"Somente Curador pode importar artigos."`

### `ImportArticle(title, description, sourceID, authorsCSV, languageCode, publicationDate, doi, externalID)`

Importa o PDF selecionado por `SelectPDF` e cadastra seus metadados. O PDF é copiado para `storage/articles/<hash SHA-256>.pdf` e o artigo nasce com a situação `PENDING`, vinculado ao usuário que importou.

- **Acesso:** Curador
- **Pré-requisito:** chamar `SelectPDF` antes

| Parâmetro | Tipo | Obrigatório | Descrição |
| --- | --- | --- | --- |
| `title` | string | Sim | Título |
| `description` | string | Não | Descrição |
| `sourceID` | number | Sim | Identificador da fonte (`ListSources`) |
| `authorsCSV` | string | Não | Autores separados por vírgula; vazios são ignorados e repetidos são unificados |
| `languageCode` | string | Não | Idioma, ex.: `pt-BR` |
| `publicationDate` | string | Não | `AAAA-MM-DD`; deve ser uma data real e não futura |
| `doi` | string | Não | DOI |
| `externalID` | string | Não | Identificador externo |

- **Retorno:** `ActionResponse` (com o `id` do artigo)
- **Respostas de erro:**
  - `"Selecione um arquivo PDF antes de importar."`
  - `"Título, fonte e PDF válido são obrigatórios."` (também quando o arquivo não é um PDF)
  - `"Data de publicação inválida; informe uma data real que não seja futura."`
  - `"Este PDF já foi importado anteriormente."` (o hash SHA-256 do arquivo é único)
  - `"Somente Curador pode importar artigos."`
- **Sucesso:** `"Artigo importado e persistido com sucesso."`
- **Histórico:** registra `IMPORTACAO_ARTIGO` (detalhes: título)

```ts
await SelectPDF()
const r = await ImportArticle('Perfectoid pure and signature pairs', '', 2, 'Hanlin Ca', 'en', '2026-10-18', '', '')
// r = { success: true, message: "Artigo importado e persistido com sucesso.", id: 7 }
```

### `ImportPDFBatch(sourceID, languageCode)`

Abre o seletor de múltiplos arquivos e importa todos os PDFs escolhidos. O título de cada artigo é o nome do arquivo. Uma falha em um arquivo não interrompe os demais.

- **Acesso:** Curador
- **Parâmetros:** `sourceID` (number, obrigatório); `languageCode` (string; se vazio, usa `pt-BR`)
- **Retorno:** `BatchImportResponse`. `success` é `true` quando ao menos um arquivo foi importado e nenhum falhou.
- **Respostas:**
  - `"Importação em lote concluída: N importados, N duplicados e N falhas."`
  - `"Os PDFs selecionados já estavam no acervo."`
  - `"Nenhum PDF selecionado."`
  - `"Fonte inválida para importação em lote."`
  - `"Somente Curador pode importar artigos."`
- **Histórico:** registra `IMPORTACAO_LOTE` (detalhes: `"N importados · N duplicados · N falhas"`)

### `UpdateArticle(id, title, description, sourceID, authorsCSV, languageCode, publicationDate, doi, externalID)`

Atualiza os metadados e os autores de um artigo. O PDF não é substituído. Título e fonte continuam obrigatórios; a data segue a mesma regra da importação.

- **Acesso:** Curador
- **Parâmetros:** `id` (number) seguido dos mesmos campos de `ImportArticle`
- **Retorno:** `ActionResponse`
- **Respostas:** `"Metadados atualizados com sucesso."`, `"Falha ao atualizar artigo: <motivo>"`, `"Somente Curador pode editar artigos."`
- **Histórico:** registra `EDICAO_ARTIGO`

### `DeleteArticle(id)`

Exclui o artigo, o índice do documento, os vínculos com autores e o PDF armazenado em disco. Se a remoção do arquivo falhar, a falha é informada.

- **Acesso:** Curador
- **Parâmetros:** `id` (number)
- **Retorno:** `ActionResponse`
- **Respostas:** `"Artigo removido com sucesso."`, `"Falha ao excluir artigo: <motivo>"`, `"Somente Curador pode excluir artigos."`
- **Histórico:** registra `EXCLUSAO_ARTIGO`

### `ExportArticlePDF(id)`

Exporta o PDF de um artigo. Abre o diálogo "Salvar como" já preenchido com o **título do artigo** como nome do arquivo, e copia o PDF armazenado para o local escolhido. O arquivo original continua em `storage/articles/`.

O nome sugerido mantém o título; apenas os caracteres que o Windows não aceita são ajustados (`:` vira ` -`; `/`, `\` e `|` viram `-`; `" * ? < >` são removidos), e títulos acima de 150 caracteres são cortados.

- **Acesso:** autenticado
- **Parâmetros:** `id` (number)
- **Retorno:** `ExportResponse`
- **Respostas:** `"PDF exportado como \"<nome>.pdf\"."`, `"Exportação cancelada."`, `"Artigo não encontrado."`, `"O arquivo PDF deste artigo não foi encontrado no armazenamento."`, `"Não foi possível gravar o arquivo."`, `"Usuário não autenticado."`
- **Histórico:** registra `EXPORTACAO_PDF` (detalhes: título)

---

## 8. Indexação e busca

### `StartIndexing(workers, reindexAll)`

Executa a indexação paralela. Os workers (goroutines) leem o PDF, extraem o texto, normalizam e tokenizam; um único writer grava o índice invertido no MySQL. A operação é síncrona: a chamada só retorna quando a indexação termina.

- **Acesso:** Curador
- **Parâmetros:**
  - `workers` (number): quantidade de workers. É ajustada ao intervalo de 1 até o dobro dos núcleos da máquina (mínimo de 4 e teto de 32).
  - `reindexAll` (boolean): se `true`, reindexa também os artigos já indexados; se `false`, indexa apenas `PENDING` e `ERROR`.
- **Retorno:** `IndexingResponse`
- **Respostas:** `"Indexação concluída."`, `"Indexação concluída com alguns erros."`, `"Nenhum artigo pendente para indexar."`, `"Falha durante a indexação: <motivo>"`, `"Somente Curador pode iniciar a indexação."`
- **Histórico:** registra `INDEXACAO` (detalhes: `"N workers · N indexados · N falhas · N ms"`)

```json
{
  "success": true,
  "message": "Indexação concluída.",
  "report": { "workers": 4, "total": 3, "indexed": 3, "failed": 0, "duration_ms": 1840, "reindexed_all": false, "errors": [] }
}
```

### `GetIndexingOverview()`

Situação do índice.

- **Acesso:** autenticado
- **Retorno:** `IndexingOverview` (zerado se não houver acesso)

### `SearchBM25(query, limit)`

Busca ranqueada por BM25, sem filtros e com busca aproximada desligada. Atalho para `SearchBM25Filtered` com apenas o limite.

- **Acesso:** autenticado
- **Parâmetros:** `query` (string), `limit` (number)
- **Retorno:** `SearchResult[]`
- **Histórico:** registra `BUSCA` quando a consulta não é vazia

### `SearchBM25Filtered(query, source, author, language, dateFrom, dateTo, fuzzy, limit)`

Busca ranqueada por relevância sobre o índice invertido, com filtros e busca aproximada.

- **Acesso:** autenticado

| Parâmetro | Tipo | Descrição |
| --- | --- | --- |
| `query` | string | Consulta. É normalizada (minúsculas, sem acentos, sem palavras comuns e sem repetições) |
| `source` | string | Filtra pela fonte (sem diferenciar maiúsculas). Vazio = qualquer |
| `author` | string | Mantém artigos cujos autores contêm o texto |
| `language` | string | Código do idioma, ex.: `pt-BR` |
| `dateFrom` | string | Data de publicação inicial (`AAAA-MM-DD`) |
| `dateTo` | string | Data de publicação final (`AAAA-MM-DD`) |
| `fuzzy` | boolean | Liga a busca aproximada |
| `limit` | number | Máximo de resultados. Padrão 50; valores menores que 1 ou maiores que 100 usam 50 |

- **Retorno:** `SearchResult[]`, ordenados por `score` decrescente
- **Regras:**
  - Pontuação BM25 com `k1 = 1,2` e `b = 0,75`.
  - Só artigos **indexados** entram no ranking.
  - Os filtros são cumulativos; artigos sem data de publicação ficam de fora quando há filtro de data.
  - Busca aproximada: aplicada apenas a termos ausentes do índice, com distância de Levenshtein de até 2 (1 para termos de até 4 caracteres) e peso 0,72 para o termo corrigido.
- **Histórico:** registra `BUSCA` (detalhes: a consulta; `result_count`: quantidade de resultados) quando a consulta não é vazia

```ts
const results = await SearchBM25Filtered('indexação paralela', 'arxiv', '', 'pt-BR', '2024-01-01', '2026-12-31', true, 20)
```

```json
[
  {
    "id": 7, "title": "Indexação paralela de documentos", "description": "",
    "filename": "2610.22059v1.pdf", "source_name": "arxiv", "authors": "Hanlin Ca",
    "language_code": "pt-BR", "publication_date": "2026-10-18", "index_status": "INDEXED",
    "score": 4.2713, "matched_terms": "indexacao, paralela"
  }
]
```

### `RunIndexingBenchmark()`

Compara o desempenho da indexação com 1, 2, 4 e 8 workers, separando o tempo de processamento (etapa paralela) do tempo de gravação no MySQL (etapa serial).

- **Acesso:** Curador
- **Parâmetros:** nenhum
- **Retorno:** `BenchmarkResponse`
- **Respostas:** `"Benchmark concluído."`, `"Falha no benchmark: <motivo>"`, `"Somente Curador pode executar benchmark."`
- **Histórico:** registra `BENCHMARK`

---

## 9. Histórico de operações

### `ListOperationLogs(operation, dateFrom, dateTo, limit)`

Lista os registros do histórico, do mais recente para o mais antigo.

- **Acesso:** autenticado
- **Parâmetros:**
  - `operation` (string): código exato da operação (veja a [seção 11](#11-operações-registradas-no-histórico)); vazio = todas
  - `dateFrom` (string): data inicial (`AAAA-MM-DD`), inclusiva
  - `dateTo` (string): data final (`AAAA-MM-DD`), inclusiva (considera o dia inteiro)
  - `limit` (number): máximo de registros; padrão 100, e valores menores que 1 ou maiores que 500 usam 100
- **Retorno:** `OperationLog[]`. O nome do usuário vem de uma associação com a tabela `users`; fica vazio se o usuário foi excluído.

```ts
const logs = await ListOperationLogs('BUSCA', '2026-10-01', '2026-10-31', 200)
```

### `GetLogSummary()`

Indicadores consolidados do histórico: total de registros, registros do dia, contagem por tipo de operação e termos mais pesquisados.

- **Acesso:** autenticado
- **Retorno:** `LogSummaryResponse`
- **Respostas:** `"Histórico consolidado."`, `"Falha ao consolidar o histórico."`

### `ExportLogsCSV(operation, dateFrom, dateTo, limit)`

Exporta o histórico filtrado em um arquivo CSV (separador `;` e codificação UTF-8 com BOM, para abrir corretamente no Excel). Abre o diálogo "Salvar como" com o nome sugerido `historico-concordia-AAAA-MM-DD.csv`.

- **Acesso:** Curador
- **Parâmetros:** os mesmos de `ListOperationLogs`
- **Retorno:** `ExportResponse`
- **Colunas do arquivo:** `id`, `data_hora`, `usuario`, `operacao`, `entidade`, `entidade_id`, `detalhes`, `resultados`
- **Respostas:** `"N registros exportados."`, `"Nenhum registro para exportar com os filtros atuais."`, `"Exportação cancelada."`, `"Somente Curador pode exportar o histórico."`
- **Histórico:** registra `EXPORTACAO_CSV`

---

## 10. Gestão de usuários

### `ListUsers()`

Lista todos os usuários, sem o hash de senha.

- **Acesso:** Administrador
- **Retorno:** `User[]` (lista vazia para quem não é Administrador)

### `UpdateUserRole(id, role)`

Altera a função de um usuário. Só aceita `PESQUISADOR` ou `CURADOR`; contas `ADMIN` não podem ser alteradas.

- **Acesso:** Administrador
- **Parâmetros:** `id` (number), `role` (string: `PESQUISADOR` ou `CURADOR`)
- **Retorno:** `ActionResponse`
- **Respostas:** `"Função de <nome> alterada para <função>."`, `"Função inválida. Use Pesquisador ou Curador."`, `"Usuário não encontrado."`, `"Administradores não podem ser alterados."`, `"Somente o administrador pode alterar funções."`

### `DeleteUser(id)`

Exclui um usuário. Os artigos que ele importou e os registros do histórico permanecem, sem responsável. Contas `ADMIN` não podem ser excluídas.

- **Acesso:** Administrador
- **Parâmetros:** `id` (number)
- **Retorno:** `ActionResponse`
- **Respostas:** `"Usuário <nome> excluído."`, `"Usuário não encontrado."`, `"Administradores não podem ser excluídos."`, `"Somente o administrador pode excluir usuários."`

---

## 11. Operações registradas no histórico

O registro é feito pelo backend quando a operação se conclui, com o usuário da sessão. Uma falha ao gravar o registro nunca interrompe a operação principal.

| Código | Gerado por |
| --- | --- |
| `LOGIN` | `Login` |
| `LOGOUT` | `Logout` |
| `CADASTRO_USUARIO` | `Register` |
| `IMPORTACAO_ARTIGO` | `ImportArticle` |
| `IMPORTACAO_LOTE` | `ImportPDFBatch` |
| `EDICAO_ARTIGO` | `UpdateArticle` |
| `EXCLUSAO_ARTIGO` | `DeleteArticle` |
| `INDEXACAO` | `StartIndexing` |
| `BENCHMARK` | `RunIndexingBenchmark` |
| `BUSCA` | `SearchBM25` e `SearchBM25Filtered` (com consulta não vazia) |
| `EXPORTACAO_CSV` | `ExportLogsCSV` |
| `EXPORTACAO_PDF` | `ExportArticlePDF` |

## 12. Resumo dos métodos

| # | Método | Acesso | Módulo |
| --- | --- | --- | --- |
| 1 | `Health` | Público | Sistema e sessão |
| 2 | `ReconnectDatabase` | Público | Sistema e sessão |
| 3 | `NeedsAdminSetup` | Público | Sistema e sessão |
| 4 | `CreateInitialAdmin` | Público (só sem Administrador) | Sistema e sessão |
| 5 | `Register` | Público | Sistema e sessão |
| 6 | `Login` | Público | Sistema e sessão |
| 7 | `Logout` | Público | Sistema e sessão |
| 8 | `GetCurrentUser` | Público | Sistema e sessão |
| 9 | `ListSources` | Autenticado | Artigos |
| 10 | `ListArticles` | Autenticado | Artigos |
| 11 | `SearchArticles` | Autenticado | Artigos |
| 12 | `GetArticle` | Autenticado | Artigos |
| 13 | `GetCollectionStats` | Autenticado | Artigos |
| 14 | `ExportArticlePDF` | Autenticado | Artigos |
| 15 | `SelectPDF` | Curador | Artigos |
| 16 | `ImportArticle` | Curador | Artigos |
| 17 | `ImportPDFBatch` | Curador | Artigos |
| 18 | `UpdateArticle` | Curador | Artigos |
| 19 | `DeleteArticle` | Curador | Artigos |
| 20 | `StartIndexing` | Curador | Indexação e busca |
| 21 | `GetIndexingOverview` | Autenticado | Indexação e busca |
| 22 | `SearchBM25` | Autenticado | Indexação e busca |
| 23 | `SearchBM25Filtered` | Autenticado | Indexação e busca |
| 24 | `RunIndexingBenchmark` | Curador | Indexação e busca |
| 25 | `ListOperationLogs` | Autenticado | Histórico |
| 26 | `GetLogSummary` | Autenticado | Histórico |
| 27 | `ExportLogsCSV` | Curador | Histórico |
| 28 | `ListUsers` | Administrador | Gestão de usuários |
| 29 | `UpdateUserRole` | Administrador | Gestão de usuários |
| 30 | `DeleteUser` | Administrador | Gestão de usuários |
