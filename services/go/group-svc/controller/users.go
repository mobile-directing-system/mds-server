package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
)

// CreateUser creates the user with the given id.
func (c *Controller) CreateUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	err := c.Store.CreateUser(ctx, tx, userID)
	if err != nil {
		return meh.Wrap(err, "create user in store", meh.Details{"user_id": userID})
	}
	return nil
}
