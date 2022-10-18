package controller

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"time"
)

// CreateIntelDeliveryAttempt creates the given
// store.AcceptedIntelDeliveryAttempt, radio-delivery and notifies. If the
// referenced radio channel cannot be found, no attempt will be created.
func (c *Controller) CreateIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, attempt store.AcceptedIntelDeliveryAttempt) error {
	_, err := c.store.RadioChannelByID(ctx, tx, attempt.Channel)
	if err != nil {
		if meh.ErrorCode(err) == meh.ErrNotFound {
			// Ignore.
			return nil
		}
		return meh.Wrap(err, "radio channel by id from store", meh.Details{"channel_id": attempt.Channel})
	}
	attempt.AcceptedAt = time.Now()
	// Add to store.
	err = c.store.CreateAcceptedIntelDeliveryAttempt(ctx, tx, attempt)
	if err != nil {
		return meh.Wrap(err, "create accepted intel delivery attempt", nil)
	}
	err = c.store.CreateRadioDelivery(ctx, tx, attempt.ID)
	if err != nil {
		return meh.Wrap(err, "create radio delivery", nil)
	}
	createdRadioDelivery, err := c.store.RadioDeliveryByAttempt(ctx, tx, attempt.ID)
	if err != nil {
		return meh.Wrap(err, "retrieve created radio delivery from store", meh.Details{"attempt_id": attempt.ID})
	}
	// Notify about accepted attempt.
	err = c.notifier.NotifyRadioDeliveryReadyForPickup(ctx, tx, attempt, createdRadioDelivery.Note)
	if err != nil {
		return meh.Wrap(err, "notify radio delivery ready for pickup", meh.Details{"attempt_id": attempt.Delivery})
	}
	c.connUpdateNotifier.scheduleNotifyUpdatesForOperations(ctx, attempt.IntelOperation)
	return nil
}

// UpdateIntelDeliveryAttemptStatus updates the status for the given
// intel-delivery-attempt. If it is not active anymore but has an associated
// active radio delivery, it will be marked as failed.
func (c *Controller) UpdateIntelDeliveryAttemptStatus(ctx context.Context, tx pgx.Tx, newStatus store.AcceptedIntelDeliveryAttemptStatus) error {
	// Update attempt.
	err := c.store.UpdateAcceptedIntelDeliveryAttemptStatus(ctx, tx, newStatus)
	if err != nil {
		if meh.ErrorCode(err) == meh.ErrNotFound {
			// Skip.
			return nil
		}
		return meh.Wrap(err, "update accepted intel delivery attempt status in store", meh.Details{"new_status": newStatus})
	}
	// Only update radio delivery, if attempt is not active anymore and discard
	// other changes.
	if newStatus.IsActive {
		// No changes for radio deliveries -> skip.
		return nil
	}
	// Attempt not active anymore. Update radio delivery if still active.
	radioDelivery, err := c.store.RadioDeliveryByAttempt(ctx, tx, newStatus.ID)
	if err != nil {
		return meh.Wrap(err, "radio delivery for attempt from store", meh.Details{"id": newStatus.ID})
	}
	if radioDelivery.Success.Valid {
		// Inactive -> skip.
		return nil
	}
	err = c.store.UpdateRadioDeliveryStatusByAttempt(ctx, tx, newStatus.ID, nulls.NewBool(false), "attempt not active anymore")
	if err != nil {
		return meh.Wrap(err, "update radio delivery status by attempt", meh.Details{"attempt_id": newStatus.ID})
	}
	// Retrieve updated.
	radioDelivery, err = c.store.RadioDeliveryByAttempt(ctx, tx, newStatus.ID)
	if err != nil {
		return meh.Wrap(err, "retrieve updated radio delivery", meh.Details{"attempt_id": newStatus.ID})
	}
	// Notify.
	err = c.notifier.NotifyRadioDeliveryFinished(ctx, tx, radioDelivery)
	if err != nil {
		return meh.Wrap(err, "notify about finished radio delivery", meh.Details{"radio_delivery": radioDelivery})
	}
	return nil
}
