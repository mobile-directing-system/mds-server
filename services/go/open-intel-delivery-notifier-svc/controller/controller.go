package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
	"sync"
)

// Controller for core application logic.
type Controller struct {
	logger                                    *zap.Logger
	db                                        pgutil.DBTxSupplier
	store                                     Store
	openIntelDeliveryWatchersByOperation      map[uuid.UUID]*openIntelDeliveryWatcher
	openIntelDeliveryWatchersByOperationMutex sync.RWMutex
}

// NewController creates a new Controller.
func NewController(logger *zap.Logger, db pgutil.DBTxSupplier, store Store) *Controller {
	return &Controller{
		logger:                               logger,
		db:                                   db,
		store:                                store,
		openIntelDeliveryWatchersByOperation: make(map[uuid.UUID]*openIntelDeliveryWatcher),
	}
}

// Store for persistence.
type Store interface {
	// CreateUser adds the given store.User to the store.
	CreateUser(ctx context.Context, tx pgx.Tx, create store.User) error
	// UpdateUser updates the given store.User, identified by its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, update store.User) error
	// CreateOperation creates the given store.Operation.
	CreateOperation(ctx context.Context, tx pgx.Tx, create store.Operation) error
	// UpdateOperation updates the given store.Operation.
	UpdateOperation(ctx context.Context, tx pgx.Tx, update store.Operation) error
	// UpdateOperationMembersByOperation updates the operation members for the given
	// operation.
	UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error
	// CreateActiveIntelDelivery creates the given ActiveIntelDelivery in the store.
	CreateActiveIntelDelivery(ctx context.Context, tx pgx.Tx, create store.ActiveIntelDelivery) error
	// DeleteActiveIntelDeliveryByID deletes the ActiveIntelDelivery with the given
	// id.
	DeleteActiveIntelDeliveryByID(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) error
	// CreateActiveIntelDeliveryAttempt creates the given ActiveIntelDeliveryAttempt.
	CreateActiveIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, create store.ActiveIntelDeliveryAttempt) error
	// DeleteActiveIntelDeliveryAttemptByID deletes the ActiveIntelDeliveryAttempt
	// with the given id.
	DeleteActiveIntelDeliveryAttemptByID(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) error
	// IsAutoIntelDeliveryEnabledForEntry checks whether auto-intel-delivery is
	// enabled for the address book entry with the given id.
	IsAutoIntelDeliveryEnabledForEntry(ctx context.Context, tx pgx.Tx, addressBookEntryID uuid.UUID) (bool, error)
	// SetAutoIntelDeliveryEnabledForEntry sets the auto-intel-delivery flag for the
	// address book entry with the given id.
	SetAutoIntelDeliveryEnabledForEntry(ctx context.Context, tx pgx.Tx, addressBookEntryID uuid.UUID, enabled bool) error
	// OpenIntelDeliveriesByOperation retrieves the intel deliveries that are active
	// but have no active delivery attempt and are not marked for
	// auto-intel-delivery.
	OpenIntelDeliveriesByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) ([]store.OpenIntelDeliverySummary, error)
	// IntelByID retrieves the Intel with the given id from the store.
	IntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) (store.Intel, error)
	// CreateIntel creates the given Intel in the store.
	CreateIntel(ctx context.Context, tx pgx.Tx, create store.Intel) error
	// InvalidateIntelByID sets the Intel with the given id to invalid.
	InvalidateIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error
	// IntelOperationByDeliveryAttempt retrieves the operation id for the intel the
	// delivery associated with the given attempt is for.
	IntelOperationByDeliveryAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (uuid.UUID, error)
	// IntelOperationsByActiveIntelDeliveryRecipient retrieves a list of distinct
	// operation ids. These operations have intel that currently has active intel
	// deliveries with the given address book entry being the recipient.
	IntelOperationsByActiveIntelDeliveryRecipient(ctx context.Context, tx pgx.Tx, addressBookEntryID uuid.UUID) ([]uuid.UUID, error)
	// IntelOperationByDelivery retrieves the operation id for the intel the
	// given delivery is associated with. This is mainly used for
	// reducing database calls when trying to get the operation id for change
	// notifications.
	IntelOperationByDelivery(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (uuid.UUID, error)
}

// OpenIntelDeliveriesListener can be registered via
// Controller.ServeOpenIntelDeliveriesListener and is notified when open intel
// deliveries change.
type OpenIntelDeliveriesListener interface {
	NotifyOpenIntelDeliveries(ctx context.Context, openDeliveries []store.OpenIntelDeliverySummary) bool
}
