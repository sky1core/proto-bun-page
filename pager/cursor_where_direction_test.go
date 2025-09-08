package pager

import (
    "testing"
)

// Validate inequality operator per key is chosen independently per direction.
func TestBuildCursorWhere_MixedDirections_Inequalities(t *testing.T) {
    plan := &OrderPlan{Items: []OrderItem{{Column: "score", Direction: "ASC"}, {Column: "name", Direction: "DESC"}, {Column: "id", Direction: "ASC"}}}
    cd := &CursorData{Values: map[string]interface{}{"score": 90, "name": "Bob", "id": 2}}
    where, args, err := BuildCursorWhere(cd, plan)
    if err != nil { t.Fatal(err) }
    // Expect: (score > ?) OR (score = ? AND name < ?) OR (score = ? AND name = ? AND id > ?)
    want := "((score > ?) OR (score = ? AND name < ?) OR (score = ? AND name = ? AND id > ?))"
    if where != want {
        t.Fatalf("unexpected WHERE.\nwant: %s\n got: %s", want, where)
    }
    if len(args) != 6 { t.Fatalf("expected 6 args, got %d", len(args)) }
}

