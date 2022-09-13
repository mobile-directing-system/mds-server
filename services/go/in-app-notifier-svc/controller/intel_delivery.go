package controller

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"time"
)

// CreateIntelDeliveryAttempt handles a new intel-delivery-attempt. It checks,
// whether the channel is supported by us and accepts it then. Otherwise, it is
// ignored.
func (c *Controller) CreateIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx,
	attempt store.AcceptedIntelDeliveryAttempt, intelToDeliver store.IntelToDeliver) error {
	_, err := c.store.NotificationChannelByID(ctx, tx, attempt.Channel)
	if err != nil {
		if meh.ErrorCode(err) == meh.ErrNotFound {
			// Ignore.
			return nil
		}
		return meh.Wrap(err, "notification channel by id from store", meh.Details{"channel_id": attempt.Channel})
	}
	attempt.AcceptedAt = time.Now()
	// Add to store.
	err = c.store.CreateAcceptedIntelDeliveryAttempt(ctx, tx, attempt)
	if err != nil {
		return meh.Wrap(err, "create accepted intel delivery attempt", nil)
	}
	err = c.store.CreateIntelToDeliver(ctx, tx, intelToDeliver)
	if err != nil {
		return meh.Wrap(err, "create intel-to-deliver in store", meh.Details{"intel_to_deliver": intelToDeliver})
	}
	// Notify about accepted attempt.
	err = c.notifier.NotifyIntelDeliveryNotificationPending(ctx, tx, attempt.ID, attempt.AcceptedAt)
	if err != nil {
		return meh.Wrap(err, "notify intel-delivery-notification pending", meh.Details{"attempt_id": attempt.Delivery})
	}
	// Schedule looking after the assigned user.
	if attempt.AssignedToUser.Valid {
		err = c.scheduleLookAfterUserNotifications(ctx, attempt.AssignedToUser.UUID)
		if err != nil {
			return meh.Wrap(err, "schedule look after user-notifications", meh.Details{"user_id": attempt.AssignedTo})
		}
	}
	return nil
}

// UpdateIntelDeliveryAttemptStatus updates the intel-delivery-attempt-status
// for the associated intel-delivery-attempt. The intel-delivery-attempt does
// not need to be accepted, as we ignore meh.ErrNotFound errors.
func (c *Controller) UpdateIntelDeliveryAttemptStatus(ctx context.Context, tx pgx.Tx, newStatus store.AcceptedIntelDeliveryAttemptStatus) error {
	err := c.store.UpdateAcceptedIntelDeliveryAttemptStatus(ctx, tx, newStatus)
	if err != nil && meh.ErrorCode(err) != meh.ErrNotFound {
		return meh.Wrap(err, "update attempt status in store", meh.Details{"new_status": newStatus})
	}
	return nil
}
