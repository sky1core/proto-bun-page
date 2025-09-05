# proto-bun-page (한국어)

Bun 기반 오프셋/커서 페이지네이션 유틸리티. 단일 Protobuf `Page` 메시지 계약과 Bun 백엔드(OR-체인 WHERE)로 정확하고 일관된 결과를 제공합니다. 구현은 단일 PK 전제를 따릅니다.

## 주요 기능
- 단일 오더 경로: 오프셋/커서 공통 정렬 플랜
- 커서 = 마지막 행의 단일 PK(opaque); 앵커 조회 + 배타(exclusive) 경계
- 항상 단일 PK를 타이브레이커로 자동 추가(빈 오더 시 PK DESC)
- `AllowedOrderKeys` 화이트리스트, `DefaultOrderSpecs` 지원
- 리밋 검증: 0/미지정 → 기본값, 상한 초과 → clamp (Warn)
- Proto 어댑터: `pagerpb.Page` 요청/응답으로 바로 사용

## 설치
```
go get github.com/sky1core/proto-bun-page@latest
```

## 빠른 시작 (Proto)
```go
pg := pager.New(&pager.Options{
    DefaultLimit: 20,
    MaxLimit:     100,
    LogLevel:     "info",
    AllowedOrderKeys:  []string{"created_at", "name", "id"},
    DefaultOrderSpecs: []pager.OrderSpec{{Key: "created_at", Desc: true}},
})

in := &pagerpb.Page{Limit: 20, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
var rows []Model
q := db.NewSelect().Model(&Model{})
out, err := pg.ApplyAndScan(ctx, q, in, &rows)
if err != nil { /* handle */ }

// 다음 페이지 (커서 기반)
if out.Cursor != "" {
    in2 := &pagerpb.Page{Limit: 20, Order: in.Order, Cursor: out.Cursor}
    var next []Model
    _, _ = pg.ApplyAndScan(ctx, db.NewSelect().Model(&Model{}), in2, &next)
}
```

## 프로토 코드 생성
- `protoc` + `protoc-gen-go` 설치 후, 루트에서 `make proto` 실행
- `.pb.go`는 CI에서 생성하며 레포에 포함하지 않습니다

- `AllowedOrderKeys`: 정렬에 허용되는 bun 컬럼명 목록(공백이면 모델 필드 모두 허용)
- `DefaultOrderSpecs`: 비어있을 때 사용할 기본 오더(예: `[]OrderSpec{{Key:"created_at", Desc:true}}`), 미설정이면 PK DESC
- `DefaultLimit`/`MaxLimit`: 리밋 기본/상한(clamp)
- `UseMySQLTupleWhenAligned`: 추후 최적화 예약(현재 미구현)

## 정렬 규칙
- 페이지/커서 공통 정렬 플랜 사용
- PK 방향은 마지막 사용자 지정 키를 따름; 사용자 오더가 없으면 PK DESC
- 오더 뒤에 PK 자동 추가로 전순서 보장

- 커서 = 이전 응답 마지막 행의 “단일 PK” 값 (base64 URL-safe, opaque)
- 서버: 커서(PK)로 앵커 조회 → (정렬키…, PK)로 OR-체인 WHERE 구성 → exclusive 경계

## 로깅
- 리밋 기본값 대체/상한 클램프 시 Warn
- 비허용/모델 미존재 정렬 키 입력 시 에러

## 테스트
- `go test ./...`
- 경계/타이/풀스캔/어댑터 테스트 포함

## 예제
- `example/main.go`: Proto 어댑터 사용 예시 포함

## 개발 참고
- 내부 개발 지침은 `local/` 경로에 있고, 버전관리에서 제외됩니다.
