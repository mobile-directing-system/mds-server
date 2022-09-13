package store

import (
	"github.com/doug-martin/goqu/v9"
)

// Mall provides all store access methods.
type Mall struct {
	dialect goqu.DialectWrapper
}

// NewMall creates a new Mall with postgres dialect.
func NewMall() *Mall {
	m := &Mall{
		dialect: goqu.Dialect("postgres"),
	}
	return m
}
