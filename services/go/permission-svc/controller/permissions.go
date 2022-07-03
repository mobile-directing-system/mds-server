package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// PermissionsByUser retrieves the permissions for the user with the given id.
func (c *Controller) PermissionsByUser(ctx context.Context, userID uuid.UUID) ([]store.Permission, error) {
	var permissions []store.Permission
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Assure exists.
		err := c.Store.AssureUserExists(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "assure user exists", meh.Details{"user_id": userID})
		}
		// Retrieve permissions.
		permissions, err = c.Store.PermissionsByUser(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "permissions by user from store", meh.Details{"user_id": userID})
		}
		return nil
	})
	if err != nil {
		return nil, meh.Wrap(err, "run in tx", nil)
	}
	return permissions, nil
}

// UpdatePermissionsByUser updates and notifies about changed permissions for
// the user with the given id.
func (c *Controller) UpdatePermissionsByUser(ctx context.Context, userID uuid.UUID, permissions []store.Permission) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Assure exists.
		err := c.Store.AssureUserExists(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "assure user exists", meh.Details{"user_id": userID})
		}
		// Update in store.
		err = c.Store.UpdatePermissionsByUser(ctx, tx, userID, permissions)
		if err != nil {
			return meh.Wrap(err, "update permissions in store", meh.Details{
				"user_id":     userID,
				"permissions": permissions,
			})
		}
		// Notify.
		err = c.Notifier.NotifyPermissionsUpdated(userID, permissions)
		if err != nil {
			return meh.Wrap(err, "notify permissions updated", meh.Details{
				"user_id":     userID,
				"permissions": permissions,
			})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
