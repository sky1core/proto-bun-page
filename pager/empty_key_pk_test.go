package pager

import (
    "testing"
    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
)

// Empty key should default to PK, and Asc flag must control its direction.
func TestBuildOrderPlan_EmptyKeyDefaultsToPK_WithAsc(t *testing.T) {
    info, err := InferModelInfo(&TestModel{})
    if err != nil { t.Fatal(err) }

    // Asc=true -> id ASC
    planAsc, err := BuildOrderPlan([]OrderSpecInterface{&pagerpb.Order{Key: "", Asc: true}}, info, nil)
    if err != nil { t.Fatal(err) }
    if len(planAsc.Items) != 1 { t.Fatalf("expected 1 item, got %d", len(planAsc.Items)) }
    if planAsc.Items[0].Column != "id" || planAsc.Items[0].Direction != "ASC" {
        t.Fatalf("expected id ASC, got %+v", planAsc.Items[0])
    }

    // Asc=false -> id DESC
    planDesc, err := BuildOrderPlan([]OrderSpecInterface{&pagerpb.Order{Key: "", Asc: false}}, info, nil)
    if err != nil { t.Fatal(err) }
    if len(planDesc.Items) != 1 { t.Fatalf("expected 1 item, got %d", len(planDesc.Items)) }
    if planDesc.Items[0].Column != "id" || planDesc.Items[0].Direction != "DESC" {
        t.Fatalf("expected id DESC, got %+v", planDesc.Items[0])
    }

    // AllowedOrderKeys should not block explicit PK via empty key
    planAllowed, err := BuildOrderPlan([]OrderSpecInterface{&pagerpb.Order{Key: "", Asc: true}}, info, []string{"name"})
    if err != nil { t.Fatalf("unexpected error with empty key and AllowedOrderKeys: %v", err) }
    if planAllowed.Items[0].Column != "id" || planAllowed.Items[0].Direction != "ASC" {
        t.Fatalf("expected id ASC with allowed-keys present, got %+v", planAllowed.Items[0])
    }
}

