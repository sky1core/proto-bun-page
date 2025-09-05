# proto-bun-page

[한국어 문서 보기](README.ko.md)

Bun-based offset/cursor pagination utility with a unified request/response contract and a proto adapter. Focuses on DB-agnostic correctness (OR-chain WHERE) with optional knobs.

## Features
- Offset and cursor pagination with a single ordering plan
- Cursor = last page's primary key(s); anchor fetch + strict exclusive boundary
- Always appends PK as tiebreaker; supports composite PK
- AllowedOrderKeys filter and DefaultOrder support
- Limit clamping and non-positive defaulting with warnings
- Proto adapter: use `pagerpb.Page` request/response without requiring clients to know Bun

## Install
```
go get github.com/sky1core/proto-bun-page@latest
```

## Quick Start (Proto)
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

// Next cursor request
if out.Cursor != "" {
    in2 := &pagerpb.Page{Limit: 20, Order: in.Order, Cursor: out.Cursor}
    var next []Model
    _, _ = pg.ApplyAndScan(ctx, db.NewSelect().Model(&Model{}), in2, &next)
}
```

## Proto Adapter
- Schema: `proto/pager/v1/pager.proto`
 - Code generation required: generate `.pb.go` from `proto/pager/v1/pager.proto`.

```go
in := &pagerpb.Page{
    Limit: 20,
    Order: []*pagerpb.Order{{Key: "created_at", Desc: true}},
    Cursor: "", // start from the beginning
}
var rows []Model
q := db.NewSelect().Model(&Model{})
out, err := pg.ApplyAndScan(ctx, q, in, &rows)
```

### Codegen (`.pb.go`)
Generating code from proto is optional (the repo ships with hand-written types for convenience), but recommended for strict schema alignment.

1) Install protoc and the Go plugin

- macOS (Homebrew):
  - `brew install protobuf`
- Debian/Ubuntu:
  - `sudo apt-get update && sudo apt-get install -y protobuf-compiler`
- Go plugin (any OS):
  - `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`

2) Generate code

- From repo root: `make proto`
- Output uses `paths=source_relative` honoring `option go_package` in proto.

## Options
- AllowedOrderKeys: logical keys allowed in `order`. Empty → all model fields allowed.
- DefaultOrderSpecs: used when no order is specified (e.g., []OrderSpec{{Key:"created_at", Desc:true}}). If empty, defaults to PK DESC.
- DefaultLimit/MaxLimit: limit handling with clamping and non-positive defaulting.
- UseMySQLTupleWhenAligned: reserved for future optimization; not implemented.
  
Notes:
- Disallowed order keys are skipped and logged at Warn level.
- You may set `AllowedOrderKeys` to a normalized snake_case set (field names are normalized internally).

## Ordering Rules
- The same ordering plan applies to both offset and cursor.
- PK direction follows the last effective key; if none, PK DESC.
- Composite PK: all PK columns appended as tiebreakers and included in the cursor.
  - When no user order is provided, all PK columns are appended with DESC.

## Cursor Semantics
- Cursor is the last row's PK tuple from the previous page.
- Server fetches anchor row by PK, derives `(keys..., pk)` values, and builds a DB-agnostic OR-chain WHERE with exclusive boundary.

## Logging
- Warn when limit is clamped or non-positive is replaced by default.
- Warn when disallowed order keys are provided (they are skipped).

## Testing
- Run: `go test ./...`
- Includes boundary tests for page/limit, stale cursor, and allowed-keys filtering.

## Example
See `example/main.go` for offset, cursor, and proto adapter usage.

```go
func intPtr(i int) *int { return &i }

## Development Notes
- Internal development guidelines have been moved to `local/` (ignored by VCS).
- Public usage instructions are maintained in this README.
```
