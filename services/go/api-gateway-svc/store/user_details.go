package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
)

// UserDetails holds details regarding a user. This will be converted to
// auth.UserDetails and added to forwarded requests.
type UserDetails struct {
	// Username of the user that performed the request.
	Username string `json:"username"`
	// IsAuthenticated describes whether the user is currently logged in.
	IsAuthenticated bool `json:"is_authenticated"`
	// IsAdmin describes whether the user is an admin.
	IsAdmin bool `json:"is_admin"`
	// Permissions the user was granted.
	Permissions []permission.Permission `json:"permissions"`
}

// PassByUsername retrieves the hashed password for the user with the given
// username.
func (m *Mall) PassByUsername(ctx context.Context, tx pgx.Tx, username string) ([]byte, error) {
	// Build query.
	q, _, err := goqu.From(goqu.T("users")).
		Select(goqu.C("pass")).
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
	if !rows.Next() {
		return nil, meh.NewNotFoundErr("user not found", nil)
	}
	var pass []byte
	err = rows.Scan(&pass)
	if err != nil {
		return nil, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return pass, nil
}

// UserDetailsByUsername retrieves the UserDetails for the user with the given
// username.
func (m *Mall) UserDetailsByUsername(ctx context.Context, tx pgx.Tx, username string) (UserDetails, error) {
	// Build user details query.
	userQ, _, err := m.Dialect.From(goqu.T("users")).
		Select(goqu.C("username"),
			goqu.C("is_admin")).
		Where(goqu.C("username").Eq(username)).ToSQL()
	if err != nil {
		return UserDetails{}, meh.NewInternalErrFromErr(err, "user query to sql", nil)
	}
	// Query user details.
	userRows, err := tx.Query(ctx, userQ)
	if err != nil {
		return UserDetails{}, mehpg.NewQueryDBErr(err, "query user from db", userQ)
	}
	defer userRows.Close()
	// Scan user details.
	if !userRows.Next() {
		return UserDetails{}, meh.NewNotFoundErr("user not found", nil)
	}
	var userDetails UserDetails
	err = userRows.Scan(&userDetails.Username,
		&userDetails.IsAdmin)
	if err != nil {
		return UserDetails{}, mehpg.NewScanRowsErr(err, "scan user row", userQ)
	}
	userRows.Close()
	// Retrieve permissions.
	permissions, err := m.PermissionsByUsername(ctx, tx, username)
	if err != nil {
		return UserDetails{}, meh.Wrap(err, "permissions by username", meh.Details{"username": username})
	}
	userDetails.Permissions = permissions
	return userDetails, nil
}
