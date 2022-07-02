package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
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
	// AssureUserExists makes sure that the user with the given id exists.
	AssureUserExists(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// CreateUser creates a user with the given id.
	CreateUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// DeleteUserByID deletes the user with the given id.
	//
	// Warning: Keep in mind, that the user must not have any permissions assigned!
	DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// PermissionsByUser retrieves the store.Permission list for the user with the
	// given id.
	PermissionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]store.Permission, error)
	// UpdatePermissionsByUser updates the permissions for the user with the given
	// id.
	UpdatePermissionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, permissions []store.Permission) error
}

// Notifier sends event messages.
type Notifier interface {
	// NotifyPermissionsUpdated notifies that permissions for the user with the
	// given id have been updated.
	NotifyPermissionsUpdated(userID uuid.UUID, permissions []store.Permission) error
}
