package eventport

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"reflect"
)

func init() {
	assureChannelTypesSupported()
}

// assureChannelTypesSupported assures that all channel types are supported.
func assureChannelTypesSupported() {
	for channelType := range store.ChannelTypeSupplier.ChannelTypes {
		// Check type.
		_, err := mapChannelType(channelType)
		if err != nil {
			panic(fmt.Sprintf("unsupported channel type in event type mapper: %v", channelType))
		}
		// Check details.
		_, err = mapChannelDetails(channelType)
		if err != nil {
			panic(fmt.Sprintf("unsupported channel type in event details mapper: %v", channelType))
		}
	}
}

// NotifyAddressBookEntryCreated emits an event.TypeAddressBookEntryCreated
// event.
func (p *Port) NotifyAddressBookEntryCreated(ctx context.Context, tx pgx.Tx, entry store.AddressBookEntry) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.AddressBookTopic,
		Key:       entry.ID.String(),
		EventType: event.TypeAddressBookEntryCreated,
		Value: event.AddressBookEntryCreated{
			ID:          entry.ID,
			Label:       entry.Label,
			Description: entry.Description,
			Operation:   entry.Operation,
			User:        entry.User,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyAddressBookEntryUpdated emits an event.TypeAddressBookEntryUpdated
// event.
func (p *Port) NotifyAddressBookEntryUpdated(ctx context.Context, tx pgx.Tx, entry store.AddressBookEntry) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.AddressBookTopic,
		Key:       entry.ID.String(),
		EventType: event.TypeAddressBookEntryUpdated,
		Value: event.AddressBookEntryUpdated{
			ID:          entry.ID,
			Label:       entry.Label,
			Description: entry.Description,
			Operation:   entry.Operation,
			User:        entry.User,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyAddressBookEntryDeleted emits an event.TypeAddressBookEntryDeleted
// event.
func (p *Port) NotifyAddressBookEntryDeleted(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.AddressBookTopic,
		Key:       entryID.String(),
		EventType: event.TypeAddressBookEntryDeleted,
		Value: event.AddressBookEntryDeleted{
			ID: entryID,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// mapChannelType maps store.ChannelType to event.AddressBookEntryChannelType.
func mapChannelType(channelType store.ChannelType) (event.AddressBookEntryChannelType, error) {
	var mappedType event.AddressBookEntryChannelType
	switch channelType {
	case store.ChannelTypeDirect:
		mappedType = event.AddressBookEntryChannelTypeDirect
	case store.ChannelTypeEmail:
		mappedType = event.AddressBookEntryChannelTypeEmail
	case store.ChannelTypeForwardToGroup:
		mappedType = event.AddressBookEntryChannelTypeForwardToGroup
	case store.ChannelTypeForwardToUser:
		mappedType = event.AddressBookEntryChannelTypeForwardToUser
	case store.ChannelTypeInAppNotification:
		mappedType = event.AddressBookEntryChannelTypeInAppNotification
	case store.ChannelTypePhoneCall:
		mappedType = event.AddressBookEntryChannelTypePhoneCall
	case store.ChannelTypeRadio:
		mappedType = event.AddressBookEntryChannelTypeRadio
	default:
		return "", meh.NewInternalErr("unsupported channel type", nil)
	}
	return mappedType, nil
}

type channelDetailsMapper func(detailsRaw store.ChannelDetails) (json.RawMessage, error)

func mapChannelDetails(channelType store.ChannelType) (channelDetailsMapper, error) {
	var mapper channelDetailsMapper
	switch channelType {
	case store.ChannelTypeDirect:
		mapper = mapDirectChannelDetails
	case store.ChannelTypeEmail:
		mapper = mapEmailChannelDetails
	case store.ChannelTypeForwardToGroup:
		mapper = mapForwardToGroupChannelDetails
	case store.ChannelTypeForwardToUser:
		mapper = mapForwardToUserChannelDetails
	case store.ChannelTypeInAppNotification:
		mapper = mapInAppNotificationChannelDetails
	case store.ChannelTypePhoneCall:
		mapper = mapPhoneCallChannelDetails
	case store.ChannelTypeRadio:
		mapper = mapRadioChannelDetails
	default:
		return nil, meh.NewInternalErr("unsupported channel type", nil)
	}
	return mapper, nil
}

// mapDirectChannelDetails maps store.ChannelDetails with
// store.ChannelTypeDirect to event.AddressBookEntryDirectChannelDetails.
func mapDirectChannelDetails(detailsRaw store.ChannelDetails) (json.RawMessage, error) {
	details, ok := detailsRaw.(store.DirectChannelDetails)
	if !ok {
		return nil, meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	mappedDetails := event.AddressBookEntryDirectChannelDetails{
		Info: details.Info,
	}
	mappedDetailsRaw, err := json.Marshal(mappedDetails)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "marshal mapped details", meh.Details{"details": mappedDetails})
	}
	return mappedDetailsRaw, nil
}

// mapEmailChannelDetails maps store.ChannelDetails with
// store.ChannelTypeEmail to event.AddressBookEntryEmailChannelDetails.
func mapEmailChannelDetails(detailsRaw store.ChannelDetails) (json.RawMessage, error) {
	details, ok := detailsRaw.(store.EmailChannelDetails)
	if !ok {
		return nil, meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	mappedDetails := event.AddressBookEntryEmailChannelDetails{
		Email: details.Email,
	}
	mappedDetailsRaw, err := json.Marshal(mappedDetails)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "marshal mapped details", meh.Details{"details": mappedDetails})
	}
	return mappedDetailsRaw, nil
}

// mapForwardToGroupChannelDetails maps store.ChannelDetails with
// store.ChannelTypeForwardToGroup to event.AddressBookEntryForwardToGroupChannelDetails.
func mapForwardToGroupChannelDetails(detailsRaw store.ChannelDetails) (json.RawMessage, error) {
	details, ok := detailsRaw.(store.ForwardToGroupChannelDetails)
	if !ok {
		return nil, meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	mappedDetails := event.AddressBookEntryForwardToGroupChannelDetails{
		ForwardToGroup: details.ForwardToGroup,
	}
	mappedDetailsRaw, err := json.Marshal(mappedDetails)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "marshal mapped details", meh.Details{"details": mappedDetails})
	}
	return mappedDetailsRaw, nil
}

// mapForwardToUserChannelDetails maps store.ChannelDetails with
// store.ChannelTypeForwardToUser to event.AddressBookEntryForwardToUserChannelDetails.
func mapForwardToUserChannelDetails(detailsRaw store.ChannelDetails) (json.RawMessage, error) {
	details, ok := detailsRaw.(store.ForwardToUserChannelDetails)
	if !ok {
		return nil, meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	mappedDetails := event.AddressBookEntryForwardToUserChannelDetails{
		ForwardToUser: details.ForwardToUser,
	}
	mappedDetailsRaw, err := json.Marshal(mappedDetails)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "marshal mapped details", meh.Details{"details": mappedDetails})
	}
	return mappedDetailsRaw, nil
}

