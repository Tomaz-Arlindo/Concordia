# Sprint 2 — Concordia

## Objetivo

Formalizar arquitetura, modelagem do banco, componentes principais e base técnica do Concordia.

## Entregáveis previstos

- arquitetura do sistema;
- diagrama de classes/componentes;
- MER;
- modelo relacional;
- protótipos/telas principais;
- banco de dados criado;
- repositório GitHub estruturado.

## Arquitetura adotada

```text
React + TypeScript
        ↓
       Wails
        ↓
        Go
        ↓
Services / Repositories
      ↙       ↘
   MySQL     filesystem
```

## Banco de dados

Banco: `concordia`

Tabelas principais:
1. `users`
2. `articles`
3. `article_sources`
4. `authors`
5. `article_authors`
6. `terms`
7. `term_occurrences`
8. `article_index_stats`

## Perfis

### PESQUISADOR
- login;
- pesquisa;
- filtros;
- visualização de detalhes.

### CURADOR
- funções do Pesquisador;
- importação de PDFs;
- gerenciamento do acervo;
- indexação;
- configuração de workers;
- métricas e benchmark.

## Implementação já funcional

- aplicação desktop Wails;
- autenticação;
- perfis;
- integração real com MySQL;
- importação individual e em lote;
- SHA-256;
- extração e tokenização;
- índice invertido;
- worker pool com goroutines;
- busca BM25;
- busca aproximada;
- filtros;
- benchmark 1/2/4/8 workers;
- testes automatizados na camada de serviços.

## Benchmark

| Workers | Processamento | Speedup proc. | Eficiência proc. | MySQL | Total |
|---:|---:|---:|---:|---:|---:|
| 1 | 895 ms | 1,00x | 100,0% | 34.046 ms | 34.941 ms |
| 2 | 707 ms | 1,27x | 63,3% | 32.639 ms | 33.346 ms |
| 4 | 440 ms | 2,03x | 50,9% | 30.655 ms | 31.095 ms |
| 8 | 284 ms | 3,15x | 39,4% | 29.614 ms | 29.898 ms |

A etapa paralelizável caiu de 895 ms para 284 ms, com speedup de 3,15x. A persistência MySQL permanece como principal gargalo do tempo fim-a-fim.

## Checklist da Sprint

- [x] Arquitetura definida
- [x] MER e modelo relacional
- [x] Banco MySQL criado
- [x] Interface funcional
- [x] Integração Wails/Go/React/MySQL
- [x] Indexação paralela
- [x] Busca por relevância
- [x] Benchmark
- [x] Documentação técnica inicial
- [x] Estrutura Git preparada
