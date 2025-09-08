package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "time"

    "github.com/sky1core/proto-bun-page/pager"
    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
    "github.com/uptrace/bun"
    "github.com/uptrace/bun/dialect/sqlitedialect"
    "github.com/uptrace/bun/driver/sqliteshim"
)

// Helper that implements OrderSpecInterface
type orderSpec struct {
    key string
    asc bool
}

func (o orderSpec) GetKey() string { return o.key }
func (o orderSpec) GetAsc() bool { return o.asc }

type Product struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name"`
	Price     float64   `bun:"price"`
	CreatedAt time.Time `bun:"created_at"`
}

func main() {
	// Setup database
	sqlDB, err := sql.Open(sqliteshim.ShimName, ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	db := bun.NewDB(sqlDB, sqlitedialect.New())
	ctx := context.Background()

	// Create table
	_, err = db.NewCreateTable().Model((*Product)(nil)).Exec(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Insert sample data
	products := []Product{
		{Name: "Laptop", Price: 999.99, CreatedAt: time.Now().Add(-5 * time.Hour)},
		{Name: "Mouse", Price: 29.99, CreatedAt: time.Now().Add(-4 * time.Hour)},
		{Name: "Keyboard", Price: 79.99, CreatedAt: time.Now().Add(-3 * time.Hour)},
		{Name: "Monitor", Price: 299.99, CreatedAt: time.Now().Add(-2 * time.Hour)},
		{Name: "Headphones", Price: 199.99, CreatedAt: time.Now().Add(-1 * time.Hour)},
	}
	_, err = db.NewInsert().Model(&products).Exec(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize pager
    pg := pager.New(&pager.Options{
        DefaultLimit: 2,
        MaxLimit:     10,
        LogLevel:     "info",
        AllowedOrderKeys:  []string{"created_at", "price", "name", "id"},
        DefaultOrderSpecs: []pager.OrderSpecInterface{orderSpec{"created_at", false}},
    })

    fmt.Println("=== Proto Adapter Example ===")
    demonstrateProtoAdapter(ctx, db, pg)
}

func demonstrateProtoAdapter(ctx context.Context, db *bun.DB, pg *pager.Pager) {
    in := &pagerpb.Page{
        Limit: 2,
        Order: []*pagerpb.Order{{Key: "created_at", Asc: false}},
        Selector: &pagerpb.Page_Cursor{Cursor: ""},
    }
    var products []Product
    q := db.NewSelect().Model(&Product{})
    out, err := pg.ApplyAndScan(ctx, q, in, &products)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Proto first batch:")
    for _, p := range products {
        fmt.Printf("  - %s (created: %s)\n", p.Name, p.CreatedAt.Format("15:04:05"))
    }
    if cursor, ok := out.Selector.(*pagerpb.Page_Cursor); ok && cursor.Cursor != "" {
        in2 := &pagerpb.Page{
            Limit: 2, 
            Order: in.Order, 
            Selector: &pagerpb.Page_Cursor{Cursor: cursor.Cursor},
        }
        var next []Product
        q2 := db.NewSelect().Model(&Product{})
        out2, err := pg.ApplyAndScan(ctx, q2, in2, &next)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Println("Proto next batch:")
        for _, p := range next {
            fmt.Printf("  - %s (created: %s)\n", p.Name, p.CreatedAt.Format("15:04:05"))
        }
        _ = out2
    }
}
