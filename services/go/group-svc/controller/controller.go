package controller

import (
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Controller manages all operations regarding groups.
type Controller struct {
	Logger   *zap.Logger
	DB       pgutil.DBTxSupplier
	Store    Store
	Notifier Notifier
}

// Store for persistence.
type Store interface {
	// CreateUser creates a user with the given id.
	CreateUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// DeleteUserByID deletes the user with the given id.
	DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// CreateOperation creates the operation with the given id.
	CreateOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) error
	// CreateGroup creates the given group and returns the one with assigned id.
	CreateGroup(ctx context.Context, tx pgx.Tx, create store.Group) (store.Group, error)
	// UpdateGroup updates the group identified by its id.
	UpdateGroup(ctx context.Context, tx pgx.Tx, update store.Group) error
	// GroupByID retrieves a group by its id.
	GroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) (store.Group, error)
	// Groups retrieves a paginated store.Group list with optional
	// store.GroupFilters.
	Groups(ctx context.Context, tx pgx.Tx, filters store.GroupFilters, params pagination.Params) (pagination.Paginated[store.Group], error)
	// AssureUserExists assures that the user with the given id exists.
	AssureUserExists(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// DeleteGroupByID deletes the group with the given id.
	DeleteGroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error
	// OperationMembersByOperation retrieves all users that are member of the
	// operation with the given id.
	OperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) ([]uuid.UUID, error)
	// UpdateOperationMembersByOperation updates the member list for the operation
	// with the given id.
	UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error
}

// Notifier sends event messages.
type Notifier interface {
	// NotifyGroupCreated creates an event.TypeGroupCreated event.
	NotifyGroupCreated(ctx context.Context, tx pgx.Tx, group store.Group) error
	// NotifyGroupUpdated creates an event.TypeGroupUpdated event.
	NotifyGroupUpdated(ctx context.Context, tx pgx.Tx, group store.Group) error
	// NotifyGroupDeleted creates an event.TypeGroupDeleted event.
	NotifyGroupDeleted(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error
}
