package eventport

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
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
	// CreateIntel creates the given store.Intel.
	CreateIntel(ctx context.Context, create store.Intel) error
	// InvalidateIntelByID invalidates the intel with the given id.
	InvalidateIntelByID(ctx context.Context, intelID uuid.UUID) error
	// CreateActiveIntelDelivery creates the given store.ActiveIntelDelivery in the
	// store.
	CreateActiveIntelDelivery(ctx context.Context, create store.ActiveIntelDelivery) error
	// DeleteActiveIntelDeliveryByID deletes the intel delivery with the given id
	// from the store.
	DeleteActiveIntelDeliveryByID(ctx context.Context, deliveryID uuid.UUID) error
	// CreateActiveIntelDeliveryAttempt creates the given
	// store.ActiveIntelDeliveryAttempt.
	CreateActiveIntelDeliveryAttempt(ctx context.Context, create store.ActiveIntelDeliveryAttempt) error
	// DeleteActiveIntelDeliveryAttemptByID deletes the
	// store.ActiveIntelDeliveryAttempt with the given id.
	DeleteActiveIntelDeliveryAttemptByID(ctx context.Context, attemptID uuid.UUID) error
	// SetAutoIntelDeliveryEnabledForAddressBookEntry sets the auto-intel-delivery
	// flag for the address book entry with the given id.
	SetAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID, enabled bool) error
}

// HandlerFn for handling messages.
func (p *Port) HandlerFn(handler Handler) kafkautil.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, message kafkautil.InboundMessage) error {
		switch message.Topic {
		case event.OperationsTopic:
			return meh.NilOrWrap(p.handleOperationsTopic(ctx, tx, handler, message), "handle operations topic", nil)
		case event.UsersTopic:
			return meh.NilOrWrap(p.handleUsersTopic(ctx, tx, handler, message), "handle users topic", nil)
		case event.IntelTopic:
			return meh.NilOrWrap(p.handleIntelTopic(ctx, tx, handler, message), "handle intel topic", nil)
		case event.IntelDeliveriesTopic:
			return meh.NilOrWrap(p.handleIntelDeliveriesTopic(ctx, tx, handler, message), "handle intel-deliveries topic", nil)
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

func (p *Port) handleIntelTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypeIntelCreated:
		return meh.NilOrWrap(p.handleIntelCreated(ctx, tx, handler, message), "handle intel created", nil)
	case event.TypeIntelInvalidated:
		return meh.NilOrWrap(p.handleIntelInvalidated(ctx, tx, handler, message), "handle intel invalidated", nil)
	}
	return nil
}

// handleIntelCreated handles an event.TypeIntelCreated event.
func (p *Port) handleIntelCreated(ctx context.Context, _ pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var intelCreatedEvent event.IntelCreated
	err := json.Unmarshal(message.RawValue, &intelCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	create := store.Intel{
		ID:         intelCreatedEvent.ID,
		CreatedAt:  intelCreatedEvent.CreatedAt,
		CreatedBy:  intelCreatedEvent.CreatedBy,
		Operation:  intelCreatedEvent.Operation,
		Importance: intelCreatedEvent.Importance,
		IsValid:    intelCreatedEvent.IsValid,
	}
	err = handler.CreateIntel(ctx, create)
	if err != nil {
		return meh.Wrap(err, "create intel", meh.Details{"create": create})
	}
	return nil
}

// handleIntelInvalidated handles an event.TypeIntelInvalidated event.
func (p *Port) handleIntelInvalidated(ctx context.Context, _ pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var intelInvalidatedEvent event.IntelInvalidated
	err := json.Unmarshal(message.RawValue, &intelInvalidatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"was": string(message.RawValue)})
	}
	err = handler.InvalidateIntelByID(ctx, intelInvalidatedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "invalidate intel by id", meh.Details{"intel_id": intelInvalidatedEvent.ID})
	}
	return nil
}

// handleIntelDeliveriesTopic handles events for event.IntelDeliveriesTopic.
func (p *Port) handleIntelDeliveriesTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypeIntelDeliveryCreated:
		return meh.NilOrWrap(p.handleIntelDeliveryCreated(ctx, tx, handler, message), "handle intel delivery created", nil)
	case event.TypeIntelDeliveryAttemptCreated:
		return meh.NilOrWrap(p.handleIntelDeliveryAttemptCreated(ctx, tx, handler, message), "handle intel delivery attempt created", nil)
	case event.TypeIntelDeliveryAttemptStatusUpdated:
		return meh.NilOrWrap(p.handleIntelDeliveryAttemptStatusUpdated(ctx, tx, handler, message), "handle intel delivery attempt status updated", nil)
	case event.TypeIntelDeliveryStatusUpdated:
		return meh.NilOrWrap(p.handleIntelDeliveryStatusUpdated(ctx, tx, handler, message), "handle intel delivery status updated", nil)
	case event.TypeAddressBookEntryAutoDeliveryUpdated:
		return meh.NilOrWrap(p.handleAddressBookEntryAutoDeliveryUpdated(ctx, tx, handler, message), "handle address book entry auto delivery updated", nil)
	}
	return nil
}

