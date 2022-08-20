package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
)

// OperationByID retrieves a store.Operation by its id.
func (c *Controller) OperationByID(ctx context.Context, operationID uuid.UUID) (store.Operation, error) {
	var operation store.Operation
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		operation, err = c.Store.OperationByID(ctx, tx, operationID)
		if err != nil {
			return meh.Wrap(err, "operation by id from store", meh.Details{"operation_id": operationID})
		}
		return nil
	})
	if err != nil {
		return store.Operation{}, meh.Wrap(err, "run in tx", nil)
	}
	return operation, nil
}

// Operations retrieves a paginated store.Operation list. If limit-to-user is
// set, then only // operations are returned, the given user is member of.
func (c *Controller) Operations(ctx context.Context, operationFilters store.OperationRetrievalFilters,
	paginationParams pagination.Params) (pagination.Paginated[store.Operation], error) {
	var operations pagination.Paginated[store.Operation]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		operations, err = c.Store.Operations(ctx, tx, operationFilters, paginationParams)
		if err != nil {
			return meh.Wrap(err, "operations from store", meh.Details{"params": paginationParams})
		}
		return nil
	})
	if err != nil {
		return pagination.Paginated[store.Operation]{}, meh.Wrap(err, "run in tx", nil)
	}
	return operations, nil
}

// CreateOperation creates and notifies about the given store.Operation.
func (c *Controller) CreateOperation(ctx context.Context, operation store.Operation) (store.Operation, error) {
	var created store.Operation
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Create in store.
		var err error
		created, err = c.Store.CreateOperation(ctx, tx, operation)
		if err != nil {
			return meh.Wrap(err, "create operation in store", meh.Details{"operation": operation})
		}
		// Notify.
		err = c.Notifier.NotifyOperationCreated(ctx, tx, created)
		if err != nil {
			return meh.Wrap(err, "notify operation updated", meh.Details{"created": created})
		}
		err = c.Notifier.NotifyOperationMembersUpdated(ctx, tx, created.ID, []uuid.UUID{})
		if err != nil {
			return meh.Wrap(err, "ntoify operation members updated", meh.Details{"created_id": created.ID})
		}
		return nil
	})
	if err != nil {
		return store.Operation{}, meh.Wrap(err, "run in tx", nil)
	}
	return created, nil
}

// UpdateOperation updates and notifies about the given store.Operation,
// identified by its id.
func (c *Controller) UpdateOperation(ctx context.Context, operation store.Operation) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Update in store.
		err := c.Store.UpdateOperation(ctx, tx, operation)
		if err != nil {
			return meh.Wrap(err, "update in store", meh.Details{"operation": operation})
		}
		// Notify.
		err = c.Notifier.NotifyOperationUpdated(ctx, tx, operation)
		if err != nil {
			return meh.Wrap(err, "notify", meh.Details{"operation": operation})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// SearchOperations searches for operations with the given search.Params.
func (c *Controller) SearchOperations(ctx context.Context, operationFilters store.OperationRetrievalFilters,
	searchParams search.Params) (search.Result[store.Operation], error) {
	var result search.Result[store.Operation]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		result, err = c.Store.SearchOperations(ctx, tx, operationFilters, searchParams)
		if err != nil {
			return meh.Wrap(err, "search operations", meh.Details{"params": searchParams})
		}
		return nil
	})
	if err != nil {
		return search.Result[store.Operation]{}, meh.Wrap(err, "run in tx", nil)
	}
	return result, nil
}

// RebuildOperationSearch asynchronously rebuilds the operation-search.
func (c *Controller) RebuildOperationSearch(ctx context.Context) {
	c.Logger.Debug("rebuilding operation-search...")
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.RebuildOperationSearch(ctx, tx)
		if err != nil {
			return meh.Wrap(err, "rebuild operation-search in store", nil)
		}
		return nil
	})
	if err != nil {
		mehlog.Log(c.Logger, meh.Wrap(meh.Wrap(err, "run in tx", nil), "rebuild operation-search", nil))
		return
	}
	c.Logger.Debug("operation-search rebuilt")
}
