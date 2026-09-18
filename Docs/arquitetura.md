# Arquitetura do Concordia

```mermaid
flowchart LR
    UI[React + TypeScript] --> W[Wails]
    W --> APP[Go / App APIs]
    APP --> S[Services]
    S --> R[Repositories]
    R --> DB[(MySQL 8)]
    S --> FS[(Filesystem de PDFs)]
```

## Camadas

- **Frontend:** interface, formulários, busca, filtros e métricas.
- **Wails:** ponte entre frontend e Go.
- **APIs Go:** autenticação, artigos, indexação e benchmark.
- **Services:** regras de negócio.
- **Repositories:** acesso ao MySQL.
- **Filesystem:** armazenamento dos PDFs.

## Pipeline de indexação

```mermaid
flowchart LR
    A[PDFs] --> B[Worker pool]
    B --> C[Extração]
    C --> D[Normalização]
    D --> E[Tokenização]
    E --> F[Índice local]
    F --> G[Canal de resultados]
    G --> H[Writer controlado]
    H --> I[(MySQL)]
```

## Decisão de concorrência

Escritas concorrentes no MySQL causaram deadlocks. A solução foi manter processamento pesado paralelo e serializar a escrita compartilhada por um writer controlado.
