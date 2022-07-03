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

// User contains all stored user information.
type User struct {
	// ID identifies the user.
	ID uuid.UUID
	// Username for the user.
	Username string
	// FirstName of the user.
	FirstName string
	// LastName of the user.
	LastName string
	// IsAdmin describes whether the User is an administrator.
	IsAdmin bool
}

// UserWithPass is a User with a Pass field.
type UserWithPass struct {
	User
	// Pass is the hashed password for the user.
	Pass []byte
}

// UserByID retrieves a User by its User.ID.
func (m *Mall) UserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (User, error) {
	// Build query.
	q, _, err := m.dialect.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("first_name"),
			goqu.C("last_name"),
			goqu.C("is_admin")).
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
		&user.FirstName,
		&user.LastName,
		&user.IsAdmin)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// UserByUsername retrieves a User by its User.Username.
func (m *Mall) UserByUsername(ctx context.Context, tx pgx.Tx, username string) (User, error) {
	// Build query.
	q, _, err := m.dialect.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("first_name"),
			goqu.C("last_name"),
			goqu.C("is_admin")).
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
		&user.FirstName,
		&user.LastName,
		&user.IsAdmin)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// Users retrieves all known users.
func (m *Mall) Users(ctx context.Context, tx pgx.Tx, params pagination.Params) (pagination.Paginated[User], error) {
	// Build query.
	q, _, err := pagination.QueryToSQLWithPagination(m.dialect.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("first_name"),
			goqu.C("last_name"),
			goqu.C("is_admin")).
		Order(goqu.C("username").Asc()), params, pagination.FieldMap{
		"username":   goqu.C("username"),
		"first_name": goqu.C("first_name"),
		"last_name":  goqu.C("last_name"),
		"is_admin":   goqu.C("is_admin"),
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
			&user.IsAdmin,
			&total)
		if err != nil {
			return pagination.Paginated[User]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		users = append(users, user)
	}
	return pagination.NewPaginated(params, users, total), nil
}

// CreateUser creates the given user.
func (m *Mall) CreateUser(ctx context.Context, tx pgx.Tx, user UserWithPass) (User, error) {
	// Build query.
	q, _, err := m.dialect.Insert(goqu.T("users")).Rows(goqu.Record{
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"is_admin":   user.IsAdmin,
		"pass":       user.Pass,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return User{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return User{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return User{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return User{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	err = rows.Scan(&user.ID)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user.User, nil
}

// UpdateUser updates the given User, identified by its User.ID. This will not
// change the password!
func (m *Mall) UpdateUser(ctx context.Context, tx pgx.Tx, user User) error {
	// Build query.
	q, _, err := m.dialect.Update(goqu.T("users")).Set(goqu.Record{
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"is_admin":   user.IsAdmin,
	}).Where(goqu.C("id").Eq(user.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("user not found", nil)
	}
	return nil
}

// DeleteUserByID deletes the user with the given id.
func (m *Mall) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Build query.
	q, _, err := m.dialect.Delete(goqu.T("users")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("user not found", nil)
	}
	return nil
}

// UpdateUserPassByUserID updates the hashed password of the user with the given
// id.
func (m *Mall) UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, pass []byte) error {
	// Build query.
	q, _, err := m.dialect.Update(goqu.T("users")).Set(goqu.Record{
		"pass": pass,
	}).Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("user not found", nil)
	}
	return nil
}
