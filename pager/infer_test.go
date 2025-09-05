package pager

import "testing"

type cm struct {
    K1 int64 `bun:"k1,pk"`
    K2 int64 `bun:"k2,pk"`
}

func TestInferModelInfo_CompositePKNotSupported(t *testing.T) {
    _, err := InferModelInfo(&cm{})
    if err == nil {
        t.Fatal("expected error for composite primary key")
    }
}

