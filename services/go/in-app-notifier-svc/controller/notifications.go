package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

// periodicNotificationCheckInterval is the interval in which to perform
// notification checks for all connected users in
// Controller.runPeriodicNotificationChecks.
const periodicNotificationCheckInterval = 16 * time.Second

func (c *Controller) runPeriodicNotificationChecks(lifetime context.Context) error {
	for {
		err := c.scheduleLookAfterAllUserNotifications(lifetime)
		if err != nil {
			err = meh.Wrap(err, "schedule look after all user notifications in periodic notification checks", nil)
			mehlog.Log(c.logger, err)
		}
		select {
		case <-lifetime.Done():
			return nil
		case <-time.After(periodicNotificationCheckInterval):
		}
	}
}

// scheduleLookAfterAllUserNotifications calls
// scheduleLookAfterUserNotifications for all currently connected users.
func (c *Controller) scheduleLookAfterAllUserNotifications(ctx context.Context) error {
	c.connectionsByUserMutex.RLock()
	defer c.connectionsByUserMutex.RUnlock()
	for userID := range c.connectionsByUser {
		err := c.scheduleLookAfterUserNotifications(ctx, userID)
		if err != nil {
			return meh.Wrap(err, "schedule look after user notifications", meh.Details{"user_id": userID})
		}
	}
	return nil
}

// scheduleLookAfterUserNotifications schedules a check for user notifications
// for the user with the given id.
func (c *Controller) scheduleLookAfterUserNotifications(ctx context.Context, userID uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.lookAfterUserNotificationRequests <- userID:
	}
	return nil
}

const notificationSchedulerWorkers = 64

// runLookAfterUserNotificationsScheduler runs workers that handle requests from
// lookAfterUserNotificationRequests.
func (c *Controller) runLookAfterUserNotificationsScheduler(lifetime context.Context) error {
	logger := c.logger.Named("look-after-user-notifs-scheduler")
	worker := func(lifetime context.Context) error {
		for {
			var request uuid.UUID
			// Wait for next request.
			select {
			case <-lifetime.Done():
				return nil
			case request = <-c.lookAfterUserNotificationRequests:
			}
			// Handle.
			err := c.lookAfterUserNotifications(lifetime, request)
			if err != nil {
				mehlog.Log(logger, meh.Wrap(err, "look after user notifications", meh.Details{"user_id": request}))
				continue
			}
		}
	}
	// Launch workers.
	eg, egCtx := errgroup.WithContext(lifetime)
	for i := 0; i < notificationSchedulerWorkers; i++ {
		eg.Go(func() error {
			return worker(egCtx)
		})
	}
	return eg.Wait()
}

// lookAfterUserNotifications checks for and sends notifications for pending
// ones for the user with the given id.
//
// Warning: You normally should not call this manually, except in
// runLookAfterUserNotificationsScheduler, as this is done by workers
// concurrently by the Controller. Instead, use
// scheduleLookAfterUserNotifications.
func (c *Controller) lookAfterUserNotifications(ctx context.Context, userID uuid.UUID) error {
	c.connectionsByUserMutex.RLock()
	defer c.connectionsByUserMutex.RUnlock()
	userConns, ok := c.connectionsByUser[userID]
	if !ok || len(userConns) == 0 {
		// No connections -> nothing to do.
		return nil
	}
	more := true
	for more {
		err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
			// Retrieve oldest attempt that needs notification.
			var attemptID uuid.UUID
			var err error
			attemptID, more, err = c.store.OldestPendingAttemptToNotifyByUser(ctx, tx, userID)
			if err != nil {
				return meh.Wrap(err, "pending attempts to notify by user", meh.Details{"user_id": userID})
			}
			if !more {
				return nil
			}
			// Try to send.
			notif, err := c.store.OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip(ctx, tx, attemptID)
			if err != nil {
				return meh.Wrap(err, "retrieve outgoing notification for attempt", meh.Details{"attempt_id": attemptID})
			}
			sentTS := time.Now()
			c.broadcastOutgoingNotificationToUser(ctx, notif, userConns)
			// Mark as sent.
			err = c.store.CreateIntelNotificationHistoryEntry(ctx, tx, attemptID, sentTS)
			if err != nil {
				return meh.Wrap(err, "create intel-notification-history-entry", meh.Details{"attempt_id": attemptID})
			}
			// Notify via Notifier.NotifyIntelDeliveryNotificationSent
			err = c.notifier.NotifyIntelDeliveryNotificationSent(ctx, tx, notif.DeliveryAttempt.ID, sentTS)
			if err != nil {
				return meh.Wrap(err, "notify intel delivery notification sent", nil)
			}
			return nil
		})
		if err != nil {
			return meh.Wrap(err, "run in tx", nil)
		}
	}
	return nil
}

// broadcastOutgoingNotificationToUser broadcasts the given
// store.OutgoingIntelDeliveryNotification concurrently to the Connection list.
func (c *Controller) broadcastOutgoingNotificationToUser(ctx context.Context, notif store.OutgoingIntelDeliveryNotification,
	conns []Connection) {
	var wg sync.WaitGroup
	for _, conn := range conns {
		conn := conn
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := conn.Notify(ctx, notif)
			if err != nil {
				mehlog.Log(c.logger, meh.Wrap(err, "notify via connection", meh.Details{
					"attempt_id":    notif.DeliveryAttempt.ID,
					"delivery_id":   notif.DeliveryAttempt.Delivery,
					"assigned_user": notif.DeliveryAttempt.AssignedToUser.UUID,
				}))
				return
			}
		}()
	}
	wg.Wait()
}
