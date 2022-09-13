package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
)

// UpdateNotificationChannelsByEntry deletes and recreates the
// notification-channels for the address book entry with the given id.
func (c *Controller) UpdateNotificationChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID, create []store.NotificationChannel) error {
	// Delete old channels.
	err := c.store.DeleteNotificationChannelsByEntry(ctx, tx, entryID)
	if err != nil {
		return meh.Wrap(err, "delete notification-channels by entry in store", meh.Details{"entry_id": entryID})
	}
	// Create new channels.
	for _, newChannel := range create {
		err = c.store.CreateNotificationChannel(ctx, tx, newChannel)
		if err != nil {
			return meh.Wrap(err, "create notification-channel in store", meh.Details{"create": create})
		}
	}
	return nil
}

// DeleteNotificationChannelsByEntry deletes all notification-channels for the
// entry with the given id.
func (c *Controller) DeleteNotificationChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	err := c.store.DeleteNotificationChannelsByEntry(ctx, tx, entryID)
	if err != nil {
		return meh.Wrap(err, "delete notification channels by entry in store", meh.Details{"entry_id": entryID})
	}
	return nil
}