// mapInAppNotificationChannelDetails maps store.ChannelDetails with
// store.ChannelTypeInAppNotification to
// event.AddressBookEntryInAppNotificationChannelDetails.
func mapInAppNotificationChannelDetails(detailsRaw store.ChannelDetails) (json.RawMessage, error) {
	_, ok := detailsRaw.(store.InAppNotificationChannelDetails)
	if !ok {
		return nil, meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	mappedDetails := event.AddressBookEntryInAppNotificationChannelDetails{}
	mappedDetailsRaw, err := json.Marshal(mappedDetails)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "marshal mapped details", meh.Details{"details": mappedDetails})
	}
	return mappedDetailsRaw, nil
}

// mapPhoneCallChannelDetails maps store.ChannelDetails with
// store.ChannelTypePhoneCall to event.AddressBookEntryPhoneCallChannelDetails.
func mapPhoneCallChannelDetails(detailsRaw store.ChannelDetails) (json.RawMessage, error) {
	details, ok := detailsRaw.(store.PhoneCallChannelDetails)
	if !ok {
		return nil, meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	mappedDetails := event.AddressBookEntryPhoneCallChannelDetails{
		Phone: details.Phone,
	}
	mappedDetailsRaw, err := json.Marshal(mappedDetails)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "marshal mapped details", meh.Details{"details": mappedDetails})
	}
	return mappedDetailsRaw, nil
}

// mapRadioChannelDetails maps store.ChannelDetails with
// store.ChannelTypeRadio to event.AddressBookEntryRadioChannelDetails.
func mapRadioChannelDetails(detailsRaw store.ChannelDetails) (json.RawMessage, error) {
	details, ok := detailsRaw.(store.RadioChannelDetails)
	if !ok {
		return nil, meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	mappedDetails := event.AddressBookEntryRadioChannelDetails{
		Info: details.Info,
	}
	mappedDetailsRaw, err := json.Marshal(mappedDetails)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "marshal mapped details", meh.Details{"details": mappedDetails})
	}
	return mappedDetailsRaw, nil
}

