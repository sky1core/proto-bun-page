#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# Ensure Docker compose uses Golang toolchain PATH inside bench container
# (bench container uses golang:1.22 which has /usr/local/go/bin)
cd "$ROOT_DIR/bench/mysql"

echo "[bench] Building and running MySQL + benchmarks (COUNT/BENCHTIME configurable)..."
docker-compose up --build --abort-on-container-exit bench
echo "[bench] Done. See bench/mysql/out/: bench_*.txt, benchstat_*.txt"
