package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"time"
)

// RadioDelivery represents a request for delivering intel over radio.
type RadioDelivery struct {
	// Attempt is the id of the delivery-attempt this radio-delivery is for.
	Attempt uuid.UUID
	// PickedUpBy is the id of the user that is responsible for this delivery.
	PickedUpBy uuid.NullUUID
	// PickedUpAt is the timestamp when the user from PickedUpBy was assigned.
	PickedUpAt nulls.Time
	// Success is not set, when the delivery is still awaiting results or open.
	Success nulls.Bool
	// SuccessTS is the timestamp of the last update of Success.
	SuccessTS time.Time
	// Note holds information regarding the delivery and/or Success.
	Note string
}

// RadioDeliveryByAttempt retrieves the RadioDelivery for the attempt with the
// given id.
func (m *Mall) RadioDeliveryByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (RadioDelivery, error) {
	q, _, err := m.dialect.From(goqu.T("radio_deliveries")).
		Select(goqu.C("attempt"),
			goqu.C("picked_up_by"),
			goqu.C("picked_up_at"),
			goqu.C("success"),
			goqu.C("success_ts"),
			goqu.C("note")).
		Where(goqu.C("attempt").Eq(attemptID)).ToSQL()
	if err != nil {
		return RadioDelivery{}, meh.NewBadInputErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return RadioDelivery{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return RadioDelivery{}, meh.NewNotFoundErr("not found", nil)
	}
	var rd RadioDelivery
	err = rows.Scan(&rd.Attempt,
		&rd.PickedUpBy,
		&rd.PickedUpAt,
		&rd.Success,
		&rd.SuccessTS,
		&rd.Note)
	if err != nil {
		return RadioDelivery{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return rd, nil
}

// CreateRadioDelivery creates a RadioDelivery for the attempt with the given
// id.
func (m *Mall) CreateRadioDelivery(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) error {
	q, _, err := m.dialect.Insert(goqu.T("radio_deliveries")).Rows(goqu.Record{
		"attempt":      attemptID,
		"picked_up_by": uuid.NullUUID{},
		"picked_up_at": nulls.Time{},
		"success":      nulls.Bool{},
		"success_ts":   time.Now(),
		"note":         "waiting for pickup",
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

// MarkRadioDeliveryAsPickedUpByAttempt marks the radio-delivery for the given
// attempt as picked-up by the user with the given id. If no user id is given,
// it will be unassigned.
func (m *Mall) MarkRadioDeliveryAsPickedUpByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID,
	by uuid.NullUUID, newNote string) error {
	var pickedUpAt nulls.Time
	if by.Valid {
		pickedUpAt = nulls.NewTime(time.Now())
	}
	q, _, err := m.dialect.Update(goqu.T("radio_deliveries")).Set(goqu.Record{
		"picked_up_by": by,
		"picked_up_at": pickedUpAt,
		"note":         newNote,
	}).Where(goqu.C("attempt").Eq(attemptID)).ToSQL()
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

// UpdateRadioDeliveryStatusByAttempt updates the radio-delivery for the given
// attempt with the given new success-status and note.
func (m *Mall) UpdateRadioDeliveryStatusByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID,
	newSuccess nulls.Bool, newNote string) error {
	q, _, err := m.dialect.Update(goqu.T("radio_deliveries")).Set(goqu.Record{
		"success":    newSuccess,
		"success_ts": time.Now(),
		"note":       newNote,
	}).Where(goqu.C("attempt").Eq(attemptID)).ToSQL()
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

// ActiveRadioDelivery is used in Mall.ActiveRadioDeliveriesAndLockOrWait for
// retrieving all active deliveries in order to choose the next one for
// assigning to a user.
type ActiveRadioDelivery struct {
	// Attempt is the id of the referenced attempt for the radio-delivery.
	Attempt uuid.UUID
	// PickedUpAt is the set, when the delivery is picked up.
	PickedUpAt nulls.Time
	// IntelOperation is the assigned operation of the referenced intel to deliver.
	IntelOperation uuid.UUID
	// IntelImportance is the importance of the intel to deliver for priorization.
	IntelImportance int
	// AttemptCreatedAt is the timestamp when the delivery-attempt was created in
	// order to avoid starvation.
	AttemptCreatedAt time.Time
}

// ActiveRadioDeliveriesAndLockOrWait locks and retrieves all radio-deliveries
// being active (success is NULL) from the database.
func (m *Mall) ActiveRadioDeliveriesAndLockOrWait(ctx context.Context, tx pgx.Tx, byOperation uuid.NullUUID) ([]ActiveRadioDelivery, error) {
	qb := m.dialect.From(goqu.T("radio_deliveries")).
		InnerJoin(goqu.T("accepted_intel_delivery_attempts"),
			goqu.On(goqu.I("accepted_intel_delivery_attempts.id").Eq(goqu.I("radio_deliveries.attempt")))).
		Select(goqu.I("radio_deliveries.attempt"),
			goqu.I("radio_deliveries.picked_up_at"),
			goqu.I("accepted_intel_delivery_attempts.intel_operation"),
			goqu.I("accepted_intel_delivery_attempts.intel_importance"),
			goqu.I("accepted_intel_delivery_attempts.created_at")).
		ForUpdate(exp.Wait).
		Where(goqu.I("radio_deliveries.success").IsNull())
	if byOperation.Valid {
		qb = qb.Where(goqu.I("accepted_intel_delivery_attempts.intel_operation").Eq(byOperation.UUID))
	}
	q, _, err := qb.ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	radioDeliveries := make([]ActiveRadioDelivery, 0)
	for rows.Next() {
		var rd ActiveRadioDelivery
		err = rows.Scan(&rd.Attempt,
			&rd.PickedUpAt,
			&rd.IntelOperation,
			&rd.IntelImportance,
			&rd.AttemptCreatedAt)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		radioDeliveries = append(radioDeliveries, rd)
	}
	return radioDeliveries, nil
}
