package store

import (
	"github.com/doug-martin/goqu/v9"
)

// Mall provides all store access methods.
type Mall struct {
	dialect goqu.DialectWrapper
}

// NewMall creates a new Mall.
func NewMall() *Mall {
	return &Mall{
		dialect: goqu.Dialect("postgres"),
	}
}
