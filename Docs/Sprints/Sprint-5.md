# Sprint 5 — Concordia

## 1. Identificação da equipe

- **Grupo 17**

- **Integrantes:**
    - Maria Karolina Rodrigues de Andrade
    - Pedro Felipe Oliveira Tavares
    - Pedro Oliveira Barros Batista de Lima
    - Tomaz Arlindo Silva Ribeiro
    - Vladison Lucas Costa Dos Santos
- **Turma:** CC-8-MB
- **Nome do projeto:** Concordia

## 2. Segundo módulo completo

O segundo módulo do Concordia corresponde ao fluxo do perfil Curador, responsável por **formar e manter o acervo** que o Pesquisador consulta. Ele reúne tudo o que o Pesquisador faz e acrescenta as funções de curadoria: importação de artigos, gerenciamento de metadados, indexação paralela e acompanhamento de métricas.

### Funcionalidades

| Funcionalidade | Descrição |
| --- | --- |
| Importação individual | Seleção de um PDF e cadastro dos metadados (título, descrição, fonte, autores, idioma, data de publicação, DOI e identificador externo) |
| Importação em lote | Seleção de vários PDFs de uma só vez para montar rapidamente um acervo; o nome do arquivo vira o título, a fonte é `local` e o idioma padrão é `pt-BR`. Ao final, é exibido um relatório com selecionados, importados, duplicados e falhas |
| Gerenciamento de artigos | Listagem do acervo, edição de metadados e exclusão (com confirmação) |
| Indexação paralela | Processamento dos PDFs por um *worker pool* de goroutines, com número de workers definido pelo Curador, e opção de reindexar documentos já indexados |
| Visão geral do índice | Quantidade de artigos pendentes e indexados, total de termos e de ocorrências |
| Métricas e benchmark | Indicadores do acervo e comparação de desempenho com 1, 2, 4 e 8 workers |

###  Fluxo

```
Login como Curador
  → Importar PDF (individual ou em lote)
  → Artigo cadastrado com status PENDING
  → Indexação paralela (escolha do número de workers)
  → Artigo com status INDEXED
  → Consulta pelo Pesquisador (Módulo 1)
  → Métricas e benchmark
```

###  Telas do perfil Curador

| Tela | Acesso | Conteúdo |
| --- | --- | --- |
| Artigos | Menu "Artigos" | Lista do acervo com opções de detalhes e exclusão |
| Importar | Menu "Importar" | Formulário de importação individual e importação em lote |
| Editar artigo | Botão "Editar metadados" nos detalhes | Formulário de atualização de metadados; o PDF não é substituído |
| Indexação | Menu "Indexação" | Contadores do índice, número de workers, opção de reindexar tudo, relatório da execução e erros por documento |
| Métricas | Menu "Métricas" | Indicadores do acervo e tabela do benchmark de paralelismo |

O Curador também possui as telas de Visão geral e Pesquisa do Módulo 1.

## 3. Integração com o banco de dados

Todas as operações do Curador são persistidas em MySQL e no sistema de arquivos local, seguindo a mesma estrutura em camadas do restante do sistema (interface → serviços → repositórios → MySQL).

| Operação | Tabelas e armazenamento afetados |
| --- | --- |
| Importar artigo | Insere em `articles`; insere ou reaproveita registros em `authors`; vincula em `article_authors`; copia o PDF para `storage/articles/` |
| Editar artigo | Atualiza `articles`; remove e recria os vínculos em `article_authors` |
| Excluir artigo | Remove de `articles` (os dados de índice e vínculos de autores são removidos em cascata) e apaga o PDF do disco |
| Indexar | Grava em `terms`, `term_occurrences` e `article_index_stats`; atualiza `index_status` e `indexed_at` em `articles` |
| Consultar métricas | Contagens em `articles`, `article_authors`, `article_sources`, `terms` e `term_occurrences` |

A persistência dos arquivos e dos dados do Curador é coordenada entre o sistema de arquivos e o MySQL. Na importação, o PDF é copiado para `storage/articles/` com o nome `<hash SHA-256>.pdf`, e o banco guarda o caminho relativo e o hash; se a transação do banco falhar, o arquivo recém-copiado é removido, evitando arquivos órfãos, e as fontes padrão (`local`, `arxiv`, `pubmed` e `openalex`) são garantidas na inicialização do sistema. A criação e a atualização de artigos, incluindo a gravação dos autores, ocorrem em uma única transação, com *rollback* em caso de erro, e a gravação do índice de cada documento também é transacional: os dados antigos do documento são removidos e os novos são inseridos e confirmados juntos, de modo que o status só passa para `INDEXED` ao final da mesma transação. Na indexação, a etapa paralela é separada da escrita: os *workers* (goroutines) leem o PDF, extraem o texto, normalizam, tokenizam e montam o índice local do documento sem escrever no banco, enquanto um **único *writer*** recebe os resultados e os grava no MySQL, e ao final a frequência de documentos de cada termo (`document_frequency`) é recalculada e os termos que ficaram sem nenhum documento são removidos. Os termos são gravados com `INSERT ... ON DUPLICATE KEY UPDATE`, o que permite que vários documentos compartilhem o mesmo termo sem duplicidade no vocabulário, e o ciclo de vida do status de cada artigo é `PENDING` → `PROCESSING` → `INDEXED` ou `ERROR`.

