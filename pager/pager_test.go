package pager

import (
	"context"
	"database/sql"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
)

type TestModel struct {
	ID        int64  `bun:"id,pk,autoincrement"`
	Name      string `bun:"name"`
	CreatedAt int64  `bun:"created_at"`
	Score     int    `bun:"score"`
}

// Composite PK types are intentionally unsupported by the library.

func setupTestDB(t *testing.T) *bun.DB {
	sqlDB, err := sql.Open(sqliteshim.ShimName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	db := bun.NewDB(sqlDB, sqlitedialect.New())
	
	ctx := context.Background()
	_, err = db.NewCreateTable().Model((*TestModel)(nil)).Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	testData := []TestModel{
		{Name: "Alice", CreatedAt: 1000, Score: 90},
		{Name: "Bob", CreatedAt: 2000, Score: 85},
		{Name: "Charlie", CreatedAt: 3000, Score: 95},
		{Name: "David", CreatedAt: 4000, Score: 80},
		{Name: "Eve", CreatedAt: 5000, Score: 88},
	}

	_, err = db.NewInsert().Model(&testData).Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestBuildOrderPlan(t *testing.T) {
	model := &TestModel{}
	info, err := InferModelInfo(model)
	if err != nil {
		t.Fatal(err)
	}

    tests := []struct {
        name     string
        specs    []OrderSpecInterface
        expected []OrderItem
    }{
        {
            name:     "empty order defaults to id DESC",
            specs:    nil,
            expected: []OrderItem{{Column: "id", Direction: "DESC"}},
        },
        {
            name:     "single field ascending",
            specs:    []OrderSpecInterface{&pagerpb.Order{Key: "name", Asc: true}},
            expected: []OrderItem{
                {Column: "name", Direction: "ASC"},
                {Column: "id", Direction: "ASC"},
            },
        },
        {
            name:     "single field descending",
            specs:    []OrderSpecInterface{&pagerpb.Order{Key: "created_at", Asc: false}},
            expected: []OrderItem{
                {Column: "created_at", Direction: "DESC"},
                {Column: "id", Direction: "DESC"},
            },
        },
        {
            name:     "multiple fields mixed",
            specs:    []OrderSpecInterface{&pagerpb.Order{Key: "score", Asc: false}, &pagerpb.Order{Key: "name", Asc: true}},
            expected: []OrderItem{
                {Column: "score", Direction: "DESC"},
                {Column: "name", Direction: "ASC"},
                {Column: "id", Direction: "ASC"},
            },
        },
    }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
            plan, err := BuildOrderPlan(tt.specs, info, nil)
			if err != nil {
				t.Fatal(err)
			}

			if len(plan.Items) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(plan.Items))
			}

			for i, item := range plan.Items {
				if i >= len(tt.expected) {
					break
				}
				if item.Column != tt.expected[i].Column || item.Direction != tt.expected[i].Direction {
					t.Errorf("item %d: expected %+v, got %+v", i, tt.expected[i], item)
				}
			}
		})
	}
}

func TestInferModelInfo(t *testing.T) {
	model := &TestModel{}
	info, err := InferModelInfo(model)
	if err != nil {
		t.Fatal(err)
	}

	if len(info.PKColumns) != 1 || info.PKColumns[0] != "id" {
		t.Errorf("expected PK column 'id', got %v", info.PKColumns)
	}

	expectedMappings := map[string]string{
		"name":       "name",
		"created_at": "created_at",
		"score":      "score",
	}

	for key, expectedColumn := range expectedMappings {
		if column, ok := info.KeyToColumn[key]; !ok || column != expectedColumn {
			t.Errorf("expected key '%s' to map to '%s', got '%s'", key, expectedColumn, column)
		}
	}
}

// Ensure ORDER uses exact bun tag column names; no snake_case fallback.
func TestOrderByExactBunTagRequired(t *testing.T) {
    model := &TestModel{}
    info, err := InferModelInfo(model)
    if err != nil { t.Fatal(err) }

    // Exact bun tag works (default is DESC now)
    plan, err := BuildOrderPlan([]OrderSpecInterface{&pagerpb.Order{Key: "created_at", Asc: false}}, info, nil)
    if err != nil { t.Fatal(err) }
    if plan.Items[0].Column != "created_at" || plan.Items[0].Direction != "DESC" {
        t.Fatalf("expected created_at DESC, got %+v", plan.Items[0])
    }

    // Non-exact key should error
    if _, err := BuildOrderPlan([]OrderSpecInterface{&pagerpb.Order{Key: "CreatedAt", Asc: false}}, info, nil); err == nil {
        t.Fatal("expected error for non-exact bun column key")
    }

    // Explicit ASC works
    planAsc, err := BuildOrderPlan([]OrderSpecInterface{&pagerpb.Order{Key: "created_at", Asc: true}}, info, nil)
    if err != nil { t.Fatal(err) }
    if planAsc.Items[0].Column != "created_at" || planAsc.Items[0].Direction != "ASC" {
        t.Fatalf("expected created_at ASC, got %+v", planAsc.Items[0])
    }

    // PK has tag with options; should parse up to comma and work
    plan2, err := BuildOrderPlan([]OrderSpecInterface{&pagerpb.Order{Key: "id", Asc: false}}, info, nil)
    if err != nil { t.Fatal(err) }
    if plan2.Items[0].Column != "id" || plan2.Items[0].Direction != "DESC" {
        t.Fatalf("expected id DESC, got %+v", plan2.Items[0])
    }
}

// Model with a field that lacks bun tag; ordering by that name must error.
type noTagModel struct {
    ID    int64  `bun:"id,pk"`
    NoTag string // no bun tag
}

func TestOrderByFieldWithoutBunTagErrors(t *testing.T) {
    info, err := InferModelInfo(&noTagModel{})
    if err != nil { t.Fatal(err) }
    if _, err := BuildOrderPlan([]OrderSpecInterface{&pagerpb.Order{Key: "no_tag", Asc: false}}, info, nil); err == nil {
        t.Fatal("expected error for field without bun tag")
    }
}

func TestAllowedOrderKeysEnforced(t *testing.T) {
    model := &TestModel{}
    info, err := InferModelInfo(model)
    if err != nil { t.Fatal(err) }
    specs := []OrderSpecInterface{&pagerpb.Order{Key: "score", Asc: false}, &pagerpb.Order{Key: "created_at", Asc: false}}
    if _, err := BuildOrderPlan(specs, info, []string{"created_at"}); err == nil {
        t.Fatal("expected error for unsupported order key")
    }
}

// Composite PK is not supported: library assumes a single PK (default: id)
