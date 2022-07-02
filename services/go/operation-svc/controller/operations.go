package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
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

// Operations retrieves a paginated store.Operation list.
func (c *Controller) Operations(ctx context.Context, params pagination.Params) (pagination.Paginated[store.Operation], error) {
	var operations pagination.Paginated[store.Operation]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		operations, err = c.Store.Operations(ctx, tx, params)
		if err != nil {
			return meh.Wrap(err, "operations from store", meh.Details{"params": params})
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
		err = c.Notifier.NotifyOperationCreated(created)
		if err != nil {
			return meh.Wrap(err, "notify", meh.Details{"created": created})
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
		err = c.Notifier.NotifyOperationUpdated(operation)
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
