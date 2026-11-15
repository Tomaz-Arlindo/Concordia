# Sprint 7 — Concordia

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
- **Repositório:** https://github.com/Tomaz-Arlindo/Concordia

*Motor de busca paralelo para artigos científicos*

---

## 1. Entregas da Sprint

Dos seis itens obrigatórios da Sprint 07, quatro já haviam sido implementados em sprints anteriores e foram verificados nesta etapa: dashboard, pesquisas, filtros e relatórios de indexação. Os dois restantes foram desenvolvidos agora: o histórico de operações e a exportação de dados. Esta documentação detalha, portanto, as duas implementações novas e as evidências de funcionamento de todo o conjunto.

---

## 2. Histórico de operações

O sistema passou a registrar automaticamente as operações realizadas pelos usuários na tabela `operation_logs`, persistida no MySQL. Cada registro guarda o usuário responsável, o tipo de operação, a entidade afetada, detalhes da ação e a data e hora.

São registrados dez tipos de operação: login, logout, cadastro de usuário, importação individual, importação em lote, indexação, busca, edição de artigo, exclusão de artigo e exportação de dados.

O registro é feito na camada de API, no momento em que a operação se conclui com sucesso, a partir do usuário autenticado na sessão. A gravação do log nunca interrompe a operação principal: em caso de falha ao registrar, o erro é reportado no console da aplicação e o fluxo de negócio prossegue normalmente.

### 2.1 Estrutura da tabela

| Coluna | Tipo | Conteúdo |
| --- | --- | --- |
| `id` | BIGINT UNSIGNED | Identificador sequencial |
| `user_id` | BIGINT UNSIGNED | Usuário responsável (chave estrangeira para `users`) |
| `operation` | VARCHAR(40) | Tipo da operação |
| `entity` | VARCHAR(40) | Entidade afetada (`articles`, `users`) |
| `entity_id` | BIGINT UNSIGNED | Registro específico, quando aplicável |
| `details` | VARCHAR(500) | Descrição da ação |
| `result_count` | INT | Quantidade de resultados, nas buscas |
| `created_at` | DATETIME | Data e hora do registro |

A chave estrangeira usa `ON DELETE SET NULL`: se um usuário for removido, o histórico das ações dele permanece, apenas sem a vinculação. Apagar o rastro junto com o usuário anularia o propósito da auditoria.

---

## 3. Relatórios

A tela de Histórico apresenta dois relatórios gerados a partir dos dados registrados:

- **Operações por tipo:** consolida a quantidade de ocorrências de cada operação, permitindo identificar o uso efetivo de cada funcionalidade.
- **Termos mais pesquisados:** agrupa as buscas realizadas e ordena pelos termos mais frequentes, oferecendo um indicador de interesse dos pesquisadores sobre o acervo.

A tela de Métricas mantém os indicadores do acervo (total de artigos, pendentes, indexados, autores, termos e ocorrências do índice invertido) e o relatório comparativo do benchmark de paralelismo.

---

## 4. Filtros e exportação

O histórico pode ser filtrado por tipo de operação e por período (data inicial e final). Os filtros são aplicados diretamente na consulta SQL, e não em memória, de modo que a quantidade de registros não compromete o tempo de resposta.

---

## 5. Evidências de funcionamento

As evidências foram geradas pela execução real das funcionalidades do sistema, com os perfis Curador, Pesquisador e Administrador, e cobrem o histórico de operações e a persistência dos dados após fechar o sistema.

### Testes de falha

Além dos fluxos de sucesso, foram executados cenários de falha para verificar o comportamento do sistema diante de entradas inválidas, operações não permitidas e situações que impedem a conclusão da ação.

---

## 6. Repositório

https://github.com/Tomaz-Arlindo/Concordia
