# Roteiro de apresentação — Concordia

## 1. Problema

O volume de artigos científicos é grande e a busca sequencial em documentos não escala bem.

## 2. Proposta

Aplicação desktop para curadoria, indexação paralela e pesquisa por relevância em artigos científicos.

## 3. Tecnologias

```text
Wails
React + TypeScript
Go
MySQL
filesystem local
```

## 4. Demonstração sugerida

### Login

Mostre Pesquisador e Curador.

### Curadoria

Importe ou mostre um PDF já importado e explique:

```text
PDF → SHA-256 → storage/articles → metadados no MySQL
```

### Indexação

Mostre a tela e explique o worker pool:

```text
jobs → goroutines → extração/tokenização → índice invertido
```

Mostre as tabelas `terms`, `term_occurrences` e `article_index_stats`.

### Busca

Pesquise uma expressão existente dentro do PDF e mostre:

- status INDEXED;
- score BM25;
- termos encontrados;
- filtros.

### Paralelismo

Abra Métricas e execute ou mostre um benchmark já registrado com 1, 2, 4 e 8 workers.

Explique que speedup pode deixar de crescer por causa de overhead, I/O e quantidade limitada de documentos.

## 5. Frase técnica curta

> O Concordia utiliza um worker pool em Go para processar documentos concorrentemente, persiste um índice invertido no MySQL e usa BM25 para ordenar os artigos por relevância.

## 6. Pontos que não devem ser exagerados

- com poucos PDFs, benchmark não prova escalabilidade;
- PDF escaneado sem camada de texto ainda precisa de OCR;
- o sistema atual é uma prova funcional acadêmica, não um motor de busca distribuído em produção.
