package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"time"
)

// Intel is a summary of relevant information of MDS intel.
type Intel struct {
	// ID identifies the Intel.
	ID uuid.UUID
	// CreatedAt is the timestamp when the Intel was created.
	CreatedAt time.Time
	// CreatedBy is the id of the user that created the Intel.
	CreatedBy uuid.UUID
	// Operation is the id of the operation, the Intel is associated with.
	Operation uuid.UUID
	// Importance of the Intel.
	Importance int
	// IsValid is false when the Intel is considered deleted.
	IsValid bool
}

// IntelByID retrieves the Intel with the given id from the store.
func (m *Mall) IntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) (Intel, error) {
	q, _, err := m.dialect.From(goqu.T("intel")).
		Select(goqu.C("id"),
			goqu.C("created_at"),
			goqu.C("created_by"),
			goqu.C("operation"),
			goqu.C("importance"),
			goqu.C("is_valid")).
		Where(goqu.C("id").Eq(intelID)).ToSQL()
	if err != nil {
		return Intel{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Intel{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return Intel{}, meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	var intel Intel
	err = rows.Scan(&intel.ID,
		&intel.CreatedAt,
		&intel.CreatedBy,
		&intel.Operation,
		&intel.Importance,
		&intel.IsValid)
	if err != nil {
		return Intel{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return intel, nil
}

// CreateIntel creates the given Intel in the store.
func (m *Mall) CreateIntel(ctx context.Context, tx pgx.Tx, create Intel) error {
	q, _, err := m.dialect.Insert(goqu.T("intel")).Rows(goqu.Record{
		"id":         create.ID,
		"created_at": create.CreatedAt.UTC(),
		"created_by": create.CreatedBy,
		"operation":  create.Operation,
		"importance": create.Importance,
		"is_valid":   create.IsValid,
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

// InvalidateIntelByID sets the Intel with the given id to invalid.
func (m *Mall) InvalidateIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error {
	q, _, err := m.dialect.Update(goqu.T("intel")).Set(goqu.Record{
		"is_valid": false,
	}).Where(goqu.C("id").Eq(intelID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	return nil
}
