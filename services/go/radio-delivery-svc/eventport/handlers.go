package eventport

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"time"
)

// Handler for received messages.
type Handler interface {
	// CreateUser creates the user with the given id.
	CreateUser(ctx context.Context, tx pgx.Tx, userID store.User) error
	// UpdateUser updates the given store.user, identified by its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// DeleteRadioChannelsByEntry deletes all radio-channels for the entry with the
	// given id.
	DeleteRadioChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error
	// UpdateRadioChannelsByEntry updates the radio-channels for the address book
	// entry with the given id.
	UpdateRadioChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID, newChannels []store.RadioChannel) error
	// CreateIntelDeliveryAttempt handles a new intel-delivery-attempt. It checks,
	// whether the channel is supported by us and accepts it then. Otherwise, it is
	// ignored.
	CreateIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, attempt store.AcceptedIntelDeliveryAttempt) error
	// UpdateIntelDeliveryAttemptStatus updates the intel-delivery-attempt-status
	// for the associated intel-delivery-attempt. The attempt does not need to be
	// accepted.
	UpdateIntelDeliveryAttemptStatus(ctx context.Context, tx pgx.Tx, newStatus store.AcceptedIntelDeliveryAttemptStatus) error
	// UpdateOperationMembersByOperation replaces the associated operation members
	// for the operation with the given id with the new given ones.
	UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error
}

// HandlerFn for handling messages.
func (p *Port) HandlerFn(handler Handler) kafkautil.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, message kafkautil.InboundMessage) error {
		switch message.Topic {
		case event.AddressBookTopic:
			return meh.NilOrWrap(p.handleAddressBookTopic(ctx, tx, handler, message), "handle address book topic", nil)
		case event.IntelDeliveriesTopic:
			return meh.NilOrWrap(p.handleIntelDeliveriesTopic(ctx, tx, handler, message), "handle intel-deliveries topic", nil)
		case event.UsersTopic:
			return meh.NilOrWrap(p.handleUsersTopic(ctx, tx, handler, message), "handle users topic", nil)
		case event.OperationsTopic:
			return meh.NilOrWrap(p.handleOperationsTopic(ctx, tx, handler, message), "handle operations topic", nil)
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

// handleAddressBookTopic handles the event.AddressBookTopic.
func (p *Port) handleAddressBookTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypeAddressBookEntryDeleted:
		return meh.NilOrWrap(p.handleAddressBookEntryDeleted(ctx, tx, handler, message), "handle address book entry deleted", nil)
	case event.TypeAddressBookEntryChannelsUpdated:
		return meh.NilOrWrap(p.handleAddressBookEntryChannelsUpdated(ctx, tx, handler, message), "handle address book entry channels updated", nil)
	}
	return nil
}

// handleAddressBookEntryDeleted handles an event.TypeAddressBookEntryDeleted
// event.
func (p *Port) handleAddressBookEntryDeleted(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var entryDeletedEvent event.AddressBookEntryDeleted
	err := json.Unmarshal(message.RawValue, &entryDeletedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	err = handler.DeleteRadioChannelsByEntry(ctx, tx, entryDeletedEvent.ID)
	if err != nil {
		return meh.Wrap(err, "delete radio-channels by entry", meh.Details{"entry_id": entryDeletedEvent.ID})
	}
	return nil
}

// handleAddressBookEntryChannelsUpdated handles an
// event.TypeAddressBookEntryChannelsUpdated event.
func (p *Port) handleAddressBookEntryChannelsUpdated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var channelsUpdatedEvent event.AddressBookEntryChannelsUpdated
	err := json.Unmarshal(message.RawValue, &channelsUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	newRadioChannels := make([]store.RadioChannel, 0)
	for _, newChannel := range channelsUpdatedEvent.Channels {
		if newChannel.Type != event.AddressBookEntryChannelTypeRadio {
			continue
		}
		// Check the entry id as we rely on them to be correct.
		if newChannel.Entry != channelsUpdatedEvent.Entry {
			return meh.NewInternalErr("radio-channel entry id differs from event entry id", meh.Details{
				"event_entry_id":   channelsUpdatedEvent.Entry,
				"channel_entry_id": newChannel.Entry,
			})
		}
		newRadioChannels = append(newRadioChannels, store.RadioChannel{
			ID:      newChannel.ID,
			Entry:   newChannel.Entry,
			Label:   newChannel.Label,
			Timeout: newChannel.Timeout,
		})
	}
	err = handler.UpdateRadioChannelsByEntry(ctx, tx, channelsUpdatedEvent.Entry, newRadioChannels)
	if err != nil {
		return meh.Wrap(err, "update radio-channels by entry", meh.Details{
			"entry_id":     channelsUpdatedEvent.Entry,
			"new_channels": newRadioChannels,
		})
	}
	return nil
}

// handleIntelDeliveriesTopic handles the event.IntelDeliveriesTopic.
func (p *Port) handleIntelDeliveriesTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypeIntelDeliveryAttemptCreated:
		return meh.NilOrWrap(p.handleIntelDeliveryAttemptCreated(ctx, tx, handler, message), "handle intel-delivery-attempt created", nil)
	case event.TypeIntelDeliveryAttemptStatusUpdated:
		return meh.NilOrWrap(p.handleIntelDeliveryAttemptStatusUpdated(ctx, tx, handler, message), "handle intel-delivery-attempt-status updated", nil)
	}
	return nil
}

