# Plano de testes — Concordia

## Testes automáticos

Execute:

```powershell
go test ./...
```

Os testes incluídos cobrem:

- remoção de stop words;
- eliminação de termos duplicados da consulta;
- frequência e posições no índice;
- distância de Levenshtein usada na busca aproximada;
- aplicação de filtros de busca.

## Testes funcionais recomendados

### 1. Autenticação

- cadastrar Pesquisador;
- sair e entrar novamente;
- alterar um usuário de teste para `CURADOR` no banco e validar menu;
- confirmar que Pesquisador não possui operações de curadoria.

### 2. Importação

- importar PDF válido;
- confirmar arquivo em `storage/articles`;
- confirmar `checksum_sha256` no banco;
- tentar importar o mesmo PDF novamente e validar bloqueio.

### 3. Indexação

- importar vários PDFs;
- executar com 4 workers;
- validar status `INDEXED`;
- conferir `terms`, `term_occurrences` e `article_index_stats`.

### 4. Pesquisa

- pesquisar um termo presente dentro do PDF;
- validar score BM25;
- validar termos encontrados;
- aplicar filtro de fonte;
- aplicar autor/idioma/data quando houver metadados;
- testar pequeno erro de digitação com busca aproximada marcada.

### 5. Benchmark

Importe vários PDFs antes de medir.

Execute o benchmark e registre:

```text
Workers | Tempo | Speedup | Eficiência
1       | ...   | 1.00x   | 100%
2       | ...   | ...     | ...
4       | ...   | ...     | ...
8       | ...   | ...     | ...
```

Não conclua que “mais workers sempre é melhor”. O resultado depende do número de documentos, tamanho dos PDFs, extração e overhead de concorrência.

## Sprint 09 — execução técnica e resultados

Comando executado na pasta App: `go test ./...`.

Saída inicial, antes dos novos testes:

```text
?    desktop                    [no test files]
?    desktop/internal/config    [no test files]
?    desktop/internal/models    [no test files]
?    desktop/internal/repositories [no test files]
ok   desktop/internal/services  (cached)
```

Saída final após as alterações:

```text
ok   desktop                    3.767s
?    desktop/internal/config    [no test files]
?    desktop/internal/models    [no test files]
?    desktop/internal/repositories [no test files]
ok   desktop/internal/services  (cached)
```

## Casos automatizados

| ID | Módulo | Descrição | Esperado | Obtido | Status |
|---|---|---|---|---|---|
| A01 | Suíte Go | Execução inicial de `go test ./...` | Testes existentes passam | Passaram; pacotes app/config/models/repositories ainda sem testes | Passou |
| A02 | E-mail | Cadastro com `abc`, `a@b`, `teste@` e endereços válidos | Inválidos recusados, válidos aceitos | Os três exemplos inválidos recusados; válidos aceitos | Passou |
| A03 | Autenticação | Cadastro, login correto/incorreto, senha curta, duplicidade e logout | Fluxo correto, erros recusados e logout encerra sessão | Comportamentos esperados observados | Passou |
| A04 | Sessão/perfis | Acesso sem login e permissões de Admin, Curador e Pesquisador | Sem login bloqueado; Admin gerencia; Curador autorizado; Pesquisador sem privilégios | Comportamentos esperados observados | Passou |
| A05 | Gestão de Usuários | Admin lista, promove a Curador e exclui usuário de teste | Apenas Admin autorizado e alterações persistidas | Operações concluídas; usuário removido | Passou |
| A06 | Artigos | Importar PDF, consultar, buscar, editar e excluir | CRUD e busca funcionais; PDF excluído | Fluxo concluído e PDF ausente após exclusão | Passou |
| A07 | Data de publicação | Datas impossíveis, formato fora de ISO e data futura | Recusar antes da persistência | Rejeição confirmada na validação, importação e edição; data válida aceita | Passou |
| A08 | Exportação | Copiar PDF e tentar origem igual ao destino | Conteúdo copiado; sobrescrita de si mesmo recusada | Ambos os comportamentos observados no helper | Passou |
| A09 | Remoção do PDF | Simular falha de remoção no filesystem | Falha propagada ao chamador | Erro propagado pelo serviço | Passou |
| A10 | Reconexão | Reconectar instância sem conexão | Banco conectado após reconexão | Status `ok`, banco `connected` | Passou |
| A11 | Admin inicial | E-mail inválido e senha menor que 8 caracteres | Dados recusados com mensagem clara para e-mail | Ambos recusados; mensagem específica exibida pela API | Passou |

