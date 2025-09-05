package pager

import (
    "context"
    "testing"

    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
)

func TestPagerIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	pager := New(nil)

    t.Run("offset pagination - first page", func(t *testing.T) {
        var results []TestModel
        in := &pagerpb.Page{Page: 1, Limit: 2, Order: []*pagerpb.Order{{Key: "score", Desc: true}}}
        q := db.NewSelect().Model(&TestModel{})
        out, err := pager.ApplyAndScan(ctx, q, in, &results)
        if err != nil {
            t.Fatal(err)
        }

		if len(results) != 2 {
			t.Errorf("expected 2 items, got %d", len(results))
		}

        if out.Page != 1 {
            t.Error("expected page echo 1")
        }

		// Check order (highest scores first)
		if results[0].Score != 95 {
			t.Errorf("expected first item score 95, got %d", results[0].Score)
		}
	})

    t.Run("cursor pagination", func(t *testing.T) {
        var results []TestModel
        
        in := &pagerpb.Page{Limit: 2, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
        q := db.NewSelect().Model(&TestModel{})
        out, err := pager.ApplyAndScan(ctx, q, in, &results)
        if err != nil {
            t.Fatal(err)
        }

        t.Logf("First page results: %d items, nextCursor: %v", 
                len(results), out.Cursor)
		
        if out.Cursor == "" {
            t.Fatal("expected cursor to be set")
        }

        var nextResults []TestModel
        in2 := &pagerpb.Page{Cursor: out.Cursor, Limit: 2, Order: in.Order}
        q2 := db.NewSelect().Model(&TestModel{})
        _, err = pager.ApplyAndScan(ctx, q2, in2, &nextResults)
        if err != nil {
            t.Fatal(err)
        }

		if len(nextResults) != 2 {
			t.Errorf("expected 2 items, got %d", len(nextResults))
		}

		// Check that we got different items
		if nextResults[0].CreatedAt >= results[1].CreatedAt {
			t.Error("cursor pagination should return older items")
		}
	})

    t.Run("empty result", func(t *testing.T) {
        var results []TestModel
        in := &pagerpb.Page{Page: 100, Limit: 10}
        q := db.NewSelect().Model(&TestModel{})
        _, err := pager.ApplyAndScan(ctx, q, in, &results)
        if err != nil {
            t.Fatal(err)
        }

		if len(results) != 0 {
			t.Errorf("expected 0 items, got %d", len(results))
		}

    })

	t.Run("mixed order", func(t *testing.T) {
		var results []TestModel
        in := &pagerpb.Page{Limit: 3, Order: []*pagerpb.Order{{Key: "score", Desc: true}, {Key: "name", Desc: false}}}
        q := db.NewSelect().Model(&TestModel{})
        _, err := pager.ApplyAndScan(ctx, q, in, &results)
        if err != nil {
            t.Fatal(err)
        }

		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}

		// Verify order
		for i := 1; i < len(results); i++ {
			if results[i].Score > results[i-1].Score {
				t.Error("scores should be descending")
			}
			if results[i].Score == results[i-1].Score && results[i].Name < results[i-1].Name {
				t.Error("names should be ascending when scores are equal")
			}
		}
    })

    t.Run("stale cursor should error", func(t *testing.T) {
        var results []TestModel
        in := &pagerpb.Page{Limit: 2, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
        q := db.NewSelect().Model(&TestModel{})
        out, err := pager.ApplyAndScan(ctx, q, in, &results)
        if err != nil {
            t.Fatal(err)
        }
        if out.Cursor == "" {
            t.Fatal("expected next cursor")
        }
        info, err := InferModelInfo(&TestModel{})
        if err != nil { t.Fatal(err) }
        cd, err := DecodeCursor(out.Cursor, info)
        if err != nil {
            t.Fatal(err)
        }
        // extract id in a tolerant way
        var id int64
        switch v := cd.Values["id"].(type) {
        case int64:
            id = v
        case int:
            id = int64(v)
        case float64:
            id = int64(v)
        default:
            t.Fatalf("unexpected id type: %T", cd.Values["id"]) 
        }
        // delete anchor
        if _, err := db.NewDelete().Model((*TestModel)(nil)).Where("id = ?", id).Exec(ctx); err != nil {
            t.Fatal(err)
        }
        // use stale cursor
        var next []TestModel
        in2 := &pagerpb.Page{Cursor: out.Cursor, Limit: 2, Order: in.Order}
        q2 := db.NewSelect().Model(&TestModel{})
        _, err = pager.ApplyAndScan(ctx, q2, in2, &next)
        if err == nil {
            t.Fatal("expected error for stale cursor")
        }
    })

    t.Run("limit boundaries: default and clamp", func(t *testing.T) {
        // default limit on non-positive (proto omits limit -> default)
        pg := New(&Options{DefaultLimit: 2, MaxLimit: 3, LogLevel: "error"})
        var results []TestModel
        in := &pagerpb.Page{Limit: 0}
        q := db.NewSelect().Model(&TestModel{})
        _, err := pg.ApplyAndScan(ctx, q, in, &results)
        if err != nil { t.Fatal(err) }
        if len(results) != 2 { t.Fatalf("expected default 2, got %d", len(results)) }

        // clamp to max
        results = nil
        in = &pagerpb.Page{Limit: 100}
        q = db.NewSelect().Model(&TestModel{})
        _, err = pg.ApplyAndScan(ctx, q, in, &results)
        if err != nil { t.Fatal(err) }
        if len(results) != 3 { t.Fatalf("expected clamped 3, got %d", len(results)) }
    })

    t.Run("exclusive boundary on ties", func(t *testing.T) {
        // Insert two rows with the same created_at as an existing middle value
        // Choose created_at=3000 (Charlie) as a tie target
        _, err := db.NewInsert().Model(&[]TestModel{{Name: "Tie1", CreatedAt: 3000, Score: 50}, {Name: "Tie2", CreatedAt: 3000, Score: 51}}).Exec(ctx)
        if err != nil { t.Fatal(err) }

        var first []TestModel
        in := &pagerpb.Page{Limit: 3, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}}
        q := db.NewSelect().Model(&TestModel{})
        out, err := pager.ApplyAndScan(ctx, q, in, &first)
        if err != nil { t.Fatal(err) }
        if out.Cursor == "" { t.Fatal("need next cursor") }
        lastFirst := first[len(first)-1]

        var second []TestModel
        q2 := db.NewSelect().Model(&TestModel{})
        _, err = pager.ApplyAndScan(ctx, q2, &pagerpb.Page{Cursor: out.Cursor, Limit: 3, Order: in.Order}, &second)
        if err != nil { t.Fatal(err) }
        // Ensure last item of first page is not repeated in second page
        for _, r := range second {
            if r.ID == lastFirst.ID {
                t.Fatalf("exclusive boundary violated: repeated id %d on second page", r.ID)
            }
        }
    })

    t.Run("full cursor scan without duplicates", func(t *testing.T) {
        // Count total rows
        total, err := db.NewSelect().Model((*TestModel)(nil)).Count(ctx)
        if err != nil { t.Fatal(err) }
        seen := make(map[int64]bool)
        var cursor string
        fetched := 0
        for {
            var batch []TestModel
            in := &pagerpb.Page{Limit: 3, Order: []*pagerpb.Order{{Key: "created_at", Desc: true}}, Cursor: cursor}
            q := db.NewSelect().Model(&TestModel{})
            out, err := pager.ApplyAndScan(ctx, q, in, &batch)
            if err != nil { t.Fatal(err) }
            for _, r := range batch {
                if seen[r.ID] { t.Fatalf("duplicate id encountered: %d", r.ID) }
                seen[r.ID] = true
                fetched++
            }
            if out.Cursor == "" {
                break
            }
            cursor = out.Cursor
        }
        if fetched != total {
            t.Fatalf("scanned %d rows but total is %d", fetched, total)
        }
    })

    t.Run("invalid cursor token returns error", func(t *testing.T) {
        var results []TestModel
        in := &pagerpb.Page{Limit: 2, Cursor: "!!!not-base64"}
        q := db.NewSelect().Model(&TestModel{})
        if _, err := pager.ApplyAndScan(ctx, q, in, &results); err == nil {
            t.Fatal("expected invalid cursor error")
        }
    })

    t.Run("default order applied when order empty", func(t *testing.T) {
        // Pager with DefaultOrder = -created_at
        pg := New(&Options{DefaultLimit: 2, MaxLimit: 10, DefaultOrderSpecs: []OrderSpec{{Key: "created_at", Desc: true}}, LogLevel: "error"})
        var rows []TestModel
        in := &pagerpb.Page{Limit: 2}
        q := db.NewSelect().Model(&TestModel{})
        _, err := pg.ApplyAndScan(ctx, q, in, &rows)
        if err != nil { t.Fatal(err) }
        if len(rows) != 2 { t.Fatalf("expected 2 rows, got %d", len(rows)) }
        // created_at should be descending; first row has the largest CreatedAt
        if rows[0].CreatedAt < rows[1].CreatedAt {
            t.Fatal("expected default order -created_at to apply")
        }
    })

    t.Run("order requires exact bun column key", func(t *testing.T) {
        pg := New(&Options{DefaultLimit: 3, MaxLimit: 10, AllowedOrderKeys: []string{"name"}, LogLevel: "error"})
        var rows []TestModel
        in := &pagerpb.Page{Limit: 3, Order: []*pagerpb.Order{{Key: "Name", Desc: false}}}
        q := db.NewSelect().Model(&TestModel{})
        if _, err := pg.ApplyAndScan(ctx, q, in, &rows); err == nil {
            t.Fatal("expected error for non-exact order key 'Name'")
        }
    })
}

