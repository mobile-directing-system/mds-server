package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
)

// UpdateOperationMembersByOperation updates the operation members for the given
// operation.
func (c *Controller) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error {
	err := c.store.UpdateOperationMembersByOperation(ctx, tx, operationID, newMembers)
	if err != nil {
		return meh.Wrap(err, "update operation members", nil)
	}
	return nil
}
