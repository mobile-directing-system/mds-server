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

// UpdateOperationMembersByOperation updates the members for the operation with
// the given id and notifies of the updated member list.
func (c *Controller) UpdateOperationMembersByOperation(ctx context.Context, operationID uuid.UUID, members []uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Assure operation exists.
		_, err := c.Store.OperationByID(ctx, tx, operationID)
		if err != nil {
			return meh.Wrap(err, "operation by id", meh.Details{"operation_id": operationID})
		}
		// Update in store.
		err = c.Store.UpdateOperationMembersByOperation(ctx, tx, operationID, members)
		if err != nil {
			return meh.Wrap(err, "update operation members in store", meh.Details{
				"operation_id": operationID,
				"membres":      members,
			})
		}
		// Notify.
		err = c.Notifier.NotifyOperationMembersUpdated(operationID, members)
		if err != nil {
			return meh.Wrap(err, "notify operation members updated", meh.Details{
				"operation_id": operationID,
				"members":      members,
			})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// OperationMembersByOperation retrieves a paginated store.user list for the
// operation with the given id.
func (c *Controller) OperationMembersByOperation(ctx context.Context, operationID uuid.UUID,
	paginationParams pagination.Params) (pagination.Paginated[store.User], error) {
	var members pagination.Paginated[store.User]
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		// Assure operation exists.
		_, err = c.Store.OperationByID(ctx, tx, operationID)
		if err != nil {
			return meh.Wrap(err, "operation from store", meh.Details{"operation_id": operationID})
		}
		// Retrieve members.
		members, err = c.Store.OperationMembersByOperation(ctx, tx, operationID, paginationParams)
		if err != nil {
			return meh.Wrap(err, "operation members from store", meh.Details{
				"operation_id":      operationID,
				"pagination_params": paginationParams,
			})
		}
		return nil
	})
	if err != nil {
		return pagination.Paginated[store.User]{}, meh.Wrap(err, "run in tx", nil)
	}
	return members, nil
}
