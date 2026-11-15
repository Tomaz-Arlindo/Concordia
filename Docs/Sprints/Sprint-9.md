# Sprint 9 — Concordia

## Identificação da equipe

- **Grupo 17**

- **Integrantes:**
    - Maria Karolina Rodrigues de Andrade
    - Pedro Felipe Oliveira Tavares
    - Pedro Oliveira Barros Batista de Lima
    - Tomaz Arlindo Silva Ribeiro
    - Vladison Lucas Costa Dos Santos
- **Scrum Master:** [PREENCHER]
- **Turma:** CC-8-MB
- **Nome do projeto:** Concordia
- **Repositório:** https://github.com/Tomaz-Arlindo/Concordia

*Motor de busca paralelo para artigos científicos*

---

## 1. Visão geral da Sprint

A Sprint 09 tem como objetivo verificar o sistema como um todo, corrigir os problemas encontrados e atualizar a documentação técnica. Os testes foram executados em duas frentes:

- **Testes automatizados em Go**, executados com o banco MySQL local de testes ligado, cobrindo sessão, perfis, autenticação, gestão de usuários, artigos, busca, datas, exportação, remoção de PDF e reconexão ao banco.
- **Testes manuais na interface**, com evidências em captura de tela, concentrados nos testes de validação do cadastro e da importação de artigos, e na verificação dos menus dos perfis Administrador e Pesquisador.

Os fluxos que não foram executados na interface estão identificados nas seções 3 e 4, com a indicação do teste automatizado que os cobre. A cobertura de código dos testes automatizados passou de 15,3% para 70,8% dos statements (seção 2.4).

---

## 2. Testes funcionais

### 2.1 Ambiente e critérios

Os testes que usam banco de dados rodam contra o MySQL local configurado no `.env`, que é utilizado apenas para testes. Os registros criados pelos testes recebem o prefixo `TESTE_` e são removidos ao final de cada teste. Não são usados `DROP` nem `TRUNCATE`. Se o banco estiver indisponível, os testes de integração são ignorados com `t.Skip` e uma mensagem clara, sem falhar a suíte.

Antes desta Sprint, apenas o pacote `internal/services` tinha testes (tokenizer, funções auxiliares de busca, nome do arquivo de exportação e benchmark). Foram acrescentados testes de sessão, perfis, autenticação, gestão de usuários, artigos, datas, exportação, remoção de PDF e reconexão, além de testes de unidade e de integração para repositórios, serviços de busca, indexação, histórico e configuração.

### 2.2 Casos de teste automatizados

