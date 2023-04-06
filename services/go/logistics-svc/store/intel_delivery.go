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

// IntelDeliveryStatus is the status of an IntelDelivery.
type IntelDeliveryStatus string

const (
	// IntelDeliveryStatusOpen for deliveries, not being picked up, yet.
	IntelDeliveryStatusOpen IntelDeliveryStatus = "open"
	// IntelDeliveryStatusAwaitingDelivery for deliveries that have been picked up
	// by a mail carrier and are now awaiting delivery.
	IntelDeliveryStatusAwaitingDelivery IntelDeliveryStatus = "awaiting-delivery"
	// IntelDeliveryStatusDelivering for deliveries that are currently delivering
	// (for example ongoing phone calls).
	IntelDeliveryStatusDelivering IntelDeliveryStatus = "delivering"
	// IntelDeliveryStatusAwaitingAck for deliveries that are currently awaiting ACK
	// from the recipient (for example push notifications or email).
	IntelDeliveryStatusAwaitingAck IntelDeliveryStatus = "awaiting-ack"
	// IntelDeliveryStatusDelivered for deliveries that are successfully delivered
	// and acknowledged by the recipient.
	IntelDeliveryStatusDelivered IntelDeliveryStatus = "delivered"
	// IntelDeliveryStatusTimeout for deliveries that timed out (based on timeout
	// specified in channel properties).
	IntelDeliveryStatusTimeout IntelDeliveryStatus = "timeout"
	// IntelDeliveryStatusCanceled for manually cancelled deliveries.
	IntelDeliveryStatusCanceled IntelDeliveryStatus = "canceled"
	// IntelDeliveryStatusFailed is used for failed deliveries. An example might be
	// an invalid phone number that cannot be called.
	IntelDeliveryStatusFailed IntelDeliveryStatus = "failed"
)

// IntelDelivery for delivering Intel to an address book entry.
type IntelDelivery struct {
	// ID identifies the delivery.
	ID uuid.UUID
	// Intel is the id of the intel to deliver.
	Intel uuid.UUID
	// To is the id of the address book entry, the delivery is for.
	To uuid.UUID
	// IsActive describes, whether the delivery is still active and should be
	// checked by the scheduler/controller.
	IsActive bool
	// Success when delivery was successful.
	Success bool
	// Note contains optional human-readable information regarding the delivery.
	Note nulls.String `json:"note"`
}

// IntelDeliveryAttempt is an attempt for IntelDelivery for a specific channel.
type IntelDeliveryAttempt struct {
	// ID identifies the attempt.
	ID uuid.UUID
	// Delivery is the id of the referenced delivery.
	Delivery uuid.UUID
	// Channel is the id of the channel to use for this attempt.
	Channel uuid.UUID
	// CreatedAt is the timestamp when the attempt was started.
	CreatedAt time.Time
	// IsActive describes whether the attempt is still ongoing.
	IsActive bool
	// Status is the current/last status of the attempt.
	Status IntelDeliveryStatus
	// StatusTS is the timestamp when the Status was last updated.
	StatusTS time.Time
	// Note contains optional human-readable information regarding the attempt.
	Note nulls.String
}

