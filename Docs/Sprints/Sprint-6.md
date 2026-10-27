# Sprint 6 — Concordia

## Identificação da equipe

- **Grupo 17**

- **Integrantes:**
    - Maria Karolina Rodrigues de Andrade
    - Pedro Felipe Oliveira Tavares
    - Pedro Oliveira Barros Batista de Lima
    - Tomaz Arlindo Silva Ribeiro
    - Vladison Lucas Costa Dos Santos
- **Turma:** CC-8-MB
- **Nome do projeto:** Concordia

---

## 1. Correções apontadas na Pré-Banca

### 1.1 Observações recebidas e ações realizadas

| Sprint | Observação da avaliação | Ação realizada |
| --- | --- | --- |
| 1 | Não deixar o escopo crescer além do necessário e manter a documentação alinhada ao que é implementado | O escopo foi organizado em três módulos (Pesquisador, Curador, Administração e Histórico) e este documento descreve apenas funcionalidades presentes no sistema |
| 2 | Finalizar a criação e o versionamento do banco | `database/schema.sql` versiona as 9 tabelas do sistema (`users`, `article_sources`, `authors`, `articles`, `article_authors`, `terms`, `term_occurrences`, `article_index_stats` e `operation_logs`) e cadastra as fontes padrão. `database/migration_operation_logs.sql` atualiza bancos já existentes e `database/consultas_validacao.sql` reúne consultas de conferência |
| 2 | Mostrar a estrutura real do projeto | Estrutura de pastas apresentada na seção 1.2 |
| 3 | O schema do banco deve estar versionado e o deploy local dependia de criação manual | O banco é criado a partir do `schema.sql` versionado. O sistema exibe o estado da conexão com o MySQL e possui botão de reconexão, que refaz a conexão e os serviços sem reiniciar o aplicativo |
| 3 | O perfil Curador só podia ser obtido alterando o banco diretamente | A primeira execução exige a criação de um Administrador, que altera a função dos usuários entre Pesquisador e Curador pela tela Usuários (seção 2) |
| 3 | Login, cadastro e CRUD funcionando | Login, cadastro, cadastro de artigos (importação individual e em lote), listagem, edição e exclusão implementados, além da gestão de usuários |
| 3 | Incluir o número do grupo | Grupo 17 na identificação |
| 4 | Testes | O projeto possui testes automatizados em Go (8 funções, executadas com `go test ./...`) que cobrem tokenização, índice invertido, busca aproximada, filtros de busca e limites do número de workers |
| 4 | Bugs com situação final e teste de confirmação | Tabela na seção 1.3 |
| 4 | Benchmark com 1, 2, 4 e 8 workers | A tela Métricas executa o benchmark com 1, 2, 4 e 8 workers e exibe, para cada um, tempo de processamento, speedup, eficiência, tempo no MySQL, tempo total e percentual de MySQL |
| 5 | Persistência dos dados | Usuários, artigos, índice e histórico são gravados no MySQL e os PDFs em `storage/articles/`, e permanecem disponíveis quando o aplicativo é reaberto |
| 5 | Identificar o que é protótipo e o que é sistema funcionando | Mapa de telas na seção 4.2: todas as telas listadas são do sistema funcionando |

### Estrutura do projeto

```
App/
├── app.go, article_api.go, indexing_api.go   Rotas expostas à interface (Wails)
├── admin_setup.go, user_management_api.go    Módulo de Administração
├── log_api.go                                Histórico de operações
├── reconnect.go                              Reconexão com o banco
├── main.go, wails.json, go.mod
├── internal/
│   ├── config/         Conexão com o MySQL
│   ├── models/         Modelos de dados
│   ├── repositories/   Acesso ao banco (usuários, artigos, índice, fontes, histórico)
│   └── services/       Regras de negócio (autenticação, artigos, indexação, busca,
│                       benchmark, administração, usuários, histórico) e testes
├── database/           schema.sql, migration_operation_logs.sql, consultas de validação
├── frontend/src/       Interface React/TypeScript (App.tsx, HistoryPanel.tsx)
├── storage/articles/   PDFs importados (nome = hash SHA-256)
└── docs/               Documentação técnica
```

### Bugs registrados e situação final

| Problema | Causa | Correção | Situação final | Confirmação |
| --- | --- | --- | --- | --- |
| Deadlocks no MySQL (erro 1213) durante a indexação | Transações concorrentes inseriam os mesmos termos na tabela `terms` | Os workers apenas processam os documentos; um único writer grava no banco | Corrigido | Indexação paralela com vários workers, pela tela Indexação e pelo benchmark com 1, 2, 4 e 8 workers |
| Chave primária duplicada em `term_occurrences` | "Martinez" e "Martínez" eram termos distintos no Go e o mesmo registro no MySQL (collation `utf8mb4_unicode_ci`) | Normalização dos termos: minúsculas, decomposição Unicode e remoção de acentos | Corrigido | Testes automatizados `TestNormalizeTermCollapsesAccentAndCaseVariants` e `TestBuildIndexedDocumentMergesCollationEquivalentTokens`, em `tokenizer_test.go` |

