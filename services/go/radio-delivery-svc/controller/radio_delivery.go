package controller

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"sort"
	"time"
)

// radioDeliveryTimeoutCheckInterval is the check-interval for
// Controller.runPeriodicRadioDeliveryTimeoutChecks.
const radioDeliveryTimeoutCheckInterval = 10 * time.Second

// PickUpNextRadioDelivery selects the next available radio delivery for the
// given operation by the user.
func (c *Controller) PickUpNextRadioDelivery(ctx context.Context, operationID uuid.UUID,
	by uuid.UUID) (store.AcceptedIntelDeliveryAttempt, bool, error) {
	var attempt store.AcceptedIntelDeliveryAttempt
	var ok bool
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		// Assure user member of operation.
		assignedOperations, err := c.store.OperationsByMember(ctx, tx, by)
		if err != nil {
			return meh.Wrap(err, "operations by member", meh.Details{"member_id": by})
		}
		isMemberOfOperation := false
		for _, assignedOperation := range assignedOperations {
			if assignedOperation == operationID {
				isMemberOfOperation = true
				break
			}
		}
		if !isMemberOfOperation {
			return meh.NewForbiddenErr("user not member of operation", nil)
		}
		// Retrieve active deliveries.
		activeDeliveries, err := c.store.ActiveRadioDeliveriesAndLockOrWait(ctx, tx, nulls.NewUUID(operationID))
		if err != nil {
			return meh.Wrap(err, "active radio deliveries from store", nil)
		}
		openDeliveries := make([]store.ActiveRadioDelivery, 0)
		for _, activeDelivery := range activeDeliveries {
			if activeDelivery.PickedUpAt.Valid {
				continue
			}
			openDeliveries = append(openDeliveries, activeDelivery)
		}
		if len(openDeliveries) == 0 {
			attempt = store.AcceptedIntelDeliveryAttempt{}
			ok = false
			return nil
		}
		// Choose priority-mode.
		inProgressDeliveryCount := len(activeDeliveries) - len(openDeliveries)
		var lowToHigh func(i, j int) bool
		if inProgressDeliveryCount%2 == 0 {
			// Use importance.
			lowToHigh = func(i, j int) bool {
				return openDeliveries[i].IntelImportance < openDeliveries[j].IntelImportance
			}
		} else {
			// Use created-date.
			lowToHigh = func(i, j int) bool {
				return openDeliveries[i].AttemptCreatedAt.After(openDeliveries[j].AttemptCreatedAt)
			}
		}
		sort.Slice(openDeliveries, lowToHigh)
		attemptID := openDeliveries[len(openDeliveries)-1].Attempt
		ok = true
		// Pick up.
		err = c.store.MarkRadioDeliveryAsPickedUpByAttempt(ctx, tx, attemptID, nulls.NewUUID(by), "picked up")
		if err != nil {
			return meh.Wrap(err, "mark radio delivery as picked up in store", meh.Details{
				"attempt_id": attemptID,
				"by":         by,
			})
		}
		attempt, err = c.store.AcceptedIntelDeliveryAttemptByID(ctx, tx, attemptID)
		if err != nil {
			return meh.Wrap(err, "retrieve picked up intel delivery attempt from store", meh.Details{"attempt_id": attemptID})
		}
		updatedRadioDelivery, err := c.store.RadioDeliveryByAttempt(ctx, tx, attemptID)
		if err != nil {
			return meh.Wrap(err, "updated radio delivery from store", meh.Details{"attempt_id": attemptID})
		}
		// Notify.
		err = c.notifier.NotifyRadioDeliveryPickedUp(ctx, tx, attemptID, by, updatedRadioDelivery.PickedUpAt.Time)
		if err != nil {
			return meh.Wrap(err, "notify", meh.Details{
				"attempt_id":   attemptID,
				"by":           by,
				"picked_up_at": updatedRadioDelivery.PickedUpAt.Time,
			})
		}
		return nil
	})
	if err != nil {
		return store.AcceptedIntelDeliveryAttempt{}, false, meh.Wrap(err, "run in tx", nil)
	}
	return attempt, ok, nil
}

