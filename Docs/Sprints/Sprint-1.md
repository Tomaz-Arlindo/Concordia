# Documento de Definição do Projeto — Sprint 1

## 1. Identificação da equipe

- **Integrantes:** 
    - Maria Karolina Rodrigues de Andrade
    - Pedro Felipe Oliveira Tavares
    - Pedro Oliveira Barros Batista de Lima
    - Tomaz Arlindo Silva Ribeiro
    - Vladison Lucas Costa Dos Santos
- **Turma:** CC-8-MB
- **Nome inicial do projeto:** Concordia (_Sugeito a alterção_)

## 2. Escolha do tema

Sistema de indexação e busca paralela aplicado a uma base de artigos científicos abertos (ex.: arXiv, PubMed ou OpenAlex). O projeto une conceitos de recuperação de informação (indexação, busca por relevância) com computação paralela, aplicados a um problema real de pesquisa acadêmica.

## 3. Definição do problema

O volume de artigos científicos publicados cresce continuamente, tornando cada vez mais difícil localizar rapidamente conteúdo relevante dentro de uma grande base de documentos. Buscas sequenciais, palavra por palavra, em coleções grandes são lentas e não escalam.

Isso é relevante porque:
- Pesquisadores e estudantes precisam de ferramentas rápidas para localizar literatura relevante.
- É um problema real, mensurável e onde o ganho de paralelismo é claro e demonstrável (indexar uma base grande é uma tarefa que se beneficia diretamente de múltiplos núcleos).
- A mesma técnica (indexação + busca) é a base de sistemas modernos de IA que combinam busca com modelos de linguagem (arquitetura conhecida como RAG — Retrieval-Augmented Generation), o que reforça a relevância prática do tema.

## 4. Objetivos do sistema

**Objetivo geral:** desenvolver um sistema que indexe uma base de artigos científicos de forma paralela e permita buscas rápidas e tolerantes a erro de digitação sobre essa base, com autenticação e perfis de usuário distintos.

**Principais resultados esperados:**
- Índice invertido construído em paralelo, com tempo de indexação mensurável e comparável entre diferentes números de threads.
- Busca por termo com resultados ranqueados por relevância (BM25) e tolerância a erros de digitação (fuzzy search).
- Sistema com autenticação, perfis de usuário, persistência em banco de dados e interface de acesso.
- Documentação do processo de desenvolvimento e dos resultados de desempenho obtidos com paralelismo.

## 5. Público-alvo

- **Pesquisadores e estudantes** que precisam localizar artigos relevantes rapidamente dentro de uma base.
- **Curadores da base** (perfil interno do sistema) responsáveis por manter os artigos atualizados.
- Indiretamente, **sistemas de IA/agentes** que poderiam consumir esse tipo de busca como fonte de contexto.

## 6. Requisitos Funcionais

| ID | Descrição |
|---|---|
| RF01 | Cadastro e autenticação de usuários (login/senha) |
| RF02 | Diferenciação de perfis de acesso: **Curador** (gerencia a base e o índice) e **Pesquisador** (realiza buscas) |
| RF03 | Importação/atualização da base de artigos científicos (perfil Curador) |
| RF04 | Geração/atualização do índice invertido de forma paralela, com opção de configurar o número de núcleos/threads utilizados (perfil Curador) |
| RF05 | Exibição de métricas técnicas do processo de indexação (tempo de execução, nº de threads, comparação de desempenho) (perfil Curador) |
| RF06 | Busca de artigos por termo, com resultados ranqueados por relevância (perfil Pesquisador) |
| RF07 | Suporte a busca tolerante a erros de digitação (fuzzy search) |
| RF08 | Visualização de metadados/detalhes do artigo retornado na busca |
| RF09 | Persistência de usuários, artigos, metadados e índices em banco de dados |
| RF10 | CRUD completo da entidade Artigo (cadastrar, consultar, atualizar e excluir) |
| RF11 | Dashboard com indicadores de uso e desempenho (nº de artigos indexados, tempo de indexação, comparação entre execuções com diferentes números de threads) |
| RF12 | Relatórios de desempenho da indexação e de uso do sistema (ex.: buscas mais frequentes, artigos mais acessados) |
| RF13 | Filtros de busca por critérios adicionais (ex.: data de publicação, autor, categoria/área) |
| RF14 | Exportação de resultados de busca e/ou relatórios em CSV ou PDF |
| RF15 | Registro de log/histórico das principais operações (buscas realizadas, execuções de indexação e parâmetros utilizados) |
| RF16 | Exposição das funcionalidades de busca e indexação por meio de API documentada |

## 7. Requisitos Não Funcionais

| ID | Descrição |
|---|---|
| RNF01 | Núcleo do sistema desenvolvido em linguagem compilada (Go) |
| RNF02 | Uso de computação paralela na etapa de indexação (e potencialmente na busca fuzzy), mesmo que isso não maximize a eficiência — o foco é o entendimento do funcionamento do paralelismo |
| RNF03 | Desempenho mensurável: o sistema deve permitir comparar tempos de execução com diferentes números de threads (benchmark) |
| RNF04 | Segurança básica de autenticação (armazenamento seguro de senhas, controle de sessão) |
| RNF05 | Interface simples e objetiva, priorizando as funcionalidades centrais em vez de complexidade visual |
| RNF06 | Persistência confiável dos dados em banco de dados relacional |
| RNF07 | Código modular, testável e documentado |
| RNF08 | Portabilidade — o sistema deve rodar de forma consistente independentemente do número de núcleos disponíveis na máquina |
| RNF09 | Cobertura de testes automatizados nos módulos principais (indexação, busca e autenticação) |
| RNF10 | Aplicação executável em ambiente local, com instruções de instalação e configuração documentadas (README) |
| RNF11 | APIs do sistema documentadas (rotas, parâmetros e respostas) |

