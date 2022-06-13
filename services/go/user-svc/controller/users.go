package controller

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
)

// CreateUser creates the given store.UserWithPass and notifies via
// Notifier.NotifyUserCreated.
func (c *Controller) CreateUser(ctx context.Context, user store.UserWithPass) (store.UserWithPass, error) {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Create in store.
		createdUser, err := c.Store.CreateUser(ctx, tx, user)
		if err != nil {
			return meh.Wrap(err, "create user in store", nil)
		}
		// Notify.
		user.User = createdUser
		err = c.Notifier.NotifyUserCreated(user)
		if err != nil {
			return meh.Wrap(err, "notify user created", nil)
		}
		return nil
	})
	if err != nil {
		return store.UserWithPass{}, meh.Wrap(err, "run in tx", nil)
	}
	return user, nil
}

// UpdateUser updates the given store.User in the Store and notifies via
// Notifier.NotifyUserUpdated. If admin changing is not allowed, any changes to
// the is-admin-field result in an meh.ErrUnauthorized error.
func (c *Controller) UpdateUser(ctx context.Context, user store.User, allowAdminChange bool) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Retrieve user for checks regarding admins.
		was, err := c.Store.UserByID(ctx, tx, user.ID)
		if err != nil {
			return meh.Wrap(err, "user by id from store", meh.Details{"user_id": user.ID})
		}
		if was.IsAdmin != user.IsAdmin && !allowAdminChange {
			return meh.NewForbiddenErr("admin change not allowed", nil)
		}
		// Admin checks.
		if was.Username == adminUsername {
			if user.Username != was.Username {
				// Assure the admin username not being changed.
				return meh.NewBadInputErr("admin username cannot be changed", nil)
			}
			// Assure the admin user not being set to non-admin.
			if !user.IsAdmin {
				return meh.NewBadInputErr("admin user cannot be set to non-admin", nil)
			}
		}
		// Update in store.
		err = c.Store.UpdateUser(ctx, tx, user)
		if err != nil {
			return meh.Wrap(err, "update user in store", meh.Details{"user": user})
		}
		// Notify.
		err = c.Notifier.NotifyUserUpdated(user)
		if err != nil {
			return meh.Wrap(err, "notify user updated", meh.Details{"user": user})
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
		// Notify.
		err = c.Notifier.NotifyUserPassUpdated(userID, newPass)
		if err != nil {
			return meh.Wrap(err, "notify user pass updated", meh.Details{"user_id": userID})
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
		// Retrieve user in order to deny deleting the admin user.
		was, err := c.Store.UserByID(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "user by id", meh.Details{"user_id": userID})
		}
		if was.Username == adminUsername {
			return meh.NewBadInputErr("admin user cannot be deleted", nil)
		}
		// Delete in store.
		err = c.Store.DeleteUserByID(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "delete user by id in store", meh.Details{"user_id": userID})
		}
		// Notify.
		err = c.Notifier.NotifyUserDeleted(userID)
		if err != nil {
			return meh.Wrap(err, "notify user deleted", meh.Details{"user_id": userID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// UserByID retrieves a store.User by its id.
func (c *Controller) UserByID(ctx context.Context, userID uuid.UUID) (store.User, error) {
	var user store.User
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		user, err = c.Store.UserByID(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "user by id from store", meh.Details{"user_id": userID})
		}
		return nil
	})
	if err != nil {
		return store.User{}, meh.Wrap(err, "run in tx", nil)
	}
	return user, nil
}

// Users retrieves a paginated store.User list.
func (c *Controller) Users(ctx context.Context, params pagination.Params) (pagination.Paginated[store.User], error) {
	var users pagination.Paginated[store.User]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		users, err = c.Store.Users(ctx, tx, params)
		if err != nil {
			return meh.Wrap(err, "users from store", meh.Details{"params": params})
		}
		return nil
	})
	if err != nil {
		return pagination.Paginated[store.User]{}, meh.Wrap(err, "run in tx", nil)
	}
	return users, nil
}
