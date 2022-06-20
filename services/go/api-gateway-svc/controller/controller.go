package controller

import (
	"context"
	"github.com/google/uuid"
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
	UserIDBySessionToken(ctx context.Context, token string) (uuid.UUID, error)
	// StoreUserIDBySessionToken stores the given user id for the session token.
	StoreUserIDBySessionToken(ctx context.Context, token string, userID uuid.UUID) error
	// GetAndDeleteUserIDBySessionToken gets and then deletes the mapping of the
	// given session token to a user id.
	GetAndDeleteUserIDBySessionToken(ctx context.Context, token string) (uuid.UUID, error)
	// PassByUsername retrieves the hashed password for the user with the given
	// username.
	PassByUsername(ctx context.Context, tx pgx.Tx, username string) ([]byte, error)
	// UserByUsername retrieves the store.User with the given username.
	UserByUsername(ctx context.Context, tx pgx.Tx, username string) (store.User, error)
	// UserByID retrieves the store.user with the given id.
	UserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (store.User, error)
	// CreateUser creates the given store.User.
	CreateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// UpdateUser updates the given User, identified by its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// DeleteUserByID deletes the user with the given id.
	DeleteUserByID(ctx context.Context, tx pgx.Tx, userID string) error
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
