# Análise do benchmark — Sprint 2

## Objetivo

Avaliar o efeito do número de workers na etapa paralelizável da indexação e separar esse ganho do custo de persistência no MySQL.

## Configurações medidas

- 1 worker
- 2 workers
- 4 workers
- 8 workers

## Métricas registradas

- tempo de processamento paralelo;
- speedup de processamento;
- eficiência de processamento;
- tempo de persistência MySQL;
- tempo total;
- speedup total;
- percentual do tempo no MySQL;
- número de falhas.

## Interpretação

O processamento paralelo corresponde à leitura dos PDFs, extração do texto, normalização, tokenização e construção do índice local.

A persistência inclui a gravação do vocabulário, postings, estatísticas dos documentos e recálculo de document frequency.

A implementação usa um writer controlado na etapa de persistência porque testes anteriores com múltiplas transações concorrentes provocaram deadlocks no MySQL. Assim, o benchmark permite mostrar separadamente:

1. onde as goroutines produzem ganho;
2. quanto o banco representa do tempo total;
3. por que o speedup fim-a-fim pode ser menor que o speedup da etapa paralelizável.

## Resultado da Sprint 2

Preencha esta tabela com os valores exibidos pelo Concordia após a execução final:

| Workers | Processamento | Speedup proc. | Eficiência proc. | MySQL | Total | Speedup total | % MySQL | Falhas |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 |  | 1,00x | 100% |  |  | 1,00x |  |  |
| 2 |  |  |  |  |  |  |  |  |
| 4 |  |  |  |  |  |  |  |  |
| 8 |  |  |  |  |  |  |  |  |

## Conclusão sugerida

O Concordia aplica paralelismo na fase de processamento dos documentos por meio de um pool de goroutines. A medição por fases permite distinguir o ganho dessa etapa do custo serial de persistência. Esse comportamento é compatível com a Lei de Amdahl: o ganho fim-a-fim é limitado pela fração do sistema que permanece serial.
