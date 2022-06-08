package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
)

// PermissionsByUsername retrieves a list of granted permission.Permission for
// the user with the given username.
func (m *Mall) PermissionsByUsername(ctx context.Context, tx pgx.Tx, username string) ([]permission.Permission, error) {
	// Build query.
	q, _, err := m.Dialect.From(goqu.T("permissions")).
		Select(goqu.C("permission")).
		Where(goqu.C("username").Eq(username)).ToSQL()
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
		err = rows.Scan(&perm)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}
