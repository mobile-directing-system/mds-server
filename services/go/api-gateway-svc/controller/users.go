package controller

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// CreateUser creates the given store.User in the store.
func (c *Controller) CreateUser(ctx context.Context, user store.User) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.CreateUser(ctx, tx, user)
		if err != nil {
			return meh.Wrap(err, "create user", nil)
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