| ID | Módulo | Descrição | Resultado esperado | Resultado obtido | Status |
| --- | --- | --- | --- | --- | --- |
| A01 | Suíte Go | Execução inicial, antes das alterações | Testes existentes passam | Passaram; os pacotes principais ainda não tinham testes | Passou |
| A02 | E-mail | Cadastro com `abc`, `a@b`, `teste@` e endereços válidos | Recusar os inválidos e aceitar os válidos | Os três formatos inválidos foram recusados; os válidos foram aceitos | Passou |
| A03 | Autenticação | Cadastro, login correto e incorreto, senha curta, duplicidade e logout | Validar login e cadastro, recusar os erros e encerrar a sessão no logout | Comportamentos esperados observados | Passou |
| A04 | Sessão e perfis | Acesso sem login e permissões de Admin, Curador e Pesquisador | Bloquear acesso anônimo; permitir gestão ao Admin e curadoria ao Curador | Comportamentos esperados observados | Passou |
| A05 | Gestão de Usuários | Admin lista, promove a Curador e exclui um usuário de teste | Alterações disponíveis só ao Admin e persistidas | Operações concluídas; usuário de teste removido | Passou |
| A06 | Artigos | Importar PDF de teste, consultar, buscar, editar e excluir | CRUD e busca funcionais; PDF removido na exclusão | Fluxo concluído; PDF ausente após a exclusão | Passou |
| A07 | Datas | Datas impossíveis, formato diferente de AAAA-MM-DD e data futura | Recusar antes da persistência | Importação e edição recusaram as entradas inválidas | Passou |
| A08 | Exportação | Copiar o PDF e rejeitar origem igual ao destino | Copiar o conteúdo; não sobrescrever o próprio arquivo | Ambos os comportamentos observados no helper de cópia | Passou |
| A09 | Remoção de PDF | Simular falha de remoção no sistema de arquivos | Propagar o erro | Erro propagado pelo serviço | Passou |
| A10 | Reconexão | Reconectar uma instância sem conexão | Retornar banco conectado | Status ok, banco conectado | Passou |
| A11 | Admin inicial | E-mail inválido e senha com menos de 8 caracteres | Recusar, com mensagem clara para o e-mail | Ambos recusados; o e-mail recebeu mensagem específica | Passou |
| A12 | Importação de artigos | Validar título, fonte, caminho do PDF, usuário, extensão e cabeçalho | Rejeitar campos ausentes, extensão não PDF e cabeçalho inválido | Cada entrada inválida retornou `ErrInvalidArticle` | Passou |
| A13 | API de artigos | Importar sem PDF, sem título, sem fonte, com extensão incorreta e com cabeçalho inválido | Recusar e retornar mensagem apropriada, sem persistir artigo | Todas as entradas inválidas foram recusadas pela API | Passou |
| A14 | Edição e exclusão de artigos | Editar ou excluir ID inexistente; repetir a exclusão | Retornar `ErrArticleNotFound` e não informar sucesso | O serviço retornou o erro esperado e a API recusou as operações | Passou |
| A15 | Permissões | Exercitar as rotas como anônimo, Pesquisador, Curador e Admin | Curador e Admin autorizados para curadoria; apenas Admin para gestão de usuários | Perfis sem permissão foram negados; os permitidos avançaram até a operação | Passou |
| A16 | Gestão de usuários | Perfil inválido, usuário inexistente, proteção do Admin, mudança de perfil e exclusão | Retornar os erros corretos, impedir alteração e exclusão de Admin e permitir as operações válidas | Erros e operações esperados confirmados no MySQL | Passou |
| A17 | Repositórios MySQL | Usuários, artigos, fontes, índice e histórico: criação, consulta, atualização e exclusão | Persistir e consultar os dados, proteger o Admin e remover os registros de teste | Operações verificadas; registros `TESTE_` removidos e limpeza conferida | Passou |
| A18 | Serviços | Autenticação, setup do Admin, buscas, indexação e CSV do histórico | Aplicar regras, filtros e erros esperados; produzir CSV válido | Cenários unitários e de integração passaram | Passou |

A exportação (A08) foi testada no nível da cópia do arquivo. O diálogo de escolha do destino, do Wails, não foi validado nos testes automatizados.

### 2.3 Execução da suíte

Comando executado na pasta `App`, sem cache:

```
go test -count=1 ./...
```

Resultado, com a suíte completa já ampliada (casos A01 a A18) e o MySQL ligado:

```
ok      desktop 4.724s
ok      desktop/internal/config 4.300s
?       desktop/internal/models [no test files]
ok      desktop/internal/repositories   4.956s
ok      desktop/internal/services       4.462s
```

Os testes de integração foram executados, não ignorados. O pacote `models` continua sem testes próprios.

### 2.4 Cobertura de código

A cobertura foi medida antes e depois da ampliação dos testes, a partir da pasta `App`, com o MySQL ligado, usando `go test -count=1 -coverprofile=cover.out ./...` e `go tool cover -func=cover.out`.

| Pacote | Antes | Depois |
| --- | --- | --- |
| `desktop` (App e APIs) | 26,7% | 50,9% |
| `desktop/internal/config` | 0,0% | 92,3% |
| `desktop/internal/models` | sem testes próprios | sem testes próprios |
| `desktop/internal/repositories` | 0,0% | 83,0% |
| `desktop/internal/services` | 15,9% | 77,1% |
| **Total de statements** | **15,3%** | **70,8%** |

