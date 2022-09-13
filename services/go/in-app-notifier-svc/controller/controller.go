package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

// Controller manages all operations regarding intelligence.
type Controller struct {
	logger   *zap.Logger
	db       pgutil.DBTxSupplier
	store    Store
	notifier Notifier
	// connectionsByUser holds a Connection list for each user by its id.
	connectionsByUser map[uuid.UUID][]Connection
	// connectionsByUserMutex locks connectionsByUser.
	connectionsByUserMutex sync.RWMutex
	// lookAfterUserNotificationRequests is a buffered channel, holding pending
	// requests for looking after user notifications. Use
	// scheduleLookAfterUserNotifications for sending to this channel. Requests are
	// read by workers in runLookAfterUserNotificationsScheduler.
	lookAfterUserNotificationRequests chan uuid.UUID
}

// NewController creates a new Controller. Do not forget to call Controller.Run.
func NewController(logger *zap.Logger, db pgutil.DBTxSupplier, store Store, notifier Notifier) *Controller {
	return &Controller{
		logger:                            logger,
		db:                                db,
		store:                             store,
		notifier:                          notifier,
		connectionsByUser:                 make(map[uuid.UUID][]Connection),
		lookAfterUserNotificationRequests: make(chan uuid.UUID, 256),
	}
}

// Run schedulers and periodic operations until the given context is done.
func (c *Controller) Run(lifetime context.Context) error {
	eg, egCtx := errgroup.WithContext(lifetime)
	// Run scheduler.
	eg.Go(func() error {
		err := c.runLookAfterUserNotificationsScheduler(egCtx)
		if err != nil {
			return meh.Wrap(err, "run look-after-user-notifications-scheduler", nil)
		}
		return nil
	})
	// Run periodic checks.
	eg.Go(func() error {
		err := c.runPeriodicNotificationChecks(egCtx)
		if err != nil {
			return meh.Wrap(err, "run periodic notification checks", nil)
		}
		return nil
	})
	return eg.Wait()
}

// Store for Controller.
type Store interface {
	// OldestPendingAttemptToNotifyByUser retrieves the id of the oldest attempt
	// that needs notification to the user with the given id.
	//
	// This is meant to be used when the user makes a connection in order to send
	// all pending notifications.
	OldestPendingAttemptToNotifyByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (uuid.UUID, bool, error)
	// OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip retrieves the
	// store.OutgoingIntelDeliveryNotification and the delivery-attempt for update
	// or skips it. This means, that when the returned error is meh.ErrNotFound, it
	// might also me locked by someone else. It is also assured that the attempt has
	// no tries in the notification history. This is required if someone else
	// already processed this attempt.
	OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip(ctx context.Context, tx pgx.Tx,
		attemptID uuid.UUID) (store.OutgoingIntelDeliveryNotification, error)
	// CreateIntelNotificationHistoryEntry creates an entry in the history for
	// keeping log of sent notifications for attempts.
	CreateIntelNotificationHistoryEntry(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, ts time.Time) error
	// CreateUser creates the given store.User.
	CreateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// UpdateUser updates the user details of the given store.User, identified by
	// its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// CreateNotificationChannel creates the given store.NotificationChannel in the
	// store.
	CreateNotificationChannel(ctx context.Context, tx pgx.Tx, create store.NotificationChannel) error
	// DeleteNotificationChannelsByEntry deletes the notification channels in the
	// store associated with the address book entry with the given id.
	DeleteNotificationChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error
	// CreateAcceptedIntelDeliveryAttempt creates the given
	// store.AcceptedIntelDeliveryAttempt.
	CreateAcceptedIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, create store.AcceptedIntelDeliveryAttempt) error
	// UpdateAcceptedIntelDeliveryAttemptStatus updates the given
	// store.AcceptedIntelDeliveryAttemptStatus, identified by its id.
	UpdateAcceptedIntelDeliveryAttemptStatus(ctx context.Context, tx pgx.Tx, update store.AcceptedIntelDeliveryAttemptStatus) error
	// NotificationChannelByID retrieves the store.NotificationChannel with the
	// given id.
	NotificationChannelByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (store.NotificationChannel, error)
	// CreateIntelToDeliver creates the given store.IntelToDeliver in the store.
	CreateIntelToDeliver(ctx context.Context, tx pgx.Tx, create store.IntelToDeliver) error
}

// Notifier for Controller.
type Notifier interface {
	// NotifyIntelDeliveryNotificationPending notifies that an in-app-notification
	// for an intel-delivery-attempt is pending.
	NotifyIntelDeliveryNotificationPending(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, since time.Time) error
	// NotifyIntelDeliveryNotificationSent notifies that an in-app-notification for
	// an intel-delivery-attempt was sent.
	NotifyIntelDeliveryNotificationSent(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, sentTS time.Time) error
}
