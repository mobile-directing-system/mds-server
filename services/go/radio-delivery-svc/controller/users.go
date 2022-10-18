package controller

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
)

// CreateUser creates the user with the given id.
func (c *Controller) CreateUser(ctx context.Context, tx pgx.Tx, create store.User) error {
	err := c.store.CreateUser(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create user in store", meh.Details{"create": create})
	}
	return nil
}

// UpdateUser updates the given store.User in the store.
func (c *Controller) UpdateUser(ctx context.Context, tx pgx.Tx, update store.User) error {
	err := c.store.UpdateUser(ctx, tx, update)
	if err != nil {
		return meh.Wrap(err, "update user in store", meh.Details{"update": update})
	}
	return nil
}