---

## 2. Terceiro módulo implementado

O terceiro módulo é **Administração e Histórico de Operações**. Ele completa o ciclo de uso do sistema: define quem acessa e com que perfil, e registra o que cada usuário fez.

### Funcionalidades

| Funcionalidade | Descrição |
| --- | --- |
| Criação do Administrador na primeira execução | Se não existir nenhum usuário `ADMIN`, um pop-up obriga a criação de um (nome, email, senha com no mínimo 8 caracteres e confirmação). Fechar o pop-up, pelo botão ou pela tecla Esc, pede confirmação e, se confirmada, encerra o aplicativo. Depois de criado, o pop-up não aparece mais |
| Perfil Administrador | Possui todas as permissões do Curador (importar, editar, excluir, indexar e executar o benchmark) e, em exclusivo, a gestão de usuários |
| Gestão de usuários | Tabela com id, nome, email, data de criação formatada e função. Busca em tempo real por qualquer uma dessas colunas, sem diferenciar acentos, com vários termos combinados |
| Alteração de função | Select na própria linha da tabela, somente entre Pesquisador e Curador, com confirmação que mostra a mudança (de X para Y) |
| Exclusão de usuário | Botão com confirmação. Os artigos importados pelo usuário continuam no acervo, sem responsável |
| Proteção de administradores | Contas `ADMIN` não podem ter a função alterada nem ser excluídas, o que impede deixar o sistema sem administrador. A regra é validada no backend |
| Histórico de operações | Registro persistente no MySQL (`operation_logs`) com usuário, operação, entidade, detalhes, quantidade de resultados e data e hora |
| Tela Histórico | Indicadores consolidados, filtros por operação e por período, relatório por tipo de operação, termos mais pesquisados e tabela de registros |
| Exportação do histórico | Arquivo CSV (separador `;` e UTF-8 com BOM, para abrir corretamente no Excel), com os filtros aplicados; restrita a Curador e Administrador |
| Status do banco e reconexão | Indicador "conectado" ou "desconectado" com botão que tenta reconectar sem reiniciar o aplicativo |

### Operações registradas no histórico

| Operação | Quando é registrada |
| --- | --- |
| `LOGIN`, `LOGOUT` | Entrada e saída do usuário |
| `CADASTRO_USUARIO` | Cadastro de novo usuário |
| `IMPORTACAO_ARTIGO`, `IMPORTACAO_LOTE` | Importação individual e em lote |
| `EDICAO_ARTIGO`, `EXCLUSAO_ARTIGO` | Edição e exclusão de artigo |
| `INDEXACAO` | Execução da indexação |
| `BENCHMARK` | Execução do benchmark |
| `BUSCA` | Pesquisa com texto informado (guarda o termo e a quantidade de resultados) |
| `EXPORTACAO_CSV` | Exportação do histórico |

O registro nunca interrompe a operação principal: se a gravação do histórico falhar, o erro é apenas escrito no console e a operação segue normalmente.

## 3. Integração entre módulos

### Pontos de integração

| Origem | Destino | Como a integração ocorre |
| --- | --- | --- |
| Administração | Curador e Pesquisador | A função definida pelo Administrador na tela Usuários é lida no banco no próximo login e define o menu e as permissões do usuário |
| Administração | Curador | O Administrador herda todas as permissões do Curador: a validação de curadoria aceita os perfis `CURADOR` e `ADMIN` |
| Curador | Pesquisador | Os artigos importados e indexados pelo Curador são os que aparecem nos resultados da busca BM25 do Pesquisador |
| Todos os módulos | Histórico | Login, cadastro, importação, edição, exclusão, indexação, benchmark e busca chamam o mesmo serviço de registro, que grava o usuário autenticado no momento da ação |
| Histórico | Administração | O histórico exibe o nome do usuário por associação com a tabela `users` |
| Administração | Curador (dados) | Ao excluir um usuário, os artigos que ele importou e os registros do histórico permanecem, com o responsável em branco (`ON DELETE SET NULL`) |
| Banco de dados | Administração | Ao reconectar com o banco, o sistema verifica de novo se existe Administrador e atualiza os dados da tela |

### Fluxo principal integrado

