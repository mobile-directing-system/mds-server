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

// scheduleDeliveriesForIntel schedules intel-deliveries for the intel with the
// given id.
func (c *Controller) scheduleDeliveriesForIntel(ctx context.Context, tx pgx.Tx, intelID uuid.UUID, recipientEntries []uuid.UUID) error {
	// Assure intel exists.
	_, err := c.Store.IntelByID(ctx, tx, intelID)
	if err != nil {
		return meh.Wrap(err, "intel by id from store", meh.Details{"intel_id": intelID})
	}
	// Create deliveries.
	for _, entryID := range recipientEntries {
		// Create.
		deliveryToCreate := store.IntelDelivery{
			Intel:    intelID,
			To:       entryID,
			IsActive: true,
			Success:  false,
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
	// No attempts are active anymore, so we check if auto-delivery is enabled for
	// this delivery.
	isAutoDeliveryEnabled, err := c.Store.IsAutoDeliveryEnabledForAddressBookEntry(ctx, tx, delivery.To)
	if err != nil {
		return meh.Wrap(err, "check if auto-delivery enabled for address book entry in store", meh.Details{"entry_id": delivery.To})
	}
	if !isAutoDeliveryEnabled {
		// Nothing to do.
		return nil
	}
	// Check for the next channel, that could be used for the next delivery attempt.
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
	_, err = c.createIntelDeliveryAttempt(ctx, tx, delivery.ID, nextChannel.ID)
	if err != nil {
		return meh.Wrap(err, "create intel delivery attempt", meh.Details{
			"delivery_id":     deliveryID,
			"next_channel_id": nextChannel.ID,
		})
	}
	return nil
}

// createIntelDeliveryAttempt creates and notifies about the given
// store.IntelDeliveryAttempt. If the delivery is inactive, a meh.ErrBadInput
// will be returned. Keep in mind, that we will not check, whether other attempts
// are ongoing/active.
func (c *Controller) createIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID, channelID uuid.UUID) (store.IntelDeliveryAttempt, error) {
	attemptToCreate := store.IntelDeliveryAttempt{
		Delivery:  deliveryID,
		Channel:   channelID,
		CreatedAt: time.Now(),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusOpen,
		StatusTS:  time.Now(),
		Note:      nulls.String{},
	}
	delivery, err := c.Store.IntelDeliveryByID(ctx, tx, deliveryID)
	if err != nil {
		return store.IntelDeliveryAttempt{}, meh.Wrap(err, "intel delivery by id from store", meh.Details{"delivery_id": deliveryID})
	}
	if !delivery.IsActive {
		return store.IntelDeliveryAttempt{}, meh.NewBadInputErr("delivery inactive", meh.Details{"delivery": delivery})
	}
	createdAttempt, err := c.Store.CreateIntelDeliveryAttempt(ctx, tx, attemptToCreate)
	if err != nil {
		return store.IntelDeliveryAttempt{}, meh.Wrap(err, "create intel delivery attempt", meh.Details{"to_create": attemptToCreate})
	}
	intel, err := c.Store.IntelByID(ctx, tx, delivery.Intel)
	if err != nil {
		return store.IntelDeliveryAttempt{}, meh.Wrap(err, "intel by id from store", meh.Details{"intel_id": delivery.Intel})
	}
	assignedEntry, err := c.Store.AddressBookEntryByID(ctx, tx, delivery.To, uuid.NullUUID{})
	if err != nil {
		return store.IntelDeliveryAttempt{}, meh.Wrap(err, "address book entry from store", meh.Details{"entry_id": delivery.To})
	}
	err = c.Notifier.NotifyIntelDeliveryAttemptCreated(ctx, tx, createdAttempt, delivery, assignedEntry, intel)
	if err != nil {
		return store.IntelDeliveryAttempt{}, meh.Wrap(err, "notify intel delivery attempt created", meh.Details{"created": createdAttempt})
	}
	return createdAttempt, nil
}

// CreateIntelDeliveryAttempt schedules a delivery attempt for the delivery with
// the given id using the given channel.
func (c *Controller) CreateIntelDeliveryAttempt(ctx context.Context, deliveryID uuid.UUID, channelID uuid.UUID) (store.IntelDeliveryAttempt, error) {
	var createdAttempt store.IntelDeliveryAttempt
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.LockIntelDeliveryByIDOrWait(ctx, tx, deliveryID)
		if err != nil {
			return meh.Wrap(err, "lock intel-delivery by id or wait", meh.Details{"delivery_id": deliveryID})
		}
		// Check for active attempts as parallel delivery attempts are not allowed.
		activeAttempts, err := c.Store.ActiveIntelDeliveryAttemptsByDelivery(ctx, tx, deliveryID)
		if err != nil {
			return meh.Wrap(err, "active intel delivery attempts by delivery", meh.Details{"delivery_id": deliveryID})
		}
		if len(activeAttempts) > 0 {
			return meh.NewBadInputErr(fmt.Sprintf("%d attempts still active", len(activeAttempts)),
				meh.Details{"active_attempts": len(activeAttempts)})
		}
		// Create.
		createdAttempt, err = c.createIntelDeliveryAttempt(ctx, tx, deliveryID, channelID)
		if err != nil {
			return meh.Wrap(err, "create intel delivery attempt", meh.Details{
				"delivery_id": deliveryID,
				"channel_id":  channelID,
			})
		}
		return nil
	})
	if err != nil {
		return store.IntelDeliveryAttempt{}, meh.Wrap(err, "run in tx", nil)
	}
	return createdAttempt, nil
}

