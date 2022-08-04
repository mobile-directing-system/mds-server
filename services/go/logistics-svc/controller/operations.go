package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// CreateOperation creates the operation with the given id.
func (c *Controller) CreateOperation(ctx context.Context, create store.Operation) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.CreateOperation(ctx, tx, create)
		if err != nil {
			return meh.Wrap(err, "create operation in store", meh.Details{"create": create})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// UpdateOperation updates the given operation, identified by its id.
func (c *Controller) UpdateOperation(ctx context.Context, update store.Operation) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.UpdateOperation(ctx, tx, update)
		if err != nil {
			return meh.Wrap(err, "update operation in store", meh.Details{"update": update})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// UpdateOperationMembersByOperation updates the operation members for the given
// operation.
func (c *Controller) UpdateOperationMembersByOperation(ctx context.Context, operationID uuid.UUID, newMembers []uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.UpdateOperationMembersByOperation(ctx, tx, operationID, newMembers)
		if err != nil {
			return meh.Wrap(err, "update operation members", nil)
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
