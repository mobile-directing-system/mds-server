package pagination

import (
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/lefinal/nulls"
)

// DefaultLimit is the default limit to use for example in QueryWithPagination
// when no other limit was specified.
var DefaultLimit = 20

// FieldMap is used for mapping fields to SQL selectors.
type FieldMap map[string]exp.Orderable

// OrderDir is the direction for ordering.
type OrderDir string

// Order directions.
const (
	OrderDirAsc  OrderDir = "asc"
	OrderDirDesc OrderDir = "desc"
)

// Paginated is a container for a paginated result for retrieved Entries.
type Paginated[T any] struct {
	// Total is the total amount of available entries.
	Total int `json:"total"`
	// Limit that was applied for retrieving the Entries.
	Limit int `json:"limit"`
	// Offset that was applied for retrieving the Entries.
	Offset int `json:"offset"`
	// OrderedBy is the optional field the entries are ordered by.
	OrderedBy nulls.String `json:"ordered_by"`
	// OrderDirection is the direction, the fields are ordered by, if OrderedBy is
	// set.
	OrderDirection OrderDir `json:"order_direction"`
	// Retrieves is the amount of entries in Entries.
	Retrieved int `json:"retrieved"`
	// Entries are the actually retrieved entries.
	Entries []T `json:"entries"`
}

// Params is a container for common ways of retrieving paginated results.
type Params struct {
	// Limit for the amount of retrieved results.
	Limit int `json:"limit"`
	// Offset is the offset for retrieving results with the set Limit.
	Offset int `json:"offset"`
	// OrderBy is an optional field to order results by.
	OrderBy nulls.String `json:"order_by"`
	// OrderDirection is the optional direction to use for ordering with OrderBy.
	OrderDirection OrderDir `json:"order_dir"`
}

// NewPaginated builds a Paginated from the given Params and entry
// list.
func NewPaginated[T any](params Params, entries []T, totalCount int) Paginated[T] {
	return Paginated[T]{
		Total:          totalCount,
		Limit:          params.Limit,
		Offset:         params.Offset,
		Retrieved:      len(entries),
		OrderedBy:      params.OrderBy,
		OrderDirection: params.OrderDirection,
		Entries:        entries,
	}
}

// MapPaginated maps Paginated from one entry type to another.
func MapPaginated[From any, To any](paginatedFrom Paginated[From], mapFn func(from From) To) Paginated[To] {
	mappedEntries := make([]To, 0, len(paginatedFrom.Entries))
	for _, from := range paginatedFrom.Entries {
		mappedEntries = append(mappedEntries, mapFn(from))
	}
	return Paginated[To]{
		Total:          paginatedFrom.Total,
		Limit:          paginatedFrom.Limit,
		Offset:         paginatedFrom.Offset,
		Retrieved:      paginatedFrom.Retrieved,
		OrderedBy:      paginatedFrom.OrderedBy,
		OrderDirection: paginatedFrom.OrderDirection,
		Entries:        mappedEntries,
	}
}
