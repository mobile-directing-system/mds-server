package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"go.uber.org/zap"
	"time"
)

// AddressBookEntryByID retrieves the store.AddressBookEntryDetailed with the
// given id. If visible-by is given, an meh.ErrNotFound will be returned, if the
// entry is associated with a user, that is not part of any operation, the
// client (visible-by) is part of.
func (c *Controller) AddressBookEntryByID(ctx context.Context, entryID uuid.UUID, visibleBy uuid.NullUUID) (store.AddressBookEntryDetailed, error) {
	var entry store.AddressBookEntryDetailed
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		entry, err = c.Store.AddressBookEntryByID(ctx, tx, entryID, visibleBy)
		if err != nil {
			return meh.Wrap(err, "address book entry from store", meh.Details{"entry_id": entryID})
		}
		return nil
	})
	if err != nil {
		return store.AddressBookEntryDetailed{}, meh.Wrap(err, "run in tx", nil)
	}
	return entry, nil
}

// AddressBookEntries retrieves a paginated store.AddressBookEntryDetailed list.
func (c *Controller) AddressBookEntries(ctx context.Context, filters store.AddressBookEntryFilters,
	paginationParams pagination.Params) (pagination.Paginated[store.AddressBookEntryDetailed], error) {
	var entries pagination.Paginated[store.AddressBookEntryDetailed]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		entries, err = c.Store.AddressBookEntries(ctx, tx, filters, paginationParams)
		if err != nil {
			return meh.Wrap(err, "address book entries from store", meh.Details{
				"filters":           filters,
				"pagination_params": paginationParams,
			})
		}
		return nil
	})
	if err != nil {
		return pagination.Paginated[store.AddressBookEntryDetailed]{}, meh.Wrap(err, "run in tx", nil)
	}
	return entries, nil
}

// CreateAddressBookEntry creates the given store.AddressBookEntry and notifies
// via Notifier.NotifyAddressBookEntryCreated.
func (c *Controller) CreateAddressBookEntry(ctx context.Context, create store.AddressBookEntry) (store.AddressBookEntryDetailed, error) {
	var created store.AddressBookEntryDetailed
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		created, err = c.Store.CreateAddressBookEntry(ctx, tx, create)
		if err != nil {
			return meh.Wrap(err, "create entry in store", meh.Details{"entry": create})
		}
		err = c.Notifier.NotifyAddressBookEntryCreated(ctx, tx, created.AddressBookEntry)
		if err != nil {
			return meh.Wrap(err, "notify", meh.Details{"created": created.AddressBookEntry})
		}
		return nil
	})
	if err != nil {
		return store.AddressBookEntryDetailed{}, meh.Wrap(err, "run in tx", nil)
	}
	return created, nil
}

