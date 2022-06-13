package pagination

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/lefinal/meh"
)

// TotalCountColumn is the column name in queries for retrieving the total count.
var TotalCountColumn = "__total_count"

// QueryToSQLWithPagination alters the query builder using QueryWithPagination
// and then builds SQL from the query. In case of failure, an meh.ErrInternal
// will be returned including the Params.
func QueryToSQLWithPagination(qb *goqu.SelectDataset, paginationParams Params, fieldMap FieldMap) (string, any, error) {
	qb, err := QueryWithPagination(qb, paginationParams, fieldMap)
	if err != nil {
		return "", nil, meh.Wrap(err, "query with pagination", meh.Details{"params": paginationParams})
	}
	q, params, err := qb.ToSQL()
	if err != nil {
		return "", nil, meh.NewInternalErrFromErr(err, "query to sql", meh.Details{"params": paginationParams})
	}
	return q, params, nil
}

// QueryWithPagination uses the given Params to alter the
// goqu.SelectDataset. A select-field with TotalCountColumn is added that allows
// reading the total available amount of entries. Based on the Params,
// offset and limit as well as order-by-clauses are added using the
// PaginationParams.FieldMap.
func QueryWithPagination(qb *goqu.SelectDataset, params Params, fieldMap FieldMap) (*goqu.SelectDataset, error) {
	var err error
	// Add total count.
	qb = addTotalCountSelectToQuery(qb)
	// Add limit.
	if params.Limit.Valid {
		qb, err = addLimitToQuery(qb, params.Limit.Int)
		if err != nil {
			return nil, meh.Wrap(err, "add limit to query", meh.Details{"limit": params.Limit.Int})
		}
	}
	// Add offset.
	qb, err = addOffsetToQuery(qb, params.Offset)
	if err != nil {
		return nil, meh.Wrap(err, "add offset to query", meh.Details{"offset": params.Offset})
	}
	// Add ordering.
	if params.OrderBy.Valid {
		qb, err = addOrderingToQuery(qb, params.OrderBy.String, params.OrderDirection, fieldMap)
		if err != nil {
			return nil, meh.Wrap(err, "add ordering to query", meh.Details{
				"order_by":  params.OrderBy.String,
				"order_dir": params.OrderDirection,
				"field_map": fieldMap,
			})
		}
	}
	return qb, nil
}

// addTotalCountSelectToQuery appends a selection for total count of available
// entries despite offset or limit. The column will have the name specified in
// TotalCountColumn.
func addTotalCountSelectToQuery(qb *goqu.SelectDataset) *goqu.SelectDataset {
	return qb.SelectAppend(goqu.L("COUNT(*) OVER()").As(TotalCountColumn))
}

// addLimitToQuery adds a limit-clause to the given goqu.SelectDataset if
// needed.
func addLimitToQuery(qb *goqu.SelectDataset, limit int) (*goqu.SelectDataset, error) {
	if limit < 0 {
		return nil, meh.NewBadInputErr("limit must not be negative", nil)
	}
	if limit == 0 {
		limit = DefaultLimit
	}
	return qb.Limit(uint(limit)), nil
}

// addOffsetToQuery adds an offset-clause to the given goqu.SelectDataset if
// needed.
func addOffsetToQuery(qb *goqu.SelectDataset, offset int) (*goqu.SelectDataset, error) {
	if offset < 0 {
		return nil, meh.NewBadInputErr("offset must not be negative", nil)
	}
	if offset == 0 {
		return qb, nil
	}
	return qb.Offset(uint(offset)), nil
}

// addOrderingToQuery prepens the given goqu.SelectDataset with an
// order-by-clause based on the given parameters.
func addOrderingToQuery(qb *goqu.SelectDataset, orderBy string, orderDirection string, fieldMap FieldMap) (*goqu.SelectDataset, error) {
	// Retrieve selector from field map.
	selector, ok := fieldMap[orderBy]
	if !ok {
		return nil, meh.NewBadInputErr("order field not found", nil)
	}
	// Order direction.
	var orderExp exp.OrderedExpression
	if orderDirection == "desc" {
		orderExp = selector.Desc()
	} else {
		orderExp = selector.Asc()
	}
	return qb.OrderPrepend(orderExp), nil
}
