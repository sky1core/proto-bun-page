MySQL 8 Pagination Bench (Repo-tracked)

[한국어 문서](./README.ko.md)

This directory contains the tracked benchmark harness for MySQL, moved out of `local/` so it can be committed safely.

- Orchestration: `docker-compose.yml`
- Entrypoint: `bench_entrypoint.sh`
- Bench package: `bench/` (separate Go module `bench-local`)
- SQL init: `sql/`
- Outputs: `out/` (gitignored)

Usage
- From repo root: `bash run_mysql_bench.sh`
- Environment knobs: `COUNT` (default 10), `BENCHTIME` (default 3s), `LABEL` (optional A/B/C)
- Results: `bench/mysql/out/bench_*.txt`, `bench/mysql/out/benchstat_*.txt`

Notes
- For A/B comparisons, run twice with `LABEL=A` then `LABEL=B`. A third `LABEL=C` can be used to check stability.
- The bench package imports `github.com/sky1core/proto-bun-page/pager` from the repo.
