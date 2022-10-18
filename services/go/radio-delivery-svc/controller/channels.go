package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
)

// UpdateRadioChannelsByEntry deletes and recreates the radio-channels for the
// address book entry with the given id.
func (c *Controller) UpdateRadioChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID, create []store.RadioChannel) error {
	// Delete old channels.
	err := c.store.DeleteRadioChannelsByEntry(ctx, tx, entryID)
	if err != nil {
		return meh.Wrap(err, "delete radio-channels by entry in store", meh.Details{"entry_id": entryID})
	}
	// Create new channels.
	for _, newChannel := range create {
		err = c.store.CreateRadioChannel(ctx, tx, newChannel)
		if err != nil {
			return meh.Wrap(err, "create radio-channel in store", meh.Details{"create": create})
		}
	}
	return nil
}

// DeleteRadioChannelsByEntry deletes all radio-channels for the entry with the
// given id.
func (c *Controller) DeleteRadioChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	err := c.store.DeleteRadioChannelsByEntry(ctx, tx, entryID)
	if err != nil {
		return meh.Wrap(err, "delete radio channels by entry in store", meh.Details{"entry_id": entryID})
	}
	return nil
}
