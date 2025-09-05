package pager

import "testing"

func TestBuildCursorWhere_SingleAsc_Full(t *testing.T) {
    plan := &OrderPlan{Items: []OrderItem{{Column: "created_at", Direction: "ASC"}, {Column: "id", Direction: "ASC"}}}
    cd := &CursorData{Values: map[string]interface{}{"created_at": 1000, "id": 5}}
    where, args, err := BuildCursorWhere(cd, plan)
    if err != nil { t.Fatal(err) }
    expected := "((created_at > ?) OR (created_at = ? AND id > ?))"
    if where != expected {
        t.Fatalf("expected %s, got %s", expected, where)
    }
    if len(args) != 3 { t.Fatalf("expected 3 args, got %d", len(args)) }
}

func TestBuildCursorWhere_SingleDesc_Full(t *testing.T) {
    plan := &OrderPlan{Items: []OrderItem{{Column: "created_at", Direction: "DESC"}, {Column: "id", Direction: "DESC"}}}
    cd := &CursorData{Values: map[string]interface{}{"created_at": 2000, "id": 7}}
    where, args, err := BuildCursorWhere(cd, plan)
    if err != nil { t.Fatal(err) }
    expected := "((created_at < ?) OR (created_at = ? AND id < ?))"
    if where != expected {
        t.Fatalf("expected %s, got %s", expected, where)
    }
    if len(args) != 3 { t.Fatalf("expected 3 args, got %d", len(args)) }
}

func TestBuildCursorWhere_Mixed_Full(t *testing.T) {
    plan := &OrderPlan{Items: []OrderItem{{Column: "score", Direction: "DESC"}, {Column: "name", Direction: "ASC"}, {Column: "id", Direction: "ASC"}}}
    cd := &CursorData{Values: map[string]interface{}{"score": 90, "name": "Bob", "id": 2}}
    where, args, err := BuildCursorWhere(cd, plan)
    if err != nil { t.Fatal(err) }
    expected := "((score < ?) OR (score = ? AND name > ?) OR (score = ? AND name = ? AND id > ?))"
    if where != expected {
        t.Fatalf("expected %s, got %s", expected, where)
    }
    if len(args) != 6 { t.Fatalf("expected 6 args, got %d", len(args)) }
}

