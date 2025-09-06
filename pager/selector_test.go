package pager

import (
    "context"
    "testing"

    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
)

// Covers: page/cursor presence semantics
func TestSelectorSemantics(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    ctx := context.Background()
    p := New(&Options{DefaultLimit: 2, MaxLimit: 10, LogLevel: "error", DefaultOrderSpecs: []OrderSpec{{Key: "created_at", Desc: true}}})

    // 1) Neither page nor cursor set -> defaults to cursor mode (from start)
    {
        in := &pagerpb.Page{Limit: 2}
        var rows []TestModel
        out, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows)
        if err != nil { t.Fatal(err) }
        if len(rows) != 2 { t.Fatalf("expected 2 rows, got %d", len(rows)) }
        if out.Cursor == "" { t.Fatal("expected next cursor when defaulting to cursor mode") }
    }

    // 2) Page explicitly set to 0 -> invalid
    {
        in := &pagerpb.Page{Limit: 2, Page: 0}
        in.PageSet = true
        var rows []TestModel
        if _, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows); err == nil {
            t.Fatal("expected error for page < 1 when page explicitly set")
        }
    }

    // 3) Page explicitly set to 1 -> offset mode, no cursor in response
    {
        in := &pagerpb.Page{Limit: 2, Page: 1}
        in.PageSet = true
        var rows []TestModel
        out, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows)
        if err != nil { t.Fatal(err) }
        if len(rows) != 2 { t.Fatalf("expected 2 rows, got %d", len(rows)) }
        if out.Cursor != "" { t.Fatal("did not expect cursor in offset mode") }
        if out.Page != 1 { t.Fatal("expected echo page=1") }
    }

    // 4) Cursor explicitly set to empty -> cursor mode from start
    {
        in := &pagerpb.Page{Limit: 2, Cursor: ""}
        in.CursorSet = true
        var rows []TestModel
        out, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows)
        if err != nil { t.Fatal(err) }
        if len(rows) != 2 { t.Fatalf("expected 2 rows, got %d", len(rows)) }
        if out.Cursor == "" { t.Fatal("expected next cursor in cursor mode (explicit empty)") }
    }

    // 5) Both specified -> error
    {
        in := &pagerpb.Page{Limit: 2, Page: 1, Cursor: "token"}
        in.PageSet, in.CursorSet = true, true
        var rows []TestModel
        if _, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows); err == nil {
            t.Fatal("expected error when both page and cursor specified")
        }
    }
}

