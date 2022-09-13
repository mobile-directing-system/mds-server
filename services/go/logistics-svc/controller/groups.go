package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
)

// CreateGroup creates the given store.Group.
func (c *Controller) CreateGroup(ctx context.Context, tx pgx.Tx, create store.Group) error {
	err := c.Store.CreateGroup(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create group in store", meh.Details{"create": create})
	}
	return nil
}

// UpdateGroup updates the given store.Group, identified by its id.
func (c *Controller) UpdateGroup(ctx context.Context, tx pgx.Tx, update store.Group) error {
	err := c.Store.UpdateGroup(ctx, tx, update)
	if err != nil {
		return meh.Wrap(err, "update group in store", meh.Details{"update": update})
	}
	return nil
}

// DeleteGroupByID deletes the group with the given id.
func (c *Controller) DeleteGroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	// Delete any channels, forwarding to this group.
	affectedEntries, err := c.Store.DeleteForwardToGroupChannelsByGroup(ctx, tx, groupID)
	if err != nil {
		return meh.Wrap(err, "delete forward-to-group-channels by group", meh.Details{"group_id": groupID})
	}
	// Delete group itself.
	err = c.Store.DeleteGroupByID(ctx, tx, groupID)
	if err != nil {
		return meh.Wrap(err, "delete group in store", meh.Details{"group_id": groupID})
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
	for entryID, channels := range channelsByEntry {
		err := c.Notifier.NotifyAddressBookEntryChannelsUpdated(ctx, tx, entryID, channels)
		if err != nil {
			return meh.Wrap(err, "notify channels updated", meh.Details{
				"entry_id": entryID,
				"channels": channels,
			})
		}
	}
	return nil
}
