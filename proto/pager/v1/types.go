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
    Page   uint32    `json:"page,omitempty"`
    Cursor string    `json:"cursor,omitempty"`
}