Antes da ampliação, estavam sem cobertura (0%) as seguintes áreas:

- **APIs:** importação (individual e em lote), edição, exclusão e consulta de artigos, seleção de PDF, exportação de PDF, estatísticas do acervo, indexação, busca BM25 (com e sem filtros), benchmark, listagem de logs, resumo e exportação CSV do histórico, e a verificação de setup do Admin.
- **Serviços:** setup do Admin, artigos, autenticação (registro, login e validação de e-mail), benchmark, indexação, logs, extração de texto de PDF, busca e gestão de usuários.

Continuam em 0% após os testes: `NewApp`, `startup`, `shutdown`, `main` e `NewIndexingService` (inicialização e ligação dos componentes), além de `RecalculateDocumentFrequency`, que altera globalmente frequências e termos do índice e por isso ficou fora do escopo seguro dos testes com o banco local.

Para permitir testes isolados do serviço de indexação, foi acrescentada uma interface interna em `indexing_service.go`, sem alterar a assinatura do construtor público.

---

## 3. Testes de validação

Os testes abaixo foram executados na interface do sistema, com o banco conectado.

| ID | Tela | Dados informados | Resultado esperado | Resultado obtido | Status |
| --- | --- | --- | --- | --- | --- |
| M13 | Cadastro | Nome `teste`, e-mail `abc` | Recusar o cadastro e mostrar mensagem | O envio foi bloqueado com o aviso de que faltava o "@" no endereço | Passou |
| M12 | Cadastro | Nome `teste`, e-mail vazio, senha com 8 caracteres | Recusar o cadastro e mostrar mensagem | Apareceu "Preencha nome e email e use uma senha com pelo menos 8 caracteres." | Passou |
| M15 | Cadastro | E-mail já existente (`teste@teste.com`) | Recusar por duplicidade | Apareceu "Este email já está cadastrado." | Passou |
| M18 | Importar artigo (perfil Administrador) | PDF selecionado, título e autores preenchidos, data de publicação `30/11/2026` (futura) | Recusar a importação e mostrar mensagem | Apareceu "Data de publicação inválida; informe uma data real que não seja futura." | Passou |

**Observação sobre M13:** a mensagem exibida nesse caso é a validação nativa do campo de e-mail do navegador. A validação de formato feita pelo backend (bug B01) foi comprovada pelos testes automatizados A02 e A11.

### 3.1 Validações cobertas por teste automatizado

Os casos abaixo constam do roteiro manual, mas não foram executados na interface. Estão cobertos pelos testes automatizados indicados.

| ID | Cenário | Cobertura |
| --- | --- | --- |
| M14 | Cadastro com senha menor que 8 caracteres | A03 e A11 |
| M16 | Login com senha incorreta | A03 |
| M17 | Importação sem título, fonte ou PDF | A12 e A13 (título, fonte, PDF, extensão e cabeçalho) |
| M19 | Edição de artigo com data impossível ou futura | A07 |
| M20 | Erro do backend durante importação ou edição | Não executado; registrado como pendência (seção 8) |

---

## 4. Testes de navegação

Os menus laterais dos perfis Administrador e Pesquisador foram verificados na interface.

O menu do **Administrador**, exibido na tela de importação, apresenta as opções Visão geral, Pesquisar, Histórico, Artigos, Importar, Indexação, Métricas e Usuários. A tela identifica o usuário logado como `admin`, com o perfil Administrador, e mostra o indicador de banco conectado.

O menu do **Pesquisador**, exibido na tela Visão geral, apresenta apenas as opções Visão geral, Pesquisar e Histórico. A tela identifica o usuário logado como `pedro`, com o perfil Pesquisador. As opções Artigos, Importar, Indexação, Métricas e Usuários não aparecem para esse perfil.

| ID | Tela | Resultado esperado | Resultado obtido | Status |
| --- | --- | --- | --- | --- |
| M21 | Visão geral (perfil Pesquisador) | Menu sem as opções de curadoria e de gestão de usuários | Menu com apenas Visão geral, Pesquisar e Histórico | Passou |

