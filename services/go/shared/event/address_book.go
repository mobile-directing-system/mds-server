package event

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"time"
)

// TypeAddressBookEntryCreated is used when an entry in the address book was
// created.
const TypeAddressBookEntryCreated Type = "address-book-entry-created"

// AddressBookEntryCreated is the value for TypeAddressBookEntryCreated.
type AddressBookEntryCreated struct {
	// ID identifies the entry.
	ID uuid.UUID `json:"id"`
	// Label for better human-readability.
	Label string `json:"label"`
	// Description for better human-readability.
	//
	// Example use-case: Multiple entries for high-rank groups are created. However,
	// each one targets slightly different people. This can be used in order to pick
	// the right one.
	Description string `json:"description"`
	// Operation holds the id of an optionally assigned operation.
	Operation uuid.NullUUID `json:"operation"`
	// User is the id of an optionally assigned user.
	User uuid.NullUUID `json:"user"`
}

// TypeAddressBookEntryUpdated is used when an entry in the address book was
// created.
const TypeAddressBookEntryUpdated Type = "address-book-entry-updated"

// AddressBookEntryUpdated is the value for TypeAddressBookEntryUpdated.
type AddressBookEntryUpdated struct {
	// ID identifies the entry.
	ID uuid.UUID `json:"id"`
	// Label for better human-readability.
	Label string `json:"label"`
	// Description for better human-readability.
	//
	// Example use-case: Multiple entries for high-rank groups are created. However,
	// each one targets slightly different people. This can be used in order to pick
	// the right one.
	Description string `json:"description"`
	// Operation holds the id of an optionally assigned operation.
	Operation uuid.NullUUID `json:"operation"`
	// User is the id of an optionally assigned user.
	User uuid.NullUUID `json:"user"`
}

// TypeAddressBookEntryDeleted is used when an entry in the address book along
// with its channels was deleted.
const TypeAddressBookEntryDeleted Type = "address-book-entry-deleted"

// AddressBookEntryDeleted is the value for TypeAddressBookEntryDeleted.
type AddressBookEntryDeleted struct {
	// ID of the deleted entry.
	ID uuid.UUID `json:"id"`
}

// AddressBookEntryChannelType in channels.Type that specifies the type as well
// as the content of the channel details.
type AddressBookEntryChannelType string

const (
	// AddressBookEntryChannelTypeDirect is used for direct communication without any (electronic)
	// medium. This equals talking in-person.
	AddressBookEntryChannelTypeDirect AddressBookEntryChannelType = "direct"
	// AddressBookEntryChannelTypeEmail is used for communicating via email.
	AddressBookEntryChannelTypeEmail AddressBookEntryChannelType = "email"
	// AddressBookEntryChannelTypeForwardToGroup is used for forwarding to another
	// group.
	AddressBookEntryChannelTypeForwardToGroup AddressBookEntryChannelType = "forward-to-group"
	// AddressBookEntryChannelTypeForwardToUser is used for forwarding to another
	// user.
	AddressBookEntryChannelTypeForwardToUser AddressBookEntryChannelType = "forward-to-user"
	// AddressBookEntryChannelTypeInAppNotification is used for notifying via in-app
	// notifications, similar to push-notifications.
	AddressBookEntryChannelTypeInAppNotification AddressBookEntryChannelType = "in-app-notification"
	// AddressBookEntryChannelTypePhoneCall is used for communicating via phone
	// calls.
	AddressBookEntryChannelTypePhoneCall AddressBookEntryChannelType = "phone-call"
	// AddressBookEntryChannelTypeRadio is used for communicating via radio.
	AddressBookEntryChannelTypeRadio AddressBookEntryChannelType = "radio"
)

// AddressBookEntryDirectChannelDetails holds channel details for
// AddressBookEntryChannelTypeDirect.
type AddressBookEntryDirectChannelDetails struct {
	// Info holds any free-text information.
	Info string `json:"info"`
}

// AddressBookEntryEmailChannelDetails holds channel details for
// AddressBookEntryChannelTypeEmail.
type AddressBookEntryEmailChannelDetails struct {
	// Email is the email address.
	Email string `json:"email"`
}

// AddressBookEntryForwardToGroupChannelDetails holds channel details for
// AddressBookEntryChannelTypeForwardToGroup.
type AddressBookEntryForwardToGroupChannelDetails struct {
	// ForwardToGroup is the id of the group that should be forwarded to.
	ForwardToGroup []uuid.UUID `json:"forward_to_group"`
}

// AddressBookEntryForwardToUserChannelDetails holds channel details for
// AddressBookEntryChannelTypeForwardToUser.
type AddressBookEntryForwardToUserChannelDetails struct {
	// ForwardToUser is the id of the user that should be forwarded to.
	ForwardToUser []uuid.UUID `json:"forward_to_user"`
}

// AddressBookEntryInAppNotificationChannelDetails holds channel details for
// AddressBookEntryChannelTypeInAppNotification.
type AddressBookEntryInAppNotificationChannelDetails struct {
}

// AddressBookEntryPhoneCallChannelDetails holds channel details for
// AddressBookEntryChannelTypePhoneCall.
type AddressBookEntryPhoneCallChannelDetails struct {
	// Phone is the phone number.
	Phone string `json:"phone"`
}

// AddressBookEntryRadioChannelDetails holds channel details for
// AddressBookEntryChannelTypeRadio.
type AddressBookEntryRadioChannelDetails struct {
	// Info holds any free-text information until radio communication is further
	// specified.
	Info string `json:"info"`
}

// TypeAddressBookEntryChannelsUpdated is used when the channels for an address
// book entry have been updated.
const TypeAddressBookEntryChannelsUpdated Type = "address-book-entry-channels-updated"

// AddressBookEntryChannelsUpdated is the value for
// TypeAddressBookEntryChannelsUpdated.
type AddressBookEntryChannelsUpdated struct {
	// Entry is the id of the entry the channel is assigned to.
	Entry uuid.UUID `json:"entry"`
	// Channels holds the updated channels.
	Channels []AddressBookEntryChannelsUpdatedChannel `json:"channels"`
}

// AddressBookEntryChannelsUpdatedChannel is used in
// AddressBookEntryChannelsUpdated.Channels
type AddressBookEntryChannelsUpdatedChannel struct {
	// ID identifies the channel.
	ID uuid.UUID `json:"id"`
	// Entry is the id of the entry the channel is assigned to.
	Entry uuid.UUID `json:"entry"`
	// Label of the channel for better human-readability.
	Label string `json:"label"`
	// Type of the channel.
	Type AddressBookEntryChannelType `json:"type"`
	// Priority is a unique priority of the channel.
	Priority int32 `json:"priority"`
	// MinImportance of information in order to use this channel.
	MinImportance float64 `json:"min_importance"`
	// Details for the channel, for example AddressBookEntryPhoneCallChannelDetails,
	// based on Type.
	Details json.RawMessage `json:"details"`
	// Timeout after which a message delivery over this channel is considered as
	// timed out.
	Timeout time.Duration `json:"timeout"`
}
