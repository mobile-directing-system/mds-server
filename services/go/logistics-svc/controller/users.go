package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"golang.org/x/sync/errgroup"
)

// CreateUser creates the given store.User.
func (c *Controller) CreateUser(ctx context.Context, tx pgx.Tx, create store.User) error {
	err := c.Store.CreateUser(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create user in store", meh.Details{"user": create})
	}
	return nil
}

// UpdateUser updates the given store.User, identified by its id.
func (c *Controller) UpdateUser(ctx context.Context, tx pgx.Tx, update store.User) error {
	err := c.Store.UpdateUser(ctx, tx, update)
	if err != nil {
		return meh.Wrap(err, "update user in store", meh.Details{"user": update})
	}
	return nil
}

// DeleteUserByID deletes the user with the given id.
func (c *Controller) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Delete any channels, forwarding to this user.
	affectedEntries, err := c.Store.DeleteForwardToUserChannelsByUser(ctx, tx, userID)
	if err != nil {
		return meh.Wrap(err, "delete forward-to-user-channels by user", meh.Details{"user_id": userID})
	}
	// Delete user itself.
	err = c.Store.DeleteUserByID(ctx, tx, userID)
	if err != nil {
		return meh.Wrap(err, "delete user in store", meh.Details{"user_id": userID})
	}
	// Retrieve updated channels.
	channelsByEntry := make(map[uuid.UUID][]store.Channel)
	for _, affectedEntryID := range affectedEntries {
		channels, err := c.Store.ChannelsByAddressBookEntry(ctx, tx, affectedEntryID)
		if err != nil {
			return meh.Wrap(err, "channels by address book entry from store",
				meh.Details{"entry_id": affectedEntryID})
		}
		channelsByEntry[affectedEntryID] = channels
	}
	// Notify about updated channels.
	var eg errgroup.Group
	for entryID, channels := range channelsByEntry {
		eID := entryID
		chs := channels
		eg.Go(func() error {
			err := c.Notifier.NotifyAddressBookEntryChannelsUpdated(ctx, tx, eID, chs)
			if err != nil {
				return meh.Wrap(err, "notify channels updated", meh.Details{
					"entry_id": eID,
					"channels": chs,
				})
			}
			return nil
		})
	}
	return eg.Wait()
}
