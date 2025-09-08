package pager

// Test helper that implements OrderSpecInterface for testing
type testOrderSpec struct {
    key string
    asc bool
}

func (o testOrderSpec) GetKey() string { return o.key }
func (o testOrderSpec) GetAsc() bool { return o.asc }