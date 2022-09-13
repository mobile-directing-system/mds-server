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

// AcceptedIntelDeliveryAttempt is an attempt for intel-delivery for a specific
// channel.
type AcceptedIntelDeliveryAttempt struct {
	// ID identifies the attempt.
	ID uuid.UUID
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

// CreateIntelNotificationHistoryEntry creates an entry in the history-table for
// keeping log of sent notifications for attempts.
func (m *Mall) CreateIntelNotificationHistoryEntry(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, ts time.Time) error {
	q, _, err := m.dialect.Insert(goqu.T("intel_notification_history")).Rows(goqu.Record{
		"attempt": attemptID,
		"ts":      ts.UTC(),
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

// OutgoingIntelDeliveryNotification contains all relevant information for
// sending a notification to the user.
type OutgoingIntelDeliveryNotification struct {
	// IntelToDeliver is the intel that needs to be delivered to the user.
	IntelToDeliver IntelToDeliver
	// DeliveryAttempt contains information regarding the delivery-attempt. This is
	// also used for confirming delivery.
	DeliveryAttempt AcceptedIntelDeliveryAttempt
	// Channel holds details regarding the channel to use for delivery.
	Channel NotificationChannel
	// CreatorDetails holds user information for the creator.
	CreatorDetails User
	// RecipientDetails holds the user information for the optionally assigned
	// recipient (from the address book entry).
	RecipientDetails nulls.JSONNullable[User]
}

// OldestPendingAttemptToNotifyByUser retrieves the id of the oldest attempt
// that needs notification to the user with the given id.
//
// This is meant to be used when the user makes a connection in order to send
// all pending notifications.
func (m *Mall) OldestPendingAttemptToNotifyByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (uuid.UUID, bool, error) {
	q, _, err := m.dialect.From(goqu.T("accepted_intel_delivery_attempts").As("attempts")).
		Select(goqu.I("attempts.id")).
		Where(goqu.I("attempts.is_active").IsTrue(),
			goqu.I("attempts.assigned_to_user").Eq(userID),
			goqu.I("attempts.id").NotIn(
				// Exclude attempts, that notifications where already sent for.
				m.dialect.From(goqu.T("intel_notification_history").As("history")).
					Select(goqu.I("history.attempt")).
					Where(goqu.I("history.attempt").Eq(goqu.I("attempts.id"))))).
		Order(goqu.I("attempts.created_at").Asc()).
		Limit(1).ToSQL()
	if err != nil {
		return uuid.Nil, false, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return uuid.Nil, false, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return uuid.Nil, false, nil
	}
	var attemptID uuid.UUID
	err = rows.Scan(&attemptID)
	if err != nil {
		return uuid.Nil, false, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	return attemptID, true, nil
}

// OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip retrieves the
// OutgoingIntelDeliveryNotification and the delivery-attempt for update or
// skips it. This means, that when the returned error is meh.ErrNotFound, it
// might also me locked by someone else. It is also assured that the attempt has
// no tries in the notification history. This is required if someone else
// already processed this attempt.
func (m *Mall) OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip(ctx context.Context, tx pgx.Tx,
	attemptID uuid.UUID) (OutgoingIntelDeliveryNotification, error) {
	var notif OutgoingIntelDeliveryNotification
	// Find attempts to send notifications for.
	attemptQuery, _, err := m.dialect.From(goqu.T("accepted_intel_delivery_attempts")).
		Select(goqu.C("id"),
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
		ForUpdate(exp.SkipLocked).
		Where(goqu.C("id").Eq(attemptID),
			goqu.C("is_active").IsTrue(),
			goqu.C("id").NotIn(
				// Retrieve entries in history for the attempt.
				m.dialect.From(goqu.T("intel_notification_history").As("history")).
					Select(goqu.I("history.attempt")).
					Where(goqu.I("history.attempt").Eq(attemptID)))).ToSQL()
	if err != nil {
		return OutgoingIntelDeliveryNotification{}, meh.NewInternalErrFromErr(err, "attempt-query to sql", nil)
	}
	rows, err := tx.Query(ctx, attemptQuery)
	if err != nil {
		return OutgoingIntelDeliveryNotification{}, mehpg.NewQueryDBErr(err, "exec attempt-query", attemptQuery)
	}
	defer rows.Close()
	if !rows.Next() {
		return OutgoingIntelDeliveryNotification{}, meh.NewNotFoundErr("not found", nil)
	}
	err = rows.Scan(&notif.DeliveryAttempt.ID,
		&notif.DeliveryAttempt.AssignedTo,
		&notif.DeliveryAttempt.AssignedToLabel,
		&notif.DeliveryAttempt.AssignedToUser,
		&notif.DeliveryAttempt.Delivery,
		&notif.DeliveryAttempt.Channel,
		&notif.DeliveryAttempt.CreatedAt,
		&notif.DeliveryAttempt.IsActive,
		&notif.DeliveryAttempt.StatusTS,
		&notif.DeliveryAttempt.Note,
		&notif.DeliveryAttempt.AcceptedAt)
	if err != nil {
		return OutgoingIntelDeliveryNotification{}, mehpg.NewScanRowsErr(err, "scan row", attemptQuery)
	}
	rows.Close()
	// Retrieve the intel.
	notif.IntelToDeliver, err = m.IntelToDeliverByAttempt(ctx, tx, attemptID)
	if err != nil {
		err = meh.ApplyCode(err, meh.ErrInternal)
		err = meh.Wrap(err, "intel to deliver by attempt", meh.Details{"attempt_id": attemptID})
		return OutgoingIntelDeliveryNotification{}, err
	}
	// Retrieve the channel.
	notif.Channel, err = m.NotificationChannelByID(ctx, tx, notif.DeliveryAttempt.Channel)
	if err != nil {
		err = meh.ApplyCode(err, meh.ErrInternal)
		err = meh.Wrap(err, "notification channel by id", meh.Details{"channel_id": notif.DeliveryAttempt.Channel})
		return OutgoingIntelDeliveryNotification{}, err
	}
	// Retrieve user information.
	notif.CreatorDetails, err = m.UserByID(ctx, tx, notif.IntelToDeliver.CreatedBy)
	if err != nil {
		err = meh.ApplyCode(err, meh.ErrInternal)
		err = meh.Wrap(err, "retrieve user details for creator", meh.Details{"user_id": notif.IntelToDeliver.CreatedBy})
		return OutgoingIntelDeliveryNotification{}, err
	}
	if notif.DeliveryAttempt.AssignedToUser.Valid {
		recipientDetails, err := m.UserByID(ctx, tx, notif.DeliveryAttempt.AssignedToUser.UUID)
		if err != nil {
			err = meh.ApplyCode(err, meh.ErrInternal)
			err = meh.Wrap(err, "retrieve user details for recipient", meh.Details{"user_id": notif.DeliveryAttempt.AssignedTo})
			return OutgoingIntelDeliveryNotification{}, err
		}
		notif.RecipientDetails = nulls.NewJSONNullable(recipientDetails)
	}
	return notif, nil
}
