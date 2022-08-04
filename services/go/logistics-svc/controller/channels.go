package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// ChannelsByAddressBookEntry retrieves the store.Channel list for the address
// book entry with the given id. If limit-to-user is set, an meh.ErrNotFound is
// returned, if the entry does not belong to this user.
func (c *Controller) ChannelsByAddressBookEntry(ctx context.Context, entryID uuid.UUID, limitToUser uuid.NullUUID) ([]store.Channel, error) {
	var channels []store.Channel
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		entry, err := c.Store.AddressBookEntryByID(ctx, tx, entryID, uuid.NullUUID{})
		if err != nil {
			return meh.Wrap(err, "address book entry from store", meh.Details{"entry_id": entryID})
		}
		if limitToUser.Valid {
			if !entry.User.Valid || entry.User.UUID != limitToUser.UUID {
				return meh.NewNotFoundErr("limited to user", meh.Details{
					"limit_to_user":  limitToUser.UUID,
					"entry_user":     entry.User.UUID,
					"entry_user_set": entry.User.Valid,
				})
			}
		}
		channels, err = c.Store.ChannelsByAddressBookEntry(ctx, tx, entryID)
		if err != nil {
			return meh.Wrap(err, "channels by address book entry from store", meh.Details{"entry_id": entryID})
		}
		return nil
	})
	if err != nil {
		return nil, meh.Wrap(err, "run in tx", nil)
	}
	return channels, nil
}

// UpdateChannelsByAddressBookEntry updates the channels for the entry with the
// given id. It tries to preserve existing channel ids for ongoing deliveries.
// This is accomplished by using channel ids as identifiers. If limit-to-user is
// set, a meh.ErrForbidden will be returned, if the entry is not associated with
// the user.
func (c *Controller) UpdateChannelsByAddressBookEntry(ctx context.Context, entryID uuid.UUID, newChannels []store.Channel,
	limitToUser uuid.NullUUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		entry, err := c.Store.AddressBookEntryByID(ctx, tx, entryID, uuid.NullUUID{})
		if err != nil {
			return meh.Wrap(err, "address book entry by id", meh.Details{"entry_id": entryID})
		}
		if limitToUser.Valid {
			if !entry.User.Valid || entry.User.UUID != limitToUser.UUID {
				return meh.NewForbiddenErr("user not associated with entry", meh.Details{
					"limit_to_user":             limitToUser.UUID,
					"entry_associated_user_set": entry.User.Valid,
					"entry_associated_user":     entry.User.UUID,
				})
			}
		}
		// Retrieve existing channels.
		oldChannels, err := c.Store.ChannelsByAddressBookEntry(ctx, tx, entryID)
		if err != nil {
			return meh.Wrap(err, "retrieve old channels", meh.Details{"entry_id": entryID})
		}
		oldChannelsHitByID := make(map[uuid.UUID]bool, len(oldChannels))
		for _, oldChannel := range oldChannels {
			oldChannelsHitByID[oldChannel.ID] = false
		}
		// Analyze which channels need to be created/removed/updated.
		create := make([]store.Channel, 0)
		update := make([]store.Channel, 0)
		for _, newChannel := range newChannels {
			if newChannel.ID.IsNil() {
				create = append(create, newChannel)
			} else if hit, ok := oldChannelsHitByID[newChannel.ID]; ok {
				if hit {
					return meh.NewBadInputErr("duplicate channel ids", meh.Details{"channel_id": newChannel.ID})
				}
				oldChannelsHitByID[newChannel.ID] = true
				update = append(update, newChannel)
			} else if !ok {
				return meh.NewBadInputErr("unknown channel", meh.Details{"channel_id": newChannel.ID})
			}
		}
		// Remove old channels, that are not used anymore.
		for _, oldChannel := range oldChannels {
			if oldChannelsHitByID[oldChannel.ID] {
				continue
			}
			err = c.Store.DeleteChannelWithDetailsByID(ctx, tx, oldChannel.ID, oldChannel.Type)
			if err != nil {
				return meh.Wrap(err, "delete channel with details", meh.Details{
					"channel_id":   oldChannel.ID,
					"channel_type": oldChannel.Type,
				})
			}
		}
		// Create channels.
		for _, channelToCreate := range create {
			err = c.Store.CreateChannelWithDetails(ctx, tx, channelToCreate)
			if err != nil {
				return meh.Wrap(err, "create channel with details", meh.Details{"channel": channelToCreate})
			}
		}
		// Update channels.
		for _, channelToUpdate := range update {
			err = c.Store.UpdateChannelWithDetails(ctx, tx, channelToUpdate)
			if err != nil {
				return meh.Wrap(err, "update channel with details", meh.Details{"channel": channelToUpdate})
			}
		}
		// Notify.
		finalUpdatedChannels, err := c.Store.ChannelsByAddressBookEntry(ctx, tx, entryID)
		if err != nil {
			return meh.Wrap(err, "final channels by entry from store", meh.Details{"entry_id": entryID})
		}
		err = c.Notifier.NotifyAddressBookEntryChannelsUpdated(entryID, finalUpdatedChannels)
		if err != nil {
			return meh.Wrap(err, "notify", meh.Details{
				"entry_id": entryID,
				"channels": finalUpdatedChannels,
			})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