// ReleasePickedUpRadioDelivery releases the picked up delivery for the attempt
// with the given id. If limit is set, the delivery must be picked up by the
// user with the given id.
func (c *Controller) ReleasePickedUpRadioDelivery(ctx context.Context, attemptID uuid.UUID, limitToPickedUpBy uuid.NullUUID) error {
	var notifyForOperation uuid.NullUUID
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		delivery, err := c.store.RadioDeliveryByAttempt(ctx, tx, attemptID)
		if err != nil {
			return meh.Wrap(err, "radio delivery by attempt from store", meh.Details{"attempt_id": attemptID})
		}
		attempt, err := c.store.AcceptedIntelDeliveryAttemptByID(ctx, tx, attemptID)
		if err != nil {
			return meh.Wrap(err, "accepted intel-delivery-attempt from store", meh.Details{"attempt_id": attemptID})
		}
		// Assure release ok.
		if delivery.Success.Valid {
			return meh.NewBadInputErr("delivery not active", meh.Details{
				"delivery_success":    delivery.Success.Bool,
				"delivery_success_ts": delivery.SuccessTS,
			})
		}
		if !delivery.PickedUpBy.Valid {
			return meh.NewBadInputErr("delivery not picked up", nil)
		}
		if limitToPickedUpBy.Valid && delivery.PickedUpBy.UUID != limitToPickedUpBy.UUID {
			return meh.NewForbiddenErr("delivery not picked up by user", meh.Details{
				"picked_up_by":       delivery.PickedUpBy.UUID,
				"release_limited_to": limitToPickedUpBy.UUID,
			})
		}
		// Release.
		err = c.store.MarkRadioDeliveryAsPickedUpByAttempt(ctx, tx, attemptID, uuid.NullUUID{}, "waiting for pickup (released)")
		if err != nil {
			return meh.Wrap(err, "unmark radio delivery as picked up in store", meh.Details{"attempt": attemptID})
		}
		// Notify.
		err = c.notifier.NotifyRadioDeliveryReleased(ctx, tx, attemptID, time.Now())
		if err != nil {
			return meh.Wrap(err, "notify", nil)
		}
		notifyForOperation = nulls.NewUUID(attempt.IntelOperation)
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	if notifyForOperation.Valid {
		c.connUpdateNotifier.scheduleNotifyUpdatesForOperations(ctx, notifyForOperation.UUID)
	}
	return nil
}

