package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"golang.org/x/sync/errgroup"
)

// CreateOperation creates the operation with the given id.
func (c *Controller) CreateOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) error {
	err := c.Store.CreateOperation(ctx, tx, operationID)
	if err != nil {
		return meh.Wrap(err, "create operation in store", meh.Details{"operation_id": operationID})
	}
	return nil
}

// UpdateOperationMembersByOperation updates the members for the operation with
// the given id.
func (c *Controller) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error {
	// Check for members that were unassigned from the operation.
	oldMembers, err := c.Store.OperationMembersByOperation(ctx, tx, operationID)
	if err != nil {
		return meh.Wrap(err, "old operation members from store", meh.Details{"operation_id": operationID})
	}
	newMembersMapped := make(map[uuid.UUID]struct{}, len(oldMembers))
	for _, newMember := range newMembers {
		newMembersMapped[newMember] = struct{}{}
	}
	removed := make([]uuid.UUID, 0)
	for _, oldMember := range oldMembers {
		if _, ok := newMembersMapped[oldMember]; !ok {
			removed = append(removed, oldMember)
		}
	}
	updatedGroups := make([]store.Group, 0)
	if len(removed) > 0 {
		// Remove from groups that are associated with the operation.
		operationGroups, err := c.Store.Groups(ctx, tx, store.GroupFilters{
			ForOperation:  nulls.NewUUID(operationID),
			ExcludeGlobal: true,
		}, pagination.Params{Limit: 0})
		if err != nil {
			return meh.Wrap(err, "groups for operation", meh.Details{"operation_id": operationID})
		}
		for _, operationGroup := range operationGroups.Entries {
			newGroupMembers := make([]uuid.UUID, 0, len(operationGroup.Members))
			needsUpdate := false
			for _, groupMember := range operationGroup.Members {
				isInRemoved := false
				for _, removedMember := range removed {
					if groupMember == removedMember {
						isInRemoved = true
						break
					}
				}
				if isInRemoved {
					needsUpdate = true
					continue
				}
				newGroupMembers = append(newGroupMembers, groupMember)
			}
			if !needsUpdate {
				continue
			}
			// Update group.
			updatedGroup := operationGroup
			updatedGroup.Members = newGroupMembers
			err = c.Store.UpdateGroup(ctx, tx, updatedGroup)
			if err != nil {
				return meh.Wrap(err, "update group", meh.Details{"group": updatedGroup})
			}
			updatedGroups = append(updatedGroups, updatedGroup)
		}
	}
	// Update operation members.
	err = c.Store.UpdateOperationMembersByOperation(ctx, tx, operationID, newMembers)
	if err != nil {
		return meh.Wrap(err, "update operation members in store", meh.Details{
			"operation_id": operationID,
			"new_members":  newMembers,
		})
	}
	// Notify of updated groups.
	if len(updatedGroups) == 0 {
		return nil
	}
	var eg errgroup.Group
	for _, updatedGroup := range updatedGroups {
		group := updatedGroup
		eg.Go(func() error {
			err := c.Notifier.NotifyGroupUpdated(ctx, tx, group)
			if err != nil {
				return meh.Wrap(err, "notify group updated", meh.Details{"group": group})
			}
			return nil
		})
	}
	return eg.Wait()
}