## Mensagens de erro da interface

- Cadastro/login: `handleAuthSubmit` coloca a resposta do backend em `message`, renderizada no formulário. Exceções de comunicação também geram mensagem visível. O campo e-mail usa `type="email"`, que pode barrar formatos no próprio navegador.
- Importação/edição: respostas de `ImportArticle` e `UpdateArticle` vão para `workspaceMessage`, exibida em `.workspace-message`. Respostas de erro da API aparecem na tela. Exceções lançadas pela ponte Wails não são capturadas nesses handlers (há `finally`, mas não `catch`), então podem não gerar mensagem visível.
- Criação inicial do Admin: o modal mostra a mensagem em `.admin-setup-error`.

## Roteiro manual da interface

Os IDs continuam a numeração dos casos automatizados. Use contas e dados de teste; registre prints da execução real.

| ID | Entrada / ação | Resultado esperado | Tela do print |
|---|---|---|---|
| M12 | Cadastro com nome, e-mail e senha vazios | Cadastro recusado; validação nativa ou mensagem visível | Formulário de cadastro |
| M13 | E-mail `abc`, `a@b` ou `teste@`, senha `SenhaTeste09` | Navegador pode bloquear; se enviado, backend mostra e-mail inválido e não cria usuário | Cadastro com mensagem |
| M14 | E-mail novo válido e senha `abc123` | Recusa com indicação de senha mínima de 8 caracteres | Cadastro |
| M15 | Repetir e-mail de conta existente | Recusa por duplicidade; conta original continua disponível | Cadastro com mensagem |
| M16 | E-mail existente e senha `Errada123` | Login recusado com “Email ou senha inválidos” | Login |
| M17 | Curador envia importação sem título, fonte ou PDF | Importação recusada e mensagem visível; nenhum artigo criado | Formulário de importação |
| M18 | Curador tenta importar com data `31/02/2026` e depois com data futura, PDF válido | Data impossível pode ser barrada pelo campo nativo; data futura recusada pelo backend; nenhum artigo persistido | Importação com mensagem |
| M19 | Editar artigo com data impossível ou futura e salvar | Edição recusada com mensagem visível; metadados permanecem | Formulário de edição |
| M20 | Induzir erro de backend durante importação/edição | Resposta de erro da API aparece; se a ponte lançar exceção, anotar ausência de mensagem local | Importação/edição |
| M21 | Entrar como Pesquisador | Pesquisa/detalhes disponíveis; menus de curadoria e gestão de usuários ausentes | Dashboard e menu |
| M22 | Entrar como Curador criado pelo Admin | Importação, artigos e indexação disponíveis; Gestão de Usuários ausente | Dashboard e menu |
| M23 | Entrar como Admin | Gestão de Usuários visível; promover usuário de teste a Curador e conferir perfil | Gestão de Usuários |
| M24 | Sair e tentar abrir área autenticada | Retorno à autenticação; ações protegidas indisponíveis | Tela de login |
| M25 | Acionar logout | Sessão encerrada e tela de autenticação exibida | Tela após logout |
| M26 | Banco desligado e reconexão; ligar banco e repetir | Primeiro estado desconectado; segundo conectado | Indicador de conexão |

## Registro de bugs

| ID | Descrição | Onde | Gravidade | Correção | Status |
|---|---|---|---|---|---|
| B01 | Cadastro aceitava `abc`, `a@b` e `teste@` | AuthService e AdminSetupService | Média | Validar formato no backend, retornar mensagem específica e testar regressão | Corrigido e testado |
| B02 | Datas inválidas, fora de ISO ou futuras não eram validadas antes do SQL | ArticleService, importação/edição | Média | Validar calendário, formato e limite de hoje; teste unitário e integração | Corrigido e testado |
| B03 | Erro de remoção do PDF era ignorado | ArticleService.DeleteArticle | Média | Propagar falha e testar cenário de erro | Corrigido; limitação residual abaixo |

