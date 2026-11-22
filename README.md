# Concordia

**Motor de busca full-text paralelo para artigos científicos.**

O Concordia é um aplicativo desktop que forma um acervo de artigos científicos em PDF, indexa o texto desses artigos em paralelo e permite pesquisá-los por relevância, com tolerância a erros de digitação. É um projeto acadêmico do **Grupo 17** (turma **CC-8-MB**), desenvolvido para as disciplinas de Fábrica de Software e Tópicos Avançados em Ciência da Computação.

## Sumário

- [Objetivo](#objetivo)
- [Funcionalidades](#funcionalidades)
- [Tecnologias utilizadas](#tecnologias-utilizadas)
- [Requisitos](#requisitos)
- [Instalação e configuração](#instalação-e-configuração)
- [Configuração do banco de dados](#configuração-do-banco-de-dados)
- [Como executar](#como-executar)
- [Primeiro uso](#primeiro-uso)
- [Testes](#testes)
- [Estrutura do projeto](#estrutura-do-projeto)
- [Documentação adicional](#documentação-adicional)
- [Equipe](#equipe)

## Objetivo

Oferecer uma ferramenta útil para a pesquisa acadêmica: reunir artigos científicos em um acervo local e encontrar rapidamente os mais relevantes para uma consulta, a partir do **conteúdo** dos documentos e não apenas de seus títulos.

Do ponto de vista técnico, o projeto demonstra o uso de **computação paralela** em um problema real: a indexação dos PDFs é feita por um *worker pool* de goroutines, e a busca usa um **índice invertido** persistido no banco e o ranqueamento **BM25**, a mesma base de técnicas usada na recuperação de informação que alimenta sistemas de RAG.

## Funcionalidades

O sistema possui três perfis de usuário:

| Perfil | O que faz |
| --- | --- |
| **Pesquisador** | Pesquisa por relevância (BM25) com busca aproximada e filtros; abre os detalhes e exporta o PDF de um artigo; acompanha o histórico de operações |
| **Curador** | Tudo o que o Pesquisador faz, mais: importa PDFs (individual e em lote), edita e exclui artigos, executa a indexação paralela com número de workers configurável, vê as métricas e o benchmark de paralelismo e exporta o histórico em CSV |
| **Administrador** | Tudo o que o Curador faz, mais a gestão de usuários: lista, busca, altera a função (entre Pesquisador e Curador) e exclui usuários |

Principais recursos:

- **Indexação paralela:** os workers leem o PDF, extraem o texto, normalizam e tokenizam, enquanto um único writer grava o índice no MySQL.
- **Busca ranqueada:** BM25 sobre o índice invertido, com pontuação e termos encontrados em cada resultado.
- **Busca aproximada (fuzzy):** termos ausentes do índice são substituídos pelo termo indexado mais próximo (distância de Levenshtein).
- **Filtros:** fonte, autor, idioma e período de publicação.
- **Benchmark de paralelismo:** compara a indexação com 1, 2, 4 e 8 workers, com tempo de processamento, speedup, eficiência e tempo no MySQL.
- **Dashboard:** indicadores do acervo e da indexação, gráficos e últimos artigos adicionados.
- **Histórico de operações:** registro persistente das ações dos usuários, com filtros e exportação em CSV.
- **Controle de permissões** validado no backend, e não apenas na interface.

## Tecnologias utilizadas

| Camada | Tecnologia |
| --- | --- |
| Núcleo do sistema | **Go** 1.27 (linguagem compilada, com concorrência nativa) |
| Aplicativo desktop | **Wails** v2.17 (Go + WebView) |
| Interface | **React** 19, **TypeScript** 5 e **Vite** 7 |
| Banco de dados | **MySQL** 8 (driver `go-sql-driver/mysql` 1.10) |
| Autenticação | **bcrypt** (`golang.org/x/crypto`) |
| Leitura de PDF | `github.com/giraffesyo/pdf` |
| Normalização de texto | `golang.org/x/text` |
| Configuração | `godotenv` (arquivo `.env`) |

## Requisitos

Para executar o projeto são necessários:

- **Go 1.27** ou superior (versão declarada no `go.mod`);
- **Node.js 20.19+** (ou 22.12+) e **npm**, requisito do Vite 7;
- **Wails CLI v2**;
- **MySQL 8.0** ou superior, em execução;
- Git.

O projeto foi desenvolvido em Windows (que exige o WebView2 Runtime, normalmente já instalado no Windows 10 e 11). O Wails também suporta macOS e Linux, com dependências próprias de cada sistema.

Instalação do Wails CLI e verificação do ambiente:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

O comando `wails doctor` indica o que ainda falta no ambiente.

## Instalação e configuração

**1. Clone o repositório e entre na pasta do aplicativo:**

```bash
git clone https://github.com/Tomaz-Arlindo/Concordia.git
cd Concordia/App
```

**2. Instale as dependências do frontend** (o `wails dev` também faz isso sozinho na primeira execução):

```bash
cd frontend
npm install
cd ..
```

**3. Crie o arquivo de configuração `.env`** a partir do exemplo:

Windows (cmd):

```cmd
copy .env.example .env
```

Linux ou macOS:

```bash
cp .env.example .env
```

Em seguida edite o `.env` e informe os dados do seu MySQL:

| Variável | Descrição | Valor padrão |
| --- | --- | --- |
| `DB_HOST` | Endereço do servidor MySQL | `localhost` |
| `DB_PORT` | Porta do MySQL | `3306` |
| `DB_USER` | Usuário do MySQL | `root` |
| `DB_PASSWORD` | Senha do usuário | *(sem padrão: obrigatória se o usuário tiver senha)* |
| `DB_NAME` | Nome do banco de dados | `concordia` |

O arquivo `.env` contém a senha do banco e **não deve ser versionado** (ele já está no `.gitignore`).

## Configuração do banco de dados

O aplicativo **não cria o banco nem as tabelas sozinho**: é preciso executar o script `database/schema.sql` uma vez. Ele cria o banco `concordia`, as 9 tabelas, os índices, as chaves estrangeiras e as quatro fontes padrão (`local`, `arxiv`, `pubmed` e `openalex`). O script usa `CREATE ... IF NOT EXISTS`, então pode ser executado novamente sem apagar dados.

Com o MySQL em execução, na pasta `App`, rode:

```bash
mysql -u root -p -e "source database/schema.sql"
```

O comando pede a senha do MySQL. Ele funciona no cmd, no PowerShell, no Linux e no macOS, desde que o executável `mysql` esteja no `PATH`.

> **No Windows**, se aparecer "`mysql` não é reconhecido", use o caminho completo do executável, por exemplo:
> `"C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe" -u root -p -e "source database/schema.sql"`
>
> Dentro do cliente `mysql`, o comando `SOURCE` deve receber o caminho **sem aspas** e com barras normais (`/`).

**Alternativa gráfica:** no MySQL Workbench, conecte ao servidor, abra `database/schema.sql` em *File > Open SQL Script* e execute o script inteiro.

**Para conferir** se deu certo:

```sql
USE concordia;
SHOW TABLES;
```

Devem aparecer 9 tabelas: `article_authors`, `article_index_stats`, `article_sources`, `articles`, `authors`, `operation_logs`, `term_occurrences`, `terms` e `users`.

Outros scripts da pasta `database/`:

| Arquivo | Uso |
| --- | --- |
| `schema.sql` | Criação completa do banco (instalação nova) |
| `migration_operation_logs.sql` | Acrescenta a tabela `operation_logs` a um banco criado antes do histórico de operações |
| `consultas_validacao.sql` | Consultas para conferir os dados |
| `limpeza_autores_orfaos_opcional.sql` | Remove autores sem artigo vinculado (opcional) |

Se o MySQL estiver desligado ou o `.env` estiver incorreto, o aplicativo abre mostrando **"Banco desconectado"**. Depois de corrigir o problema, o botão de reconexão ao lado do indicador refaz a conexão sem reiniciar o programa.

## Como executar

Todos os comandos abaixo são executados na pasta `App`.

**Modo de desenvolvimento** (com atualização automática da interface):

```bash
wails dev
```

**Gerar o executável:**

```bash
wails build
```

O executável é gerado em `build/bin/` (no Windows, `build/bin/desktop.exe`).

> Execute o programa sempre a partir da pasta do projeto. Os PDFs importados são guardados em `storage/articles/`, em um caminho relativo ao diretório de onde o aplicativo foi iniciado.

## Primeiro uso

1. Na primeira execução, se ainda não existir um Administrador, o sistema exibe um pop-up obrigatório para criá-lo (nome, e-mail, senha com no mínimo 8 caracteres e confirmação).
2. Entre como Administrador. Novos usuários se cadastram pela tela de login e começam como **Pesquisador**.
3. Na tela **Usuários**, o Administrador promove quem for atuar como **Curador**.
4. O Curador importa PDFs em **Importar** (individual ou em lote) e executa a indexação em **Indexação**.
5. Depois de indexados, os artigos aparecem nos resultados da tela **Pesquisar**.

## Testes

O projeto possui 38 funções de teste em Go, distribuídas em 14 arquivos, que cobrem autenticação, permissões, validação de artigos, indexação, busca (BM25, filtros e busca aproximada), exportação, histórico e repositórios.

```bash
go test ./... -v
```

Observações:

- O pacote da raiz embute a pasta `frontend/dist`, que precisa existir. Gere-a com `wails build` ou, dentro de `frontend`, com `npm run build`. Para rodar apenas os testes do núcleo, sem o frontend: `go test ./internal/... -v`.
- Os **testes de integração** usam o MySQL configurado no `.env` e são ignorados (`t.Skip`) quando o banco não está disponível. Os dados criados usam o prefixo `TESTE_`, são removidos ao final, e nenhum teste usa `DROP` ou `TRUNCATE`.
- Para medir a cobertura:

```bash
go test ./... -coverprofile=cover.out
go tool cover -func=cover.out
```

O plano de testes e os resultados registrados pela equipe estão em [`docs/TESTES.md`](docs/TESTES.md).

## Estrutura do projeto

```
App/
├── main.go, app.go                      Ponto de entrada e sessão (login, cadastro, saúde do banco)
├── article_api.go                       Rotas de artigos (listar, importar, editar, excluir)
├── article_export.go                    Exportação do PDF de um artigo
├── indexing_api.go                      Indexação paralela, busca BM25 e benchmark
├── log_api.go                           Histórico de operações e exportação em CSV
├── user_management_api.go               Gestão de usuários (Administrador)
├── admin_setup.go                       Criação do primeiro Administrador
├── reconnect.go                         Reconexão com o banco
├── internal/
│   ├── config/                          Conexão com o MySQL
│   ├── models/                          Modelos de dados
│   ├── repositories/                    Acesso ao banco de dados
│   └── services/                        Regras de negócio e testes
├── database/                            Scripts SQL (schema, migração e consultas)
├── frontend/
│   ├── src/                             Interface em React e TypeScript
│   └── wailsjs/                         Ligações (bindings) geradas pelo Wails
├── storage/articles/                    PDFs importados (nome = hash SHA-256)
├── docs/                                Documentação técnica
├── .env.example                         Modelo do arquivo de configuração
├── go.mod, wails.json
└── README.md
```

O código segue uma arquitetura em camadas: **interface (React) → rotas (Go) → serviços → repositórios → MySQL**. A comunicação entre a interface e o backend não usa HTTP: o Wails expõe os métodos Go diretamente ao frontend (veja [`docs/API.md`](docs/API.md)).

## Documentação adicional

| Documento | Conteúdo |
| --- | --- |
| [`App/docs/API.md`](App/docs/API.md) | Documentação das rotas (métodos expostos pelo backend) |
| [`App/docs/ARQUITETURA_FINAL.md`](App/docs/ARQUITETURA_FINAL.md) | Arquitetura do sistema |
| [`App/docs/MODELO_RELACIONAL.md`](App/docs/MODELO_RELACIONAL.md) | Modelo relacional |
| [`App/docs/TESTES.md`](App/docs/TESTES.md) | Plano de testes e resultados |

## Equipe

**Grupo 17** · Turma **CC-8-MB**

- Maria Karolina Rodrigues de Andrade
- Pedro Felipe Oliveira Tavares
- Pedro Oliveira Barros Batista de Lima
- Tomaz Arlindo Silva Ribeiro
- Vladison Lucas Costa Dos Santos
