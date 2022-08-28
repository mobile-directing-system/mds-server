package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
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
		err = c.Notifier.NotifyUserCreated(ctx, tx, user)
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
// the is-admin-field result in an meh.ErrForbidden error. This also applies to
// active-state changes without being allowed.
func (c *Controller) UpdateUser(ctx context.Context, user store.User, allowAdminChange bool, allowActiveStateChange bool) error {
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
			if user.IsActive == false {
				return meh.NewBadInputErr("admin cannot be set to inactive", nil)
			}
			// Assure the admin user not being set to non-admin.
			if !user.IsAdmin {
				return meh.NewBadInputErr("admin user cannot be set to non-admin", nil)
			}
		}
		if was.IsActive != user.IsActive && !allowActiveStateChange {
			return meh.NewForbiddenErr("active-state change not allowed", nil)
		}
		// Update in store.
		err = c.Store.UpdateUser(ctx, tx, user)
		if err != nil {
			return meh.Wrap(err, "update user in store", meh.Details{"user": user})
		}
		// Notify.
		err = c.Notifier.NotifyUserUpdated(ctx, tx, user)
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
		err = c.Notifier.NotifyUserPassUpdated(ctx, tx, userID, newPass)
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

// SetUserInactiveByID sets the user with the given id to inactive in the store
// and notifies via Notifier.NotifyUserUpdated.
func (c *Controller) SetUserInactiveByID(ctx context.Context, userID uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Retrieve user in order to deny deleting the admin user.
		user, err := c.Store.UserByID(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "user by id", meh.Details{"user_id": userID})
		}
		if user.Username == adminUsername {
			return meh.NewBadInputErr("admin user cannot be set inactive", nil)
		}
		// Update in store.
		user.IsActive = false
		err = c.Store.UpdateUser(ctx, tx, user)
		if err != nil {
			return meh.Wrap(err, "update user by id in store", meh.Details{"update": user})
		}
		// Notify.
		err = c.Notifier.NotifyUserUpdated(ctx, tx, user)
		if err != nil {
			return meh.Wrap(err, "notify user updated", meh.Details{"updated": user})
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
func (c *Controller) Users(ctx context.Context, filters store.UserFilters, params pagination.Params) (pagination.Paginated[store.User], error) {
	var users pagination.Paginated[store.User]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		users, err = c.Store.Users(ctx, tx, filters, params)
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

// SearchUsers searches for users with the given search.Params.
func (c *Controller) SearchUsers(ctx context.Context, filters store.UserFilters, searchParams search.Params) (search.Result[store.User], error) {
	var result search.Result[store.User]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		result, err = c.Store.SearchUsers(ctx, tx, filters, searchParams)
		if err != nil {
			return meh.Wrap(err, "search users", meh.Details{"params": searchParams})
		}
		return nil
	})
	if err != nil {
		return search.Result[store.User]{}, meh.Wrap(err, "run in tx", nil)
	}
	return result, nil
}

// RebuildUserSearch asynchronously rebuilds the user-search.
func (c *Controller) RebuildUserSearch(ctx context.Context) {
	c.Logger.Debug("rebuilding user-search...")
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.RebuildUserSearch(ctx, tx)
		if err != nil {
			return meh.Wrap(err, "rebuild user search in store", nil)
		}
		return nil
	})
	if err != nil {
		mehlog.Log(c.Logger, meh.Wrap(meh.Wrap(err, "run in tx", nil), "rebuild user search", nil))
		return
	}
	c.Logger.Debug("user-search rebuilt")
}
