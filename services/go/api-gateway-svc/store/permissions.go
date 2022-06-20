package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
)

// PermissionsByUserID retrieves a list of granted permission.Permission for the
// user with the given id.
func (m *Mall) PermissionsByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]permission.Permission, error) {
	// Build query.
	q, _, err := m.dialect.From(goqu.T("permissions")).
		Select(goqu.C("name"),
			goqu.C("options")).
		Where(goqu.C("user").Eq(userID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	permissions := make([]permission.Permission, 0)
	for rows.Next() {
		var perm permission.Permission
		err = rows.Scan(&perm.Name,
			&perm.Options)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

// UpdatePermissionsByUser updates the permissions for the user with the given
// id.
func (m *Mall) UpdatePermissionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, permissions []permission.Permission) error {
	// Delete current permissions.
	deleteQuery, _, err := goqu.Delete(goqu.T("permissions")).
		Where(goqu.C("user").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "delete-query to sql", nil)
	}
	_, err = tx.Exec(ctx, deleteQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec delete-query", deleteQuery)
	}
	if len(permissions) == 0 {
		return nil
	}
	// Create new permissions.
	records := make([]interface{}, 0, len(permissions))
	for _, p := range permissions {
		records = append(records, goqu.Record{
			"user":    userID,
			"name":    p.Name,
			"options": p.Options,
		})
	}
	createQuery, _, err := goqu.Insert(goqu.T("permissions")).Rows(records...).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "create-query to sql", nil)
	}
	_, err = tx.Exec(ctx, createQuery)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec create-query", createQuery)
	}
	return nil
}
