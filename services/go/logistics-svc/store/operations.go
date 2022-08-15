package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"time"
)

// Operation contains details for detailed address book entries and
// associations.
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

// CreateOperation creates the operation with the given id.
func (m *Mall) CreateOperation(ctx context.Context, tx pgx.Tx, operation Operation) error {
	q, _, err := goqu.Insert(goqu.T("operations")).Rows(goqu.Record{
		"id":          operation.ID,
		"title":       operation.Title,
		"description": operation.Description,
		"start_ts":    operation.Start,
		"end_ts":      operation.End,
		"is_archived": operation.IsArchived,
	}).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// UpdateOperation updates the given Operation.
func (m *Mall) UpdateOperation(ctx context.Context, tx pgx.Tx, update Operation) error {
	q, _, err := goqu.Update(goqu.T("operations")).Set(goqu.Record{
		"title":       update.Title,
		"description": update.Description,
		"start_ts":    update.Start,
		"end_ts":      update.End,
		"is_archived": update.IsArchived,
	}).Where(goqu.C("id").Eq(update.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	return nil
}

// UpdateOperationMembersByOperation updates the operation members for the given
// operation.
func (m *Mall) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error {
	// Delete old members.
	deleteQuery, _, err := m.dialect.Delete(goqu.T("operation_members")).
		Where(goqu.C("operation").Eq(operationID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "delete-query to sql", nil)
	}
	_, err = tx.Exec(ctx, deleteQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec delete-query", deleteQuery)
	}
	// Create new members.
	if len(newMembers) == 0 {
		return nil
	}
	rows := make([]any, 0, len(newMembers))
	for _, member := range newMembers {
		rows = append(rows, goqu.Record{
			"operation": operationID,
			"user":      member,
		})
	}
	insertQuery, _, err := m.dialect.Insert(goqu.T("operation_members")).Rows(rows...).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "insert-query to sql", nil)
	}
	_, err = tx.Exec(ctx, insertQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec insert-query", insertQuery)
	}
	return nil
}
