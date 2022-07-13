package eventport

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// Handler for received messages.
type Handler interface {
	// CreateOperation creates the operation with the given id.
	CreateOperation(ctx context.Context, operationID uuid.UUID) error
	// CreateUser creates the user with the given id.
	CreateUser(ctx context.Context, userID uuid.UUID) error
	// DeleteUserByID deletes the user with the given id and notfies of updated
	// groups.
	DeleteUserByID(ctx context.Context, userID uuid.UUID) error
	// UpdateOperationMembersByOperation updates the members for the operation with
	// the given id.
	UpdateOperationMembersByOperation(ctx context.Context, operationID uuid.UUID, newMembers []uuid.UUID) error
}

// HandlerFn for handling messages.
func (p *Port) HandlerFn(handler Handler) kafkautil.HandlerFunc {
	return func(ctx context.Context, message kafkautil.Message) error {
		switch message.Topic {
		case event.OperationsTopic:
			return meh.NilOrWrap(p.handleOperationsTopic(ctx, handler, message), "handle operations topic", nil)
		case event.UsersTopic:
			return meh.NilOrWrap(p.handleUsersTopic(ctx, handler, message), "handle users topic", nil)
		}
		return nil
	}
}

// handleOperationsTopic handles the event.OperationsTopic.
func (p *Port) handleOperationsTopic(ctx context.Context, handler Handler, message kafkautil.Message) error {
	switch message.EventType {
	case event.TypeOperationCreated:
		return meh.NilOrWrap(p.handleOperationCreated(ctx, handler, message), "handle operation created", nil)
	case event.TypeOperationMembersUpdated:
		return meh.NilOrWrap(p.handleOperationMembersUpdated(ctx, handler, message), "handle operation members updated", nil)
	}
	return nil
}

// handleOperationCreated handles an event.TypeOperationCreated event.
func (p *Port) handleOperationCreated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var operationCreatedEvent event.OperationCreated
	err := json.Unmarshal(message.RawValue, &operationCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	err = handler.CreateOperation(ctx, operationCreatedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "create operation", meh.Details{"operation_id": operationCreatedEvent.ID})
	}
	return nil
}

// handleOperationMembersUpdated handles an event.TypeOperationMembersUpdated
// event.
func (p *Port) handleOperationMembersUpdated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var operationMembersUpdatedEvent event.OperationMembersUpdated
	err := json.Unmarshal(message.RawValue, &operationMembersUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	err = handler.UpdateOperationMembersByOperation(ctx, operationMembersUpdatedEvent.Operation, operationMembersUpdatedEvent.Members)
	if err != nil {
		return meh.Wrap(err, "update operation members", meh.Details{
			"operation_id": operationMembersUpdatedEvent.Operation,
			"new_members":  operationMembersUpdatedEvent.Members,
		})
	}
	return nil
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

// handleUserCreated handles an event.TypeUserCreated event.
func (p *Port) handleUserCreated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var userCreatedEvent event.UserCreated
	err := json.Unmarshal(message.RawValue, &userCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
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
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	err = handler.DeleteUserByID(ctx, userDeletedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete user", meh.Details{"user_id": userDeletedEvent.ID})
	}
	return nil
}
