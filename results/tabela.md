| Solução | Estratégia | Execuções | Tempo total médio (s) | Tempo total máx. (s) | Espera média (s) | Maior espera (s) | Máx. comendo | Ciclos por filósofo |
|---|---|---|---|---|---|---|---|---|
| V0 | sequencial | 5 | 500.30 | 500.38 | 0.000 | 0.000 | 1 | [20,20,20,20,20] |
| V2 | semaforo4 | 5 | 302.12 | 302.12 | 9.803 | 12.006 | 1 | [20,20,20,20,20] |
| V3 | monitor | 5 | 159.30 | 167.11 | 2.451 | 12.009 | 2 | [20,20,20,20,20] |
| V3 | ordered | 5 | 166.48 | 173.09 | 2.349 | 9.005 | 2 | [20,20,20,20,20] |

Médias e máximos sobre as execuções de cada solução. "Maior espera" = maior valor observado em todas as execuções.