A comparação entre os dois menus confirma, na interface, a restrição de acesso por perfil. As permissões também foram verificadas por teste automatizado (A04 e A15): o acesso sem login é bloqueado, o Pesquisador não acessa a curadoria, o Curador não gerencia usuários e o Admin tem acesso à Gestão de Usuários. A gestão de usuários, incluindo a proteção do perfil Admin, foi testada em A16. Pelo levantamento do código, o frontend oculta os menus por perfil e o backend também confere as permissões nas operações protegidas.

Os casos abaixo constam do roteiro manual, mas não foram executados na interface. Estão cobertos pelos testes automatizados indicados.

| ID | Cenário | Cobertura |
| --- | --- | --- |
| M22 | Menu do Curador | A04 e A15; não executado na interface |
| M23 | Gestão de Usuários pelo Admin | A05 e A16; não executado na interface |
| M24 e M25 | Logout e acesso a área autenticada sem login | A03 e A04; não executados na interface |
| M26 | Desligar e religar o banco e reconectar | A10; não executado na interface |

---

## 5. Registro de bugs e correções

| ID | Descrição | Onde | Gravidade | Correção | Status |
| --- | --- | --- | --- | --- | --- |
| B01 | O cadastro aceitava formatos de e-mail inválidos, como `abc`, `a@b` e `teste@` | `AuthService` e configuração inicial do Admin | Média | Validação de formato no backend, mensagem específica e testes de regressão | Corrigido e testado |
| B02 | A importação e a edição de artigos não validavam formato, calendário nem data futura da publicação antes de enviar ao SQL | `ArticleService` | Média | Validação estrita: data real no formato AAAA-MM-DD, sem datas futuras, com mensagem ao usuário e testes de data | Corrigido e testado |
| B03 | A exclusão de artigo ignorava falhas ao remover o PDF do disco | `ArticleService.DeleteArticle` | Média | O erro passou a ser propagado, com teste de regressão | Corrigido e testado |

**Observações:**

- **B02:** a rejeição de datas futuras é uma regra de negócio adotada nesta Sprint. Não foi executada a versão anterior contra o banco, portanto não se afirma que o MySQL gravaria datas inválidas; a correção garante que a aplicação as recuse antes da persistência.
- **B03:** a falha foi reproduzida em teste. Quando a remoção do arquivo falha, o registro do banco já foi excluído, e o PDF pode permanecer no disco, exigindo limpeza manual. Não há retentativa automática.

A ampliação dos testes (A12 a A18) não identificou novos bugs, portanto não há correções adicionais a registrar.

Arquivos de código alterados nas correções: `auth_service.go`, `admin_setup_service.go` e `article_service.go` (em `App/internal/services`), e `app.go`, `admin_setup.go` e `article_api.go` (mensagens ao usuário, em `App`).

---

## 6. Atualização da documentação

A documentação técnica de testes foi atualizada em `App/docs/TESTES.md`, com acréscimos referentes aos testes e às correções desta Sprint (incluindo os casos A12 a A18), sem remover conteúdo existente. O `README.md`, o `App/README.md` e o `App/docs/ARQUITETURA_FINAL.md` foram mantidos no padrão das sprints anteriores, sem alterações nesta Sprint. O comportamento validado e corrigido nesta Sprint (validação de e-mail, senha e data, rejeição de datas futuras e erro ao remover o PDF) está descrito nas seções 3 e 5 deste documento.

Os testes novos foram adicionados nos seguintes arquivos:

- `App/app_test.go`
- `App/integration_mysql_test.go`
- `App/internal/config/database_test.go`
- `App/internal/repositories/mysql_integration_test.go`
- `App/internal/services/admin_setup_service_test.go`
- `App/internal/services/article_validation_test.go`
- `App/internal/services/auth_service_test.go`
- `App/internal/services/indexing_service_test.go`
- `App/internal/services/log_service_test.go`
- `App/internal/services/user_management_integration_test.go`

---

## 7. Repositório

https://github.com/Tomaz-Arlindo/Concordia