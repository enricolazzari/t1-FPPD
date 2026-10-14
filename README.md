# T1 FPPD — Filósofos Jantando em Go

| Versão | Pasta | Estratégia |
|---|---|---|
| V0 | `v0/` | sequencial (baseline) |
| V1 | `v1/` | concorrente, cada um pega esquerdo → direito (sujeita a deadlock; `vigia()` detecta) |
| V2 | `v2/` | semáforo (canal com buffer 4) antes dos garfos — sem deadlock, concorrência limitada |
| V3 | `v3/` | monitor com estados PENSANDO/FAMINTO/COMENDO (padrão) ou ordenação de garfos (`-strategy=ordered`) |

Todas: 5 filósofos, 20 ciclos, 2 s pensando, 3 s comendo. Não há `go.mod`: cada versão é um `main` de arquivo único.

## Rodar cada versão

```bash
go run v0/v0.go
go run v1/v1.go
go run v2/v2.go
go run v3/v3.go                      # monitor
go run v3/v3.go -strategy=ordered    # ordenação de garfos
```

Flags da V3: `-strategy=monitor|ordered`, `-think=2s`, `-eat=3s`, `-cycles=20`, `-runs=5` (N execuções, uma linha CSV cada), `-csv=arquivo.csv`, `-v` (log por ciclo — distorce o tempo), `-env` (imprime o ambiente).

Compilar: `go build -o build/v3 v3/v3.go` (idem para as outras). Verificar data race: `go run -race v3/v3.go -think=20ms -eat=30ms`.

## Experimentos e relatório

```bash
./scripts/run_experiments.sh          # v0 v1 v2 v3, 5 execuções cada (~1h50)
./scripts/run_experiments.sh v3       # só a V3
python3 scripts/report.py             # results/tabela.md + PNGs
```

- Cada versão imprime uma linha `CSV:versao,estrategia,execucao,tempo_total_ms,espera_media_ms,espera_max_ms,max_comendo,ciclos_f0..f4` no fim. V0, V1 e V2 ainda não imprimem: os patches em `patches/` acrescentam só essa linha (e `rounds := 20` na V0), sem mudar lógica — `git apply patches/v2-csv.patch`.
- V1 roda uma vez; a saída do `vigia()` vai para `results/v1_deadlock.txt` (evidência do deadlock).
- Requisitos: Go ≥ 1.21, `python3` + `matplotlib` (`pip3 install matplotlib`); no macOS, `gtimeout` (`brew install coreutils`) é opcional.

## Ambiente de execução (`go run v3/v3.go -env`)

```
Linguagem:   Go 1.27.2
SO:          Darwin 27.0.0 arm64 (darwin/arm64)
Processador: Apple M1 Pro
Cores:       10
Compilar:    go build -o build/v3 v3/v3.go
Executar:    ./build/v3 [-strategy=monitor|ordered] [-runs=5] [-csv=results/v3.csv] [-v]
```
