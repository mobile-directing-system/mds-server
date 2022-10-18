package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"time"
)

// AcceptedIntelDeliveryAttempt is an attempt for intel-delivery for a specific
// channel.
type AcceptedIntelDeliveryAttempt struct {
	// ID identifies the attempt.
	ID uuid.UUID
	// Intel is the id of the referenced intel.
	Intel uuid.UUID
	// IntelOperation is the operation the referenced intel is assigned to.
	IntelOperation uuid.UUID
	// IntelImportance is the importance of the referenced intel.
	IntelImportance int
	// AssignedTo is the id of the assigned address book entry.
	AssignedTo uuid.UUID
	// AssignedToLabel is the label of the assigned address book entry.
	AssignedToLabel string
	// AssignedToUser is the id of the optionally assigned user (from the address
	// book entry).
	AssignedToUser uuid.NullUUID
	// Delivery is the id of the referenced delivery.
	Delivery uuid.UUID
	// Channel is the id of the channel to use for this attempt.
	Channel uuid.UUID
	// CreatedAt is the timestamp when the attempt was started.
	CreatedAt time.Time
	// IsActive describes whether the attempt is still ongoing.
	IsActive bool
	// StatusTS is the timestamp when the Status was last updated.
	StatusTS time.Time
	// Note contains optional human-readable information regarding the attempt.
	Note nulls.String
	// AcceptedAt is the timestamp when the attempt was accepted by the service.
	AcceptedAt time.Time
}

// CreateAcceptedIntelDeliveryAttempt creates the given
// AcceptedIntelDeliveryAttempt.
func (m *Mall) CreateAcceptedIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, create AcceptedIntelDeliveryAttempt) error {
	q, _, err := m.dialect.Insert(goqu.T("accepted_intel_delivery_attempts")).Rows(goqu.Record{
		"id":                create.ID,
		"intel":             create.Intel,
		"intel_operation":   create.IntelOperation,
		"intel_importance":  create.IntelImportance,
		"assigned_to":       create.AssignedTo,
		"assigned_to_label": create.AssignedToLabel,
		"assigned_to_user":  create.AssignedToUser,
		"delivery":          create.Delivery,
		"channel":           create.Channel,
		"created_at":        create.CreatedAt.UTC(),
		"is_active":         create.IsActive,
		"status_ts":         create.StatusTS.UTC(),
		"note":              create.Note,
		"accepted_at":       create.AcceptedAt.UTC(),
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

// AcceptedIntelDeliveryAttemptStatus holds the status information for an
// accepted intel-delivery-attempt. However, we do not care about the actual
// status-code but IsActive.
type AcceptedIntelDeliveryAttemptStatus struct {
	// ID identifies the delivery-attempt.
	ID uuid.UUID
	// IsActive describes whether the attempt is still ongoing.
	IsActive bool `json:"is_active"`
	// StatusTS is the timestamp when the Status was last updated.
	StatusTS time.Time `json:"status_ts"`
	// Note contains optional human-readable information regarding the attempt.
	Note nulls.String `json:"note"`
}

// UpdateAcceptedIntelDeliveryAttemptStatus updates the given
// AcceptedIntelDeliveryAttemptStatus, identified by its id.
func (m *Mall) UpdateAcceptedIntelDeliveryAttemptStatus(ctx context.Context, tx pgx.Tx, update AcceptedIntelDeliveryAttemptStatus) error {
	q, _, err := m.dialect.Update(goqu.T("accepted_intel_delivery_attempts")).Set(goqu.Record{
		"is_active": update.IsActive,
		"status_ts": update.StatusTS.UTC(),
		"note":      update.Note,
	}).Where(goqu.C("id").Eq(update.ID)).ToSQL()
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

// AcceptedIntelDeliveryAttemptByID retrieves the AcceptedIntelDeliveryAttempt
// with the given id from the database.
func (m *Mall) AcceptedIntelDeliveryAttemptByID(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (AcceptedIntelDeliveryAttempt, error) {
	q, _, err := m.dialect.From(goqu.T("accepted_intel_delivery_attempts")).
		Select(goqu.C("id"),
			goqu.C("intel"),
			goqu.C("intel_operation"),
			goqu.C("intel_importance"),
			goqu.C("assigned_to"),
			goqu.C("assigned_to_label"),
			goqu.C("assigned_to_user"),
			goqu.C("delivery"),
			goqu.C("channel"),
			goqu.C("created_at"),
			goqu.C("is_active"),
			goqu.C("status_ts"),
			goqu.C("note"),
			goqu.C("accepted_at")).
		Where(goqu.C("id").Eq(attemptID)).ToSQL()
	if err != nil {
		return AcceptedIntelDeliveryAttempt{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return AcceptedIntelDeliveryAttempt{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return AcceptedIntelDeliveryAttempt{}, meh.NewNotFoundErr("not found", nil)
	}
	var attempt AcceptedIntelDeliveryAttempt
	err = rows.Scan(&attempt.ID,
		&attempt.Intel,
		&attempt.IntelOperation,
		&attempt.IntelImportance,
		&attempt.AssignedTo,
		&attempt.AssignedToLabel,
		&attempt.AssignedToUser,
		&attempt.Delivery,
		&attempt.Channel,
		&attempt.CreatedAt,
		&attempt.IsActive,
		&attempt.StatusTS,
		&attempt.Note,
		&attempt.AcceptedAt)
	if err != nil {
		return AcceptedIntelDeliveryAttempt{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return attempt, nil
}
