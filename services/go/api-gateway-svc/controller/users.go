package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
)

// CreateUser creates the given store.User in the store.
func (c *Controller) CreateUser(ctx context.Context, tx pgx.Tx, user store.UserWithPass) error {
	err := c.Store.CreateUser(ctx, tx, user)
	if err != nil {
		return meh.Wrap(err, "create user", nil)
	}
	return nil
}

// UpdateUser updates the given store.User in the Store.
func (c *Controller) UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	if !user.IsActive {
		// Invalidate sessions.
		err := c.Store.DeleteSessionTokensByUser(ctx, tx, user.ID)
		if err != nil {
			return meh.Wrap(err, "delete session tokens for user in store", meh.Details{"user_id": user.ID})
		}
	}
	// Update in store.
	err := c.Store.UpdateUser(ctx, tx, user)
	if err != nil {
		return meh.Wrap(err, "update user in store", meh.Details{"user": user})
	}
	return nil
}

// UpdateUserPassByUserID updates the password for the user with the given id in
// the store and notifies via Notifier.NotifyUserPassUpdated.
func (c *Controller) UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newPass []byte) error {
	// Invalidate sessions.
	err := c.Store.DeleteSessionTokensByUser(ctx, tx, userID)
	if err != nil {
		return meh.Wrap(err, "delete session tokens for user in store", meh.Details{"user_id": userID})
	}
	// Update in store.
	err = c.Store.UpdateUserPassByUserID(ctx, tx, userID, newPass)
	if err != nil {
		return meh.Wrap(err, "update user pass by user id in store", meh.Details{"user_id": userID})
	}
	return nil
}