// markDeliveryAsFailed marks the delivery with the given id as failed and
// notifies about the updated status.
//
// Warning: Only call this when there are no more active delivery-attempts as
// this will NOT be checked by markDeliveryAsFailed!
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
		// TODO: notify for manual delviery
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
		return nil, meh.Wrap(err, "active intel-delivery attempts by channel", meh.Details{"channel_ids": deletedChannelIDs})
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

// UpdateIntelDeliveryAttemptStatusForActive updates the
// intel-delivery-attempt-status for the attempt with the given id. It assures
// that the delivery attempt is still active and does not have
// store.IntelDeliveryStatusCanceled. It then notifies via
// Notifier.NotifyIntelDeliveryAttemptStatusUpdated.
func (c *Controller) UpdateIntelDeliveryAttemptStatusForActive(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID,
	newStatus store.IntelDeliveryStatus, newNote nulls.String) error {
	// Retrieve first time in order to find the delivery-id.
	attempt, err := c.Store.IntelDeliveryAttemptByID(ctx, tx, attemptID)
	if err != nil {
		return meh.Wrap(err, "intel-delivery-attempt by id from store", meh.Details{"attempt_id": attemptID})
	}
	// Lock the delivery for manipulation so we can be sure that the status does not
	// change.
	err = c.Store.LockIntelDeliveryByIDOrWait(ctx, tx, attempt.Delivery)
	if err != nil {
		return meh.Wrap(err, "lock intel-delivery by id or wait in store", meh.Details{"delivery_id": attempt.Delivery})
	}
	// Retrieve the attempt again in order to find its updated information that will
	// not change until the delivery is unlocked.
	attempt, err = c.Store.IntelDeliveryAttemptByID(ctx, tx, attemptID)
	if err != nil {
		return meh.Wrap(err, "final intel-delivery-attempt by id from store", meh.Details{"attempt_id": attemptID})
	}
	// Assure still active.
	if !attempt.IsActive || attempt.Status == store.IntelDeliveryStatusCanceled {
		// Already done -> skip the status-update.
		c.Logger.Debug("skipping update for intel-delivery-attempt because of not being active anymore",
			zap.Any("attempt_id", attemptID),
			zap.Any("new_status", newStatus),
			zap.Any("new_note", newNote),
			zap.Any("current_is_active", attempt.IsActive),
			zap.Any("current_status", attempt.Status))
		return nil
	}
	// Update status.
	err = c.Store.UpdateIntelDeliveryAttemptStatusByID(ctx, tx, attemptID, true, newStatus, newNote)
	if err != nil {
		return meh.Wrap(err, "update intel-delivery-attempt status by id in store", meh.Details{
			"attempt_id": attemptID,
			"new_status": newStatus,
			"new_note":   newNote,
		})
	}
	updated, err := c.Store.IntelDeliveryAttemptByID(ctx, tx, attemptID)
	if err != nil {
		return meh.Wrap(err, "retrieve updated intel-delivery-attempt from store", meh.Details{"attempt_id": attemptID})
	}
	// Notify.
	err = c.Notifier.NotifyIntelDeliveryAttemptStatusUpdated(ctx, tx, updated)
	if err != nil {
		return meh.Wrap(err, "notify intel-delivery-attempt-status updated", meh.Details{"updated": updated})
	}
	return nil
}

