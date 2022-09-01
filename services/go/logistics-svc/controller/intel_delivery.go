package controller

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
	"time"
)

const periodicDeliveryChecksInterval = 30 * time.Second
const periodicDeliveryChecksDurationWarnThreshold = 1 * time.Second

func (c *Controller) runPeriodicDeliveryChecks(lifetime context.Context) error {
	runCheck := func() error {
		err := pgutil.RunInTx(lifetime, c.DB, func(ctx context.Context, tx pgx.Tx) error {
			activeLockedDeliveries, err := c.Store.ActiveIntelDeliveriesAndLockOrSkip(ctx, tx)
			if err != nil {
				return meh.Wrap(err, "active intel-deliveries from store and lock or skip", nil)
			}
			// Concurrency not possible because of locking and limited connections. If this
			// reveals to be a bottleneck, we need to change this behavior.
			for _, delivery := range activeLockedDeliveries {
				err = c.lookAfterDelivery(ctx, tx, delivery.ID)
				if err != nil {
					return meh.Wrap(err, "look after delivery", meh.Details{"delivery_id": delivery.ID})
				}
			}
			return nil
		})
		if err != nil {
			return meh.Wrap(err, "run in tx", nil)
		}
		return nil
	}
	for {
		start := time.Now()
		err := runCheck()
		if err != nil {
			mehlog.Log(c.Logger, meh.Wrap(err, "run periodic delivery checks", nil))
		} else if took := time.Since(start); took > periodicDeliveryChecksDurationWarnThreshold {
			c.Logger.Warn("periodic delivery checks took longer than expected",
				zap.Duration("took", took),
				zap.Duration("warn_threshold", periodicDeliveryChecksDurationWarnThreshold))
		}
		// Wait.
		select {
		case <-lifetime.Done():
			return nil
		case <-time.After(periodicDeliveryChecksInterval):
		}
	}
}

