package eventport

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
)

// Handler for received Kafka messages.
type Handler interface {
	// CreateUser creates the given store.User.
	CreateUser(ctx context.Context, user store.UserWithPass) error
	// UpdateUser updates the given store.User in the Store.
	UpdateUser(ctx context.Context, user store.User) error
	// UpdateUserPassByUserID updates the password for the user with the given id in
	// the store and notifies via Notifier.NotifyUserPassUpdated.
	UpdateUserPassByUserID(ctx context.Context, userID uuid.UUID, newPass []byte) error
	// DeleteUserByID deletes the user with the given id in the store and notifies
	// via Notifier.NotifyUserDeleted.
	DeleteUserByID(ctx context.Context, userID uuid.UUID) error
	// UpdatePermissionsByUser updates the permissions for the given user.
	UpdatePermissionsByUser(ctx context.Context, userID uuid.UUID, updatedPermissions []permission.Permission) error
}

// HandlerFn is the handler for Kafka messages.
func (p *Port) HandlerFn(handler Handler) kafkautil.HandlerFunc {
	return func(ctx context.Context, message kafkautil.Message) error {
		switch message.Topic {
		case event.PermissionsTopic:
			return meh.NilOrWrap(p.handlePermissionsTopic(ctx, handler, message), "handle permissions topic", nil)
		case event.UsersTopic:
			return meh.NilOrWrap(p.handleUsersTopic(ctx, handler, message), "handle users topic", nil)
		}
		return nil
	}
}

// handlePermissionsTopic handles the event.PermissionsTopic.
func (p *Port) handlePermissionsTopic(ctx context.Context, handler Handler, message kafkautil.Message) error {
	switch message.EventType {
	case event.TypePermissionsUpdated:
		return meh.NilOrWrap(p.handlePermissionsUpdated(ctx, handler, message), "handler permissions updated", nil)
	}
	return nil
}

// handlePermissionsUpdated handles an event.PermissionsUpdated.
func (p *Port) handlePermissionsUpdated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var permissionsUpdatedEvent event.PermissionsUpdated
	err := json.Unmarshal(message.RawValue, &permissionsUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.UpdatePermissionsByUser(ctx, permissionsUpdatedEvent.User, permissionsUpdatedEvent.Permissions)
	if err != nil {
		return meh.Wrap(err, "update permissions", meh.Details{
			"user_id":             permissionsUpdatedEvent.User,
			"updated_permissions": permissionsUpdatedEvent.Permissions,
		})
	}
	return nil
}

// handleUsersTopic handles the event.UsersTopic.
func (p *Port) handleUsersTopic(ctx context.Context, handler Handler, message kafkautil.Message) error {
	switch message.EventType {
	case event.TypeUserCreated:
		return meh.NilOrWrap(p.handleUserCreated(ctx, handler, message), "handle user created", nil)
	case event.TypeUserUpdated:
		return meh.NilOrWrap(p.handleUserUpdated(ctx, handler, message), "handle user updated", nil)
	case event.TypeUserPassUpdated:
		return meh.NilOrWrap(p.handleUserPassUpdated(ctx, handler, message), "handle user pass updated", nil)
	case event.TypeUserDeleted:
		return meh.NilOrWrap(p.handleUserDeleted(ctx, handler, message), "handle user deleted", nil)
	}
	return nil
}

// handleUserCreated handles an event.UserCreated.
func (p *Port) handleUserCreated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userCreatedEvent event.UserCreated
	err := json.Unmarshal(message.RawValue, &userCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.CreateUser(ctx, store.UserWithPass{
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
func (p *Port) handleUserUpdated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userUpdatedEvent event.UserUpdated
	err := json.Unmarshal(message.RawValue, &userUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.UpdateUser(ctx, store.User{
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
func (p *Port) handleUserPassUpdated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userPassUpdatedEvent event.UserPassUpdated
	err := json.Unmarshal(message.RawValue, &userPassUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.UpdateUserPassByUserID(ctx, userPassUpdatedEvent.User, userPassUpdatedEvent.NewPass)
	if err != nil {
		return meh.Wrap(err, "update user pass", nil)
	}
	return nil
}

// handleUserDeleted handles an event.UserDeleted.
func (p *Port) handleUserDeleted(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userDeletedEvent event.UserDeleted
	err := json.Unmarshal(message.RawValue, &userDeletedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	err = handler.DeleteUserByID(ctx, userDeletedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete user", nil)
	}
	return nil
}

// TODO: UPDATES, DELETE, ETC!!! for users
