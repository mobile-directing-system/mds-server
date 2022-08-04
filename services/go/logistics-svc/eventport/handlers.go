package eventport

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
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
	// CreateGroup creates the given store.Group.
	CreateGroup(ctx context.Context, create store.Group) error
	// UpdateGroup updates the given store.Group, identified by its id.
	UpdateGroup(ctx context.Context, update store.Group) error
	// DeleteGroupByID deletes the group with the given id.
	DeleteGroupByID(ctx context.Context, groupID uuid.UUID) error
	// CreateOperation creates the given store.Operation.
	CreateOperation(ctx context.Context, create store.Operation) error
	// UpdateOperation updates the given store.Operation.
	UpdateOperation(ctx context.Context, update store.Operation) error
	// UpdateOperationMembers updates the operation members for the given operation.
	UpdateOperationMembersByOperation(ctx context.Context, operationID uuid.UUID, newMembers []uuid.UUID) error
}

// HandlerFn for handling messages.
func (p *Port) HandlerFn(handler Handler) kafkautil.HandlerFunc {
	return func(ctx context.Context, message kafkautil.Message) error {
		switch message.Topic {
		case event.GroupsTopic:
			return meh.NilOrWrap(p.handleGroupsTopic(ctx, handler, message), "handle groups topic", nil)
		case event.OperationsTopic:
			return meh.NilOrWrap(p.handleOperationsTopic(ctx, handler, message), "handle operations topic", nil)
		case event.UsersTopic:
			return meh.NilOrWrap(p.handleUsersTopic(ctx, handler, message), "handle users topic", nil)
		}
		return nil
	}
}

// handleGroupsTopic handles the event.GroupsTopic.
func (p *Port) handleGroupsTopic(ctx context.Context, handler Handler, message kafkautil.Message) error {
	switch message.EventType {
	case event.TypeGroupCreated:
		return meh.NilOrWrap(p.handleGroupCreated(ctx, handler, message), "handle group created", nil)
	case event.TypeGroupDeleted:
		return meh.NilOrWrap(p.handleGroupDeleted(ctx, handler, message), "handle group deleted", nil)
	case event.TypeGroupUpdated:
		return meh.NilOrWrap(p.handleGroupUpdated(ctx, handler, message), "handle group updated", nil)
	}
	return nil
}

// handleGroupCreated handles an event.TypeGroupCreated event.
func (p *Port) handleGroupCreated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var groupCreatedEvent event.GroupCreated
	err := json.Unmarshal(message.RawValue, &groupCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	create := store.Group{
		ID:          groupCreatedEvent.ID,
		Title:       groupCreatedEvent.Title,
		Description: groupCreatedEvent.Description,
		Operation:   groupCreatedEvent.Operation,
		Members:     groupCreatedEvent.Members,
	}
	err = handler.CreateGroup(ctx, create)
	if err != nil {
		return meh.Wrap(err, "create group", meh.Details{"group": create})
	}
	return nil
}

// handleGroupUpdated handles an event.TypeGroupUpdated event.
func (p *Port) handleGroupUpdated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var groupUpdatedEvent event.GroupUpdated
	err := json.Unmarshal(message.RawValue, &groupUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	update := store.Group{
		ID:          groupUpdatedEvent.ID,
		Title:       groupUpdatedEvent.Title,
		Description: groupUpdatedEvent.Description,
		Operation:   groupUpdatedEvent.Operation,
		Members:     groupUpdatedEvent.Members,
	}
	err = handler.UpdateGroup(ctx, update)
	if err != nil {
		return meh.Wrap(err, "update group", meh.Details{"group": update})
	}
	return nil
}

// handleGroupDeleted handles an event.TypeGroupDeleted event.
func (p *Port) handleGroupDeleted(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var groupDeletedEvent event.GroupDeleted
	err := json.Unmarshal(message.RawValue, &groupDeletedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	err = handler.DeleteGroupByID(ctx, groupDeletedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete group", meh.Details{"group_id": groupDeletedEvent.ID})
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

// handleOperationsTopic handles the event.OperationsTopic.
func (p *Port) handleOperationsTopic(ctx context.Context, handler Handler, message kafkautil.Message) error {
	switch message.EventType {
	case event.TypeOperationCreated:
		return meh.NilOrWrap(p.handleOperationCreated(ctx, handler, message), "handle operation created", nil)
	case event.TypeOperationUpdated:
		return meh.NilOrWrap(p.handleOperationUpdated(ctx, handler, message), "handle operation updated", nil)
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
	create := store.Operation{
		ID:          operationCreatedEvent.ID,
		Title:       operationCreatedEvent.Title,
		Description: operationCreatedEvent.Description,
		Start:       operationCreatedEvent.Start,
		End:         operationCreatedEvent.End,
		IsArchived:  operationCreatedEvent.IsArchived,
	}
	err = handler.CreateOperation(ctx, create)
	if err != nil {
		return meh.Wrap(err, "create operation", meh.Details{"create": create})
	}
	return nil
}

// handleOperationUpdated handles an event.TypeOperationUpdated event.
func (p *Port) handleOperationUpdated(ctx context.Context, handler Handler, message kafkautil.Message) error {
	var operationUpdatedEvent event.OperationUpdated
	err := json.Unmarshal(message.RawValue, &operationUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	update := store.Operation{
		ID:          operationUpdatedEvent.ID,
		Title:       operationUpdatedEvent.Title,
		Description: operationUpdatedEvent.Description,
		Start:       operationUpdatedEvent.Start,
		End:         operationUpdatedEvent.End,
		IsArchived:  operationUpdatedEvent.IsArchived,
	}
	err = handler.UpdateOperation(ctx, update)
	if err != nil {
		return meh.Wrap(err, "update operation", meh.Details{"update": update})
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
			"operation":   operationMembersUpdatedEvent.Operation,
			"new_members": operationMembersUpdatedEvent.Members,
		})
	}
	return nil
}
