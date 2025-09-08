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
    p := New(&Options{DefaultLimit: 2, MaxLimit: 10, LogLevel: "error", DefaultOrderSpecs: []OrderSpec{{Key: "created_at", Asc: false}}})

    // 1) Neither page nor cursor set -> defaults to cursor mode (from start)
    {
        in := &pagerpb.Page{Limit: 2}
        var rows []TestModel
        out, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows)
        if err != nil { t.Fatal(err) }
        if len(rows) != 2 { t.Fatalf("expected 2 rows, got %d", len(rows)) }
        if cursor, ok := out.Selector.(*pagerpb.Page_Cursor); !ok || cursor.Cursor == "" {
            t.Fatal("expected next cursor when defaulting to cursor mode")
        }
    }

    // 2) Page explicitly set to 0 -> invalid
    {
        in := &pagerpb.Page{
            Limit: 2,
            Selector: &pagerpb.Page_Page{Page: 0},
        }
        var rows []TestModel
        if _, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows); err == nil {
            t.Fatal("expected error for page < 1 when page explicitly set")
        }
    }

    // 3) Page explicitly set to 1 -> offset mode, no cursor in response
    {
        in := &pagerpb.Page{
            Limit: 2,
            Selector: &pagerpb.Page_Page{Page: 1},
        }
        var rows []TestModel
        out, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows)
        if err != nil { t.Fatal(err) }
        if len(rows) != 2 { t.Fatalf("expected 2 rows, got %d", len(rows)) }
        if page, ok := out.Selector.(*pagerpb.Page_Page); !ok || page.Page != 1 {
            t.Fatal("expected page=1 in offset mode")
        }
    }

    // 4) Cursor explicitly set to empty -> cursor mode from start
    {
        in := &pagerpb.Page{
            Limit: 2,
            Selector: &pagerpb.Page_Cursor{Cursor: ""},
        }
        var rows []TestModel
        out, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), in, &rows)
        if err != nil { t.Fatal(err) }
        if len(rows) != 2 { t.Fatalf("expected 2 rows, got %d", len(rows)) }
        if cursor, ok := out.Selector.(*pagerpb.Page_Cursor); !ok || cursor.Cursor == "" {
            t.Fatal("expected next cursor in cursor mode (explicit empty)")
        }
    }

    // 5) oneof prevents both from being set simultaneously (this test is not needed with proper oneof)
}

