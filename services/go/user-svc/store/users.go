package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
)

// User contains all stored user information.
type User struct {
	// Username that identifies the user.
	Username string
	// FirstName of the user.
	FirstName string
	// LastName of the user.
	LastName string
	// IsAdmin describes whether the User is an administrator.
	IsAdmin bool
	// Pass is the salted password for the user.
	Pass []byte
}

// UserByUsername retrieves a User by its User.Username.
func (m *Mall) UserByUsername(ctx context.Context, tx pgx.Tx, username string) (User, error) {
	// Build query.
	q, _, err := m.dialect.From(goqu.T("users")).
		Select(goqu.C("username"),
			goqu.C("first_name"),
			goqu.C("last_name"),
			goqu.C("is_admin"),
			goqu.C("pass")).
		Where(goqu.C("username").Eq(username)).ToSQL()
	if err != nil {
		return User{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return User{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	if !rows.Next() {
		return User{}, meh.NewNotFoundErr("user not found", nil)
	}
	var user User
	err = rows.Scan(&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsAdmin,
		&user.Pass)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// CreateUser creates the given User.
func (m *Mall) CreateUser(ctx context.Context, tx pgx.Tx, user User) (User, error) {
	// Build query.
	q, _, err := m.dialect.Insert(goqu.T("users")).Rows(goqu.Record{
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"is_admin":   user.IsAdmin,
		"pass":       user.Pass,
	}).ToSQL()
	if err != nil {
		return User{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return User{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return user, nil
}
