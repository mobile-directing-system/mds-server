package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// CreateIntel assures that the creator of the given store.CreateIntel is member
// of the associated operation. It then creates the intel in the store and
// notifies via Notifier.NotifyIntelCreated.
func (c *Controller) CreateIntel(ctx context.Context, create store.CreateIntel) (store.Intel, error) {
	var created store.Intel
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Assure creator part of operation.
		ok, err := c.Store.IsUserOperationMember(ctx, tx, create.CreatedBy, create.Operation)
		if err != nil {
			return meh.Wrap(err, "is user operation member", meh.Details{
				"user":      create.CreatedBy,
				"operation": create.Operation,
			})
		}
		if !ok {
			return meh.NewForbiddenErr("creator is not operation member", meh.Details{
				"creator":   create.CreatedBy,
				"operation": create.Operation,
			})
		}
		// Assure assignment-recipients exist.
		for _, assignment := range create.Assignments {
			_, err = c.Store.AddressBookEntryByID(ctx, tx, assignment.To)
			if err != nil {
				if meh.ErrorCode(err) == meh.ErrNotFound {
					return meh.NewBadInputErr("assignment-to not found", meh.Details{"assignment_to": assignment.To})
				}
				return meh.Wrap(err, "address book entry by id", meh.Details{"entry_id": assignment.To})
			}
		}
		// Create in store.
		created, err = c.Store.CreateIntel(ctx, tx, create)
		if err != nil {
			return meh.Wrap(err, "create in store", meh.Details{"create": create})
		}
		// Notify.
		err = c.Notifier.NotifyIntelCreated(ctx, tx, created)
		if err != nil {
			return meh.Wrap(err, "notify intel created", meh.Details{"created": created})
		}
		return nil
	})
	if err != nil {
		return store.Intel{}, meh.Wrap(err, "run in tx", nil)
	}
	return created, nil
}

// InvalidateIntelByID invalidates the intel with the given id after assuring,
// that the user is also part of the same operation.
func (c *Controller) InvalidateIntelByID(ctx context.Context, intelID uuid.UUID, by uuid.UUID) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		intelToInvalidate, err := c.Store.IntelByID(ctx, tx, intelID)
		if err != nil {
			return meh.Wrap(err, "intel by id from store", meh.Details{"intel_id": intelID})
		}
		// Assure invalidator part of operation.
		ok, err := c.Store.IsUserOperationMember(ctx, tx, by, intelToInvalidate.Operation)
		if err != nil {
			return meh.Wrap(err, "is user operation member", meh.Details{
				"user_id":      by,
				"operation_id": intelToInvalidate.Operation,
			})
		}
		if !ok {
			return meh.NewForbiddenErr("user is not operation member", meh.Details{
				"invalidating_user": by,
				"operation":         intelToInvalidate.Operation,
			})
		}
		// Invalidate in store.
		err = c.Store.InvalidateIntelByID(ctx, tx, intelID)
		if err != nil {
			return meh.Wrap(err, "invalide intel in store", meh.Details{"intel_id": intelID})
		}
		// Notify about invalidated intel.
		err = c.Notifier.NotifyIntelInvalidated(ctx, tx, intelID, by)
		if err != nil {
			return meh.Wrap(err, "notfiy intel invalidated", meh.Details{"intel_id": intelID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}

// IntelByID retrieves the store.Intel with the given id. If the limit is given,
// it assures, that it is assigned to the user and returns a meh.ErrForbidden
// otherwise.
func (c *Controller) IntelByID(ctx context.Context, intelID uuid.UUID, limitToAssignedUser uuid.NullUUID) (store.Intel, error) {
	var intel store.Intel
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		intel, err = c.Store.IntelByID(ctx, tx, intelID)
		if err != nil {
			return meh.Wrap(err, "intel by id", meh.Details{"intel_id": intelID})
		}
		if limitToAssignedUser.Valid {
			foundInAssignments := false
			for _, assignment := range intel.Assignments {
				if assignment.To == limitToAssignedUser.UUID {
					foundInAssignments = true
					break
				}
			}
			if !foundInAssignments {
				return meh.NewForbiddenErr("intel not assigned to user", meh.Details{"target_user": limitToAssignedUser.UUID})
			}
		}
		return nil
	})
	if err != nil {
		return store.Intel{}, meh.Wrap(err, "run in tx", nil)
	}
	return intel, nil
}
