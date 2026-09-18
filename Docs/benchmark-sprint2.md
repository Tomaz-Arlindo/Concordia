# Benchmark de Paralelismo — Sprint 2

Dataset: **21 documentos**.

| Workers | Processamento | Speedup proc. | Eficiência proc. | MySQL | Total | Speedup total |
|---:|---:|---:|---:|---:|---:|---:|
| 1 | 895 ms | 1,00x | 100,0% | 34.046 ms | 34.941 ms | 1,00x |
| 2 | 707 ms | 1,27x | 63,3% | 32.639 ms | 33.346 ms | 1,05x |
| 4 | 440 ms | 2,03x | 50,9% | 30.655 ms | 31.095 ms | 1,12x |
| 8 | 284 ms | 3,15x | 39,4% | 29.614 ms | 29.898 ms | 1,17x |

## Interpretação

A etapa de processamento apresentou ganho real com o aumento de workers. Com 8 workers, o tempo caiu de 895 ms para 284 ms, produzindo speedup de 3,15x.

O speedup total é menor porque a persistência no MySQL domina o tempo fim-a-fim. Isso ilustra a Lei de Amdahl.
