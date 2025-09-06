package bench

import (
    "context"
    "database/sql"
    "os"
    "testing"
    "time"

    "github.com/uptrace/bun"
    "github.com/uptrace/bun/dialect/mysqldialect"
    _ "github.com/go-sql-driver/mysql"

    pager "github.com/sky1core/proto-bun-page/pager"
    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
)

type Item struct {
    ID        int64     `bun:"id,pk,autoincrement"`
    CreatedAt time.Time `bun:"created_at"`
    Name      string    `bun:"name"`
    Score     int       `bun:"score"`
    Status    int       `bun:"status"`
    Payload   *string   `bun:"payload"`
}

func openDB(b *testing.B) *bun.DB {
    dsn := os.Getenv("DSN")
    if dsn == "" {
        b.Fatal("DSN env not set")
    }
    sqldb, err := sql.Open("mysql", dsn)
    if err != nil {
        b.Fatal(err)
    }
    return bun.NewDB(sqldb, mysqldialect.New())
}

func newPager() *pager.Pager {
    return pager.New(&pager.Options{
        DefaultLimit:      20,
        MaxLimit:          100,
        LogLevel:          "error",
        AllowedOrderKeys:  []string{"created_at", "name", "score", "id"},
        DefaultOrderSpecs: []pager.OrderSpec{{Key: "created_at", Desc: true}},
    })
}

func benchCursorCreatedAtDesc(b *testing.B, limit int) {
    db := openDB(b)
    defer db.Close()
    p := newPager()
    ctx := context.Background()
    in := &pagerpb.Page{Limit: uint32(limit), Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var rows []Item
        q := db.NewSelect().Model(&Item{})
        if _, err := p.ApplyAndScan(ctx, q, in, &rows); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCursor_CreatedAtDesc_L20(b *testing.B)  { benchCursorCreatedAtDesc(b, 20) }
func BenchmarkCursor_CreatedAtDesc_L100(b *testing.B) { benchCursorCreatedAtDesc(b, 100) }

func BenchmarkCursor_ScoreDesc_NameAsc(b *testing.B) {
    db := openDB(b); defer db.Close()
    p := newPager(); ctx := context.Background()
    in := &pagerpb.Page{Limit: 20, Order: []*pagerpb.Order{{Key: "score", Desc: true}, {Key: "name", Desc: false}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var rows []Item
        q := db.NewSelect().Model(&Item{})
        if _, err := p.ApplyAndScan(ctx, q, in, &rows); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCursor_Filter_Status_CreatedAt(b *testing.B) {
    db := openDB(b); defer db.Close()
    p := newPager(); ctx := context.Background()
    in := &pagerpb.Page{Limit: 20, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var rows []Item
        q := db.NewSelect().Model(&Item{}).Where("status = ?", 1)
        if _, err := p.ApplyAndScan(ctx, q, in, &rows); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkOffset_CreatedAtDesc_Page1000_L20(b *testing.B) {
    db := openDB(b); defer db.Close()
    p := newPager(); ctx := context.Background()
    in := &pagerpb.Page{Limit: 20, Page: 1000, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var rows []Item
        q := db.NewSelect().Model(&Item{})
        if _, err := p.ApplyAndScan(ctx, q, in, &rows); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCursor_Covering(b *testing.B) {
    db := openDB(b); defer db.Close()
    p := newPager(); ctx := context.Background()
    in := &pagerpb.Page{Limit: 20, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var rows []Item
        q := db.NewSelect().Model(&Item{}).Column("id", "created_at", "name", "score")
        if _, err := p.ApplyAndScan(ctx, q, in, &rows); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCursor_NonCovering(b *testing.B) {
    db := openDB(b); defer db.Close()
    p := newPager(); ctx := context.Background()
    in := &pagerpb.Page{Limit: 20, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var rows []Item
        q := db.NewSelect().Model(&Item{}) // includes payload by default
        if _, err := p.ApplyAndScan(ctx, q, in, &rows); err != nil {
            b.Fatal(err)
        }
    }
}

