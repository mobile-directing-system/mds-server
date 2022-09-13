package store

import (
	"context"
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"time"
)

// IntelType is the type of intel. Also describes the content.
type IntelType string

// IntelToDeliver holds intel information.
type IntelToDeliver struct {
	// Attempt identifies the intel to deliver and is the id of the associated
	// delivery-attempt.
	Attempt uuid.UUID
	// ID identifies intel.
	ID uuid.UUID
	// CreatedAt is the timestamp, the intel was created.
	CreatedAt time.Time
	// CreatedBy is the id of the user, who created the intent.
	CreatedBy uuid.UUID
	// Operation is the id of the associated operation.
	Operation uuid.UUID
	// Type of the intel.
	Type IntelType
	// Content is the actual information.
	Content json.RawMessage
	// Importance of the intel.
	Importance int
}

// CreateIntelToDeliver creates the given IntelToDeliver in the database.
func (m *Mall) CreateIntelToDeliver(ctx context.Context, tx pgx.Tx, create IntelToDeliver) error {
	q, _, err := m.dialect.Insert(goqu.T("intel_to_deliver")).Rows(goqu.Record{
		"attempt":    create.Attempt,
		"id":         create.ID,
		"created_at": create.CreatedAt.UTC(),
		"created_by": create.CreatedBy,
		"operation":  create.Operation,
		"type":       create.Type,
		"content":    pgutil.JSONB(create.Content),
		"importance": create.Importance,
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

// IntelToDeliverByAttempt retrieves the IntelToDeliver by its associated
// delivery-attempt.
func (m *Mall) IntelToDeliverByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (IntelToDeliver, error) {
	q, _, err := m.dialect.From(goqu.T("intel_to_deliver")).
		Select(goqu.C("attempt"),
			goqu.C("id"),
			goqu.C("created_at"),
			goqu.C("created_by"),
			goqu.C("operation"),
			goqu.C("type"),
			goqu.C("content"),
			goqu.C("importance")).
		Where(goqu.C("attempt").Eq(attemptID)).ToSQL()
	if err != nil {
		return IntelToDeliver{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return IntelToDeliver{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return IntelToDeliver{}, meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	var intel IntelToDeliver
	err = rows.Scan(&intel.Attempt,
		&intel.ID,
		&intel.CreatedAt,
		&intel.CreatedBy,
		&intel.Operation,
		&intel.Type,
		pgutil.JSONB(&intel.Content),
		&intel.Importance)
	if err != nil {
		return IntelToDeliver{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	return intel, nil
}
