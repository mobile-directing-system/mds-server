package eventport

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
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
func (p *Port) NotifyAddressBookEntryCreated(entry store.AddressBookEntry) error {
	err := kafkautil.WriteMessages(p.writer, kafkautil.Message{
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
func (p *Port) NotifyAddressBookEntryUpdated(entry store.AddressBookEntry) error {
	err := kafkautil.WriteMessages(p.writer, kafkautil.Message{
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
func (p *Port) NotifyAddressBookEntryDeleted(entryID uuid.UUID) error {
	err := kafkautil.WriteMessages(p.writer, kafkautil.Message{
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
	case store.ChannelTypePhoneCall:
		mappedType = event.AddressBookEntryChannelTypePhoneCall
	case store.ChannelTypeRadio:
		mappedType = event.AddressBookEntryChannelTypeRadio
	case store.ChannelTypePush:
		mappedType = event.AddressBookEntryChannelTypePush
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
	case store.ChannelTypePhoneCall:
		mapper = mapPhoneCallChannelDetails
	case store.ChannelTypeRadio:
		mapper = mapRadioChannelDetails
	case store.ChannelTypePush:
		mapper = mapPushChannelDetails
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

// mapPushChannelDetails maps store.ChannelDetails with
// store.ChannelTypePush to event.AddressBookEntryPushChannelDetails.
func mapPushChannelDetails(detailsRaw store.ChannelDetails) (json.RawMessage, error) {
	_, ok := detailsRaw.(store.PushChannelDetails)
	if !ok {
		return nil, meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	mappedDetails := event.AddressBookEntryPushChannelDetails{}
	mappedDetailsRaw, err := json.Marshal(mappedDetails)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "marshal mapped details", meh.Details{"details": mappedDetails})
	}
	return mappedDetailsRaw, nil
}

// NotifyAddressBookEntryChannelsUpdated emits an
// event.TypeAddressBookEntryChannelsUpdated event.
func (p *Port) NotifyAddressBookEntryChannelsUpdated(entryID uuid.UUID, channels []store.Channel) error {
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
	err := kafkautil.WriteMessages(p.writer, kafkautil.Message{
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
