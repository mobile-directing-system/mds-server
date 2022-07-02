package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
)

// Controller manages all operations regarding permissions.
type Controller struct {
	Logger   *zap.Logger
	DB       pgutil.DBTxSupplier
	Store    Store
	Notifier Notifier
}

// Store for persistence.
type Store interface {
	// OperationByID retrieves an store.Operation by its id.
	OperationByID(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) (store.Operation, error)
	// Operations retrieves an store.Operation list.
	Operations(ctx context.Context, tx pgx.Tx, params pagination.Params) (pagination.Paginated[store.Operation], error)
	// CreateOperation creates the given store.Operation and returns it with its
	// assigned id.
	CreateOperation(ctx context.Context, tx pgx.Tx, operation store.Operation) (store.Operation, error)
	// UpdateOperation updates the given store.Operation, identified by its id.
	UpdateOperation(ctx context.Context, tx pgx.Tx, operation store.Operation) error
}

// Notifier sends event messages.
type Notifier interface {
	// NotifyOperationCreated emits an event.TypeOperationCreated event.
	NotifyOperationCreated(operation store.Operation) error
	// NotifyOperationUpdated emits an event.TypeOperationUpdated event.
	NotifyOperationUpdated(operation store.Operation) error
}