// handleIntelDeliveryAttemptCreated handles an
// event.TypeIntelDeliveryAttemptCreated.
func (p *Port) handleIntelDeliveryAttemptCreated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var attemptCreatedEvent event.IntelDeliveryAttemptCreated
	err := json.Unmarshal(message.RawValue, &attemptCreatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	err = handler.CreateIntelDeliveryAttempt(ctx, tx, store.AcceptedIntelDeliveryAttempt{
		ID:              attemptCreatedEvent.ID,
		Intel:           attemptCreatedEvent.Intel.ID,
		IntelOperation:  attemptCreatedEvent.Intel.Operation,
		IntelImportance: attemptCreatedEvent.Intel.Importance,
		AssignedTo:      attemptCreatedEvent.Delivery.To,
		AssignedToLabel: attemptCreatedEvent.AssignedEntry.Label,
		AssignedToUser:  attemptCreatedEvent.AssignedEntry.User,
		Delivery:        attemptCreatedEvent.Delivery.ID,
		Channel:         attemptCreatedEvent.Channel,
		CreatedAt:       attemptCreatedEvent.CreatedAt,
		IsActive:        attemptCreatedEvent.IsActive,
		StatusTS:        attemptCreatedEvent.StatusTS,
		Note:            attemptCreatedEvent.Note,
		AcceptedAt:      time.Time{}, // Set by the controller.
	})
	if err != nil {
		return meh.Wrap(err, "create intel-delivery-attempt", nil)
	}
	return nil
}

// handleIntelDeliveryAttemptCreated handles an
// event.TypeIntelDeliveryAttemptStatusUpdated.
func (p *Port) handleIntelDeliveryAttemptStatusUpdated(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	var statusUpdatedEvent event.IntelDeliveryAttemptStatusUpdated
	err := json.Unmarshal(message.RawValue, &statusUpdatedEvent)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal event", meh.Details{"raw": string(message.RawValue)})
	}
	err = handler.UpdateIntelDeliveryAttemptStatus(ctx, tx, store.AcceptedIntelDeliveryAttemptStatus{
		ID:       statusUpdatedEvent.ID,
		IsActive: statusUpdatedEvent.IsActive,
		StatusTS: statusUpdatedEvent.StatusTS,
		Note:     statusUpdatedEvent.Note,
	})
	if err != nil {
		return meh.Wrap(err, "update intel-delivery-attempt status", nil)
	}
	return nil
}

// handleOperationsTopic handles the event.OperationsTopic.
func (p *Port) handleOperationsTopic(ctx context.Context, tx pgx.Tx, handler Handler, message kafkautil.InboundMessage) error {
	switch message.EventType {
	case event.TypeOperationMembersUpdated:
		return meh.NilOrWrap(p.handleOperationMembersUpdated(ctx, tx, handler, message), "handle operation members updated", nil)
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
