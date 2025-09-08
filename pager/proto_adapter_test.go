package pager

import (
    "context"
    "testing"

    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
)

func TestProtoAdapter_BasicFlows(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    ctx := context.Background()
    pg := New(&Options{DefaultLimit: 2, MaxLimit: 10, LogLevel: "error"})

    // oneof prevents both page and cursor from being set simultaneously
    var rows []TestModel

    // page echo, no cursor
    inPage := &pagerpb.Page{
        Limit: 2,
        Selector: &pagerpb.Page_Page{Page: 1},
    }
    rows = nil
    outPage, err := pg.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), inPage, &rows)
    if err != nil { t.Fatal(err) }
    if page, ok := outPage.Selector.(*pagerpb.Page_Page); !ok || page.Page != 1 {
        t.Fatalf("expected page echo=1")
    }

    // cursor mode first page (empty cursor) should return a next cursor
    inCur := &pagerpb.Page{Limit: 2, Order: []*pagerpb.Order{{Key: "created_at", Asc: false}}}
    rows = nil
    outCur, err := pg.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), inCur, &rows)
    if err != nil { t.Fatal(err) }
    if cursor, ok := outCur.Selector.(*pagerpb.Page_Cursor); !ok || cursor.Cursor == "" && len(rows) > 0 {
        t.Fatal("expected non-empty cursor when rows returned in cursor mode")
    }
}
