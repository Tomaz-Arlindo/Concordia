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
