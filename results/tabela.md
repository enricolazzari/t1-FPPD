| Solução | Estratégia | Execuções | Tempo total médio (s) | Tempo total máx. (s) | Espera média (s) | Maior espera (s) | Máx. comendo | Ciclos por filósofo |
|---|---|---|---|---|---|---|---|---|
| V0 | sequencial | 5 | 500.32 | 500.34 | 0.000 | 0.000 | 1 | [20,20,20,20,20] |
| V2 | semaforo4 | 5 | 302.09 | 302.10 | 9.802 | 12.007 | 1 | [20,20,20,20,20] |
| V3 | monitor | 5 | 158.05 | 176.09 | 2.420 | 45.027 | 2 | [20,20,20,20,20] |
| V3 | ordered | 5 | 174.24 | 182.04 | 2.294 | 9.004 | 2 | [20,20,20,20,20] |

Médias e máximos sobre as execuções de cada solução. "Maior espera" = maior valor observado em todas as execuções.