// handleIntelDeliveryCreated handles an event.TypeIntelDeliveryCreated event.
func (p *Port) handleIntelDeliveryCreated(ctx context.Context, _ pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var createdEvent event.IntelDeliveryCreated
	err := json.Unmarshal(message.RawValue, &createdEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"was": string(message.RawValue)})
	}
	if !createdEvent.IsActive {
		return nil
	}
	create := store.ActiveIntelDelivery{
		ID:    createdEvent.ID,
		Intel: createdEvent.Intel,
		To:    createdEvent.To,
		Note:  createdEvent.Note,
	}
	err = handler.CreateActiveIntelDelivery(ctx, create)
	if err != nil {
		return meh.Wrap(err, "create active intel", meh.Details{"create": create})
	}
	return nil
}

// handleIntelDeliveryAttemptCreated handles an
// event.TypeIntelDeliveryAttemptCreated event.
func (p *Port) handleIntelDeliveryAttemptCreated(ctx context.Context, _ pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var createdEvent event.IntelDeliveryAttemptCreated
	err := json.Unmarshal(message.RawValue, &createdEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"was": string(message.RawValue)})
	}
	if !createdEvent.IsActive {
		return nil
	}
	create := store.ActiveIntelDeliveryAttempt{
		ID:       createdEvent.ID,
		Delivery: createdEvent.Delivery.ID,
	}
	err = handler.CreateActiveIntelDeliveryAttempt(ctx, create)
	if err != nil {
		return meh.Wrap(err, "create active intel delivery attempt", meh.Details{"create": create})
	}
	return nil
}

// TODO: Small chance for a race condition where the delivery attempt is updated
//  to delivered but processing or receiving the event for the delivery being
//  inactive might take longer. This might lead to listeners being notified about
//  the delivery being open for delivery and a possible redelivery being
//  scheduled.

// handleIntelDeliveryAttemptStatusUpdated handles an
// event.TypeIntelDeliveryAttemptStatusUpdated event.
func (p *Port) handleIntelDeliveryAttemptStatusUpdated(ctx context.Context, _ pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var updatedEvent event.IntelDeliveryAttemptStatusUpdated
	err := json.Unmarshal(message.RawValue, &updatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"was": string(message.RawValue)})
	}
	if updatedEvent.IsActive {
		return nil
	}
	err = handler.DeleteActiveIntelDeliveryAttemptByID(ctx, updatedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete active intel delivery attempt", meh.Details{"attempt_id": updatedEvent.ID})
	}
	return nil
}

// handleIntelDeliveryStatusUpdated handles an
// event.TypeIntelDeliveryStatusUpdated event.
func (p *Port) handleIntelDeliveryStatusUpdated(ctx context.Context, _ pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var updatedEvent event.IntelDeliveryStatusUpdated
	err := json.Unmarshal(message.RawValue, &updatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"was": string(message.RawValue)})
	}
	if updatedEvent.IsActive {
		return nil
	}
	err = handler.DeleteActiveIntelDeliveryByID(ctx, updatedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete active intel delivery", meh.Details{"delivery_id": updatedEvent.ID})
	}
	return nil
}

// handleAddressBookEntryAutoDeliveryUpdated handles an
// event.TypeAddressBookEntryAutoDeliveryUpdated event.
func (p *Port) handleAddressBookEntryAutoDeliveryUpdated(ctx context.Context, _ pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var updatedEvent event.AddressBookEntryAutoDeliveryUpdated
	err := json.Unmarshal(message.RawValue, &updatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"was": string(message.RawValue)})
	}
	err = handler.SetAutoIntelDeliveryEnabledForAddressBookEntry(ctx, updatedEvent.ID, updatedEvent.IsAutoDeliveryEnabled)
	if err != nil {
		return meh.Wrap(err, "set auto intel delivery enabled for address book entry", meh.Details{
			"entry_id":   updatedEvent.ID,
			"is_enabled": updatedEvent.IsAutoDeliveryEnabled,
		})
	}
	return nil
}
