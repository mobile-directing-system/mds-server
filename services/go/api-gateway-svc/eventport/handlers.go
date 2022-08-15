package eventport

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
)

// Handler for received Kafka messages.
type Handler interface {
	// CreateUser creates the given store.User.
	CreateUser(ctx context.Context, tx pgx.Tx, user store.UserWithPass) error
	// UpdateUser updates the given store.User in the Store.
	UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// UpdateUserPassByUserID updates the password for the user with the given id in
	// the store and notifies via Notifier.NotifyUserPassUpdated.
	UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newPass []byte) error
	// DeleteUserByID deletes the user with the given id in the store and notifies
	// via Notifier.NotifyUserDeleted.
	DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// UpdatePermissionsByUser updates the permissions for the given user.
	UpdatePermissionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, updatedPermissions []permission.Permission) error
}

// HandlerFn is the handler for Kafka messages.
func (p *Port) HandlerFn(handler Handler) kafkautil.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, message kafkautil.InboundMessage) error {
		switch message.Topic {
		case event.PermissionsTopic:
			return meh.NilOrWrap(p.handlePermissionsTopic(ctx, tx, handler, message), "handle permissions topic", nil)
		case event.UsersTopic:
			return meh.NilOrWrap(p.handleUsersTopic(ctx, tx, handler, message), "handle users topic", nil)
		}
		return nil
	}
}

// handlePermissionsTopic handles the event.PermissionsTopic.
func (p *Port) handlePermissionsTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypePermissionsUpdated:
		return meh.NilOrWrap(p.handlePermissionsUpdated(ctx, tx, handler, message), "handler permissions updated", nil)
	}
	return nil
}

// handlePermissionsUpdated handles an event.PermissionsUpdated.
func (p *Port) handlePermissionsUpdated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var permissionsUpdatedEvent event.PermissionsUpdated
	err := json.Unmarshal(message.RawValue, &permissionsUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.UpdatePermissionsByUser(ctx, tx, permissionsUpdatedEvent.User, permissionsUpdatedEvent.Permissions)
	if err != nil {
		return meh.Wrap(err, "update permissions", meh.Details{
			"user_id":             permissionsUpdatedEvent.User,
			"updated_permissions": permissionsUpdatedEvent.Permissions,
		})
	}
	return nil
}

// handleUsersTopic handles the event.UsersTopic.
func (p *Port) handleUsersTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypeUserCreated:
		return meh.NilOrWrap(p.handleUserCreated(ctx, tx, handler, message), "handle user created", nil)
	case event.TypeUserUpdated:
		return meh.NilOrWrap(p.handleUserUpdated(ctx, tx, handler, message), "handle user updated", nil)
	case event.TypeUserPassUpdated:
		return meh.NilOrWrap(p.handleUserPassUpdated(ctx, tx, handler, message), "handle user pass updated", nil)
	case event.TypeUserDeleted:
		return meh.NilOrWrap(p.handleUserDeleted(ctx, tx, handler, message), "handle user deleted", nil)
	}
	return nil
}

// handleUserCreated handles an event.UserCreated.
func (p *Port) handleUserCreated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var userCreatedEvent event.UserCreated
	err := json.Unmarshal(message.RawValue, &userCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.CreateUser(ctx, tx, store.UserWithPass{
		User: store.User{
			ID:       userCreatedEvent.ID,
			Username: userCreatedEvent.Username,
			IsAdmin:  userCreatedEvent.IsAdmin,
		},
		Pass: userCreatedEvent.Pass,
	})
	if err != nil {
		return meh.Wrap(err, "create user", nil)
	}
	return nil
}

// handleUserUpdated handles an event.UserUpdated.
func (p *Port) handleUserUpdated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var userUpdatedEvent event.UserUpdated
	err := json.Unmarshal(message.RawValue, &userUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.UpdateUser(ctx, tx, store.User{
		ID:       userUpdatedEvent.ID,
		Username: userUpdatedEvent.Username,
		IsAdmin:  userUpdatedEvent.IsAdmin,
	})
	if err != nil {
		return meh.Wrap(err, "update user", nil)
	}
	return nil
}

// handleUserPassUpdated handles an event.UserPassUpdated.
func (p *Port) handleUserPassUpdated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var userPassUpdatedEvent event.UserPassUpdated
	err := json.Unmarshal(message.RawValue, &userPassUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.UpdateUserPassByUserID(ctx, tx, userPassUpdatedEvent.User, userPassUpdatedEvent.NewPass)
	if err != nil {
		return meh.Wrap(err, "update user pass", nil)
	}
	return nil
}

// handleUserDeleted handles an event.UserDeleted.
func (p *Port) handleUserDeleted(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var userDeletedEvent event.UserDeleted
	err := json.Unmarshal(message.RawValue, &userDeletedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.DeleteUserByID(ctx, tx, userDeletedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete user", nil)
	}
	return nil
}
