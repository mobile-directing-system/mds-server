package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
)

// UpdateOperationMembersByOperation updates the members for the operation with
// the given id.
func (m *Mall) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, members []uuid.UUID) error {
	// Clear operation members.
	clearQuery, _, err := m.dialect.Delete(goqu.T("operation_members")).
		Where(goqu.C("operation").Eq(operationID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "clear-query to sql", nil)
	}
	_, err = tx.Exec(ctx, clearQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec clear-query", clearQuery)
	}
	// Add new members.
	records := make([]any, 0, len(members))
	for _, id := range members {
		records = append(records, goqu.Record{
			"operation": operationID,
			"user":      id,
		})
	}
	if len(records) == 0 {
		return nil
	}
	addQuery, _, err := goqu.Insert(goqu.T("operation_members")).
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

// OperationMembersByOperation retrieves a paginated User list for the operation
// with the given id.
func (m *Mall) OperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID,
	params pagination.Params) (pagination.Paginated[User], error) {
	// Build query.
	q, _, err := pagination.QueryToSQLWithPagination(m.dialect.From(goqu.T("operation_members")).
		InnerJoin(goqu.T("users"), goqu.On(goqu.I("users.id").Eq(goqu.I("operation_members.user")))).
		Select(goqu.I("users.id"),
			goqu.I("users.username"),
			goqu.I("users.first_name"),
			goqu.I("users.last_name")).
		Where(goqu.I("operation_members.operation").Eq(operationID)).
		Order(goqu.I("users.username").Asc()), params, pagination.FieldMap{
		"username":   goqu.I("users.username"),
		"first_name": goqu.I("users.first_name"),
		"last_name":  goqu.I("users.last_name"),
	})
	if err != nil {
		return pagination.Paginated[User]{}, meh.Wrap(err, "query to sql with pagination", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return pagination.Paginated[User]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	users := make([]User, 0)
	var total int
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&total)
		if err != nil {
			return pagination.Paginated[User]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		users = append(users, user)
	}
	return pagination.NewPaginated(params, users, total), nil
}

// OperationsByMember retrieves an Operation list for the member with the given
// id.
func (m *Mall) OperationsByMember(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]Operation, error) {
	q, _, err := m.dialect.From(goqu.T("operations")).
		InnerJoin(goqu.T("operation_members"),
			goqu.On(goqu.I("operation_members.operation").Eq(goqu.I("operations.id")))).
		Select(goqu.I("operations.id"),
			goqu.I("operations.title"),
			goqu.I("operations.description"),
			goqu.I("operations.start_ts"),
			goqu.I("operations.end_ts"),
			goqu.I("operations.is_archived")).
		Where(goqu.I("operation_members.user").Eq(userID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	operations := make([]Operation, 0)
	for rows.Next() {
		var operation Operation
		err = rows.Scan(&operation.ID,
			&operation.Title,
			&operation.Description,
			&operation.Start,
			&operation.End,
			&operation.IsArchived)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		operations = append(operations, operation)
	}
	return operations, nil
}