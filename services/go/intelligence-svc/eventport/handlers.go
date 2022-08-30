package eventport

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// Handler for received messages.
type Handler interface {
	// CreateUser creates the user with the given id.
	CreateUser(ctx context.Context, tx pgx.Tx, userID store.User) error
	// UpdateUser updates the given store.user, identified by its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// CreateOperation creates the given store.Operation.
	CreateOperation(ctx context.Context, tx pgx.Tx, create store.Operation) error
	// UpdateOperation updates the given store.Operation.
	UpdateOperation(ctx context.Context, tx pgx.Tx, update store.Operation) error
	// UpdateOperationMembersByOperation updates the operation members for the given
	// operation.
	UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error
	// CreateAddressBookEntry creates the given store.AddressBookEntry.
	CreateAddressBookEntry(ctx context.Context, tx pgx.Tx, create store.AddressBookEntry) error
	// UpdateAddressBookEntry updates the given store.AddressBookEntry.
	UpdateAddressBookEntry(ctx context.Context, tx pgx.Tx, update store.AddressBookEntry) error
	// DeleteAddressBookEntryByID deletes the address book entry with the given id.
	DeleteAddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error
}

// HandlerFn for handling messages.
func (p *Port) HandlerFn(handler Handler) kafkautil.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, message kafkautil.InboundMessage) error {
		switch message.Topic {
		case event.AddressBookTopic:
			return meh.NilOrWrap(p.handleAddressBookTopic(ctx, tx, handler, message), "handle address book topic", nil)
		case event.OperationsTopic:
			return meh.NilOrWrap(p.handleOperationsTopic(ctx, tx, handler, message), "handle operations topic", nil)
		case event.UsersTopic:
			return meh.NilOrWrap(p.handleUsersTopic(ctx, tx, handler, message), "handle users topic", nil)
		}
		return nil
	}
}

// handleUsersTopic handles the event.UsersTopic.
func (p *Port) handleUsersTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypeUserCreated:
		return meh.NilOrWrap(p.handleUserCreated(ctx, tx, handler, message), "handle user created", nil)
	case event.TypeUserUpdated:
		return meh.NilOrWrap(p.handleUserUpdated(ctx, tx, handler, message), "handle user updated", nil)
	}
	return nil
}

// handleUserCreated handles an event.TypeUserCreated event.
func (p *Port) handleUserCreated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
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
		IsActive:  userCreatedEvent.IsActive,
	}
	err = handler.CreateUser(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create user", meh.Details{"user": create})
	}
	return nil
}

// handleUserUpdated handles an event.TypeUserUpdated event.
func (p *Port) handleUserUpdated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
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
		IsActive:  userUpdatedEvent.IsActive,
	}
	err = handler.UpdateUser(ctx, tx, update)
	if err != nil {
		return meh.Wrap(err, "update user", meh.Details{"user": update})
	}
	return nil
}

// handleOperationsTopic handles the event.OperationsTopic.
func (p *Port) handleOperationsTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypeOperationCreated:
		return meh.NilOrWrap(p.handleOperationCreated(ctx, tx, handler, message), "handle operation created", nil)
	case event.TypeOperationUpdated:
		return meh.NilOrWrap(p.handleOperationUpdated(ctx, tx, handler, message), "handle operation updated", nil)
	case event.TypeOperationMembersUpdated:
		return meh.NilOrWrap(p.handleOperationMembersUpdated(ctx, tx, handler, message), "handle operation members updated", nil)
	}
	return nil
}

// handleOperationCreated handles an event.TypeOperationCreated event.
func (p *Port) handleOperationCreated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
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
	err = handler.CreateOperation(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create operation", meh.Details{"create": create})
	}
	return nil
}

// handleOperationUpdated handles an event.TypeOperationUpdated event.
func (p *Port) handleOperationUpdated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
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
	err = handler.UpdateOperation(ctx, tx, update)
	if err != nil {
		return meh.Wrap(err, "update operation", meh.Details{"update": update})
	}
	return nil
}

// handleOperationMembersUpdated handles an event.TypeOperationMembersUpdated
// event.
func (p *Port) handleOperationMembersUpdated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var operationMembersUpdatedEvent event.OperationMembersUpdated
	err := json.Unmarshal(message.RawValue, &operationMembersUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	err = handler.UpdateOperationMembersByOperation(ctx, tx, operationMembersUpdatedEvent.Operation, operationMembersUpdatedEvent.Members)
	if err != nil {
		return meh.Wrap(err, "update operation members", meh.Details{
			"operation":   operationMembersUpdatedEvent.Operation,
			"new_members": operationMembersUpdatedEvent.Members,
		})
	}
	return nil
}

// handleAddressBookTopic handles an kafkautil.InboundMessage for the
// event.AddressBookTopic.
func (p *Port) handleAddressBookTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypeAddressBookEntryCreated:
		return meh.NilOrWrap(p.handleAddressBookEntryCreated(ctx, tx, handler, message), "handle address book entry created", nil)
	case event.TypeAddressBookEntryUpdated:
		return meh.NilOrWrap(p.handleAddressBookEntryUpdated(ctx, tx, handler, message), "handle address book entry updated", nil)
	case event.TypeAddressBookEntryDeleted:
		return meh.NilOrWrap(p.handleAddressBookEntryDeleted(ctx, tx, handler, message), "handle address book entry deleted", nil)
	}
	return nil
}

// handleAddressBookEntryCreated handles an event.TypeAddressBookEntryCreated
// event.
func (p *Port) handleAddressBookEntryCreated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var entryCreatedEvent event.AddressBookEntryCreated
	err := json.Unmarshal(message.RawValue, &entryCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "parse event", meh.Details{"event_payload": message.RawValue})
	}
	create := store.AddressBookEntry{
		ID:          entryCreatedEvent.ID,
		Label:       entryCreatedEvent.Label,
		Description: entryCreatedEvent.Description,
		Operation:   entryCreatedEvent.Operation,
		User:        entryCreatedEvent.User,
	}
	err = handler.CreateAddressBookEntry(ctx, tx, create)
	if err != nil {
		return meh.Wrap(err, "create address book entry", meh.Details{"create": create})
	}
	return nil
}

// handleAddressBookEntryUpdated handles an event.TypeAddressBookEntryUpdated
// event.
func (p *Port) handleAddressBookEntryUpdated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var entryUpdatedevent event.AddressBookEntryUpdated
	err := json.Unmarshal(message.RawValue, &entryUpdatedevent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "parse event", meh.Details{"event_payload": message.RawValue})
	}
	update := store.AddressBookEntry{
		ID:          entryUpdatedevent.ID,
		Label:       entryUpdatedevent.Label,
		Description: entryUpdatedevent.Description,
		Operation:   entryUpdatedevent.Operation,
		User:        entryUpdatedevent.User,
	}
	err = handler.UpdateAddressBookEntry(ctx, tx, update)
	if err != nil {
		return meh.Wrap(err, "update address book entry", meh.Details{"update": update})
	}
	return nil
}

// handleAddressBookEntryDeleted handles an event.TypeAddressBookEntryDeleted
// event.
func (p *Port) handleAddressBookEntryDeleted(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var entryDeletedEvent event.AddressBookEntryDeleted
	err := json.Unmarshal(message.RawValue, &entryDeletedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "parse event", meh.Details{"event_payload": message.RawValue})
	}
	err = handler.DeleteAddressBookEntryByID(ctx, tx, entryDeletedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete address book entry by id", meh.Details{"id": entryDeletedEvent.ID})
	}
	return nil
}
