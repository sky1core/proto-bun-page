//go:build !pbgen
// +build !pbgen

package pagerpb

type Order struct {
    Key  string `json:"key,omitempty"`
    Desc bool   `json:"desc,omitempty"`
}

type Page struct {
    Limit  uint32    `json:"limit,omitempty"`
    Order  []*Order  `json:"order,omitempty"`
    // Selector fields; since this file is used behind a build tag (!pbgen),
    // we expose presence bits to emulate proto3 oneof semantics.
    // Notes:
    // - In generated (.pb.go) builds, oneof wrappers/Get* methods indicate presence;
    //   this shim provides equivalent presence via PageSet/CursorSet when pbgen isn't used.
    // - Callers should set PageSet/CursorSet to true when intentionally specifying the selector,
    //   even if the value is the type's zero (e.g., CursorSet=true with Cursor="").
    Page      uint32 `json:"page,omitempty"`
    PageSet   bool   `json:"-"`
    Cursor    string `json:"cursor,omitempty"`
    CursorSet bool   `json:"-"`
}

// Presence helpers to mirror oneof behavior in the pbgen build.
func (p *Page) HasPage() bool {
    if p == nil { return false }
    if p.PageSet { return true }
    // Fallback: treat page>0 as explicitly set in legacy callers
    return p.Page > 0
}

func (p *Page) HasCursor() bool {
    if p == nil { return false }
    if p.CursorSet { return true }
    // Fallback: non-empty string means present in legacy callers
    return p.Cursor != ""
}