## 8. Casos de Uso

1. Usuário se cadastra e faz login no sistema.
2. Curador importa/atualiza artigos na base.
3. Curador dispara a geração (ou atualização) do índice, escolhendo o número de núcleos/threads.
4. Sistema exibe ao Curador métricas técnicas do processo de indexação.
5. Pesquisador busca artigos por termo.
6. Sistema retorna resultados ranqueados, com tolerância a erros de digitação.
7. Pesquisador visualiza detalhes/metadados de um artigo retornado.
8. Curador visualiza dashboard com indicadores de uso e desempenho do sistema.
9. Pesquisador filtra resultados de busca por critérios adicionais (data, autor, categoria).
10. Pesquisador exporta resultados de busca ou relatórios em CSV/PDF.
11. Sistema registra log das buscas e execuções de indexação realizadas.

## 9. Product Backlog

| ID | História | Prioridade |
|---|---|---|
| PB01 | Como Curador, quero importar artigos para a base, para manter o acervo atualizado | Alta |
| PB02 | Como Curador, quero disparar a geração do índice com paralelismo, para preparar a base para busca | Alta |
| PB03 | Como usuário, quero me cadastrar e fazer login, para acessar o sistema de acordo com meu perfil | Alta |
| PB04 | Como Pesquisador, quero buscar artigos por palavra-chave, para encontrar conteúdo relevante rapidamente | Alta |
| PB05 | Como Pesquisador, quero ver os resultados ranqueados por relevância, para priorizar a leitura | Alta |
| PB06 | Como Curador, quero escolher o número de threads usados na indexação, para comparar desempenho | Média |
| PB07 | Como Curador, quero ver métricas técnicas do processo de indexação, para analisar o ganho do paralelismo | Média |
| PB08 | Como Pesquisador, quero que a busca tolere erros de digitação, para não perder resultados por pequenos erros | Média |
| PB09 | Como equipe, queremos testes automatizados dos módulos centrais, para garantir confiabilidade | Média |
| PB10 | Como equipe, queremos documentação do processo e da arquitetura, para atender à disciplina | Média |
| PB11 | Como Curador, quero um CRUD completo da entidade Artigo, para gerenciar a base independentemente da indexação | Alta |
| PB12 | Como Curador, quero visualizar um dashboard com indicadores de uso e desempenho, para acompanhar o sistema | Média |
| PB13 | Como Pesquisador, quero filtrar resultados por data/autor/categoria, para refinar minha busca | Média |
| PB14 | Como Pesquisador, quero exportar resultados de busca em CSV/PDF, para usar fora do sistema | Baixa |
| PB15 | Como equipe, queremos registrar logs das operações de busca e indexação, para rastreabilidade | Baixa |
| PB16 | Como equipe, queremos expor a busca e a indexação por meio de uma API documentada, para integração externa | Média |

## 10. Cronograma Inicial

*(baseado no cronograma oficial de Sprints divulgado pela disciplina de Fábrica de Software)*

| Sprint | Data-limite | Foco |
|---|---|---|
| 1 (atual) | 05/09 | Planejamento do projeto (este documento) e criação do repositório |
| 2 | 19/09 *(prazo alterado de 12/09)* | Arquitetura do sistema, diagrama de classes, MER, modelo relacional, protótipo das telas, banco de dados criado |
| 3 | 19/09 | Banco de dados conectado, login funcional, cadastro de usuários, controle de perfis, CRUD principal (Artigo) funcionando, primeiro deploy local |
| 4 | 26/09 | Primeiro módulo completo (indexação básica), persistência, validações, mensagens de erro, navegação |
| 5 | 03/10 | Segundo módulo completo (busca), integração com o banco de dados, regras de negócio, testes e correção de bugs |
| 6 | 17/10 | Correções apontadas na Pré-Banca, terceiro módulo (paralelismo/fuzzy search), integração entre módulos, melhorias de interface |
| 7 | 24/10 | Dashboard, relatórios, filtros, exportação, logs/histórico de operações |
| 8 | 31/10 | Todas as funcionalidades implementadas, controle de permissões por perfil, melhorias de usabilidade, revisão das regras de negócio |
| 9 | 07/11 | Testes funcionais, de validação e de navegação completos; correção de bugs; atualização da documentação |
| 10 | 14/11 | Release Candidate: sistema estável, interface final, banco de dados final, APIs documentadas, README atualizado |
| 11 | 21/11 | Manual do usuário, manual técnico, documentação final consolidada, revisão de código, organização do GitHub |
| 12 | 28/11 | Versão final, correção de todos os bugs, code freeze, preparação dos vídeos, ensaio da apresentação |
| Entrega Final | 05/12 | Sistema completo, código-fonte, documentação (manuais, arquitetura, MER), vídeo horizontal (YouTube) e vertical (Instagram) |

## 11. Repositório GitHub

- **Link:** https://github.com/Tomaz-Arlindo/Concordia
