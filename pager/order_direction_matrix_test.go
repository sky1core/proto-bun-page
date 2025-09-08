package pager

import (
    "testing"
    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
)

// Ensures each order spec's direction is applied independently per key.
func TestBuildOrderPlan_DirectionPerKey_Matrix(t *testing.T) {
    info, err := InferModelInfo(&TestModel{})
    if err != nil { t.Fatal(err) }

    cases := []struct{
        name string
        specs []OrderSpecInterface
        want  []OrderItem
    }{
        {
            name:  "one ASC",
            specs: []OrderSpecInterface{&pagerpb.Order{Key: "name", Asc: true}},
            want:  []OrderItem{{Column: "name", Direction: "ASC"}},
        },
        {
            name:  "one DESC",
            specs: []OrderSpecInterface{&pagerpb.Order{Key: "name", Asc: false}},
            want:  []OrderItem{{Column: "name", Direction: "DESC"}},
        },
        {
            name:  "two ASC,ASC",
            specs: []OrderSpecInterface{&pagerpb.Order{Key: "score", Asc: true}, &pagerpb.Order{Key: "name", Asc: true}},
            want:  []OrderItem{{Column: "score", Direction: "ASC"}, {Column: "name", Direction: "ASC"}},
        },
        {
            name:  "two ASC,DESC",
            specs: []OrderSpecInterface{&pagerpb.Order{Key: "score", Asc: true}, &pagerpb.Order{Key: "name", Asc: false}},
            want:  []OrderItem{{Column: "score", Direction: "ASC"}, {Column: "name", Direction: "DESC"}},
        },
        {
            name:  "two DESC,ASC",
            specs: []OrderSpecInterface{&pagerpb.Order{Key: "score", Asc: false}, &pagerpb.Order{Key: "name", Asc: true}},
            want:  []OrderItem{{Column: "score", Direction: "DESC"}, {Column: "name", Direction: "ASC"}},
        },
        {
            name:  "two DESC,DESC",
            specs: []OrderSpecInterface{&pagerpb.Order{Key: "score", Asc: false}, &pagerpb.Order{Key: "name", Asc: false}},
            want:  []OrderItem{{Column: "score", Direction: "DESC"}, {Column: "name", Direction: "DESC"}},
        },
        {
            name:  "dedupe last-wins preserves direction",
            specs: []OrderSpecInterface{&pagerpb.Order{Key: "name", Asc: true}, &pagerpb.Order{Key: "name", Asc: false}},
            want:  []OrderItem{{Column: "name", Direction: "DESC"}},
        },
        {
            name:  "empty key PK ASC",
            specs: []OrderSpecInterface{&pagerpb.Order{Key: "", Asc: true}},
            want:  []OrderItem{{Column: "id", Direction: "ASC"}},
        },
        {
            name:  "empty key PK DESC",
            specs: []OrderSpecInterface{&pagerpb.Order{Key: "", Asc: false}},
            want:  []OrderItem{{Column: "id", Direction: "DESC"}},
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            plan, err := BuildOrderPlan(tc.specs, info, nil)
            if err != nil { t.Fatal(err) }
            if len(plan.Items) < len(tc.want) {
                t.Fatalf("expected at least %d items, got %d", len(tc.want), len(plan.Items))
            }
            for i, w := range tc.want {
                if plan.Items[i].Column != w.Column || plan.Items[i].Direction != w.Direction {
                    t.Fatalf("item %d mismatch: want %+v, got %+v", i, w, plan.Items[i])
                }
            }
        })
    }
}

