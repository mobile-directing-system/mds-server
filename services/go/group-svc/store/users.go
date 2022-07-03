package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
)

// CreateUser creates a user with the given id.
func (m *Mall) CreateUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	q, _, err := goqu.Insert(goqu.T("users")).Rows(goqu.Record{
		"id": userID,
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