// scheduleDeliveriesForIntelAssignments schedules intel-deliveries for the
// assignments of the intel with the given id.
func (c *Controller) scheduleDeliveriesForIntelAssignments(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error {
	intel, err := c.Store.IntelByID(ctx, tx, intelID)
	if err != nil {
		return meh.Wrap(err, "intel by id from store", meh.Details{"intel_id": intelID})
	}
	// Create deliveries for all assignments.
	for _, assignment := range intel.Assignments {
		// Create.
		deliveryToCreate := store.IntelDelivery{
			Assignment: assignment.ID,
			IsActive:   true,
			Success:    false,
		}
		createdDelivery, err := c.Store.CreateIntelDelivery(ctx, tx, deliveryToCreate)
		if err != nil {
			return meh.Wrap(err, "create intel-delivery in store", meh.Details{"create": deliveryToCreate})
		}
		err = c.Notifier.NotifyIntelDeliveryCreated(ctx, tx, createdDelivery)
		if err != nil {
			return meh.Wrap(err, "notify intel-delivery created", meh.Details{"created": createdDelivery})
		}
		// Lock and look after.
		err = c.Store.LockIntelDeliveryByIDOrSkip(ctx, tx, createdDelivery.ID)
		if err != nil {
			return meh.Wrap(err, "lock created intel-delivery in store", meh.Details{"delivery_id": deliveryToCreate.ID})
		}
		err = c.lookAfterDelivery(ctx, tx, createdDelivery.ID)
		if err != nil {
			return meh.Wrap(err, "look after created delivery", meh.Details{"delivery_id": createdDelivery.ID})
		}
	}
	return nil
}

// TODO: remember locking delivery when updating attempts (for example email service posts update)

// lookAfterDelivery checks the intel-delivery with the given id. It creates and
// notifies about new attempts as required, timeouts and all other stuff that is
// relevant for the delivery.
//
// Warning: The delivery with the given id is expected to be LOCKED in the store!
func (c *Controller) lookAfterDelivery(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) error {
	// Retrieve basic information.
	delivery, err := c.Store.IntelDeliveryByID(ctx, tx, deliveryID)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "intel delivery from store", meh.Details{"delivery_id": deliveryID})
	}
	if !delivery.IsActive {
		c.Logger.Warn("look after delivery requested although not active. possibly race condition.",
			zap.Any("delivery_id", deliveryID))
		return nil
	}
	// First, we check for timed out attempts.
	err = c.handleTimedOutDeliveryAttempts(ctx, tx, deliveryID)
	if err != nil {
		return meh.Wrap(err, "handle timed out delivery attempts for delivery", meh.Details{"delivery_id": deliveryID})
	}
	// Second, we check if there are still attempts ongoing, as then, we can skip
	// further processing.
	activeAttempts, err := c.Store.ActiveIntelDeliveryAttemptsByDelivery(ctx, tx, deliveryID)
	if err != nil {
		return meh.Wrap(err, "active delivery attempts by delivery from store", meh.Details{"delivery_id": deliveryID})
	}
	if len(activeAttempts) > 0 {
		return nil
	}
	// If no attempts are active anymore, we check for the next channel, that could
	// be used for the next delivery attempt.
	nextChannel, ok, err := c.Store.NextChannelForDeliveryAttempt(ctx, tx, deliveryID)
	if err != nil {
		return meh.Wrap(err, "next channel for delivery attempt from store", meh.Details{"delivery_id": deliveryID})
	}
	if !ok {
		// No more attempts possible. We mark delivery as failed.
		err = c.markDeliveryAsFailed(ctx, tx, deliveryID, "no more channels to try")
		if err != nil {
			return meh.Wrap(err, "mark delivery as failed because of no more attempts possible",
				meh.Details{"delivery_id": deliveryID})
		}
		return nil
	}
	// Create attempt with this channel.
	attemptToCreate := store.IntelDeliveryAttempt{
		Delivery:  deliveryID,
		Channel:   nextChannel.ID,
		CreatedAt: time.Now(),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusOpen,
		StatusTS:  time.Now(),
		Note:      nulls.String{},
	}
	createdAttempt, err := c.Store.CreateIntelDeliveryAttempt(ctx, tx, attemptToCreate)
	if err != nil {
		return meh.Wrap(err, "create intel delivery attempt", meh.Details{"to_create": attemptToCreate})
	}
	err = c.Notifier.NotifyIntelDeliveryAttemptCreated(ctx, tx, createdAttempt)
	if err != nil {
		return meh.Wrap(err, "notify intel delivery attempt created", meh.Details{"created": createdAttempt})
	}
	return nil
}

// markDeliveryAsFailed marks the delivery with the given id as failed and
// notifies about the updated status.
func (c *Controller) markDeliveryAsFailed(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID, reason string) error {
	const newIsActive = false
	const newSuccess = false
	newNote := nulls.NewString(reason)
	err := c.Store.UpdateIntelDeliveryStatusByDelivery(ctx, tx, deliveryID, newIsActive, newSuccess, newNote)
	if err != nil {
		return meh.Wrap(err, "update intel delivery status",
			meh.Details{
				"delivery_id":   deliveryID,
				"new_is_active": newIsActive,
				"new_success":   newSuccess,
			})
	}
	err = c.Notifier.NotifyIntelDeliveryStatusUpdated(ctx, tx, deliveryID, newIsActive, newSuccess, newNote)
	if err != nil {
		return meh.Wrap(err, "notify intel-delivery-status updated", meh.Details{
			"delivery_id":   deliveryID,
			"new_is_active": newIsActive,
			"new_success":   newSuccess,
		})
	}
	return nil
}

