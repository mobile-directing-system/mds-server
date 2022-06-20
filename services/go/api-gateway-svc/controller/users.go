package controller

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// CreateUser creates the given store.User in the store.
func (c *Controller) CreateUser(ctx context.Context, user store.UserWithPass) error {
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

// UpdateUser updates the given store.User in the Store.
func (c *Controller) UpdateUser(ctx context.Context, user store.User) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Update in store.
		err := c.Store.UpdateUser(ctx, tx, user)
		if err != nil {
			return meh.Wrap(err, "update user in store", meh.Details{"user": user})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// UpdateUserPassByUserID updates the password for the user with the given id in
// the store and notifies via Notifier.NotifyUserPassUpdated.
func (c *Controller) UpdateUserPassByUserID(ctx context.Context, userID uuid.UUID, newPass []byte) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Update in store.
		err := c.Store.UpdateUserPassByUserID(ctx, tx, userID, newPass)
		if err != nil {
			return meh.Wrap(err, "update user pass by user id in store", meh.Details{"user_id": userID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// DeleteUserByID deletes the user with the given id in the store and notifies
// via Notifier.NotifyUserDeleted.
func (c *Controller) DeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Delete in store.
		err := c.Store.DeleteUserByID(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "delete user by id in store", meh.Details{"user_id": userID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// TODO: DELETE TOKENS FOR DELETED USESRS!!!!!!
