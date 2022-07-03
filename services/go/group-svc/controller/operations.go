package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// CreateOperation creates the operation with the given id.
func (c *Controller) CreateOperation(ctx context.Context, operationID uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.Store.CreateOperation(ctx, tx, operationID)
		if err != nil {
			return meh.Wrap(err, "create operation in store", meh.Details{"operation_id": operationID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