```
Primeira execução
  → Criação do Administrador (Administração)
  → Cadastro de um usuário (começa como Pesquisador)
  → Administrador promove o usuário a Curador (Administração)
  → Curador importa PDFs e executa a indexação paralela (Curador)
  → Pesquisador busca e filtra artigos (Pesquisador)
  → Todas as ações aparecem no Histórico (Histórico)
  → Exportação do histórico em CSV
```

| Passo | Módulo | Resultado |
| --- | --- | --- |
| 1. Criar o Administrador | Administração | O pop-up some e o Administrador entra no sistema |
| 2. Cadastrar usuário | Pesquisador | Usuário criado com função Pesquisador |
| 3. Promover a Curador | Administração | Função alterada após confirmação; no próximo login, o usuário vê os menus de curadoria |
| 4. Importar PDFs | Curador | Artigos com status `PENDING` e arquivos em `storage/articles/` |
| 5. Indexar | Curador | Artigos com status `INDEXED` |
| 6. Pesquisar | Pesquisador | Resultados com pontuação BM25 e termos encontrados |
| 7. Conferir o histórico | Histórico | Registros de login, importação, indexação e busca com o usuário correto |
| 8. Exportar o histórico | Histórico | Arquivo CSV com os registros filtrados |

---

## 4. Melhorias na interface

### Melhorias realizadas

| Melhoria | Onde | Benefício |
| --- | --- | --- |
| Menu lateral por perfil | Todas as telas | Cada usuário vê apenas o que pode usar: Visão geral, Pesquisar e Histórico para todos; Artigos, Importar, Indexação e Métricas para Curador e Administrador; Usuários só para o Administrador |
| Identificação do usuário | Rodapé do menu | Nome e função (Administrador, Curador ou Pesquisador) sempre visíveis |
| Indicador de banco com botão de reconexão | Tela de login e barra superior | O usuário percebe o problema de conexão e o corrige sem reiniciar o aplicativo |
| Pop-up de criação do Administrador | Primeira execução | Explica a obrigatoriedade e a função do Administrador, e pede confirmação antes de encerrar o aplicativo |
| Tela de gestão de usuários | Menu Usuários | Tabela com busca em tempo real, seleção de função na própria linha, confirmações em pop-up e coluna de ações alinhada |
| Tela de histórico | Menu Histórico | Filtros, relatório por tipo de operação, termos mais pesquisados, tabela e exportação |
| Mensagens de retorno e estados vazios | Telas de consulta | Mensagens de sucesso, erro e informação, e textos orientativos quando não há registros |
| Padrão visual único | Todas as telas | Mesmos cartões, botões (primário, secundário e de perigo), indicadores e janelas de confirmação |

### Mapa de telas

| Tela | Perfis | Condição |
| --- | --- | --- |
| Login e cadastro | Todos | Sistema funcionando |
| Criação do Administrador (pop-up) | Primeira execução | Sistema funcionando |
| Visão geral | Todos | Sistema funcionando |
| Pesquisar | Todos | Sistema funcionando |
| Histórico | Todos | Sistema funcionando |
| Artigos, Importar e Editar artigo | Curador e Administrador | Sistema funcionando |
| Indexação e Métricas | Curador e Administrador | Sistema funcionando |
| Usuários | Administrador | Sistema funcionando |

---

## 5. Ajustes de navegação

### Ajustes realizados

| Ajuste | Descrição |
| --- | --- |
| Navegação centralizada | Toda troca de tela passa por uma única função, que limpa mensagens antigas, fecha o painel de detalhes e atualiza a lista ao abrir Artigos |
| Acesso por perfil | As telas de curadoria e a de usuários só são exibidas ao perfil correspondente; além disso, as operações são validadas no backend |
| Destino após autenticar | Login, cadastro e criação do Administrador levam à Visão geral |
| Fluxo de importação | Após importar um artigo, o sistema abre a lista de Artigos com o novo registro |
| Fluxo de edição | O item "Artigos" permanece ativo durante a edição; salvar ou cancelar volta à lista de Artigos |
| Valores iniciais na importação | Ao abrir Importar, a fonte padrão já vem selecionada |
| Primeira execução | Enquanto não existir um Administrador, o pop-up cobre a tela e o sistema só é usado depois da criação |
| Reconexão com o banco | Ao reconectar, o sistema verifica de novo se existe Administrador e atualiza os dados da tela |
| Teclado | A tecla Esc, no pop-up do Administrador, abre a confirmação de saída |

### Telas acessíveis por perfil

| Tela | Pesquisador | Curador | Administrador |
| --- | --- | --- | --- |
| Visão geral | Sim | Sim | Sim |
| Pesquisar | Sim | Sim | Sim |
| Histórico | Sim | Sim | Sim |
| Artigos, Importar, Indexação e Métricas | Não | Sim | Sim |
| Usuários | Não | Não | Sim |