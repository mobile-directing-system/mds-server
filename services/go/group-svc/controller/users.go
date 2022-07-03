package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// CreateUser creates the user with the given id.
func (c *Controller) CreateUser(ctx context.Context, userID uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.CreateUser(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "create user in store", meh.Details{"user_id": userID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// DeleteUserByID removes the user with the given id from all groups, notifies
// of updated ones and deletes the user.
func (c *Controller) DeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Retrieve all groups where the user is a member of in order to remove from
		// these groups.
		memberGroups, err := c.Store.Groups(ctx, tx, store.GroupFilters{
			ByUser: uuid.NullUUID{UUID: userID, Valid: true},
		}, pagination.Params{})
		if err != nil {
			return meh.Wrap(err, "groups from store", nil)
		}
		// Remove user from all these groups.
		updated := make([]store.Group, 0, memberGroups.Total)
	eachMemberGroup:
		for _, group := range memberGroups.Entries {
			for i, groupMember := range group.Members {
				if groupMember != userID {
					continue
				}
				group.Members[i] = group.Members[len(group.Members)-1]
				group.Members = group.Members[:len(group.Members)-1]
				updated = append(updated, group)
				continue eachMemberGroup
			}
			return meh.NewInternalErr("user not found in group", meh.Details{
				"user_id":       userID,
				"group_members": group.Members,
			})
		}
		// Update groups in store.
		for _, group := range updated {
			err = c.Store.UpdateGroup(ctx, tx, group)
			if err != nil {
				return meh.Wrap(err, "update group in store", meh.Details{"group": group})
			}
		}
		// Delete user.
		err = c.Store.DeleteUserByID(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "delete user in store", meh.Details{"user_id": userID})
		}
		// Notify of updated groups.
		for _, group := range updated {
			err = c.Notifier.NotifyGroupUpdated(group)
			if err != nil {
				return meh.Wrap(err, "notify group updated", meh.Details{"group": group})
			}
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