// CreateIntelDelivery creates the given IntelDelivery and returns the one with
// assigned id.
func (m *Mall) CreateIntelDelivery(ctx context.Context, tx pgx.Tx, create IntelDelivery) (IntelDelivery, error) {
	q, _, err := m.dialect.Insert(goqu.T("intel_deliveries")).Rows(goqu.Record{
		"intel":     create.Intel,
		"to":        create.To,
		"is_active": create.IsActive,
		"success":   create.Success,
		"note":      create.Note,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return IntelDelivery{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return IntelDelivery{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return IntelDelivery{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return IntelDelivery{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	var deliveryID uuid.UUID
	err = rows.Scan(&deliveryID)
	if err != nil {
		return IntelDelivery{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	created, err := m.IntelDeliveryByID(ctx, tx, deliveryID)
	if err != nil {
		return IntelDelivery{}, meh.Wrap(err, "created intel delivery by id", meh.Details{"delivery_id": deliveryID})
	}
	// Update intel in search.
	err = m.addOrUpdateIntelInSearch(ctx, tx, create.Intel)
	if err != nil {
		return IntelDelivery{}, meh.Wrap(err, "update intel in search", meh.Details{"intel_id": create.Intel})
	}
	return created, nil
}

// IntelDeliveryByID retrieves the IntelDelivery with the given id.
func (m *Mall) IntelDeliveryByID(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) (IntelDelivery, error) {
	q, _, err := m.dialect.From(goqu.T("intel_deliveries")).
		Select(goqu.C("id"),
			goqu.C("intel"),
			goqu.C("to"),
			goqu.C("is_active"),
			goqu.C("success"),
			goqu.C("note")).
		Where(goqu.C("id").Eq(deliveryID)).ToSQL()
	if err != nil {
		return IntelDelivery{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return IntelDelivery{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return IntelDelivery{}, meh.NewNotFoundErr("not found", nil)
	}
	var delivery IntelDelivery
	err = rows.Scan(&delivery.ID,
		&delivery.Intel,
		&delivery.To,
		&delivery.IsActive,
		&delivery.Success,
		&delivery.Note)
	if err != nil {
		return IntelDelivery{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return delivery, nil
}

// IntelDeliveriesByIntel retrieves the IntelDelivery list for the intel with
// the given id.
func (m *Mall) IntelDeliveriesByIntel(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) ([]IntelDelivery, error) {
	q, _, err := m.dialect.From(goqu.T("intel_deliveries")).
		Select(goqu.C("id"),
			goqu.C("intel"),
			goqu.C("to"),
			goqu.C("is_active"),
			goqu.C("success"),
			goqu.C("note")).
		Where(goqu.C("intel").Eq(intelID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	deliveries := make([]IntelDelivery, 0)
	for rows.Next() {
		var delivery IntelDelivery
		err = rows.Scan(&delivery.ID,
			&delivery.Intel,
			&delivery.To,
			&delivery.IsActive,
			&delivery.Success,
			&delivery.Note)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		deliveries = append(deliveries, delivery)
	}
	return deliveries, nil
}

// CreateIntelDeliveryAttempt creates the given IntelDeliveryAttempt and returns
// the one with assigned id. Additionally, the status is set to
// IntelDeliveryStatusOpen and status-ts and created-at to the current
// timestamp.
func (m *Mall) CreateIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, create IntelDeliveryAttempt) (IntelDeliveryAttempt, error) {
	create.CreatedAt = time.Now()
	create.Status = IntelDeliveryStatusOpen
	create.StatusTS = time.Now()
	q, _, err := m.dialect.Insert(goqu.T("intel_delivery_attempts")).Rows(goqu.Record{
		"delivery":   create.Delivery,
		"channel":    create.Channel,
		"created_at": create.CreatedAt.UTC(),
		"is_active":  create.IsActive,
		"status":     create.Status,
		"status_ts":  create.StatusTS.UTC(),
		"note":       create.Note,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return IntelDeliveryAttempt{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return IntelDeliveryAttempt{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return IntelDeliveryAttempt{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return IntelDeliveryAttempt{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	err = rows.Scan(&create.ID)
	if err != nil {
		return IntelDeliveryAttempt{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return create, nil
}

// UpdateIntelDeliveryAttemptStatusByID updates the status of the intel-delivery
// attempt with the given id.
func (m *Mall) UpdateIntelDeliveryAttemptStatusByID(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, newIsActive bool,
	newStatus IntelDeliveryStatus, newNote nulls.String) error {
	q, _, err := m.dialect.Update(goqu.T("intel_delivery_attempts")).Set(goqu.Record{
		"is_active": newIsActive,
		"status":    newStatus,
		"status_ts": time.Now().UTC(),
		"note":      newNote,
	}).Where(goqu.C("id").Eq(attemptID)).ToSQL()
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

// TimedOutIntelDeliveryAttemptsByDelivery retrieves an IntelDeliveryAttempt
// list with all attempts that have timed out.
func (m *Mall) TimedOutIntelDeliveryAttemptsByDelivery(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) ([]IntelDeliveryAttempt, error) {
	q, _, err := m.dialect.From(goqu.T("intel_delivery_attempts")).
		InnerJoin(goqu.T("channels"), goqu.On(goqu.I("channels.id").Eq(goqu.I("intel_delivery_attempts.channel")))).
		Select(goqu.I("intel_delivery_attempts.id"),
			goqu.I("intel_delivery_attempts.delivery"),
			goqu.I("intel_delivery_attempts.channel"),
			goqu.I("intel_delivery_attempts.created_at"),
			goqu.I("intel_delivery_attempts.is_active"),
			goqu.I("intel_delivery_attempts.status"),
			goqu.I("intel_delivery_attempts.status_ts"),
			goqu.I("intel_delivery_attempts.note")).
		Where(goqu.I("intel_delivery_attempts.delivery").Eq(deliveryID),
			goqu.I("intel_delivery_attempts.created_at").Lt(goqu.L("now() - interval '1 ms' * channels.timeout / 1000000"))).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	attempts := make([]IntelDeliveryAttempt, 0)
	for rows.Next() {
		var attempt IntelDeliveryAttempt
		err = rows.Scan(&attempt.ID,
			&attempt.Delivery,
			&attempt.Channel,
			&attempt.CreatedAt,
			&attempt.IsActive,
			&attempt.Status,
			&attempt.StatusTS,
			&attempt.Note)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		attempts = append(attempts, attempt)
	}
	rows.Close()
	return attempts, nil
}

// IntelDeliveryAttemptByID retrieves an IntelDeliveryAttempt by its id.
func (m *Mall) IntelDeliveryAttemptByID(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (IntelDeliveryAttempt, error) {
	q, _, err := m.dialect.From(goqu.T("intel_delivery_attempts")).
		Select(goqu.C("id"),
			goqu.C("delivery"),
			goqu.C("channel"),
			goqu.C("created_at"),
			goqu.C("is_active"),
			goqu.C("status"),
			goqu.C("status_ts"),
			goqu.C("note")).
		Where(goqu.C("id").Eq(attemptID)).ToSQL()
	if err != nil {
		return IntelDeliveryAttempt{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return IntelDeliveryAttempt{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return IntelDeliveryAttempt{}, meh.NewNotFoundErr("not found", nil)
	}
	var attempt IntelDeliveryAttempt
	err = rows.Scan(&attempt.ID,
		&attempt.Delivery,
		&attempt.Channel,
		&attempt.CreatedAt,
		&attempt.IsActive,
		&attempt.Status,
		&attempt.StatusTS,
		&attempt.Note)
	if err != nil {
		return IntelDeliveryAttempt{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	return attempt, nil
}

// NextChannelForDeliveryAttempt retrieves the next channel to use for a
// delivery attempt. Choice is based on available ones, priority and past
// attempts.
func (m *Mall) NextChannelForDeliveryAttempt(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) (Channel, bool, error) {
	q, _, err := m.dialect.From(goqu.T("intel_deliveries")).
		InnerJoin(goqu.T("intel"),
			goqu.On(goqu.I("intel.id").Eq(goqu.I("intel_deliveries.intel")))).
		InnerJoin(goqu.T("channels"),
			goqu.On(goqu.I("channels.entry").Eq(goqu.I("intel_deliveries.to")))).
		LeftJoin(goqu.T("intel_delivery_attempts"),
			goqu.On(goqu.I("intel_delivery_attempts.delivery").Eq(deliveryID),
				goqu.I("intel_delivery_attempts.channel").Eq(goqu.I("channels.id")))).
		Select(goqu.I("channels.id"),
			goqu.I("channels.entry"),
			goqu.I("channels.is_active"),
			goqu.I("channels.label"),
			goqu.I("channels.type"),
			goqu.I("channels.priority"),
			goqu.I("channels.min_importance"),
			goqu.I("channels.timeout")).
		Where(goqu.I("intel_deliveries.id").Eq(deliveryID),
			goqu.I("intel_delivery_attempts.id").IsNull(),
			goqu.I("channels.is_active").IsTrue(),
			goqu.I("intel.importance").Gte(goqu.I("channels.min_importance"))).
		Order(goqu.I("channels.priority").Desc()).
		Limit(1).ToSQL()
	if err != nil {
		return Channel{}, false, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Channel{}, false, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return Channel{}, false, nil
	}
	var channel Channel
	err = rows.Scan(&channel.ID,
		&channel.Entry,
		&channel.IsActive,
		&channel.Label,
		&channel.Type,
		&channel.Priority,
		&channel.MinImportance,
		&channel.Timeout)
	if err != nil {
		return Channel{}, false, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return channel, true, nil
}

// UpdateIntelDeliveryStatusByDelivery updates the status for the delivery with
// the given id.
func (m *Mall) UpdateIntelDeliveryStatusByDelivery(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID, newIsActive bool,
	newSuccess bool, newNote nulls.String) error {
	q, _, err := m.dialect.Update(goqu.T("intel_deliveries")).Set(goqu.Record{
		"is_active": newIsActive,
		"success":   newSuccess,
		"note":      newNote,
	}).Where(goqu.C("id").Eq(deliveryID)).ToSQL()
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
	// Update intel in search.
	delivery, err := m.IntelDeliveryByID(ctx, tx, deliveryID)
	if err != nil {
		return meh.Wrap(err, "updated intel-delivery by id", meh.Details{"delivery_id": deliveryID})
	}
	err = m.addOrUpdateIntelInSearch(ctx, tx, delivery.Intel)
	if err != nil {
		return meh.Wrap(err, "update intel in search", meh.Details{"intel_id": delivery.Intel})
	}
	return nil
}

// ActiveIntelDeliveryAttemptsByDelivery retrieves an IntelDeliveryAttempt list
// with attempts for the delivery with the given id, that are currently active.
func (m *Mall) ActiveIntelDeliveryAttemptsByDelivery(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) ([]IntelDeliveryAttempt, error) {
	attempts, err := m.IntelDeliveryAttemptsByDelivery(ctx, tx, deliveryID)
	if err != nil {
		return nil, meh.Wrap(err, "intel delivery attempts by delivery", meh.Details{"delivery_id": deliveryID})
	}
	activeAttempts := make([]IntelDeliveryAttempt, 0, len(attempts))
	for _, attempt := range attempts {
		if attempt.IsActive {
			activeAttempts = append(activeAttempts, attempt)
		}
	}
	return attempts, nil
}

// IntelDeliveryAttemptsByDelivery retrieves an IntelDeliveryAttempt list with
// attempts for the delivery with the given id.
func (m *Mall) IntelDeliveryAttemptsByDelivery(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) ([]IntelDeliveryAttempt, error) {
	q, _, err := m.dialect.From(goqu.T("intel_delivery_attempts")).
		Select(goqu.C("id"),
			goqu.C("delivery"),
			goqu.C("channel"),
			goqu.C("created_at"),
			goqu.C("is_active"),
			goqu.C("status"),
			goqu.C("status_ts"),
			goqu.C("note")).
		Where(goqu.C("delivery").Eq(deliveryID)).
		Order(goqu.C("created_at").Asc()).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	attempts := make([]IntelDeliveryAttempt, 0)
	for rows.Next() {
		var attempt IntelDeliveryAttempt
		err = rows.Scan(&attempt.ID,
			&attempt.Delivery,
			&attempt.Channel,
			&attempt.CreatedAt,
			&attempt.IsActive,
			&attempt.Status,
			&attempt.StatusTS,
			&attempt.Note)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		attempts = append(attempts, attempt)
	}
	rows.Close()
	return attempts, nil
}

// LockIntelDeliveryByIDOrSkip selects the delivery with the given id in the
// database and locks it. Selection skips locked entries, so if the entry is not
// found or already locked, a meh.ErrNotFound will be returned.
func (m *Mall) LockIntelDeliveryByIDOrSkip(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) error {
	q, _, err := m.dialect.From(goqu.T("intel_deliveries")).
		Select(goqu.C("id")).
		ForUpdate(exp.SkipLocked).
		Where(goqu.C("id").Eq(deliveryID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return meh.NewNotFoundErr("not found or already locked", nil)
	}
	return nil
}

// IntelDeliveryByIDAndLockOrWait retrieves the IntelDelivery with the given id
// and locks it or waits until it is available.
func (m *Mall) IntelDeliveryByIDAndLockOrWait(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) (IntelDelivery, error) {
	q, _, err := m.dialect.From(goqu.T("intel_deliveries")).
		Select(goqu.C("id"),
			goqu.C("intel"),
			goqu.C("to"),
			goqu.C("is_active"),
			goqu.C("success"),
			goqu.C("note")).
		ForUpdate(exp.Wait).
		Where(goqu.C("id").Eq(deliveryID)).ToSQL()
	if err != nil {
		return IntelDelivery{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return IntelDelivery{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return IntelDelivery{}, meh.NewNotFoundErr("not found", nil)
	}
	var delivery IntelDelivery
	err = rows.Scan(&delivery.ID,
		&delivery.Intel,
		&delivery.To,
		&delivery.IsActive,
		&delivery.Success,
		&delivery.Note)
	if err != nil {
		return IntelDelivery{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return delivery, nil
}

// LockIntelDeliveryByIDOrWait locks the intel-delivery in the database with the
// given id or waits until it is available.
func (m *Mall) LockIntelDeliveryByIDOrWait(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) error {
	_, err := m.IntelDeliveryByIDAndLockOrWait(ctx, tx, deliveryID)
	return err
}

// ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait retrieves an
// IntelDeliveryAttempt list where each one is active and uses one of the given
// channels. It locks the associated deliveries as well as the attempts or waits
// until locked.
func (m *Mall) ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait(ctx context.Context, tx pgx.Tx,
	channelIDs []uuid.UUID) ([]IntelDeliveryAttempt, error) {
	if len(channelIDs) == 0 {
		return make([]IntelDeliveryAttempt, 0), nil
	}
	q, _, err := m.dialect.From(goqu.T("intel_delivery_attempts")).
		InnerJoin(goqu.T("intel_deliveries"),
			goqu.On(goqu.I("intel_deliveries.id").Eq(goqu.I("intel_delivery_attempts.delivery")))).
		Select(goqu.I("intel_delivery_attempts.id"),
			goqu.I("intel_delivery_attempts.delivery"),
			goqu.I("intel_delivery_attempts.channel"),
			goqu.I("intel_delivery_attempts.created_at"),
			goqu.I("intel_delivery_attempts.is_active"),
			goqu.I("intel_delivery_attempts.status"),
			goqu.I("intel_delivery_attempts.status_ts"),
			goqu.I("intel_delivery_attempts.note")).
		ForUpdate(exp.Wait).
		Where(goqu.I("intel_delivery_attempts.is_active").IsTrue(),
			goqu.I("intel_delivery_attempts.channel").In(channelIDs)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	attempts := make([]IntelDeliveryAttempt, 0)
	for rows.Next() {
		var attempt IntelDeliveryAttempt
		err = rows.Scan(&attempt.ID,
			&attempt.Delivery,
			&attempt.Channel,
			&attempt.CreatedAt,
			&attempt.IsActive,
			&attempt.Status,
			&attempt.StatusTS,
			&attempt.Note)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		attempts = append(attempts, attempt)
	}
	rows.Close()
	return attempts, nil
}

// DeleteIntelDeliveryAttemptsByChannel deletes all intel-delivery-attempts
// using the channel with the given id.
func (m *Mall) DeleteIntelDeliveryAttemptsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) error {
	q, _, err := m.dialect.Delete(goqu.T("intel_delivery_attempts")).
		Where(goqu.C("channel").Eq(channelID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// ActiveIntelDeliveriesAndLockOrSkip retrieves all active intel-deliveries and
// locks or skips them.
func (m *Mall) ActiveIntelDeliveriesAndLockOrSkip(ctx context.Context, tx pgx.Tx) ([]IntelDelivery, error) {
	q, _, err := m.dialect.From(goqu.T("intel_deliveries")).
		Select(goqu.C("id"),
			goqu.C("intel"),
			goqu.C("to"),
			goqu.C("is_active"),
			goqu.C("success"),
			goqu.C("note")).
		ForUpdate(exp.SkipLocked).
		Where(goqu.C("is_active").IsTrue()).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	deliveries := make([]IntelDelivery, 0)
	for rows.Next() {
		var delivery IntelDelivery
		err = rows.Scan(&delivery.ID,
			&delivery.Intel,
			&delivery.To,
			&delivery.IsActive,
			&delivery.Success,
			&delivery.Note)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		deliveries = append(deliveries, delivery)
	}
	rows.Close()
	return deliveries, nil
}

// ActiveIntelDeliveriesForOperation retrieves the IntelDelivery list for active
// deliveries for intel for the operation with the given id.
func (m *Mall) ActiveIntelDeliveriesForOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) ([]IntelDelivery, error) {
	q, _, err := m.dialect.From(goqu.T("intel_deliveries")).
		InnerJoin(goqu.T("intel"), goqu.On(goqu.I("intel_deliveries.intel").Eq(goqu.I("intel.id")))).
		Select(goqu.I("intel_deliveries.id"),
			goqu.I("intel_deliveries.intel"),
			goqu.I("intel_deliveries.to"),
			goqu.I("intel_deliveries.is_active"),
			goqu.I("intel_deliveries.success"),
			goqu.I("intel_deliveries.note")).
		Where(goqu.I("intel_deliveries.is_active").IsTrue(),
			goqu.I("intel.operation").Eq(operationID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	deliveries := make([]IntelDelivery, 0)
	for rows.Next() {
		var delivery IntelDelivery
		err = rows.Scan(&delivery.ID,
			&delivery.Intel,
			&delivery.To,
			&delivery.IsActive,
			&delivery.Success,
			&delivery.Note)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		deliveries = append(deliveries, delivery)
	}
	rows.Close()
	return deliveries, nil
}
