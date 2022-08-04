package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
)

// User represents a user in the store.
type User struct {
	// ID identifies the user.
	ID uuid.UUID
	// Username for logging in.
	Username string
	// FirstName of the user.
	FirstName string
	// LastName of the user.
	LastName string
}

// CreateUser adds the given User to the store.
func (m *Mall) CreateUser(ctx context.Context, tx pgx.Tx, create User) error {
	q, _, err := goqu.Insert(goqu.T("users")).Rows(goqu.Record{
		"id":         create.ID,
		"username":   create.Username,
		"first_name": create.FirstName,
		"last_name":  create.LastName,
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

// UpdateUser updates the given User, identified by its id.
func (m *Mall) UpdateUser(ctx context.Context, tx pgx.Tx, update User) error {
	q, _, err := goqu.Update(goqu.T("users")).Set(goqu.Record{
		"username":   update.Username,
		"first_name": update.FirstName,
		"last_name":  update.LastName,
	}).Where(goqu.C("id").Eq(update.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("not found", nil)
	}
	return nil
}

// DeleteUserByID deletes the user with the given id.
func (m *Mall) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	q, _, err := goqu.Delete(goqu.T("users")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("not found", nil)
	}
	return nil
}
