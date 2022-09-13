package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
)

// CreateIntel creates the given store.Intel.
func (c *Controller) CreateIntel(ctx context.Context, tx pgx.Tx, create store.Intel) error {
	err := c.Store.CreateIntel(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create intel in store", meh.Details{"create": create})
	}
	err = c.scheduleDeliveriesForIntelAssignments(ctx, tx, create.ID)
	if err != nil {
		return meh.Wrap(err, "schedule intel delivery", meh.Details{"intel_id": create.ID})
	}
	return nil
}

// InvalidateIntelByID sets the valid-field of the intel with the given id to
// false.
func (c *Controller) InvalidateIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error {
	err := c.Store.InvalidateIntelByID(ctx, tx, intelID)
	if err != nil {
		return meh.Wrap(err, "invalidate intel in store", meh.Details{"intel_id": intelID})
	}
	return nil
}