// FinishRadioDelivery finishes the picked up radio delivery for the attempt
// with the given id. The success-state and note are updated to the given ones.
// If the limit is set, the delivery is assured to be picked up by the user with
// the given id.
func (c *Controller) FinishRadioDelivery(ctx context.Context, attemptID uuid.UUID, success bool, note string,
	limitToPickedUpBy uuid.NullUUID) error {
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		// Assure picked up.
		radioDelivery, err := c.store.RadioDeliveryByAttempt(ctx, tx, attemptID)
		if err != nil {
			return meh.Wrap(err, "radio delivery by attempt from store", meh.Details{"attempt_id": attemptID})
		}
		if !radioDelivery.PickedUpBy.Valid {
			return meh.NewBadInputErr("radio delivery not picked up", nil)
		}
		// Check limit.
		if limitToPickedUpBy.Valid && radioDelivery.PickedUpBy.UUID != limitToPickedUpBy.UUID {
			return meh.NewForbiddenErr("not picked up by this user", meh.Details{
				"picked_up_by":          radioDelivery.PickedUpBy.UUID,
				"limit_to_picked_up_by": limitToPickedUpBy.UUID,
			})
		}
		// Update.
		newSuccess := nulls.NewBool(success)
		err = c.store.UpdateRadioDeliveryStatusByAttempt(ctx, tx, attemptID, newSuccess, note)
		if err != nil {
			return meh.Wrap(err, "update radio delivery status by attempt in store", meh.Details{
				"attempt_id":  attemptID,
				"new_success": newSuccess,
				"new_note":    note,
			})
		}
		updatedRadioDelivery, err := c.store.RadioDeliveryByAttempt(ctx, tx, attemptID)
		if err != nil {
			return meh.Wrap(err, "updated radio delivery from store", meh.Details{"attempt_id": attemptID})
		}
		// Notify.
		err = c.notifier.NotifyRadioDeliveryFinished(ctx, tx, updatedRadioDelivery)
		if err != nil {
			return meh.Wrap(err, "notify", meh.Details{"updated_radio_delivery": updatedRadioDelivery})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// runPeriodicRadioDeliveryTimeoutChecks periodically checks active picked up
// radio deliveries for being timed out until the given context is done. Check
// interval is taken from radioDeliveryTimeoutCheckInterval.
func (c *Controller) runPeriodicRadioDeliveryTimeoutChecks(lifetime context.Context) {
	for {
		err := c.releaseTimedOutRadioDeliveries(lifetime)
		if err != nil {
			mehlog.Log(c.logger, meh.Wrap(err, "run periodic radio delivery timeout check", nil))
		}
		select {
		case <-lifetime.Done():
			return
		case <-time.After(radioDeliveryTimeoutCheckInterval):
		}
	}
}

// releaseTimedOutRadioDeliveries checks all active deliveries for being timed
// out (based on pickedUpTimeout), releases and schedules update notifications
// accordingly and notifies.
func (c *Controller) releaseTimedOutRadioDeliveries(ctx context.Context) error {
	affectedOperations := make(map[uuid.UUID]struct{})
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		activeDeliveries, err := c.store.ActiveRadioDeliveriesAndLockOrWait(ctx, tx, uuid.NullUUID{})
		if err != nil {
			return meh.Wrap(err, "active radio deliveries from store", nil)
		}
		// Check each delivery for timeout.
		for _, activeDelivery := range activeDeliveries {
			if !activeDelivery.PickedUpAt.Valid {
				continue
			}
			if time.Since(activeDelivery.PickedUpAt.Time) < c.pickedUpTimeout {
				continue
			}
			// Release.
			newNote := fmt.Sprintf("timeout (%s) while being picked up", c.pickedUpTimeout.String())
			err := c.store.MarkRadioDeliveryAsPickedUpByAttempt(ctx, tx, activeDelivery.Attempt, uuid.NullUUID{}, newNote)
			if err != nil {
				return meh.Wrap(err, "(un)mark radio delivery as picked up by attempt in store", meh.Details{
					"attempt_id":        activeDelivery.Attempt,
					"picked_up_at":      activeDelivery.PickedUpAt.Time,
					"picked_up_timeout": c.pickedUpTimeout.String(),
				})
			}
			affectedOperations[activeDelivery.IntelOperation] = struct{}{}
			updatedIntelDeliveryAttempt, err := c.store.AcceptedIntelDeliveryAttemptByID(ctx, tx, activeDelivery.Attempt)
			if err != nil {
				return meh.Wrap(err, "updated intel-delivery-attempt from store",
					meh.Details{"attempt_id": activeDelivery.Attempt})
			}
			// Notify.
			err = c.notifier.NotifyRadioDeliveryReadyForPickup(ctx, tx, updatedIntelDeliveryAttempt, newNote)
			if err != nil {
				return meh.Wrap(err, "notify about timed out radio delivery being available for pickup again",
					meh.Details{"intel_delivery_attempt": updatedIntelDeliveryAttempt})
			}
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	// Notify about updated operations.
	operations := make([]uuid.UUID, 0, len(affectedOperations))
	for operation := range affectedOperations {
		operations = append(operations, operation)
	}
	c.connUpdateNotifier.scheduleNotifyUpdatesForOperations(ctx, operations...)
	return nil
}