// NotifyAddressBookEntryChannelsUpdated emits an
// event.TypeAddressBookEntryChannelsUpdated event.
func (p *Port) NotifyAddressBookEntryChannelsUpdated(ctx context.Context, tx pgx.Tx, entryID uuid.UUID, channels []store.Channel) error {
	value := event.AddressBookEntryChannelsUpdated{
		Entry:    entryID,
		Channels: make([]event.AddressBookEntryChannelsUpdatedChannel, 0, len(channels)),
	}
	// Map channels.
	for _, channel := range channels {
		var err error
		mappedChannel := event.AddressBookEntryChannelsUpdatedChannel{
			ID:            channel.ID,
			Entry:         channel.Entry,
			Label:         channel.Label,
			Priority:      channel.Priority,
			MinImportance: channel.MinImportance,
			Timeout:       channel.Timeout,
		}
		mappedChannel.Type, err = mapChannelType(channel.Type)
		if err != nil {
			return meh.Wrap(err, "map channel type", meh.Details{"channel_type": channel.Type})
		}
		detailsMapper, err := mapChannelDetails(channel.Type)
		if err != nil {
			return meh.Wrap(err, "get channel details mapper", meh.Details{"channel_type": channel.Type})
		}
		mappedChannel.Details, err = detailsMapper(channel.Details)
		if err != nil {
			return meh.Wrap(err, "map channel details", meh.Details{"channel_type": channel.Type})
		}
		value.Channels = append(value.Channels, mappedChannel)
	}
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.AddressBookTopic,
		Key:       entryID.String(),
		EventType: event.TypeAddressBookEntryChannelsUpdated,
		Value:     value,
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyIntelDeliveryCreated emits an event.TypeIntelDeliveryCreated event.
func (p *Port) NotifyIntelDeliveryCreated(ctx context.Context, tx pgx.Tx, created store.IntelDelivery) error {
	message := kafkautil.OutboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		Key:       created.ID.String(),
		EventType: event.TypeIntelDeliveryCreated,
		Value: event.IntelDeliveryCreated{
			ID:         created.ID,
			Assignment: created.Assignment,
			IsActive:   created.IsActive,
			Success:    created.Success,
			Note:       created.Note,
		},
		Headers: nil,
	}
	err := p.writer.AddOutboxMessages(ctx, tx, message)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": message})
	}
	return nil
}

// NotifyIntelDeliveryStatusUpdated emits an
// event.TypeIntelDeliveryStatusUpdated event.
func (p *Port) NotifyIntelDeliveryStatusUpdated(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID, newIsActive bool,
	newSuccess bool, newNote nulls.String) error {
	message := kafkautil.OutboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		Key:       deliveryID.String(),
		EventType: event.TypeIntelDeliveryStatusUpdated,
		Value: event.IntelDeliveryStatusUpdated{
			ID:       deliveryID,
			IsActive: newIsActive,
			Success:  newSuccess,
			Note:     newNote,
		},
		Headers: nil,
	}
	err := p.writer.AddOutboxMessages(ctx, tx, message)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": message})
	}
	return nil
}

// eventIntelDeliveryStatusFromStore maps store.IntelDeliveryStatus to
// event.IntelDeliveryStaus and returns a meh.ErrInternal when no mapping was
// found.
func eventIntelDeliveryStatusFromStore(s store.IntelDeliveryStatus) (event.IntelDeliveryStatus, error) {
	switch s {
	case store.IntelDeliveryStatusOpen:
		return event.IntelDeliveryStatusOpen, nil
	case store.IntelDeliveryStatusAwaitingDelivery:
		return event.IntelDeliveryStatusAwaitingDelivery, nil
	case store.IntelDeliveryStatusDelivering:
		return event.IntelDeliveryStatusDelivering, nil
	case store.IntelDeliveryStatusAwaitingAck:
		return event.IntelDeliveryStatusAwaitingAck, nil
	case store.IntelDeliveryStatusDelivered:
		return event.IntelDeliveryStatusDelivered, nil
	case store.IntelDeliveryStatusTimeout:
		return event.IntelDeliveryStatusTimeout, nil
	case store.IntelDeliveryStatusCanceled:
		return event.IntelDeliveryStatusCanceled, nil
	case store.IntelDeliveryStatusFailed:
		return event.IntelDeliveryStatusFailed, nil
	default:
		return "", meh.NewInternalErr("unsupported delivery-status", meh.Details{"state": s})
	}
}