// handleTimedOutDeliveryAttempts checks for timed out attempts for the delivery
// with the given id. It is only meant to be used in lookAfterDelivery and kept
// separate for better readability.
func (c *Controller) handleTimedOutDeliveryAttempts(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) error {
	timedOutAttempts, err := c.Store.TimedOutIntelDeliveryAttemptsByDelivery(ctx, tx, deliveryID)
	if err != nil {
		return meh.Wrap(err, "timed out intel-delivery-attempts by delivery from store", meh.Details{"delivery_id": deliveryID})
	}
	for _, timedOutAttempt := range timedOutAttempts {
		if !timedOutAttempt.IsActive {
			// Already handled.
			continue
		}
		// Retrieve the channel for providing better information output.
		channel, err := c.Store.ChannelMetadataByID(ctx, tx, timedOutAttempt.Channel)
		if err != nil {
			return meh.Wrap(err, "channel by id for timed out attempt", meh.Details{"channel_id": timedOutAttempt.Channel})
		}
		// Update status and notify.
		err = c.Store.UpdateIntelDeliveryAttemptStatusByID(ctx, tx, timedOutAttempt.ID, false, store.IntelDeliveryStatusTimeout,
			nulls.NewString(fmt.Sprintf("delivery attempt timed out (%s from channel config)", channel.Timeout.String())))
		if err != nil {
			return meh.Wrap(err, "update status for timed out delivery attempt by id", meh.Details{"attempt_id": timedOutAttempt.ID})
		}
		updatedAttempt, err := c.Store.IntelDeliveryAttemptByID(ctx, tx, timedOutAttempt.ID)
		if err != nil {
			return meh.Wrap(err, "intel delivery attempt by id after status update", meh.Details{"attempt_id": timedOutAttempt.ID})
		}
		err = c.Notifier.NotifyIntelDeliveryAttemptStatusUpdated(ctx, tx, updatedAttempt)
		if err != nil {
			return meh.Wrap(err, "notify intel delivery attempt status updated", meh.Details{"updated_attempt": updatedAttempt})
		}
	}
	return nil
}

// handleDeletedChannelsForDeliveryAttempts cancels, deletes and notifies about
// all delivery attempts that are using the given deleted channels. It then
// returns a list of affected deliveries, that lookAfterDelivery needs to be
// called for after channels have been updated!
func (c *Controller) handleDeletedChannelsForDeliveryAttempts(ctx context.Context, tx pgx.Tx, deletedChannels []store.Channel) ([]uuid.UUID, error) {
	deletedChannelIDs := make([]uuid.UUID, 0, len(deletedChannels))
	for _, deletedChannel := range deletedChannels {
		deletedChannelIDs = append(deletedChannelIDs, deletedChannel.ID)
	}
	affectedDeliveries := make(map[uuid.UUID]struct{})
	affectedActiveAttempts, err := c.Store.ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait(ctx, tx, deletedChannelIDs)
	if err != nil {
		return nil, meh.Wrap(err, "active intel delivery attempts by channel", meh.Details{"channel_ids": deletedChannelIDs})
	}
	for _, activeAttempt := range affectedActiveAttempts {
		// Notify about updated state. We do not need to update in store, as the
		// attempts will be deleted anyways.
		activeAttempt.IsActive = false
		activeAttempt.Status = store.IntelDeliveryStatusCanceled
		activeAttempt.StatusTS = time.Now()
		activeAttempt.Note = nulls.NewString("canceled because of channel deletion")
		err = c.Notifier.NotifyIntelDeliveryAttemptStatusUpdated(ctx, tx, activeAttempt)
		if err != nil {
			return nil, meh.Wrap(err, "notify intel delivery attempt status updated", meh.Details{"attempt": activeAttempt})
		}
		affectedDeliveries[activeAttempt.Delivery] = struct{}{}
	}
	// Delete for all channels.
	for _, deletedChannel := range deletedChannels {
		err = c.Store.DeleteIntelDeliveryAttemptsByChannel(ctx, tx, deletedChannel.ID)
		if err != nil {
			return nil, meh.Wrap(err, "delete intel delivery attempts by channel", meh.Details{"channel_id": deletedChannel.ID})
		}
	}
	// Return a list of all affected deliveries.
	affectedDeliveriesList := make([]uuid.UUID, 0, len(affectedDeliveries))
	for affectedDelivery := range affectedDeliveries {
		affectedDeliveriesList = append(affectedDeliveriesList, affectedDelivery)
	}
	return affectedDeliveriesList, nil
}
