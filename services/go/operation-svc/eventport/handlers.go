package eventport

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// Handler for received messages.
type Handler interface {
	// CreateUser creates the user with the given id.
	CreateUser(ctx context.Context, userID store.User) error
	// UpdateUser updates the given store.user, identified by its id.
	UpdateUser(ctx context.Context, user store.User) error
	// DeleteUserByID deletes the user with the given id and notfies of updated
	// groups.
	DeleteUserByID(ctx context.Context, userID uuid.UUID) error
}

// HandlerFn for handling messages.
func (p *Port) HandlerFn(handler Handler) kafkautil.HandlerFunc {
	return func(ctx context.Context, message kafkautil.Message) error {
		switch message.Topic {
		case event.UsersTopic:
			return meh.NilOrWrap(p.handleUsersTopic(ctx, handler, message), "handle users topic", nil)
		}
		return nil
	}
}

// handleUsersTopic handles the event.UsersTopic.
func (p *Port) handleUsersTopic(ctx context.Context, handler Handler, message kafkautil.Message) error {
	switch message.EventType {
	case event.TypeUserCreated:
		return meh.NilOrWrap(p.handleUserCreated(ctx, handler, message), "handle user created", nil)
	case event.TypeUserDeleted:
		return meh.NilOrWrap(p.handleUserDeleted(ctx, handler, message), "handle user deleted", nil)
	case event.TypeUserUpdated:
		return meh.NilOrWrap(p.handleUserUpdated(ctx, handler, message), "handle user updated", nil)
	}
	return nil
}

// handleUserCreated handles an event.TypeUserCreated event.
func (p *Port) handleUserCreated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userCreatedEvent event.UserCreated
	err := json.Unmarshal(message.RawValue, &userCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	create := store.User{
		ID:        userCreatedEvent.ID,
		Username:  userCreatedEvent.Username,
		FirstName: userCreatedEvent.FirstName,
		LastName:  userCreatedEvent.LastName,
	}
	err = handler.CreateUser(ctx, create)
	if err != nil {
		return meh.Wrap(err, "create user", meh.Details{"user": create})
	}
	return nil
}

// handleUserUpdated handles an event.TypeUserUpdated event.
func (p *Port) handleUserUpdated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userUpdatedEvent event.UserUpdated
	err := json.Unmarshal(message.RawValue, &userUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	update := store.User{
		ID:        userUpdatedEvent.ID,
		Username:  userUpdatedEvent.Username,
		FirstName: userUpdatedEvent.FirstName,
		LastName:  userUpdatedEvent.LastName,
	}
	err = handler.UpdateUser(ctx, update)
	if err != nil {
		return meh.Wrap(err, "update user", meh.Details{"user": update})
	}
	return nil
}

// handleUserDeleted handles an event.TypeUserDeleted event.
func (p *Port) handleUserDeleted(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userDeletedEvent event.UserDeleted
	err := json.Unmarshal(message.RawValue, &userDeletedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	err = handler.DeleteUserByID(ctx, userDeletedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete user", meh.Details{"user_id": userDeletedEvent.ID})
	}
	return nil
}
