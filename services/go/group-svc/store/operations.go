package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
)

// CreateOperation creates the operation with the given id.
func (m *Mall) CreateOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) error {
	q, _, err := goqu.Insert(goqu.T("operations")).Rows(goqu.Record{
		"id": operationID,
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

// OperationMembersByOperation retrieves all users that are member of the
// operation with the given id.
func (m *Mall) OperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) ([]uuid.UUID, error) {
	q, _, err := m.dialect.From(goqu.T("operation_members")).
		Select(goqu.C("user")).
		Where(goqu.C("operation").Eq(operationID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	users := make([]uuid.UUID, 0)
	for rows.Next() {
		var user uuid.UUID
		err = rows.Scan(&user)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		users = append(users, user)
	}
	return users, nil
}

// UpdateOperationMembersByOperation updates the member list for the operation
// with the given id.
func (m *Mall) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error {
	// Clear all existing ones.
	clearQuery, _, err := m.dialect.Delete(goqu.T("operation_members")).
		Where(goqu.C("operation").Eq(operationID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "clear-query to sql", nil)
	}
	_, err = tx.Exec(ctx, clearQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec clear-query", clearQuery)
	}
	// Add members.
	records := make([]any, 0, len(newMembers))
	for _, member := range newMembers {
		records = append(records, goqu.Record{
			"operation": operationID,
			"user":      member,
		})
	}
	if len(records) == 0 {
		return nil
	}
	addQuery, _, err := m.dialect.Insert(goqu.T("operation_members")).
		Rows(records...).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "add-query to sql", nil)
	}
	_, err = tx.Exec(ctx, addQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec add-query", addQuery)
	}
	return nil
}
