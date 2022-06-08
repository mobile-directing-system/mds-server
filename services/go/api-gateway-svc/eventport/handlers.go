package eventport

import (
	"context"
	"encoding/json"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// Handler for received Kafka messages.
type Handler interface {
	// CreateUser creates the given store.User.
	CreateUser(ctx context.Context, user store.User) error
}

// HandlerFn is the handler for Kafka messages.
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
	err = handler.CreateUser(ctx, store.User{
		ID:       userCreatedEvent.ID,
		Username: userCreatedEvent.Username,
		IsAdmin:  userCreatedEvent.IsAdmin,
		Pass:     userCreatedEvent.Pass,
	})
	if err != nil {
		return meh.Wrap(err, "create user", nil)
	}
	return nil
}
