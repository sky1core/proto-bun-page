package pager

import (
    "context"
    "strings"
    "testing"

    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
)

type capturingLogger struct{ msgs []string }
func (l *capturingLogger) Debug(msg string, args ...interface{}) { l.msgs = append(l.msgs, msg) }
func (l *capturingLogger) Info(msg string, args ...interface{})  { l.msgs = append(l.msgs, msg) }
func (l *capturingLogger) Warn(msg string, args ...interface{})  { l.msgs = append(l.msgs, msg) }
func (l *capturingLogger) Error(msg string, args ...interface{}) { l.msgs = append(l.msgs, msg) }

func TestErrorOnDisallowedOrderKeys(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    lg := &capturingLogger{}
    pg := New(&Options{DefaultLimit: 2, MaxLimit: 10, LogLevel: "debug", AllowedOrderKeys: []string{"created_at"}})
    // inject logger
    pg.logger = lg

    ctx := context.Background()
    in := &pagerpb.Page{Limit: 2, Order: []*pagerpb.Order{{Key: "score", Desc: true}, {Key: "created_at", Desc: false}}}
    var rows []TestModel
    _, err := pg.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows)
    if err == nil { t.Fatal("expected error for disallowed order key") }
    // Optional: check message contains 'unsupported order key'
    if err != nil && !strings.Contains(err.Error(), "unsupported order key") {
        t.Fatalf("unexpected error: %v", err)
    }
}
