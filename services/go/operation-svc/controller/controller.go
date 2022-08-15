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
	// CreateUser adds the given store.User to the store.
	CreateUser(ctx context.Context, tx pgx.Tx, create store.User) error
	// UpdateUser updates the given store.User, identified by its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, update store.User) error
	// DeleteUserByID deletes the user with the given id.
	DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// UpdateOperationMembersByOperation updates the members for the operation with
	// the given id.
	UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, members []uuid.UUID) error
	// OperationMembersByOperation retrieves a paginated store.User list for the
	// operation with the given id.
	OperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID,
		params pagination.Params) (pagination.Paginated[store.User], error)
	// OperationsByMember retrieves an Operation list for the member with the given
	// id.
	OperationsByMember(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]store.Operation, error)
}

// Notifier sends event messages.
type Notifier interface {
	// NotifyOperationCreated emits an event.TypeOperationCreated event.
	NotifyOperationCreated(ctx context.Context, tx pgx.Tx, operation store.Operation) error
	// NotifyOperationUpdated emits an event.TypeOperationUpdated event.
	NotifyOperationUpdated(ctx context.Context, tx pgx.Tx, operation store.Operation) error
	// NotifyOperationMembersUpdated emits an event.TypeOperationMembersUpdated.
	NotifyOperationMembersUpdated(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, members []uuid.UUID) error
}
