#!/usr/bin/env bash
set -euo pipefail

DSN_DEFAULT='bench:benchpass@tcp(mysql:3306)/benchdb?parseTime=true&multiStatements=true'
export DSN="${DSN:-$DSN_DEFAULT}"

OUT_DIR="/app/bench/mysql/out"
mkdir -p "$OUT_DIR"

cd /app/bench/mysql/bench
echo "[bench] DSN=$DSN"
COUNT_DEFAULT=10
BENCHTIME_DEFAULT=3s
COUNT="${COUNT:-$COUNT_DEFAULT}"
BENCHTIME="${BENCHTIME:-$BENCHTIME_DEFAULT}"
LABEL="${LABEL:-}"

FNAME="bench_all.txt"
if [ -n "$LABEL" ]; then
  FNAME="bench_${LABEL}.txt"
fi

echo "[bench] Running benchmarks with COUNT=$COUNT, BENCHTIME=$BENCHTIME, LABEL=${LABEL:-none}..."
# Single consolidated run file with repeated samples to enable CI/p-values
go test -bench=. -benchtime="$BENCHTIME" -count "$COUNT" -run=^$ ./... | tee "$OUT_DIR/$FNAME"

echo "[bench] Aggregating with benchstat for this set (n=$COUNT per benchmark)..."
/go/bin/benchstat "$OUT_DIR/$FNAME" | tee "$OUT_DIR/benchstat_${LABEL:-single}.txt"

# If A/B files exist, produce labeled comparison with p-values
if [ -f "$OUT_DIR/bench_A.txt" ] && [ -f "$OUT_DIR/bench_B.txt" ]; then
  echo "[bench] Comparing A vs B with benchstat ..."
  /go/bin/benchstat -alpha 0.05 "$OUT_DIR/bench_A.txt" "$OUT_DIR/bench_B.txt" | tee "$OUT_DIR/benchstat_ab.txt"
fi

echo "[bench] Done. Latest set: $OUT_DIR/$FNAME"

