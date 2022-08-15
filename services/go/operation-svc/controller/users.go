package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"golang.org/x/sync/errgroup"
)

// CreateUser creates the given store.User.
func (c *Controller) CreateUser(ctx context.Context, tx pgx.Tx, create store.User) error {
	err := c.Store.CreateUser(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create user in store", meh.Details{"user": create})
	}
	return nil
}

// UpdateUser updates the given store.User, identified by its id.
func (c *Controller) UpdateUser(ctx context.Context, tx pgx.Tx, update store.User) error {
	err := c.Store.UpdateUser(ctx, tx, update)
	if err != nil {
		return meh.Wrap(err, "update user in store", meh.Details{"user": update})
	}
	return nil
}

// DeleteUserByID deletes the user with the given id in the store, unassigns
// from all operations and notifies about changed operation members.
func (c *Controller) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Delete from all operations, currently assigned to.
	currentlyMemberOf, err := c.Store.OperationsByMember(ctx, tx, userID)
	if err != nil {
		return meh.Wrap(err, "operations from store", meh.Details{"user_id": userID})
	}
	// Remove from all operations.
	updatedMembersByOperation := make(map[uuid.UUID][]uuid.UUID, len(currentlyMemberOf))
	for _, operation := range currentlyMemberOf {
		operationMembers, err := c.Store.OperationMembersByOperation(ctx, tx, operation.ID, pagination.Params{Limit: 0})
		if err != nil {
			return meh.Wrap(err, "operation members for assigned operation",
				meh.Details{"operation_id": operation.ID})
		}
		newMembers := make([]uuid.UUID, 0, len(operationMembers.Entries))
		for _, member := range operationMembers.Entries {
			if member.ID != userID {
				newMembers = append(newMembers, member.ID)
			}
		}
		err = c.Store.UpdateOperationMembersByOperation(ctx, tx, operation.ID, newMembers)
		if err != nil {
			return meh.Wrap(err, "update operation members for former assigned operation", meh.Details{
				"operation_id": operation.ID,
				"new_members":  newMembers,
			})
		}
		updatedMembersByOperation[operation.ID] = newMembers
	}
	// Delete user itself.
	err = c.Store.DeleteUserByID(ctx, tx, userID)
	if err != nil {
		return meh.Wrap(err, "delete user in store", meh.Details{"user_id": userID})
	}
	// Notify of updated operations.
	var eg errgroup.Group
	for tmpOperationID := range updatedMembersByOperation {
		operationID := tmpOperationID
		newMembers := updatedMembersByOperation[tmpOperationID]
		eg.Go(func() error {
			err := c.Notifier.NotifyOperationMembersUpdated(ctx, tx, operationID, newMembers)
			if err != nil {
				return meh.Wrap(err, "notify operation members updated", meh.Details{
					"operation_id": operationID,
					"new_members":  newMembers,
				})
			}
			return nil
		})
	}
	return eg.Wait()
}