// MarkIntelDeliveryAttemptAsDelivered calls
// MarkIntelDeliveryAttemptAsDeliveredTX with an opened transaction.
func (c *Controller) MarkIntelDeliveryAttemptAsDelivered(ctx context.Context, attemptID uuid.UUID, by uuid.NullUUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		return c.MarkIntelDeliveryAttemptAsDeliveredTx(ctx, tx, attemptID, by)
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// MarkIntelDeliveryAttemptAsDeliveredTx is a shortcut for
// MarkIntelDeliveryAndAttemptAsDelivered that concludes the delivery id from
// the given attempt id.
func (c *Controller) MarkIntelDeliveryAttemptAsDeliveredTx(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, by uuid.NullUUID) error {
	attempt, err := c.Store.IntelDeliveryAttemptByID(ctx, tx, attemptID)
	if err != nil {
		return meh.Wrap(err, "intel-delivery-attempt by id from store", meh.Details{"attempt_id": attemptID})
	}
	err = c.MarkIntelDeliveryAndAttemptAsDelivered(ctx, tx, attempt.Delivery, nulls.NewUUID(attemptID), by)
	if err != nil {
		return meh.Wrap(err, "mark intel delivery and attempt as delivered", meh.Details{
			"delivery_id": attempt.Delivery,
			"attempt_id":  attemptID,
			"by":          by,
		})
	}
	return nil
}

// MarkIntelDeliveryAsDelivered is a shortcut for
// MarkIntelDeliveryAndAttemptAsDelivered.
func (c *Controller) MarkIntelDeliveryAsDelivered(ctx context.Context, deliveryID uuid.UUID, by uuid.NullUUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.MarkIntelDeliveryAndAttemptAsDelivered(ctx, tx, deliveryID, uuid.NullUUID{}, by)
		if err != nil {
			return meh.Wrap(err, "mark intel delivery and attempt as delivered", meh.Details{
				"delivery_id": deliveryID,
				"by":          by,
			})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// MarkIntelDeliveryAttemptAsFailed marks the intel-delivery-attempt with the
// given id with store.IntelDeliveryStatusFailed, if still being active.
func (c *Controller) MarkIntelDeliveryAttemptAsFailed(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, note nulls.String) error {
	// Lock delivery.
	attempt, err := c.Store.IntelDeliveryAttemptByID(ctx, tx, attemptID)
	if err != nil {
		return meh.Wrap(err, "intel-delivery-attempt by id from store", meh.Details{"attempt_id": attemptID})
	}
	_, err = c.Store.IntelDeliveryByIDAndLockOrWait(ctx, tx, attempt.Delivery)
	if err != nil {
		return meh.Wrap(err, "intel-delivery by id from store", meh.Details{"delivery_id": attempt.Delivery})
	}
	// Assure attempt still active.
	attempt, err = c.Store.IntelDeliveryAttemptByID(ctx, tx, attemptID)
	if err != nil {
		return meh.Wrap(err, "intel-delivery-attempt by id from store (after locked delivery)",
			meh.Details{"attempt_id": attemptID})
	}
	if !attempt.IsActive {
		c.Logger.Debug("skipping marking intel-delivery-attempt as failed due to not being active anymore",
			zap.Any("attempt_id", attemptID))
		return nil
	}
	// Mark as failed.
	newStatus := store.IntelDeliveryStatusFailed
	newNote := note
	err = c.Store.UpdateIntelDeliveryAttemptStatusByID(ctx, tx, attemptID, false, newStatus, newNote)
	if err != nil {
		return meh.Wrap(err, "update intel-delivery-attempt-status by id in store", meh.Details{"attempt_id": attemptID})
	}
	// Retrieve updated.
	attempt, err = c.Store.IntelDeliveryAttemptByID(ctx, tx, attemptID)
	if err != nil {
		return meh.Wrap(err, "retrieve updated intel-delivery-attempt from store", meh.Details{"attempt_id": attemptID})
	}
	// Notify.
	err = c.Notifier.NotifyIntelDeliveryAttemptStatusUpdated(ctx, tx, attempt)
	if err != nil {
		return meh.Wrap(err, "notify intel-deliver-attempt-status updated", meh.Details{"updated_attempt": attempt})
	}
	err = c.lookAfterDelivery(ctx, tx, attempt.Delivery)
	if err != nil {
		return meh.Wrap(err, "look after delivery", meh.Details{"delivery_id": attempt.Delivery})
	}
	return nil
}

// MarkIntelDeliveryAndAttemptAsDelivered marks the intel-delivery with the
// given id as delivered. If the given by-user is set, we also check whether the
// user is the assigned one. If the attempt id is provided, the attempt will be
// marked as successful. All other active attempts will be marked as canceled.
func (c *Controller) MarkIntelDeliveryAndAttemptAsDelivered(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID,
	attemptID uuid.NullUUID, by uuid.NullUUID) error {
	delivery, err := c.Store.IntelDeliveryByIDAndLockOrWait(ctx, tx, deliveryID)
	if err != nil {
		return meh.Wrap(err, "intel-delivery by id from store", meh.Details{"delivery_id": deliveryID})
	}
	// Assure assigned to given user if set.
	if by.Valid {
		if by.UUID != delivery.To {
			return meh.NewForbiddenErr("intel-delivery assigned to different user", meh.Details{
				"mark_by":              by.UUID,
				"delivery_assigned_to": delivery.To,
			})
		}
	}
	// Cancel all ongoing attempts, except for the the one with the given id.
	activeAttempts, err := c.Store.ActiveIntelDeliveryAttemptsByDelivery(ctx, tx, deliveryID)
	if err != nil {
		return meh.Wrap(err, "active intel-delivery-attempts by delivery from store", meh.Details{"delivery_id": deliveryID})
	}
	for _, attempt := range activeAttempts {
		newStatus := store.IntelDeliveryStatusCanceled
		newNote := nulls.NewString("canceled due to manual delivery-confirmation")
		if attemptID.Valid && attempt.ID == attemptID.UUID {
			newStatus = store.IntelDeliveryStatusDelivered
			newNote = nulls.String{}
		}
		err = c.Store.UpdateIntelDeliveryAttemptStatusByID(ctx, tx, attempt.ID, false, newStatus, newNote)
		if err != nil {
			return meh.Wrap(err, "update intel-delivery-attempt status by id", meh.Details{"attempt_id": attempt.ID})
		}
		updatedAttempt, err := c.Store.IntelDeliveryAttemptByID(ctx, tx, attempt.ID)
		if err != nil {
			return meh.Wrap(err, "updated intel-delivery-attemt by id", meh.Details{"attempt_id": attempt.ID})
		}
		err = c.Notifier.NotifyIntelDeliveryAttemptStatusUpdated(ctx, tx, updatedAttempt)
		if err != nil {
			return meh.Wrap(err, "notify about updated intel-delivery-attempt", meh.Details{"updated": updatedAttempt})
		}
	}
	// Mark delivery as delivered and notify.
	const newDeliveryIsActive = false
	const newDeliverySuccess = true
	newNote := nulls.NewString("delivered")
	err = c.Store.UpdateIntelDeliveryStatusByDelivery(ctx, tx, deliveryID, newDeliveryIsActive, newDeliverySuccess, newNote)
	if err != nil {
		return meh.Wrap(err, "update intel-delivery-status in store", meh.Details{"delivery_id": deliveryID})
	}
	err = c.Notifier.NotifyIntelDeliveryStatusUpdated(ctx, tx, deliveryID, newDeliveryIsActive, newDeliverySuccess, newNote)
	if err != nil {
		return meh.Wrap(err, "notify intel-delivery-status updated", meh.Details{"delivery_id": deliveryID})
	}
	return nil
}

// IntelDeliveryAttemptsByDelivery retrieves a store.IntelDeliveryAttempt list
// with attempts for the delivery with the given id.
func (c *Controller) IntelDeliveryAttemptsByDelivery(ctx context.Context, deliveryID uuid.UUID) ([]store.IntelDeliveryAttempt, error) {
	var attempts []store.IntelDeliveryAttempt
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		attempts, err = c.Store.IntelDeliveryAttemptsByDelivery(ctx, tx, deliveryID)
		if err != nil {
			return meh.Wrap(err, "intel delivery attempts by delivery", meh.Details{"delivery_id": deliveryID})
		}
		return nil
	})
	if err != nil {
		return nil, meh.Wrap(err, "run in tx", nil)
	}
	return attempts, nil
}

// TODO: notify on intel delivery creation/updates for manual delivery
