package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// CreateIntel creates the given store.Intel.
func (c *Controller) CreateIntel(ctx context.Context, create store.Intel) error {
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		err := c.store.CreateIntel(ctx, tx, create)
		if err != nil {
			return meh.Wrap(err, "create intel in store", meh.Details{"create": create})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// InvalidateIntelByID invalidates the intel with the given id.
func (c *Controller) InvalidateIntelByID(ctx context.Context, intelID uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		err := c.store.InvalidateIntelByID(ctx, tx, intelID)
		if err != nil {
			return meh.Wrap(err, "invalidate intel by id in store", meh.Details{"intel_id": intelID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
