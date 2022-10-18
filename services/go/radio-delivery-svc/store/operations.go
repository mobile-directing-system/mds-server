package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
)

// UpdateOperationMembersByOperation replaces the associated operation members
// for the operation with the given id with the new given ones.
func (m *Mall) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error {
	err := m.deleteOperationMembersByOperation(ctx, tx, operationID)
	if err != nil {
		return meh.Wrap(err, "delete operation members by operation", meh.Details{"operation_id": operationID})
	}
	rows := make([]any, 0, len(newMembers))
	for _, member := range newMembers {
		rows = append(rows, goqu.Record{
			"operation": operationID,
			"user":      member,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	q, _, err := m.dialect.Insert(goqu.T("operation_members")).
		Rows(rows...).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// deleteOperationMembersByOperation deletes the operation member associations
// for the operation with the given id.
func (m *Mall) deleteOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) error {
	q, _, err := m.dialect.Delete(goqu.T("operation_members")).
		Where(goqu.C("operation").Eq(operationID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// OperationsByMember retrieves the ids of all operations, the user with the
// given id is member of.
func (m *Mall) OperationsByMember(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]uuid.UUID, error) {
	q, _, err := m.dialect.From(goqu.T("operation_members")).
		Select(goqu.C("operation")).
		Where(goqu.C("user").Eq(userID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	operations := make([]uuid.UUID, 0)
	for rows.Next() {
		var operationID uuid.UUID
		err = rows.Scan(&operationID)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		operations = append(operations, operationID)
	}
	return operations, nil
}
