package pager

import (
    "testing"
)

func TestBuildOrderPlan_SanitizeAndDedupe(t *testing.T) {
    info, err := InferModelInfo(&TestModel{})
    if err != nil { t.Fatal(err) }

    // Duplicate key with different directions; last wins
    specs := []OrderSpecInterface{testOrderSpec{"name", false}, testOrderSpec{" score ", false}, testOrderSpec{"name", false}}
    plan, err := BuildOrderPlan(specs, info, nil)
    if err != nil { t.Fatal(err) }

    if len(plan.Items) < 3 { // score, name, pk
        t.Fatalf("expected at least 3 items (score,name,pk), got %d", len(plan.Items))
    }
    // Expect score DESC first, then name DESC (last occurrence kept)
    if plan.Items[0].Column != "score" || plan.Items[0].Direction != "DESC" {
        t.Fatalf("expected first item to be score DESC, got %+v", plan.Items[0])
    }
    if plan.Items[1].Column != "name" || plan.Items[1].Direction != "DESC" {
        t.Fatalf("expected second item to be name DESC after dedupe, got %+v", plan.Items[1])
    }
    // PK tiebreaker appended; direction follows last user key
    if plan.Items[len(plan.Items)-1].Column != "id" || plan.Items[len(plan.Items)-1].Direction != plan.Items[0].Direction {
        t.Fatalf("expected pk appended with same direction as last user key")
    }
}
