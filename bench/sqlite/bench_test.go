package sqlitebench

import (
    "context"
    "database/sql"
    "math/rand"
    "testing"
    "time"

    "github.com/uptrace/bun"
    "github.com/uptrace/bun/dialect/sqlitedialect"
    "github.com/uptrace/bun/driver/sqliteshim"

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

func openMemDB(b *testing.B) *bun.DB {
    sqldb, err := sql.Open(sqliteshim.ShimName, ":memory:")
    if err != nil {
        b.Fatal(err)
    }
    return bun.NewDB(sqldb, sqlitedialect.New())
}

func ensureSchemaAndSeed(b *testing.B, db *bun.DB, n int) {
    ctx := context.Background()
    // schema
    if _, err := db.NewCreateTable().IfNotExists().Model((*Item)(nil)).Exec(ctx); err != nil {
        b.Fatal(err)
    }
    // indexes
    must := func(err error) { if err != nil { b.Fatal(err) } }
    _, err := db.NewCreateIndex().IfNotExists().Model((*Item)(nil)).Index("idx_items_created_at_id").ColumnExpr("created_at DESC").ColumnExpr("id DESC").Exec(ctx); must(err)
    _, err = db.NewCreateIndex().IfNotExists().Model((*Item)(nil)).Index("idx_items_name_id").Column("name").Column("id").Exec(ctx); must(err)
    _, err = db.NewCreateIndex().IfNotExists().Model((*Item)(nil)).Index("idx_items_score_name_id").ColumnExpr("score DESC").Column("name").Column("id").Exec(ctx); must(err)
    _, err = db.NewCreateIndex().IfNotExists().Model((*Item)(nil)).Index("idx_items_status_created_at_id").Column("status").ColumnExpr("created_at DESC").Column("id").Exec(ctx); must(err)

    // count
    cnt, err := db.NewSelect().Model((*Item)(nil)).Count(ctx)
    if err != nil { b.Fatal(err) }
    if cnt >= n { return }

    // seed up to n
    r := rand.New(rand.NewSource(42))
    batch := 5000
    items := make([]Item, 0, batch)
    now := time.Now()
    for i := cnt; i < n; i++ {
        items = append(items, Item{
            CreatedAt: now.Add(-time.Duration(r.Intn(200000)) * time.Second),
            Name:      randomName(r, i+1),
            Score:     r.Intn(101),
            Status:    map[bool]int{true: 1, false: 0}[r.Intn(5) == 0],
        })
        if len(items) == batch || i == n-1 {
            if _, err := db.NewInsert().Model(&items).Exec(ctx); err != nil { b.Fatal(err) }
            items = items[:0]
        }
    }
}

func randomName(r *rand.Rand, i int) string {
    return "name_" + time.Unix(int64(i%1000000), 0).Format("150405")
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

// Benchmarks

func BenchmarkCursor_CreatedAtDesc_L20(b *testing.B) {
    db := openMemDB(b); defer db.Close()
    ensureSchemaAndSeed(b, db, 50000)
    p := newPager(); ctx := context.Background()
    in := &pagerpb.Page{Limit: 20, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var rows []Item
        q := db.NewSelect().Model(&Item{})
        if _, err := p.ApplyAndScan(ctx, q, in, &rows); err != nil { b.Fatal(err) }
    }
}

func BenchmarkOffset_CreatedAtDesc_Page1000_L20(b *testing.B) {
    db := openMemDB(b); defer db.Close()
    ensureSchemaAndSeed(b, db, 50000)
    p := newPager(); ctx := context.Background()
    in := &pagerpb.Page{Limit: 20, Page: 1000, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var rows []Item
        q := db.NewSelect().Model(&Item{})
        if _, err := p.ApplyAndScan(ctx, q, in, &rows); err != nil { b.Fatal(err) }
    }
}

func BenchmarkCursor_ScoreDesc_NameAsc(b *testing.B) {
    db := openMemDB(b); defer db.Close()
    ensureSchemaAndSeed(b, db, 50000)
    p := newPager(); ctx := context.Background()
    in := &pagerpb.Page{Limit: 20, Order: []*pagerpb.Order{{Key: "score", Desc: true}, {Key: "name", Desc: false}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var rows []Item
        q := db.NewSelect().Model(&Item{})
        if _, err := p.ApplyAndScan(ctx, q, in, &rows); err != nil { b.Fatal(err) }
    }
}

