package controller

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
)

// CreateUser creates the given store.User.
func (c *Controller) CreateUser(ctx context.Context, tx pgx.Tx, create store.User) error {
	err := c.Store.CreateUser(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create user in store", meh.Details{"user": create})
	}
	return nil
}

// UpdateUser updates the given store.User, identified by its id.
func (c *Controller) UpdateUser(ctx context.Context, tx pgx.Tx, update store.User) error {
	err := c.Store.UpdateUser(ctx, tx, update)
	if err != nil {
		return meh.Wrap(err, "update user in store", meh.Details{"user": update})
	}
	return nil
}
