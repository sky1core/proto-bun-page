package pager

import (
    "context"
    "testing"

    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
)

func TestApplyAndScan_DestinationValidation(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    p := New(nil)
    ctx := context.Background()

    // Non-pointer slice -> invalid
    {
        var rows []TestModel
        // Pass non-pointer intentionally
        if _, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), &pagerpb.Page{}, rows); err == nil {
            t.Fatal("expected error for non-pointer destination")
        }
    }

    // Nil destination -> invalid
    {
        var rows *[]TestModel = nil
        if _, err := p.ApplyAndScan(ctx, db.NewSelect().Model(&TestModel{}), &pagerpb.Page{}, rows); err == nil {
            t.Fatal("expected error for nil destination pointer")
        }
    }
}

