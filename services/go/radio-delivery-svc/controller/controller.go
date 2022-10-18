package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"time"
)

// Controller manages all operations regarding intelligence.
type Controller struct {
	logger   *zap.Logger
	db       pgutil.DBTxSupplier
	store    Store
	notifier Notifier
	// pickedUpTimeout is the timeout for picked up radio deliveries to be released.
	pickedUpTimeout    time.Duration
	connUpdateNotifier *updateNotifier
}

// NewController creates a new Controller. Do not forget to call Controller.Run.
func NewController(logger *zap.Logger, db pgutil.DBTxSupplier, store Store, notifier Notifier, pickedUpTimeout time.Duration) *Controller {
	return &Controller{
		logger:             logger,
		db:                 db,
		store:              store,
		notifier:           notifier,
		pickedUpTimeout:    pickedUpTimeout,
		connUpdateNotifier: newUpdateNotifier(logger.Named("conn-update-notifier"), db, store),
	}
}

// Run schedulers and periodic operations until the given context is done.
func (c *Controller) Run(lifetime context.Context) error {
	eg, egCtx := errgroup.WithContext(lifetime)
	// Run periodic timeout checks.
	eg.Go(func() error {
		c.runPeriodicRadioDeliveryTimeoutChecks(egCtx)
		return nil
	})
	// Run connection-update-notifier.
	eg.Go(func() error {
		c.connUpdateNotifier.run(egCtx)
		return nil
	})
	return eg.Wait()
}

// AcceptNewConnection handles the new given Connection via the internal
// update-notifier.
func (c *Controller) AcceptNewConnection(conn Connection) {
	c.connUpdateNotifier.AcceptNewConnection(conn)
}

// Store for Controller.
type Store interface {
	// CreateUser creates the given store.User.
	CreateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// UpdateUser updates the user details of the given store.User, identified by
	// its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// CreateRadioChannel creates the given store.RadioChannel in the store.
	CreateRadioChannel(ctx context.Context, tx pgx.Tx, create store.RadioChannel) error
	// DeleteRadioChannelsByEntry deletes the notification channels in the store
	// associated with the address book entry with the given id.
	DeleteRadioChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error
	// CreateAcceptedIntelDeliveryAttempt creates the given
	// store.AcceptedIntelDeliveryAttempt.
	CreateAcceptedIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, create store.AcceptedIntelDeliveryAttempt) error
	// UpdateAcceptedIntelDeliveryAttemptStatus updates the given
	// store.AcceptedIntelDeliveryAttemptStatus, identified by its id.
	UpdateAcceptedIntelDeliveryAttemptStatus(ctx context.Context, tx pgx.Tx, update store.AcceptedIntelDeliveryAttemptStatus) error
	// RadioChannelByID retrieves the store.RadioChannel with the
	// given id.
	RadioChannelByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (store.RadioChannel, error)
	// UpdateOperationMembersByOperation replaces the associated operation members
	// for the operation with the given id with the new given ones.
	UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error
	// OperationsByMember retrieves the ids of all operations, the user with the
	// given id is member of.
	OperationsByMember(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]uuid.UUID, error)
	// CreateRadioDelivery creates a store.RadioDelivery for the attempt with the
	// given id.
	CreateRadioDelivery(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) error
	// RadioDeliveryByAttempt retrieves the store.RadioDelivery for the attempt with
	// the given id.
	RadioDeliveryByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (store.RadioDelivery, error)
	// MarkRadioDeliveryAsPickedUpByAttempt marks the radio-delivery for the given
	// attempt as picked-up by the user with the given id. If no user id is given,
	// it will be unassigned.
	MarkRadioDeliveryAsPickedUpByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, by uuid.NullUUID, newNote string) error
	// UpdateRadioDeliveryStatusByAttempt updates the radio-delivery for the given
	// attempt with the given new success-status and note.
	UpdateRadioDeliveryStatusByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, newSuccess nulls.Bool, newNote string) error
	// ActiveRadioDeliveriesAndLockOrWait locks and retrieves all radio-deliveries
	// being active (success is NULL) from the database.
	ActiveRadioDeliveriesAndLockOrWait(ctx context.Context, tx pgx.Tx, byOperation uuid.NullUUID) ([]store.ActiveRadioDelivery, error)
	// AcceptedIntelDeliveryAttemptByID retrieves the
	// store.AcceptedIntelDeliveryAttempt with the given id from the store.
	AcceptedIntelDeliveryAttemptByID(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (store.AcceptedIntelDeliveryAttempt, error)
}

// Notifier for Controller.
type Notifier interface {
	NotifyRadioDeliveryReadyForPickup(ctx context.Context, tx pgx.Tx, intelDeliveryAttempt store.AcceptedIntelDeliveryAttempt,
		radioDeliveryNote string) error
	NotifyRadioDeliveryPickedUp(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, pickedUpBy uuid.UUID, pickedUpAt time.Time) error
	NotifyRadioDeliveryReleased(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, releasedAt time.Time) error
	NotifyRadioDeliveryFinished(ctx context.Context, tx pgx.Tx, radioDelivery store.RadioDelivery) error
}
