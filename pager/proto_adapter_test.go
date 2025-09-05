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

    // both page and cursor -> error
    inErr := &pagerpb.Page{Limit: 2, Page: 1, Cursor: "nonempty"}
    var rows []TestModel
    if _, err := pg.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), inErr, &rows); err == nil {
        t.Fatal("expected error when both page and cursor are set")
    }

    // page echo, no cursor
    inPage := &pagerpb.Page{Limit: 2, Page: 1}
    rows = nil
    outPage, err := pg.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), inPage, &rows)
    if err != nil { t.Fatal(err) }
    if outPage.Page != 1 { t.Fatalf("expected page echo=1, got %d", outPage.Page) }
    if outPage.Cursor != "" && len(rows) == 2 { /* may have next cursor depending on data; allow empty here */ }

    // cursor mode first page (empty cursor) should return a next cursor
    inCur := &pagerpb.Page{Limit: 2, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
    rows = nil
    outCur, err := pg.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), inCur, &rows)
    if err != nil { t.Fatal(err) }
    if outCur.Cursor == "" && len(rows) > 0 {
        t.Fatal("expected non-empty cursor when rows returned in cursor mode")
    }
}
