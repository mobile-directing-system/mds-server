package controller

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
	"go.uber.org/zap"
)

// Controller manages all operations regarding users.
type Controller struct {
	Logger   *zap.Logger
	DB       *pgxpool.Pool
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
	// CreateUser creates the given store.User.
	CreateUser(ctx context.Context, tx pgx.Tx, user store.User) (store.User, error)
}

// Notifier sends event messages.
type Notifier interface {
	// NotifyUserCreated notifies, that the given store.User has been created.
	NotifyUserCreated(user store.User) error
}
