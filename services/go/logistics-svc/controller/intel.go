package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"go.uber.org/zap"
	"time"
)

// CreateIntel creates the given store.Intel.
func (c *Controller) CreateIntel(ctx context.Context, create store.CreateIntel) (store.Intel, error) {
	var created store.Intel
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Assure creator part of operation.
		ok, err := c.Store.IsUserOperationMember(ctx, tx, create.CreatedBy, create.Operation)
		if err != nil {
			return meh.Wrap(err, "is user operation member", meh.Details{
				"user":      create.CreatedBy,
				"operation": create.Operation,
			})
		}
		if !ok {
			return meh.NewForbiddenErr("creator is not operation member", meh.Details{
				"creator":   create.CreatedBy,
				"operation": create.Operation,
			})
		}
		// Search text.
		create.SearchText, err = c.genSearchText(create)
		// Create in store.
		created, err = c.Store.CreateIntel(ctx, tx, create)
		if err != nil {
			return meh.Wrap(err, "create in store", meh.Details{"create": create})
		}
		// Schedule initial deliveries.
		err = c.scheduleDeliveriesForIntel(ctx, tx, created.ID, create.InitialDeliverTo)
		if err != nil {
			return meh.Wrap(err, "schedule intel delivery", meh.Details{"intel_id": created.ID})
		}
		// Notify.
		err = c.Notifier.NotifyIntelCreated(ctx, tx, created)
		if err != nil {
			return meh.Wrap(err, "notify intel created", meh.Details{"created": created})
		}
		return nil
	})
	if err != nil {
		return store.Intel{}, meh.Wrap(err, "run in tx", nil)
	}
	return created, nil

}

// InvalidateIntelByID invalidates the intel with the given id after assuring,
// that the user is also part of the same operation.
func (c *Controller) InvalidateIntelByID(ctx context.Context, intelID uuid.UUID, by uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		intelToInvalidate, err := c.Store.IntelByID(ctx, tx, intelID)
		if err != nil {
			return meh.Wrap(err, "intel by id from store", meh.Details{"intel_id": intelID})
		}
		// Assure invalidator part of operation.
		ok, err := c.Store.IsUserOperationMember(ctx, tx, by, intelToInvalidate.Operation)
		if err != nil {
			return meh.Wrap(err, "is user operation member", meh.Details{
				"user_id":      by,
				"operation_id": intelToInvalidate.Operation,
			})
		}
		if !ok {
			return meh.NewForbiddenErr("user is not operation member", meh.Details{
				"invalidating_user": by,
				"operation":         intelToInvalidate.Operation,
			})
		}
		// Invalidate in store.
		err = c.Store.InvalidateIntelByID(ctx, tx, intelID)
		if err != nil {
			return meh.Wrap(err, "invalide intel in store", meh.Details{"intel_id": intelID})
		}
		// Notify about invalidated intel.
		err = c.Notifier.NotifyIntelInvalidated(ctx, tx, intelID, by)
		if err != nil {
			return meh.Wrap(err, "notfiy intel invalidated", meh.Details{"intel_id": intelID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// limitIntelFiltersToUser adjusts store.IntelFilters in order to filter by
// delivery-for-entries. These only include ones, associated with the usre with
// the given id. If no entries can be set as filter, the second return value is
// false and true, otherwise.
func (c *Controller) limitIntelFiltersToUser(ctx context.Context, tx pgx.Tx, intelFilters store.IntelFilters,
	userID uuid.UUID) (store.IntelFilters, bool, error) {
	// Retrieve entries associated with the user.
	userEntries, err := c.Store.AddressBookEntries(ctx, tx, store.AddressBookEntryFilters{ByUser: nulls.NewUUID(userID)},
		pagination.Params{Limit: 0})
	if err != nil {
		return store.IntelFilters{}, false, meh.Wrap(err, "address book entries by user", meh.Details{"user_id": userID})
	}
	if len(userEntries.Entries) == 0 {
		// No filters can be set -> We do not perform any operation in order to not
		// expose unwanted intel.
		return store.IntelFilters{}, false, nil
	}
	// Remove all entries, not being associated with the user.
	newEntryFilter := make([]uuid.UUID, 0, len(intelFilters.OneOfDeliveryForEntries))
	for _, entryInFilter := range intelFilters.OneOfDeliveryForEntries {
		for _, userEntry := range userEntries.Entries {
			if entryInFilter != userEntry.ID {
				continue
			}
			// Entry was found -> keep it.
			newEntryFilter = append(newEntryFilter, entryInFilter)
			break
		}
	}
	intelFilters.OneOfDeliveryForEntries = newEntryFilter
	// Assure filters set.
	if len(intelFilters.OneOfDeliveryForEntries) == 0 {
		// Replace with all of the user's entries.
		for _, entry := range userEntries.Entries {
			intelFilters.OneOfDeliveryForEntries = append(intelFilters.OneOfDeliveryForEntries, entry.ID)
		}
	}
	return intelFilters, true, nil
}

// SearchIntel searches for intel with the given search.Params.
func (c *Controller) SearchIntel(ctx context.Context, intelFilters store.IntelFilters,
	searchParams search.Params, limitToUser uuid.NullUUID) (search.Result[store.Intel], error) {
	var result search.Result[store.Intel]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		// If limit to user is set, we need to set delivery-filters accordingly in order
		// to limit intel-visibility.
		if limitToUser.Valid {
			var ok bool
			intelFilters, ok, err = c.limitIntelFiltersToUser(ctx, tx, intelFilters, limitToUser.UUID)
			if err != nil {
				return meh.Wrap(err, "limit intel filters to user", meh.Details{
					"intel_filters": intelFilters,
					"limit_to_user": limitToUser.UUID,
				})
			}
			if !ok {
				// Skip search.
				result = search.Result[store.Intel]{}
				return nil
			}
		}
		result, err = c.Store.SearchIntel(ctx, tx, intelFilters, searchParams)
		if err != nil {
			return meh.Wrap(err, "search operations", meh.Details{"params": searchParams})
		}
		return nil
	})
	if err != nil {
		return search.Result[store.Intel]{}, meh.Wrap(err, "run in tx", nil)
	}
	return result, nil
}

// RebuildIntelSearch rebuilds the intel-search.
func (c *Controller) RebuildIntelSearch(ctx context.Context) {
	c.Logger.Debug("rebuilding intel-search...")
	start := time.Now()
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.RebuildIntelSearch(ctx, tx)
		if err != nil {
			return meh.Wrap(err, "rebuild intel-search in store", nil)
		}
		return nil
	})
	if err != nil {
		err = meh.Wrap(err, "run in tx", nil)
		mehlog.Log(c.Logger, err)
		return
	}
	c.Logger.Debug("intel-search rebuilt", zap.Duration("took", time.Since(start)))
}

