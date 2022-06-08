package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
)

// User holds internal details regarding a user like he is stored in the
// database.
type User struct {
	// ID identifies the user.
	ID uuid.UUID `json:"id"`
	// Username is used for logging in.
	Username string `json:"username"`
	// IsAdmin describes whether the user is an admin.
	IsAdmin bool `json:"isAdmin"`
	// Pass is the hashed password.
	Pass []byte `json:"pass"`
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

// UserByUsername retrieves the User with the given User.Username.
func (m *Mall) UserByUsername(ctx context.Context, tx pgx.Tx, username string) (User, error) {
	// Build query.
	q, _, err := goqu.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
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
	err = rows.Scan(&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.Pass)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// UserByID retrieves the User with the given User.ID.
func (m *Mall) UserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (User, error) {
	// Build query.
	q, _, err := goqu.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("is_admin"),
			goqu.C("pass")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
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
	err = rows.Scan(&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.Pass)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// CreateUser creates the given User in the database.
func (m *Mall) CreateUser(ctx context.Context, tx pgx.Tx, user User) error {
	// Build query.
	q, _, err := goqu.Insert(goqu.T("users")).Rows(goqu.Record{
		"id":       user.ID,
		"username": user.Username,
		"is_admin": user.IsAdmin,
		"pass":     user.Pass,
	}).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// UpdateUser updates the given User, identified by its User.ID.
func (m *Mall) UpdateUser(ctx context.Context, tx pgx.Tx, user User) error {
	// Build query.
	q, _, err := goqu.Update(goqu.T("users")).Set(goqu.Record{
		"username": user.Username,
		"is_admin": user.IsAdmin,
		"pass":     user.Pass,
	}).Where(goqu.C("id").Eq(user.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	// Assure found.
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("user not found", nil)
	}
	return nil
}

// DeleteUserByID deletes the user with the given id.
func (m *Mall) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID string) error {
	// Build query.
	q, _, err := goqu.Delete(goqu.T("users")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	// Assure found.
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("user not found", nil)
	}
	return nil
}
