package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
)

// UpdatePermissionsByUser updates the permissions for the given user.
func (c *Controller) UpdatePermissionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, updatedPermissions []permission.Permission) error {
	// Update in store.
	err := c.Store.UpdatePermissionsByUser(ctx, tx, userID, updatedPermissions)
	if err != nil {
		return meh.Wrap(err, "update permissions in store", meh.Details{
			"user_id":     userID,
			"permissions": updatedPermissions,
		})
	}
	return nil
}
