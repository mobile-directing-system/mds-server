package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
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
	// IsActive describes whether the user is active.
	IsActive bool `json:"isActive"`
}

// UserWithPass holds an User with the hashed password.
type UserWithPass struct {
	User
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

// UserWithPassByUsername retrieves the User with the given User.Username.
func (m *Mall) UserWithPassByUsername(ctx context.Context, tx pgx.Tx, username string) (UserWithPass, error) {
	// Build query.
	q, _, err := goqu.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("is_admin"),
			goqu.C("is_active"),
			goqu.C("pass")).
		Where(goqu.C("username").Eq(username)).ToSQL()
	if err != nil {
		return UserWithPass{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return UserWithPass{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	if !rows.Next() {
		return UserWithPass{}, meh.NewNotFoundErr("user not found", nil)
	}
	var user UserWithPass
	err = rows.Scan(&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.IsActive,
		&user.Pass)
	if err != nil {
		return UserWithPass{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// UserWithPassByID retrieves the User with the given User.ID.
func (m *Mall) UserWithPassByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (UserWithPass, error) {
	// Build query.
	q, _, err := goqu.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("is_admin"),
			goqu.C("is_active"),
			goqu.C("pass")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return UserWithPass{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return UserWithPass{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	if !rows.Next() {
		return UserWithPass{}, meh.NewNotFoundErr("user not found", nil)
	}
	var user UserWithPass
	err = rows.Scan(&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.IsActive,
		&user.Pass)
	if err != nil {
		return UserWithPass{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// CreateUser creates the given User in the database.
func (m *Mall) CreateUser(ctx context.Context, tx pgx.Tx, user UserWithPass) error {
	// Build query.
	q, _, err := goqu.Insert(goqu.T("users")).Rows(goqu.Record{
		"id":        user.ID,
		"username":  user.Username,
		"is_admin":  user.IsAdmin,
		"is_active": user.IsActive,
		"pass":      user.Pass,
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
		"username":  user.Username,
		"is_admin":  user.IsAdmin,
		"is_active": user.IsActive,
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

// UpdateUserPassByUserID updates the password for the user with the given id.
func (m *Mall) UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newPass []byte) error {
	// Build query.
	q, _, err := goqu.Update(goqu.T("users")).Set(goqu.Record{
		"pass": newPass,
	}).Where(goqu.C("id").Eq(userID)).ToSQL()
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
