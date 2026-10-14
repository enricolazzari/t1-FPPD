#!/usr/bin/env bash
# Compila e roda cada versão RUNS vezes (padrão 5), juntando as linhas "CSV:" do
# stdout em results/<versao>.csv e results/todas.csv.
#   ./scripts/run_experiments.sh            -> v0 v1 v2 v3
#   ./scripts/run_experiments.sh v3         -> só a V3
#   RUNS=3 ./scripts/run_experiments.sh v2 v3
# V1 roda uma vez só: ela trava, e a saída do vigia() vira results/v1_deadlock.txt.
set -u
cd "$(dirname "$0")/.."
RUNS=${RUNS:-5}
VERSIONS=${*:-v0 v1 v2 v3}
HEADER="versao,estrategia,execucao,tempo_total_ms,espera_media_ms,espera_max_ms,max_comendo,ciclos_f0,ciclos_f1,ciclos_f2,ciclos_f3,ciclos_f4"
mkdir -p build results

# timeout de segurança (o vigia da V1 já encerra sozinho em ~10 s)
if command -v timeout >/dev/null; then TO="timeout 900"
elif command -v gtimeout >/dev/null; then TO="gtimeout 900"
else TO=""; echo "aviso: sem timeout/gtimeout (brew install coreutils) — rodando sem limite"; fi

# duração estimada por execução, em segundos (20 ciclos, 2 s pensando, 3 s comendo)
est() { case $1 in v0) echo $((500 * RUNS));; v1) echo 15;; v2) echo $((302 * RUNS));;
        v3) echo $((2 * 180 * RUNS));; *) echo 0;; esac; }
total=0; for v in $VERSIONS; do total=$((total + $(est "$v"))); done
echo "Versões: $VERSIONS | execuções: $RUNS | duração estimada: ~$((total / 60)) min"

for v in $VERSIONS; do
  [ -f "$v/$v.go" ] || { echo "[$v] $v/$v.go não existe — pulando"; continue; }
  go build -o "build/$v" "$v/$v.go" || { echo "[$v] não compilou — pulando"; continue; }
  rows=$(mktemp)

  if [ "$v" = v1 ]; then
    echo "[v1] rodando (espera-se deadlock)..."
    $TO "./build/v1" > results/v1_deadlock.txt 2>&1
    echo "[v1] código de saída $? — saída completa em results/v1_deadlock.txt"
    grep '^CSV:' results/v1_deadlock.txt | sed 's/^CSV://' >> "$rows"
  elif [ "$v" = v3 ]; then
    for s in monitor ordered; do
      for i in $(seq "$RUNS"); do
        echo "[v3/$s] execução $i/$RUNS..."
        $TO "./build/v3" -strategy="$s" | grep '^CSV:' | sed 's/^CSV://' >> "$rows"
      done
    done
  else
    for i in $(seq "$RUNS"); do
      echo "[$v] execução $i/$RUNS..."
      $TO "./build/$v" | grep '^CSV:' | sed 's/^CSV://' >> "$rows"
    done
  fi

  if [ ! -s "$rows" ]; then
    if [ "$v" = v1 ]; then echo "[v1] travou, como esperado — sem linha CSV (só a evidência)"
    else echo "[$v] FALTOU: a versão não imprime linha 'CSV:' — aplicar patches/$v-csv.patch"; fi
    rm -f "$rows"; continue
  fi
  # renumera a coluna execucao (1..N por estratégia)
  { echo "$HEADER"; awk -F, -v OFS=, '{ $3 = ++n[$2]; print }' "$rows"; } > "results/$v.csv"
  rm -f "$rows"
  echo "[$v] -> results/$v.csv"
done

# junta todos os CSVs de versão que existirem (inclusive de rodadas anteriores)
{ echo "$HEADER"; for f in results/v[0-9].csv; do [ -f "$f" ] && tail -n +2 "$f"; done; } > results/todas.csv
echo "-> results/todas.csv ($(($(wc -l < results/todas.csv) - 1)) linhas)"
