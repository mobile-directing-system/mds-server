package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
)

// CreateUser creates the user with the given id.
func (c *Controller) CreateUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	err := c.Store.CreateUser(ctx, tx, userID)
	if err != nil {
		return meh.Wrap(err, "create user in store", meh.Details{"user_id": userID})
	}
	return nil
}

// DeleteUserByID deletes the user with the given id and notifies of unassigned
// permissions.
func (c *Controller) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Unassign permissions.
	err := c.Store.UpdatePermissionsByUser(ctx, tx, userID, []store.Permission{})
	if err != nil {
		return meh.Wrap(err, "update permissions in store", meh.Details{"user_id": userID})
	}
	// Notify.
	err = c.Notifier.NotifyPermissionsUpdated(ctx, tx, userID, []store.Permission{})
	if err != nil {
		return meh.Wrap(err, "notify permissions updated", meh.Details{"user_id": userID})
	}
	// Delete user.
	err = c.Store.DeleteUserByID(ctx, tx, userID)
	if err != nil {
		return meh.Wrap(err, "delete user in store", meh.Details{"user_id": userID})
	}
	return nil
}
