package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
)

// Controller manages all core operations of the gateway.
type Controller struct {
	Logger                *zap.Logger
	PublicAuthTokenSecret string
	AuthTokenSecret       string
	Store                 Store
	DB                    pgutil.DBTxSupplier
	Notifier              Notifier
}

// Store is an interface for store.Mall.
type Store interface {
	// PermissionsByUserID retrieves a permission.Permission list for the user with
	// the given id.
	PermissionsByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]permission.Permission, error)
	// UserIDBySessionToken returns the user id for the given session token. If the
	// token was not found, a meh.ErrNotFound will be returned.
	UserIDBySessionToken(ctx context.Context, txSupplier pgutil.DBTxSupplier, token string) (uuid.UUID, error)
	// StoreSessionTokenForUser stores the given token for the user with the given
	// id.
	StoreSessionTokenForUser(ctx context.Context, tx pgx.Tx, token string, userID uuid.UUID) error
	// GetAndDeleteUserIDBySessionToken gets and then deletes the mapping of the
	// given session token to a user id.
	GetAndDeleteUserIDBySessionToken(ctx context.Context, tx pgx.Tx, token string) (uuid.UUID, error)
	// DeleteSessionTokensByUser deletes all session tokens for the given user from
	// the database and from cache.
	DeleteSessionTokensByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// PassByUsername retrieves the hashed password for the user with the given
	// username.
	PassByUsername(ctx context.Context, tx pgx.Tx, username string) ([]byte, error)
	// UserWithPassByUsername retrieves the store.UserWithPass with the given
	// username.
	UserWithPassByUsername(ctx context.Context, tx pgx.Tx, username string) (store.UserWithPass, error)
	// UserWithPassByID retrieves the store.UserWithPass with the given id.
	UserWithPassByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (store.UserWithPass, error)
	// CreateUser creates the given store.UserWithPass.
	CreateUser(ctx context.Context, tx pgx.Tx, user store.UserWithPass) error
	// UpdateUser updates the given User, identified by its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// UpdateUserPassByUserID updates the password for the user with the given id.
	UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newPass []byte) error
	// DeleteUserByID deletes the user with the given id.
	DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// UpdatePermissionsByUser updates the permissions for the user with the given
	// id.
	UpdatePermissionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, permissions []permission.Permission) error
}

// Notifier for event notifications.
type Notifier interface {
	// NotifyUserLoggedIn notifies that a user has logged in.
	NotifyUserLoggedIn(userID uuid.UUID, username string, requestMetadata AuthRequestMetadata) error
	// NotifyUserLoggedOut notifies that a user has logged out.
	NotifyUserLoggedOut(userID uuid.UUID, username string, requestMetadata AuthRequestMetadata) error
}
