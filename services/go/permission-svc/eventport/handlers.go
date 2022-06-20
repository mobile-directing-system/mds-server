package eventport

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// Handler for received messages.
type Handler interface {
	// CreateUser creates the user with the given id.
	CreateUser(ctx context.Context, userID uuid.UUID) error
	// DeleteUserByID deletes the user with the given id and notifies of unassigned
	// permissions.
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
	}
	return nil
}

// handleUserDeleted handles an event.TypeUserCreated event.
func (p *Port) handleUserCreated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userCreatedEvent event.UserCreated
	err := json.Unmarshal(message.RawValue, &userCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	// Clear permissions.
	err = handler.CreateUser(ctx, userCreatedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "create user", meh.Details{"user_id": userCreatedEvent.ID})
	}
	return nil
}

// handleUserDeleted handles an event.TypeUserDeleted event.
func (p *Port) handleUserDeleted(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userDeletedEvent event.UserDeleted
	err := json.Unmarshal(message.RawValue, &userDeletedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", nil)
	}
	// Clear permissions.
	err = handler.DeleteUserByID(ctx, userDeletedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete user", meh.Details{"user_id": userDeletedEvent.ID})
	}
	return nil
}
