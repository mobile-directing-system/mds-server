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
// given id. All previous channels are deleted and active delivery attempts,
// using them, are canceled. If limit-to-user is set, a meh.ErrForbidden will be
// returned, if the entry is not associated with the user.
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
		affectedDeliveries, err := c.handleDeletedChannelsForDeliveryAttempts(ctx, tx, oldChannels)
		if err != nil {
			return meh.Wrap(err, "handle deleted channels for delivery attempts", nil)
		}
		// Recreate channels in store.
		err = c.Store.UpdateChannelsByEntry(ctx, tx, entryID, newChannels)
		if err != nil {
			return meh.Wrap(err, "update channels by entry in store", meh.Details{"entry_id": entryID})
		}
		// Notify.
		finalUpdatedChannels, err := c.Store.ChannelsByAddressBookEntry(ctx, tx, entryID)
		if err != nil {
			return meh.Wrap(err, "final channels by entry from store", meh.Details{"entry_id": entryID})
		}
		err = c.Notifier.NotifyAddressBookEntryChannelsUpdated(ctx, tx, entryID, finalUpdatedChannels)
		if err != nil {
			return meh.Wrap(err, "notify", meh.Details{
				"entry_id": entryID,
				"channels": finalUpdatedChannels,
			})
		}
		// Look after affected deliveries.
		for _, affectedDelivery := range affectedDeliveries {
			err = c.lookAfterDelivery(ctx, tx, affectedDelivery)
			if err != nil {
				return meh.Wrap(err, "look after affected delivery", meh.Details{"affected_delivery": affectedDelivery})
			}
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
