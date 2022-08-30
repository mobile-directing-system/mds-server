package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
)

// CreateAddressBookEntry creates the given store.AddressBookEntry.
func (c *Controller) CreateAddressBookEntry(ctx context.Context, tx pgx.Tx, create store.AddressBookEntry) error {
	err := c.Store.CreateAddressBookEntry(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create address book entry in store", meh.Details{"create": create})
	}
	return nil
}

// UpdateAddressBookEntry updates the given store.AddressBookEntry.
func (c *Controller) UpdateAddressBookEntry(ctx context.Context, tx pgx.Tx, update store.AddressBookEntry) error {
	err := c.Store.UpdateAddressBookEntry(ctx, tx, update)
	if err != nil {
		return meh.Wrap(err, "update address book entry in store", meh.Details{"update": update})
	}
	return nil
}

// DeleteAddressBookEntryByID deletes the address book entry with the given id.
func (c *Controller) DeleteAddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	err := c.Store.DeleteAddressBookEntryByID(ctx, tx, entryID)
	if err != nil {
		return meh.Wrap(err, "delete address book entry in store", meh.Details{"entry_id": entryID})
	}
	return nil
}