## 4. Atualização das regras de negócio

Com a criação do módulo do Curador, o sistema passou a ter regras de negócio voltadas à formação e à manutenção do acervo, que antes não existiam, já que o Pesquisador apenas consulta dados. Elas definem quem pode alterar o acervo, quais dados são aceitos na importação e como os artigos passam pelo processo de indexação.

| Regra | O que faz |
| --- | --- |
| Acesso restrito ao Curador | Importação, edição, exclusão, indexação e benchmark só podem ser executados por usuários com perfil Curador, validados no backend |
| Campos obrigatórios | A importação exige título, fonte e arquivo PDF; a edição mantém título e fonte obrigatórios |
| Validação do arquivo | Aceita apenas arquivos `.pdf` cujo conteúdo comece com a assinatura de PDF |
| Bloqueio de duplicidade | O hash SHA-256 do arquivo é único, impedindo importar o mesmo PDF duas vezes, mesmo com nome diferente |
| Status inicial | Todo artigo importado nasce como `PENDING` e fica vinculado ao Curador que o importou |
| Cadastro de autores | Nomes separados por vírgula; espaços e vazios são ignorados, repetidos são unificados e autores já existentes são reaproveitados |
| Importação em lote | O título vem do nome do arquivo, a fonte é `local` e o idioma padrão é `pt-BR`; falhas individuais não interrompem os demais arquivos |
| Edição de artigos | Permite alterar metadados e autores, mas não substitui o PDF |
| Exclusão de artigos | Exige confirmação e remove o registro, o índice do documento, os vínculos com autores e o arquivo em disco |
| Seleção para indexação | Por padrão, indexa apenas artigos `PENDING` ou `ERROR`; há opção para reindexar também os já indexados |
| Limite de workers | O número de workers fica entre 1 e um máximo definido pela máquina (teto de 32) |
| Tratamento do texto | Termos são normalizados (minúsculas e sem acentos); palavras comuns e termos com menos de 2 ou mais de 120 caracteres são descartados |
| PDFs sem texto | Documentos sem texto extraível, como escaneados, são marcados com erro, pois não há OCR |
| Ciclo de status | `PENDING` → `PROCESSING` → `INDEXED` ou `ERROR` |

## 5. Testes das funcionalidades

### Testes automatizados

O projeto possui testes unitários em Go (`go test ./...`), na camada de serviços:

| Arquivo | Cobertura |
| --- | --- |
| `tokenizer_test.go` | Remoção de *stop words* e duplicados na consulta; frequência e posições dos termos no índice; unificação de variações de acento e caixa; unificação de termos equivalentes pela collation do MySQL |
| `search_helpers_test.go` | Distância de Levenshtein da busca aproximada; aplicação dos filtros de busca |
| `benchmark_service_test.go` | Limites do número de workers |

### Testes funcionais do fluxo do Curador

Os cenários abaixo seguem o plano de testes documentado no projeto (`docs/TESTES.md`):

| Cenário | Resultado esperado |
| --- | --- |
| Importar um PDF válido | Artigo cadastrado; arquivo presente em `storage/articles/`; hash gravado no banco |
| Importar o mesmo PDF novamente | Operação bloqueada com a mensagem "Este PDF já foi importado anteriormente." |
| Importar vários PDFs em lote | Relatório com selecionados, importados, duplicados e falhas |
| Executar a indexação com 4 workers | Artigos com status `INDEXED`; tabelas `terms`, `term_occurrences` e `article_index_stats` preenchidas |
| Pesquisar um termo do conteúdo após indexar | Artigo retornado com pontuação BM25 e termos encontrados |
| Executar o benchmark | Resultados para 1, 2, 4 e 8 workers |
| Acessar funções de curadoria como Pesquisador | Menu de curadoria oculto e operações recusadas pelo backend |

## 6. Correção dos bugs encontrados

Os problemas abaixo foram identificados durante o desenvolvimento e estão registrados nos comentários do código e na documentação do projeto.

| Problema | Causa | Correção |
| --- | --- | --- |
| *Deadlocks* no MySQL (erro 1213) durante a indexação | Várias transações concorrentes tentavam inserir ou atualizar os mesmos termos na tabela `terms`, já que diferentes artigos compartilham palavras | Os workers passaram a executar apenas a etapa de processamento; a gravação no banco é feita por um único *writer* |
| Erro de chave primária duplicada em `term_occurrences` | Termos como "Martinez" e "Martínez" eram chaves diferentes no Go, mas apontavam para o mesmo registro no MySQL por causa da collation `utf8mb4_unicode_ci` | Normalização dos termos (minúsculas, decomposição Unicode e remoção de acentos), com testes automatizados para esse caso |

