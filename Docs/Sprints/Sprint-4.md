# Sprint 4 — Concordia

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

## 2. Primeiro módulo totalmente funcional

O primeiro módulo do Concordia corresponde ao fluxo do perfil Pesquisador, responsável por consultar o acervo de artigos científicos já indexado. O usuário realiza cadastro e autenticação, com senha armazenada em formato de hash, e passa a ter acesso a um painel com estatísticas do acervo e a uma tela de pesquisa.

A busca é feita sobre o conteúdo extraído dos PDFs, utilizando um índice invertido persistido em MySQL e o algoritmo de ranqueamento BM25, que ordena os resultados por relevância. O módulo também oferece busca aproximada, que tolera pequenos erros de digitação por meio da distância de Levenshtein, e filtros por fonte, autor, idioma e data de publicação. Ao selecionar um resultado, o usuário visualiza os detalhes do artigo, como autores, fonte, idioma, data de publicação e DOI.

### Funcionalidades

| Funcionalidade | Descrição |
| --- | --- |
| Cadastro | Criação de conta com nome, e-mail e senha; todo novo usuário recebe o perfil Pesquisador |
| Autenticação | Login e logout, com senha protegida por hash (bcrypt) |
| Painel de visão geral | Estatísticas do acervo: total de artigos, pendentes, indexados e autores |
| Pesquisa por conteúdo | Busca no texto extraído dos PDFs, com ranqueamento por relevância (BM25) |
| Busca aproximada | Tolerância a erros de digitação por meio da distância de Levenshtein |
| Filtros | Refinamento por fonte, autor, idioma e período de publicação |
| Listagem do acervo | Exibição dos artigos cadastrados quando a pesquisa é feita sem texto |
| Detalhes do artigo | Consulta aos metadados completos do documento |

### Fluxo

```
Cadastro (ou login)
  → Painel de visão geral
  → Tela de pesquisa
  → Digitar consulta e (opcional) definir filtros e busca aproximada
  → Lista de resultados ordenada por relevância (BM25)
  → Detalhes do artigo selecionado
  → Nova pesquisa ou logout
```

### Telas do perfil Pesquisador

| Tela | Acesso | Conteúdo |
| --- | --- | --- |
| Login / Cadastro | Tela inicial, antes da autenticação | Formulários de acesso e de criação de conta, alternados por um botão; indicador de conexão com o banco |
| Visão geral | Menu "Visão geral" (tela inicial após o login) | Estatísticas do acervo (artigos, pendentes, indexados e autores) e apresentação do sistema |
| Pesquisa | Menu "Pesquisar" | Campo de busca, filtros por fonte, autor, idioma e período, opção de busca aproximada e lista de resultados com pontuação BM25 e termos encontrados |
| Detalhes do artigo | Botão "Detalhes" em um artigo da lista | Janela com autores, fonte, idioma, data de publicação, status, arquivo, DOI, identificador externo e descrição |

## 3. Persistencia de dados

A persistência do Concordia é híbrida: os **metadados e o índice de busca** são armazenados em um banco **MySQL 8**, enquanto os **arquivos PDF** ficam no sistema de arquivos local, na pasta `storage/articles/`. O banco guarda apenas o caminho e o hash SHA-256 de cada arquivo, nunca o PDF em si.

