# proto-bun-page

[한국어 문서 보기](README.ko.md)

Requirements
- Go 1.21+ (uses `log/slog`)

Bun-based offset/cursor pagination utility with a unified request/response contract and a proto adapter. Focuses on DB-agnostic correctness (OR-chain WHERE) with optional knobs.

## Features
- Offset and cursor pagination with a single ordering plan
- Cursor = last page's single PK (opaque); anchor fetch + strict exclusive boundary
- Always appends single PK as tiebreaker
- AllowedOrderKeys filter and DefaultOrderSpecs support
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

Semantics (selector)
- Choose exactly one of `page` or `cursor`.
- `page` is 1-based: if explicitly set, it must be >= 1 (1 → offset=0).
- `cursor` is opaque; if explicitly set to empty string, it means "from the start".
- If neither is set, defaults to cursor mode from the start.

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
- AllowedOrderKeys: bun column names allowed in `order`. Empty → all model fields allowed.
- DefaultOrderSpecs: used when no order is specified (e.g., []OrderSpec{{Key:"created_at", Desc:true}}). If empty, defaults to PK DESC.
- DefaultLimit/MaxLimit: limit handling with clamping and non-positive defaulting.
  
Notes:
- Order keys must exactly match bun column names (case/spacing included).
- Disallowed or non-existent keys return an error.

## Ordering Rules
- The same ordering plan applies to both offset and cursor.
- PK direction follows the last effective key; if none, PK DESC.
- Composite PK: all PK columns appended as tiebreakers and included in the cursor.
  - When no user order is provided, all PK columns are appended with DESC.
 - OrderSpec sanitization: keys are trimmed and duplicate keys are de-duplicated (last occurrence wins); PK tiebreaker is always appended.

## Cursor Semantics
- Cursor is the last row's PK tuple from the previous page.
- Server fetches anchor row by PK, derives `(keys..., pk)` values, and builds a DB-agnostic OR-chain WHERE with exclusive boundary.

## Logging
- Backend: Go `log/slog` (TextHandler, stderr). `Options.LogLevel` controls minimum level.
- Warn when limit is clamped or non-positive is replaced by default.
- Error when disallowed or unknown order keys are provided.

Injecting a custom logger
```go
// Provide your own slog.Logger
lp := pager.New(nil)
lp.SetLogger(pager.NewSlogLoggerAdapter(slog.NewJSONHandler(os.Stdout, nil)))
```

Errors

| Code            | Meaning                                                                 |
|-----------------|-------------------------------------------------------------------------|
| INVALID_REQUEST | Bad inputs (both page+cursor, page<1, invalid cursor, bad order key, invalid destination, composite PK, etc.) |
| STALE_CURSOR    | Anchor row not found (e.g., deleted) — cursor no longer valid           |
| INTERNAL_ERROR  | Query execution failure or unexpected internal error                    |

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
