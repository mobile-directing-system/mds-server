package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
)

// CreateUser creates a user with the given id.
func (m *Mall) CreateUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Build query.
	q, _, err := goqu.Insert(goqu.T("users")).Rows(goqu.Record{
		"id": userID,
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

// DeleteUserByID deletes the user with the given id.
//
// Warning: Keep in mind, that the user must not have any permissions assigned!
func (m *Mall) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Build query.
	q, _, err := goqu.Delete(goqu.T("users")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
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

// AssureUserExists makes sure that the user with the given id exists.
func (m *Mall) AssureUserExists(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Build query.
	q, _, err := goqu.From(goqu.T("users")).
		Select(goqu.C("id")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return meh.NewNotFoundErr("user not found", nil)
	}
	rows.Close()
	return nil
}
