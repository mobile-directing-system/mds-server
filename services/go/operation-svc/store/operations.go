package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"time"
)

// Operation is the store representation of an operation.
type Operation struct {
	// ID identifies the operation.
	ID uuid.UUID
	// Title of the operation.
	Title string
	// Optional description. We treat a non-existent description as empty string.
	Description string
	// Start timestamp of the operation.
	Start time.Time
	// End is the optional timestamp, when the operation has finished. If End is not
	// set or set to a moment in the past, the operation is considered finished.
	End nulls.Time
	// IsArchived describes whether the operation was archived. This is used instead
	// of deleting the operation in order to avoid unintended data loss.
	IsArchived bool
}

// Validate assures that the title is set and End is not before Start.
func (o Operation) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	if o.Title == "" {
		report.AddError("title must not be empty")
	}
	if o.End.Valid && o.End.Time.UTC().Before(o.Start.UTC()) {
		report.AddError("end-time must not be before start-time")
	}
	return report, nil
}

// OperationByID retrieves an Operation by its id.
func (m *Mall) OperationByID(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) (Operation, error) {
	// Build query.
	q, _, err := m.dialect.From(goqu.T("operations")).
		Select(goqu.C("id"),
			goqu.C("title"),
			goqu.C("description"),
			goqu.C("start_ts"),
			goqu.C("end_ts"),
			goqu.C("is_archived")).
		Where(goqu.C("id").Eq(operationID)).ToSQL()
	if err != nil {
		return Operation{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Operation{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	if !rows.Next() {
		return Operation{}, meh.NewNotFoundErr("operation not found", nil)
	}
	var operation Operation
	err = rows.Scan(&operation.ID,
		&operation.Title,
		&operation.Description,
		&operation.Start,
		&operation.End,
		&operation.IsArchived)
	if err != nil {
		return Operation{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return operation, nil
}

// Operations retrieves an Operation list.
func (m *Mall) Operations(ctx context.Context, tx pgx.Tx, params pagination.Params) (pagination.Paginated[Operation], error) {
	// Build query.
	q, _, err := pagination.QueryToSQLWithPagination(m.dialect.From(goqu.T("operations")).
		Select(goqu.C("id"),
			goqu.C("title"),
			goqu.C("description"),
			goqu.C("start_ts"),
			goqu.C("end_ts"),
			goqu.C("is_archived")).
		Order(goqu.C("end_ts").Desc()), params, pagination.FieldMap{
		"title":       goqu.C("title"),
		"description": goqu.C("description"),
		"start":       goqu.C("start_ts"),
		"end":         goqu.C("end_ts"),
		"is_archived": goqu.C("is_archived"),
	})
	if err != nil {
		return pagination.Paginated[Operation]{}, meh.Wrap(err, "query to sql with pagination", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return pagination.Paginated[Operation]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	operations := make([]Operation, 0)
	var total int
	for rows.Next() {
		var operation Operation
		err = rows.Scan(&operation.ID,
			&operation.Title,
			&operation.Description,
			&operation.Start,
			&operation.End,
			&operation.IsArchived,
			&total)
		if err != nil {
			return pagination.Paginated[Operation]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		operations = append(operations, operation)
	}
	return pagination.NewPaginated(params, operations, total), nil
}

// CreateOperation creates the given Operation and returns it with its assigned
// id.
func (m *Mall) CreateOperation(ctx context.Context, tx pgx.Tx, operation Operation) (Operation, error) {
	// Build query.
	q, _, err := m.dialect.Insert(goqu.T("operations")).Rows(goqu.Record{
		"title":       operation.Title,
		"description": operation.Description,
		"start_ts":    operation.Start.UTC(),
		"end_ts":      operation.End.UTC(),
		"is_archived": operation.IsArchived,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return Operation{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Operation{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return Operation{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return Operation{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	err = rows.Scan(&operation.ID)
	if err != nil {
		return Operation{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return operation, nil
}

// UpdateOperation updates the given Operation, identified by its Operation.ID.
func (m *Mall) UpdateOperation(ctx context.Context, tx pgx.Tx, operation Operation) error {
	// Build query.
	q, _, err := m.dialect.Update(goqu.T("operations")).Set(goqu.Record{
		"title":       operation.Title,
		"description": operation.Description,
		"start_ts":    operation.Start.UTC(),
		"end_ts":      operation.End.UTC(),
		"is_archived": operation.IsArchived,
	}).Where(goqu.C("id").Eq(operation.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("operation not found", nil)
	}
	return nil
}
