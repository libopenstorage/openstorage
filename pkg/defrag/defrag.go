package defrag

import (

)

// Provider is a collection of APIs for performing different kinds of defrag
// operations on a node
type Provider interface {
}

// NewDefaultNodeDrainProvider does not any defrag related operations
func NewDefaultDefragProvider() Provider {
	return &UnsupportedDefragProvider{}
}

// UnsupportedDefragProvider unsupported implementation of defrag.
type UnsupportedDefragProvider struct {
}