## Cobertura e casos automatizados adicionados

Medição realizada com o MySQL ligado, usando os comandos abaixo no PowerShell a partir de `App`:

```powershell
go test -count=1 "-coverprofile=cover.out" ./...
go tool cover "-func=cover.out"
```

O perfil de cobertura é gerado localmente; não deve ser versionado. Comparativo de cobertura por pacote:

| Pacote | Antes | Depois |
|---|---:|---:|
| `desktop` (App e APIs) | 26,7% | 50,9% |
| `desktop/internal/config` | 0,0% | 92,3% |
| `desktop/internal/models` | sem testes próprios | sem testes próprios |
| `desktop/internal/repositories` | 0,0% | 83,0% |
| `desktop/internal/services` | 15,9% | 77,1% |
| Total de statements (`go tool cover`) | 15,3% | 70,8% |

Casos adicionados nesta execução:

| ID | Módulo | Descrição | Esperado | Obtido | Status |
|---|---|---|---|---|---|
| A12 | Importação de artigos | Validar título, fonte, caminho do PDF, usuário, extensão e cabeçalho | Rejeitar campos ausentes, extensão não PDF e cabeçalho inválido | Cada entrada inválida retornou `ErrInvalidArticle` | Passou |
| A13 | API de artigos | Importar sem PDF, sem título, sem fonte, com extensão incorreta e cabeçalho inválido | Recusar e retornar mensagem apropriada, sem persistir artigo | Todas as entradas inválidas foram recusadas pela API | Passou |
| A14 | Edição/exclusão de artigos | Editar ou excluir ID inexistente; repetir exclusão | Retornar `ErrArticleNotFound` e não informar sucesso | Serviço retornou o erro esperado e a API recusou as operações | Passou |
| A15 | Permissões | Exercitar rotas de curadoria e de gestão de usuários como anônimo, Pesquisador, Curador e Admin | Curador/Admin autorizados para curadoria; apenas Admin autorizado para usuários | Rotas negaram perfis sem permissão; perfis permitidos avançaram à operação. O seletor nativo de PDF não foi aberto no teste | Passou |
| A16 | Gestão de usuários | Perfil inválido, usuário inexistente, proteção do Admin, mudança de perfil e exclusão | Retornar erros corretos; impedir alteração/exclusão de Admin; permitir operações válidas | Erros e operações esperados confirmados no MySQL | Passou |
| A17 | Repositórios MySQL | Usuários, artigos, fontes, índice e histórico: criação, consulta, atualização e exclusão | Persistir e consultar os dados; proteger Admin e remover registros de teste | Operações verificadas no MySQL; registros `TESTE_` removidos e limpeza conferida | Passou |
| A18 | Serviços | Autenticação, setup do Admin, BM25/filtros/busca aproximada, indexação sem alvos/com alvo inválido e CSV do histórico | Aplicar regras, filtros e erros esperados; produzir CSV válido | Cenários unitários e de integração passaram | Passou |

Os testes de integração carregam o `.env` existente sem modificá-lo. Se o MySQL estiver indisponível, são ignorados com `t.Skip`; nesta execução, os testes de integração foram executados e passaram. Os dados criados usam identificadores `TESTE_`, são removidos em `t.Cleanup` e a limpeza é verificada. Nenhum teste usa `DROP` ou `TRUNCATE`.

A função `RecalculateDocumentFrequency` permanece sem cobertura: ela atualiza frequências de todos os termos e remove globalmente termos com frequência zero, o que poderia modificar dados fora do escopo `TESTE_`. Também permanecem sem cobertura `NewApp`, `NewIndexingService`, `startup`, `shutdown` e `main`, que são wiring/entrada do aplicativo.

Na execução desta medição, a conexão MySQL e os testes do histórico passaram; portanto, a tabela `operation_logs` estava disponível no banco utilizado. A observação anterior sobre a ausência dessa tabela não se aplica a este ambiente de teste.