// IntelByID retrieves the store.Intel with the given id. If limit-to-user is
// set, all deliveries for the intel are checked for target-address-book-entries
// being associated with the user with the given id.
func (c *Controller) IntelByID(ctx context.Context, intelID uuid.UUID, limitToUser uuid.NullUUID) (store.Intel, error) {
	var intel store.Intel
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		// Assure allowed to retrieve.
		if limitToUser.Valid {
			associatedUserIDs, err := c.Store.UsersWithDeliveriesByIntel(ctx, tx, intelID)
			if err != nil {
				return meh.Wrap(err, "users with deliveries by intel from store", meh.Details{"intel_id": intelID})
			}
			ok := false
			for _, associatedUser := range associatedUserIDs {
				if limitToUser.UUID != associatedUser {
					continue
				}
				ok = true
				break
			}
			if !ok {
				return meh.NewForbiddenErr("user has no associated target-delivery-address-book-entries",
					meh.Details{"user_id": limitToUser.UUID})
			}
		}
		// Retrieve intel.
		intel, err = c.Store.IntelByID(ctx, tx, intelID)
		if err != nil {
			return meh.Wrap(err, "intel by id from store", meh.Details{"intel_id": intelID})
		}
		return nil
	})
	if err != nil {
		return store.Intel{}, meh.Wrap(err, "run in tx", nil)
	}
	return intel, nil
}

// Intel retrieves a paginated store.Intel list using the given
// store.IntelFilters and pagination.Params, sorted descending by creation date.
//
// Warning: Sorting via pagination.Params is discarded!
func (c *Controller) Intel(ctx context.Context, intelFilters store.IntelFilters,
	paginationParams pagination.Params, limitToUser uuid.NullUUID) (pagination.Paginated[store.Intel], error) {
	paginationParams.OrderBy = nulls.String{}
	var result pagination.Paginated[store.Intel]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		// If limit to user is set, we need to set delivery-filters accordingly in order
		// to limit intel-visibility.
		if limitToUser.Valid {
			var ok bool
			intelFilters, ok, err = c.limitIntelFiltersToUser(ctx, tx, intelFilters, limitToUser.UUID)
			if err != nil {
				return meh.Wrap(err, "limit intel filters to user", meh.Details{
					"intel_filters": intelFilters,
					"limit_to_user": limitToUser.UUID,
				})
			}
			if !ok {
				// Skip search.
				result = pagination.NewPaginated(paginationParams, make([]store.Intel, 0), 0)
				return nil
			}
		}
		// Search.
		result, err = c.Store.Intel(ctx, tx, intelFilters, paginationParams)
		if err != nil {
			return meh.Wrap(err, "intel from store", meh.Details{
				"intel_filters":     intelFilters,
				"pagination_params": paginationParams,
			})
		}
		return nil
	})
	if err != nil {
		return pagination.Paginated[store.Intel]{}, meh.Wrap(err, "run in tx", nil)
	}
	return result, nil
}