func TestCursorWhere(t *testing.T) {
	model := &TestModel{}
	_, err := InferModelInfo(model)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("single column ASC", func(t *testing.T) {
		plan := &OrderPlan{
			Items: []OrderItem{
				{Column: "created_at", Direction: "ASC"},
				{Column: "id", Direction: "ASC"},
			},
		}

		cursorData := &CursorData{
			Values: map[string]interface{}{
				"created_at": 1000,
				"id":         5,
			},
		}

		where, args, err := BuildCursorWhere(cursorData, plan)
		if err != nil {
			t.Fatal(err)
		}

		expected := "((created_at > ?) OR (created_at = ? AND id > ?))"
		if where != expected {
			t.Errorf("expected WHERE %s, got %s", expected, where)
		}

		if len(args) != 3 {
			t.Errorf("expected 3 args, got %d", len(args))
		}
	})

	t.Run("multiple columns mixed", func(t *testing.T) {
		plan := &OrderPlan{
			Items: []OrderItem{
				{Column: "score", Direction: "DESC"},
				{Column: "name", Direction: "ASC"},
				{Column: "id", Direction: "ASC"},
			},
		}

		cursorData := &CursorData{
			Values: map[string]interface{}{
				"score": 90,
				"name":  "Bob",
				"id":    2,
			},
		}

		where, args, err := BuildCursorWhere(cursorData, plan)
		if err != nil {
			t.Fatal(err)
		}

		// Should build: (score < ?) OR (score = ? AND name > ?) OR (score = ? AND name = ? AND id > ?)
		if !contains(where, "score < ?") {
			t.Error("expected 'score < ?' in WHERE clause")
		}
		if !contains(where, "score = ? AND name > ?") {
			t.Error("expected 'score = ? AND name > ?' in WHERE clause")
		}

		if len(args) != 6 {
			t.Errorf("expected 6 args, got %d", len(args))
		}
	})
}

func intPtr(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