// NotifyIntelDeliveryAttemptCreated emits an
// event.TypeIntelDeliveryAttemptCreated event.
func (p *Port) NotifyIntelDeliveryAttemptCreated(ctx context.Context, tx pgx.Tx, created store.IntelDeliveryAttempt,
	delivery store.IntelDelivery, assignment store.IntelAssignment, assignedEntry store.AddressBookEntryDetailed, intel store.Intel) error {
	mappedStatus, err := eventIntelDeliveryStatusFromStore(created.Status)
	if err != nil {
		return meh.Wrap(err, "event intel-delivery-status from store", meh.Details{"status": created.Status})
	}
	var mappedUserDetails nulls.JSONNullable[event.IntelDeliveryAttemptCreatedAssignedEntryUserDetails]
	if assignedEntry.UserDetails.Valid {
		v := assignedEntry.UserDetails.V
		mappedUserDetails = nulls.NewJSONNullable(event.IntelDeliveryAttemptCreatedAssignedEntryUserDetails{
			ID:        v.ID,
			Username:  v.Username,
			FirstName: v.FirstName,
			LastName:  v.LastName,
			IsActive:  v.IsActive,
		})
	}
	message := kafkautil.OutboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		Key:       created.Delivery.String(),
		EventType: event.TypeIntelDeliveryAttemptCreated,
		Value: event.IntelDeliveryAttemptCreated{
			ID: created.ID,
			Delivery: event.IntelDeliveryAttemptCreatedDelivery{
				ID:         delivery.ID,
				Assignment: delivery.Assignment,
				IsActive:   delivery.IsActive,
				Success:    delivery.Success,
				Note:       delivery.Note,
			},
			Assignment: event.IntelDeliveryAttemptCreatedAssignment{
				ID:    assignment.ID,
				Intel: assignment.Intel,
				To:    assignment.To,
			},
			AssignedEntry: event.IntelDeliveryAttemptCreatedAssignedEntry{
				ID:          assignedEntry.ID,
				Label:       assignedEntry.Label,
				Description: assignedEntry.Description,
				Operation:   assignedEntry.Operation,
				User:        assignedEntry.User,
				UserDetails: mappedUserDetails,
			},
			Intel: event.IntelDeliveryAttemptCreatedIntel{
				ID:         intel.ID,
				CreatedAt:  intel.CreatedAt,
				CreatedBy:  intel.CreatedBy,
				Operation:  intel.Operation,
				Type:       event.IntelType(intel.Type),
				Content:    intel.Content,
				SearchText: intel.SearchText,
				Importance: intel.Importance,
				IsValid:    intel.IsValid,
			},
			Channel:   created.Channel,
			CreatedAt: created.CreatedAt,
			IsActive:  created.IsActive,
			Status:    mappedStatus,
			StatusTS:  created.StatusTS,
			Note:      created.Note,
		},
		Headers: nil,
	}
	err = p.writer.AddOutboxMessages(ctx, tx, message)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": message})
	}
	return nil
}

// NotifyIntelDeliveryAttemptStatusUpdated emits an
// event.TypeIntelDeliveryAttemptStatusUpdated event.
func (p *Port) NotifyIntelDeliveryAttemptStatusUpdated(ctx context.Context, tx pgx.Tx, attempt store.IntelDeliveryAttempt) error {
	mappedStatus, err := eventIntelDeliveryStatusFromStore(attempt.Status)
	if err != nil {
		return meh.Wrap(err, "event intel-delivery-status from store", meh.Details{"status": attempt.Status})
	}
	message := kafkautil.OutboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		Key:       attempt.Delivery.String(),
		EventType: event.TypeIntelDeliveryAttemptStatusUpdated,
		Value: event.IntelDeliveryAttemptStatusUpdated{
			ID:       attempt.ID,
			IsActive: attempt.IsActive,
			Status:   mappedStatus,
			StatusTS: attempt.StatusTS,
			Note:     attempt.Note,
		},
		Headers: nil,
	}
	err = p.writer.AddOutboxMessages(ctx, tx, message)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": message})
	}
	return nil
}