// UpdateAddressBookEntry updates the given store.AddressBookEntry and notifies
// via Notifier.NotifyAddressBookEntryUpdated. If limit-to-user is set, only
// entries, associated with this user can be updated.
func (c *Controller) UpdateAddressBookEntry(ctx context.Context, update store.AddressBookEntry, limitToUser uuid.NullUUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		entry, err := c.Store.AddressBookEntryByID(ctx, tx, update.ID, uuid.NullUUID{})
		if err != nil {
			return meh.Wrap(err, "entry from store by id", meh.Details{"id": update.ID})
		}
		if limitToUser.Valid {
			if !entry.User.Valid || entry.User.UUID != limitToUser.UUID {
				return meh.NewNotFoundErr("limited to user", meh.Details{
					"entry_user_id_set": entry.User.Valid,
					"entry_user_id":     entry.User.UUID,
					"limit_to_user":     limitToUser.UUID,
				})
			}
		}
		err = c.Store.UpdateAddressBookEntry(ctx, tx, update)
		if err != nil {
			return meh.Wrap(err, "update entry in store", meh.Details{"entry": update})
		}
		err = c.Notifier.NotifyAddressBookEntryUpdated(ctx, tx, update)
		if err != nil {
			return meh.Wrap(err, "notify", meh.Details{"updated": update})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// DeleteAddressBookEntryByID deletes the address book entry with the given id
// and notifies via Notifier.NotifyAddressBookEntryDeleted. If limit-to-user is
// set, only entries, associated with the given user can be deleted. Otherwise,
// an meh.ErrNotFound will be returned.
func (c *Controller) DeleteAddressBookEntryByID(ctx context.Context, entryID uuid.UUID, limitToUser uuid.NullUUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		entry, err := c.Store.AddressBookEntryByID(ctx, tx, entryID, uuid.NullUUID{})
		if err != nil {
			return meh.Wrap(err, "entry from store by id", meh.Details{"id": entryID})
		}
		if limitToUser.Valid {
			if !entry.User.Valid || entry.User.UUID != limitToUser.UUID {
				return meh.NewNotFoundErr("limited to user", meh.Details{
					"entry_user_id_set": entry.User.Valid,
					"entry_user_id":     entry.User.UUID,
					"limit_to_user":     limitToUser.UUID,
				})
			}
		}
		err = c.Store.DeleteAddressBookEntryByID(ctx, tx, entryID)
		if err != nil {
			return meh.Wrap(err, "delete entry in store", meh.Details{"entry_id": entryID})
		}
		err = c.Notifier.NotifyAddressBookEntryDeleted(ctx, tx, entryID)
		if err != nil {
			return meh.Wrap(err, "notify", meh.Details{"entry_id": entryID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// SearchAddressBookEntries searches for address book entries with the given
// store.AddressBookEntryFilters and search.Params.
func (c *Controller) SearchAddressBookEntries(ctx context.Context, entryFilters store.AddressBookEntryFilters,
	searchParams search.Params) (search.Result[store.AddressBookEntryDetailed], error) {
	var result search.Result[store.AddressBookEntryDetailed]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		result, err = c.Store.SearchAddressBookEntries(ctx, tx, entryFilters, searchParams)
		if err != nil {
			return meh.Wrap(err, "search address book entries in store", meh.Details{
				"entry_filters": entryFilters,
				"search_params": searchParams,
			})
		}
		return nil
	})
	if err != nil {
		return search.Result[store.AddressBookEntryDetailed]{}, meh.Wrap(err, "run in tx", nil)
	}
	return result, nil
}

// RebuildAddressBookEntrySearch rebuilds the address-book-entry-search.
func (c *Controller) RebuildAddressBookEntrySearch(ctx context.Context) {
	c.Logger.Debug("rebuilding address-book-entry-search...")
	start := time.Now()
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.RebuildAddressBookEntrySearch(ctx, tx)
		if err != nil {
			return meh.Wrap(err, "rebuild address-book-entry-search in store", nil)
		}
		return nil
	})
	if err != nil {
		err = meh.Wrap(err, "run in tx", nil)
		mehlog.Log(c.Logger, err)
		return
	}
	c.Logger.Debug("address-book-entry-search rebuilt", zap.Duration("took", time.Since(start)))
}

// IsAutoIntelDeliveryEnabledForAddressBookEntry checks whether auto-delivery is
// enabled for the address book entry with the given id.
func (c *Controller) IsAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID) (bool, error) {
	var isEnabled bool
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		isEnabled, err = c.Store.IsAutoDeliveryEnabledForAddressBookEntry(ctx, tx, entryID)
		if err != nil {
			return meh.Wrap(err, "is auto delivery enabled for address book entry in store", meh.Details{"entry_id": entryID})
		}
		return nil
	})
	if err != nil {
		return false, meh.Wrap(err, "run in tx", nil)
	}
	return isEnabled, nil
}

// SetAutoIntelDeliveryEnabledForAddressBookEntry sets auto-delivery
// enabled/disabled for the address book entry with the given id.
func (c *Controller) SetAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID, enabled bool) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.SetAutoDeliveryEnabledForAddressBookEntry(ctx, tx, entryID, enabled)
		if err != nil {
			return meh.Wrap(err, "set auto delivery enabled for address book entry in store", meh.Details{
				"entry_id": entryID,
				"enabled":  enabled,
			})
		}
		err = c.Notifier.NotifyAddressBookEntryAutoDeliveryUpdated(ctx, tx, entryID, enabled)
		if err != nil {
			return meh.Wrap(err, "notify address book entry auto delivery updated", meh.Details{
				"entry_id": entryID,
				"enabled":  enabled,
			})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// SetAddressBookEntriesWithAutoDeliveryEnabled sets the list of address book
// entries with auto-delivery being enabled to the given ones.
func (c *Controller) SetAddressBookEntriesWithAutoDeliveryEnabled(ctx context.Context, entryIDs []uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		disabled, err := c.Store.SetAddressBookEntriesWithAutoDeliveryEnabled(ctx, tx, entryIDs)
		if err != nil {
			return meh.Wrap(err, "set address book entries with auto-delivery enabled in store",
				meh.Details{"new_entries_with_auto_delivery_enabled": entryIDs})
		}
		for _, enabledEntryID := range entryIDs {
			err = c.Notifier.NotifyAddressBookEntryAutoDeliveryUpdated(ctx, tx, enabledEntryID, true)
			if err != nil {
				return meh.Wrap(err, "notify address book entry auto delivery enabled", meh.Details{"entry_id": enabledEntryID})
			}
		}
		for _, disabledEntryID := range disabled {
			err = c.Notifier.NotifyAddressBookEntryAutoDeliveryUpdated(ctx, tx, disabledEntryID, false)
			if err != nil {
				return meh.Wrap(err, "notify address book entry auto delivery disabled", meh.Details{"entry_id": disabledEntryID})
			}
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