A conexão é configurada por variáveis de ambiente (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` e `DB_NAME`), lidas de um arquivo `.env` que não é versionado. O sistema utiliza um pool de conexões, com no máximo 10 conexões abertas e 5 ociosas. O estado da conexão é verificado ao iniciar a aplicação e exibido na interface como "Banco conectado" ou "Banco desconectado".

O acesso aos dados segue a estrutura **interface → serviços → repositórios → MySQL**. Todo o SQL fica concentrado nos repositórios e utiliza consultas parametrizadas, o que evita injeção de SQL.

O projeto também inclui o script `database/consultas_validacao.sql`, com consultas de validação do acervo e do índice. A persistência dos usuários pode ser comprovada cadastrando uma conta, fechando a aplicação e realizando o login novamente.

## 4. Validações

As validações são feitas principalmente no backend (Go), de modo que não dependem da interface. O frontend apenas apresenta ao usuário a mensagem retornada.

### Cadastro e login

| Campo | Regra | Onde é aplicada |
| --- | --- | --- |
| Nome | Obrigatório; espaços nas extremidades são removidos | Backend |
| E-mail | Obrigatório; convertido para minúsculas e sem espaços nas extremidades; deve ser único | Backend e banco (restrição `UNIQUE`) |
| E-mail (formato) | Campo do tipo `email` no formulário | Frontend |
| Senha | Mínimo de 8 caracteres | Backend (o formulário também informa o mínimo no campo) |
| Credenciais | E-mail e senha são conferidos com o hash armazenado (bcrypt) | Backend |

O botão de envio do formulário permanece desabilitado enquanto uma requisição está em andamento ou quando o banco está desconectado.

### Controle de acesso

Todas as operações de consulta (listagem, busca, detalhes e estatísticas) exigem usuário autenticado. Sem autenticação, as funções retornam resultado vazio ou a mensagem "Usuário não autenticado.". As operações restritas ao perfil Curador (importação, edição, exclusão, indexação e benchmark) são verificadas no backend por meio da função `requireCurator()`, e não apenas ocultadas na interface.

### Pesquisa

| Regra | Comportamento |
| --- | --- |
| Texto da consulta | Normalizado (minúsculas, sem acentos), com remoção de palavras comuns (*stop words*), de termos com menos de 2 caracteres e de termos repetidos |
| Consulta vazia | A tela lista o acervo completo, sem ranqueamento |
| Consulta sem termos válidos | Retorna lista vazia |
| Limite de resultados | A interface solicita 50 resultados; o backend assume 50 como padrão e limita o máximo a 100 |
| Busca aproximada | Quando ativada, termos não encontrados no índice são substituídos pelo termo indexado mais próximo, com distância de edição de até 2 (até 1 para termos de 4 letras ou menos), e recebem peso reduzido (0,72) |
| Filtro de fonte e idioma | Comparação exata, sem diferenciar maiúsculas de minúsculas |
| Filtro de autor | Correspondência parcial, sem diferenciar maiúsculas de minúsculas |
| Filtro de data | Intervalo entre data inicial e final; artigos sem data de publicação são excluídos quando há filtro de data |

Os filtros são aplicados em conjunto com um texto de consulta. Somente artigos já indexados aparecem nos resultados ranqueados.

## 5. Mensagens de erro

As mensagens são escritas em português e exibidas ao usuário sem expor detalhes técnicos internos do banco.

### Autenticação

| Situação | Mensagem exibida |
| --- | --- |
| Nome, e-mail ou senha inválidos no cadastro | "Preencha nome e email e use uma senha com pelo menos 8 caracteres." |
| E-mail já cadastrado | "Este email já está cadastrado." |
| Falha inesperada no cadastro | "Não foi possível criar o usuário." |
| Credenciais incorretas | "Email ou senha inválidos." |
| Falha inesperada no login | "Não foi possível realizar o login." |
| Banco de dados indisponível | "Banco de dados não conectado." |
| Falha de comunicação com o backend | "Ocorreu um erro ao comunicar com o backend Go." |

Em caso de sucesso, a interface exibe "Cadastro realizado com sucesso." ou "Login realizado com sucesso.".

### Pesquisa e consulta

| Situação | Mensagem exibida |
| --- | --- |
| Pesquisa sem resultados | "Nenhum resultado BM25 encontrado. Verifique se os artigos já foram indexados." |
| Lista vazia | "Nenhum artigo encontrado. Importe um PDF ou ajuste sua pesquisa." |
| Falha na pesquisa | "Falha ao pesquisar artigos pelo índice invertido." |
| Artigo inexistente | "Artigo não encontrado." / "Não foi possível abrir o artigo." |
| Usuário não autenticado | "Usuário não autenticado." |
| Serviço indisponível | "Serviço de artigos indisponível." |
| Falha ao carregar dados | "Não foi possível atualizar os dados do projeto." |

### Estado da conexão

O estado do banco é indicado visualmente por um indicador nas telas de login e no cabeçalho do sistema ("Banco conectado" / "Banco desconectado"), o que permite identificar rapidamente falhas de conexão.

## 6. Navegação entre telas

A aplicação funciona como uma interface de página única: a tela exibida é controlada por um estado interno, sem recarregamento de página. Antes da autenticação, apenas a tela de acesso é exibida. Após o login, o usuário é direcionado ao painel e passa a ter um menu lateral fixo.

| Tela | Acesso | Conteúdo |
| --- | --- | --- |
| Login / Cadastro | Tela inicial | Formulários alternáveis por um botão ("Criar cadastro" / "Fazer login") |
| Visão geral | Menu "Visão geral" (tela inicial após o login) | Estatísticas do acervo: artigos, pendentes, indexados e autores |
| Pesquisa | Menu "Pesquisar" | Campo de busca, filtros, opção de busca aproximada e lista de resultados com pontuação BM25 e termos encontrados |
| Detalhes do artigo | Botão "Detalhes" em um resultado | Janela com autores, fonte, idioma, data de publicação, status, arquivo, DOI, identificador externo e descrição |

- O menu lateral exibe ao Pesquisador apenas "Visão geral" e "Pesquisar", além do botão "Sair". Os itens de curadoria (Artigos, Importar, Indexação e Métricas) só são exibidos ao perfil Curador, e suas telas também são protegidas na renderização.
- A pesquisa é executada ao clicar em "Pesquisar" ou ao pressionar *Enter* no campo de busca.
- Os detalhes do artigo abrem em uma janela sobreposta, que é fechada pelo botão "×" ou ao clicar fora dela.
- Ao trocar de tela, as mensagens pendentes são limpas e a janela de detalhes é fechada.
- Ao clicar em "Sair", a sessão é encerrada e o usuário retorna à tela de acesso.
- Os dados do painel e da lista de artigos são carregados automaticamente após a autenticação.

## 7. Commits organizados

O desenvolvimento do Concordia foi organizado com Git e GitHub. Cada integrante criou um *fork* do repositório original, o que gerou uma cópia independente na própria conta, e fez nela as alterações, sem mexer diretamente na branch `main` do projeto principal. As mudanças foram enviadas ao repositório original por meio de *pull requests*, o que permitiu revisar o código antes da integração e manter a `main` estável. Os commits seguem uma organização por tarefa, com mensagens descritivas que indicam o que foi alterado e por quê, facilitando o acompanhamento do histórico e a identificação da contribuição de cada membro da equipe.