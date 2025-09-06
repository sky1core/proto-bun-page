MySQL 8 페이지네이션 벤치 (레포 추적용)

이 디렉터리는 레포에서 추적되는 MySQL 벤치마크 인프라입니다. 개발용 `local/` 경로가 아닌, 커밋 가능한 위치(`bench/mysql/`)로 정리되어 있습니다.

- 오케스트레이션: `docker-compose.yml`
- 엔트리포인트: `bench_entrypoint.sh`
- 벤치 패키지: `bench/` (별도 Go 모듈 `bench-local`)
- 초기화 SQL: `sql/`
- 출력물: `out/` (gitignore 처리)

사용법
- 레포 루트에서 실행: `bash run_mysql_bench.sh`
- 환경 변수:
  - `COUNT` 기본 10 (벤치 샘플 수)
  - `BENCHTIME` 기본 3s (벤치 각 샘플 시간)
  - `LABEL` 선택 A/B/C (세트 라벨링; A/B 비교 시 p-value 산출에 사용)
- 결과 파일: `bench/mysql/out/bench_*.txt`, `bench/mysql/out/benchstat_*.txt`
  - 단일 세트 요약: `bench_all.txt`, `benchstat_single.txt`
  - 라벨 세트: `bench_A.txt`, `bench_B.txt`, `bench_C.txt`
  - 비교 결과: `benchstat_ab.txt`(A vs B), 필요 시 `benchstat_bc.txt`, `benchstat_ac.txt`

권장 절차
- 웜업: `LABEL=A`로 1회 실행(캐시 워밍)
- 측정: `LABEL=B`(필요 시 `LABEL=C`)로 실행 — 기본 `COUNT=10 BENCHTIME=3s`
- 집계: 단일 세트 CI 확인 + A/B 비교 결과에서 p-value 확인

메모
- 오프셋 vs 커서, 커버링 vs 논커버링, limit(20/100) 변화 등 케이스 포함.
- 더 안정적인 신뢰구간이 필요하면 `BENCHTIME=5s` 또는 `COUNT=12`로 조정하세요.
- 벤치 패키지는 현재 레포의 `github.com/sky1core/proto-bun-page/pager`를 직접 사용합니다.

