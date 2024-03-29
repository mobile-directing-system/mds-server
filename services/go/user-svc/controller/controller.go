package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
	"go.uber.org/zap"
)

// Controller manages all operations regarding users.
type Controller struct {
	Logger   *zap.Logger
	DB       pgutil.DBTxSupplier
	Store    Store
	Notifier Notifier
}

// Run the controller.
func (c *Controller) Run(lifetime context.Context) error {
	err := c.AssureAdminUser(lifetime)
	if err != nil {
		return meh.Wrap(err, "assure admin user", nil)
	}
	<-lifetime.Done()
	return nil
}

// Store for persistence.
type Store interface {
	// UserByID retrieves a store.User by its id.
	UserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (store.User, error)
	// UserByUsername retrieves a store.User by its username.
	UserByUsername(ctx context.Context, tx pgx.Tx, username string) (store.User, error)
	// Users retrieves all known users.
	Users(ctx context.Context, tx pgx.Tx, filters store.UserFilters, params pagination.Params) (pagination.Paginated[store.User], error)
	// CreateUser creates the given store.User.
	CreateUser(ctx context.Context, tx pgx.Tx, user store.UserWithPass) (store.User, error)
	// UpdateUser updates the given store.User, identifies by its user id. This will
	// not change the password!
	UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// UpdateUserPassByUserID updates the hashed password of the user with the given
	// id.
	UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, pass []byte) error
	// SearchUsers searches for users with the given search.Params.
	SearchUsers(ctx context.Context, tx pgx.Tx, filters store.UserFilters, searchParams search.Params) (search.Result[store.User], error)
	// RebuildUserSearch rebuilds the user search.
	RebuildUserSearch(ctx context.Context, tx pgx.Tx) error
}

// Notifier sends event messages.
type Notifier interface {
	// NotifyUserCreated notifies, that the given store.UserWithPass has been
	// created.
	NotifyUserCreated(ctx context.Context, tx pgx.Tx, user store.UserWithPass) error
	// NotifyUserUpdated notifies, that the given store.User was updated.
	NotifyUserUpdated(ctx context.Context, tx pgx.Tx, user store.User) error
	// NotifyUserPassUpdated notifies, that the the user with the given id has
	// updated its password.
	NotifyUserPassUpdated(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newPass []byte) error
}
