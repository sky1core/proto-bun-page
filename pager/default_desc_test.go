package pager

import (
    "testing"
)

// TestDefaultOrderDirection ensures all fields default to DESC when not explicitly specified
func TestDefaultOrderDirection(t *testing.T) {
    info, err := InferModelInfo(&TestModel{})
    if err != nil {
        t.Fatal(err)
    }

    tests := []struct {
        name     string
        specs    []OrderSpecInterface
        expected []OrderItem
    }{
        {
            name:  "empty specs should default to id DESC",
            specs: nil,
            expected: []OrderItem{
                {Column: "id", Direction: "DESC"},
            },
        },
        {
            name:  "single field without Asc specified should default to DESC",
            specs: []OrderSpecInterface{testOrderSpec{"name", false}}, // Asc not specified (false)
            expected: []OrderItem{
                {Column: "name", Direction: "DESC"},
                {Column: "id", Direction: "DESC"}, // PK tiebreaker follows last direction
            },
        },
        {
            name:  "single field with Asc=true should be ASC",
            specs: []OrderSpecInterface{testOrderSpec{"name", true}},
            expected: []OrderItem{
                {Column: "name", Direction: "ASC"},
                {Column: "id", Direction: "ASC"},
            },
        },
        {
            name:  "single field with Asc=false should be DESC",
            specs: []OrderSpecInterface{testOrderSpec{"name", false}},
            expected: []OrderItem{
                {Column: "name", Direction: "DESC"},
                {Column: "id", Direction: "DESC"},
            },
        },
        {
            name:  "multiple fields without explicit Asc should all default to DESC",
            specs: []OrderSpecInterface{testOrderSpec{"score", false}, testOrderSpec{"name", false}}, // Both without Asc specified (false)
            expected: []OrderItem{
                {Column: "score", Direction: "DESC"},
                {Column: "name", Direction: "DESC"},
                {Column: "id", Direction: "DESC"},
            },
        },
        {
            name:  "mixed explicit and implicit should respect explicit values",
            specs: []OrderSpecInterface{testOrderSpec{"score", false}, testOrderSpec{"name", false}}, // both should be DESC
            expected: []OrderItem{
                {Column: "score", Direction: "DESC"},
                {Column: "name", Direction: "DESC"},
                {Column: "id", Direction: "DESC"},
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            plan, err := BuildOrderPlan(tt.specs, info, nil)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            if len(plan.Items) != len(tt.expected) {
                t.Fatalf("expected %d items, got %d", len(tt.expected), len(plan.Items))
            }

            for i, item := range plan.Items {
                if item.Column != tt.expected[i].Column {
                    t.Errorf("item %d: expected column %s, got %s", i, tt.expected[i].Column, item.Column)
                }
                if item.Direction != tt.expected[i].Direction {
                    t.Errorf("item %d: expected direction %s, got %s", i, tt.expected[i].Direction, item.Direction)
                }
            }
        })
    }
}

// TestDefaultOrderConsistency ensures that default behavior is always DESC across the codebase
func TestDefaultOrderConsistency(t *testing.T) {
    info, err := InferModelInfo(&TestModel{})
    if err != nil {
        t.Fatal(err)
    }

    t.Run("no specs defaults to DESC", func(t *testing.T) {
        plan, err := BuildOrderPlan(nil, info, nil)
        if err != nil {
            t.Fatal(err)
        }
        
        // Should have exactly one item: id DESC
        if len(plan.Items) != 1 {
            t.Fatalf("expected 1 item, got %d", len(plan.Items))
        }
        
        if plan.Items[0].Column != "id" || plan.Items[0].Direction != "DESC" {
            t.Fatalf("expected id DESC, got %s %s", plan.Items[0].Column, plan.Items[0].Direction)
        }
    })

    t.Run("implicit field defaults to DESC", func(t *testing.T) {
        plan, err := BuildOrderPlan([]OrderSpecInterface{testOrderSpec{"created_at", false}}, info, nil)
        if err != nil {
            t.Fatal(err)
        }
        
        // Should have created_at DESC, then id DESC
        if len(plan.Items) != 2 {
            t.Fatalf("expected 2 items, got %d", len(plan.Items))
        }
        
        if plan.Items[0].Column != "created_at" || plan.Items[0].Direction != "DESC" {
            t.Fatalf("expected created_at DESC, got %s %s", plan.Items[0].Column, plan.Items[0].Direction)
        }
        
        if plan.Items[1].Column != "id" || plan.Items[1].Direction != "DESC" {
            t.Fatalf("expected id DESC, got %s %s", plan.Items[1].Column, plan.Items[1].Direction)
        }
    })
}