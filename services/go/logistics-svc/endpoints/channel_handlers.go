package endpoints

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"net/http"
	"reflect"
	"time"
)

// publicChannel is the public representation of store.Channel.
type publicChannel struct {
	ID            uuid.UUID       `json:"id"`
	Entry         uuid.UUID       `json:"entry"`
	Label         string          `json:"label"`
	Type          string          `json:"type"`
	Priority      int32           `json:"priority"`
	MinImportance float64         `json:"min_importance"`
	Details       json.RawMessage `json:"details"`
	Timeout       time.Duration   `json:"timeout"`
}

// publicChannelFromStore converts store.Channel to publicChannel.
func publicChannelFromStore(s store.Channel) (publicChannel, error) {
	p := publicChannel{
		ID:            s.ID,
		Entry:         s.Entry,
		Label:         s.Label,
		Type:          string(s.Type),
		Priority:      s.Priority,
		MinImportance: s.MinImportance,
		Timeout:       s.Timeout,
	}
	// Convert details.
	var marshalErr error
	switch d := s.Details.(type) {
	case store.DirectChannelDetails:
		p.Details, marshalErr = json.Marshal(publicDirectChannelDetailsFromStore(d))
	case store.EmailChannelDetails:
		p.Details, marshalErr = json.Marshal(publicEmailChannelDetailsFromStore(d))
	case store.ForwardToGroupChannelDetails:
		p.Details, marshalErr = json.Marshal(publicForwardToGroupChannelDetailsFromStore(d))
	case store.ForwardToUserChannelDetails:
		p.Details, marshalErr = json.Marshal(publicForwardToUserChannelDetailsFromStore(d))
	case store.PhoneCallChannelDetails:
		p.Details, marshalErr = json.Marshal(publicPhoneCallChannelDetailsFromStore(d))
	case store.InAppNotificationChannelDetails:
		p.Details, marshalErr = json.Marshal(publicInAppNotificationChannelDetailsFromStore(d))
	case store.RadioChannelDetails:
		p.Details, marshalErr = json.Marshal(publicRadioChannelDetailsFromStore(d))
	default:
		return publicChannel{}, meh.NewInternalErr("unsupported details type",
			meh.Details{"was": reflect.TypeOf(s.Details)})
	}
	if marshalErr != nil {
		return publicChannel{}, meh.NewInternalErrFromErr(marshalErr, "marshal channel details",
			meh.Details{"details_type": reflect.TypeOf(s.Details)})
	}
	return p, nil
}

// storeChannelDetailsFromPublic converts the given json.RawMessage to
// store.ChannelDetails with the given mapper function.
func storeChannelDetailsFromPublic[T any](details json.RawMessage, mapFn func(T) store.ChannelDetails) (store.ChannelDetails, error) {
	var v T
	err := json.Unmarshal(details, &v)
	if err != nil {
		return nil, meh.NewBadInputErrFromErr(err, "unmarshal details", meh.Details{"into": reflect.TypeOf(v)})
	}
	return mapFn(v), nil
}

// storeChannelFromPublic converts publicChannel to store.Channel.
func storeChannelFromPublic(p publicChannel) (store.Channel, error) {
	s := store.Channel{
		ID:            p.ID,
		Entry:         p.Entry,
		Label:         p.Label,
		Type:          store.ChannelType(p.Type),
		Priority:      p.Priority,
		MinImportance: p.MinImportance,
		Timeout:       p.Timeout,
	}
	// Parse details based on channel type.
	var err error
	switch s.Type {
	case store.ChannelTypeDirect:
		s.Details, err = storeChannelDetailsFromPublic(p.Details, storeDirectChannelDetailsFromPublic)
	case store.ChannelTypeEmail:
		s.Details, err = storeChannelDetailsFromPublic(p.Details, storeEmailChannelDetailsFromPublic)
	case store.ChannelTypeForwardToGroup:
		s.Details, err = storeChannelDetailsFromPublic(p.Details, storeForwardToGroupChannelDetailsFromPublic)
	case store.ChannelTypeForwardToUser:
		s.Details, err = storeChannelDetailsFromPublic(p.Details, storeForwardToUserChannelDetailsFromPublic)
	case store.ChannelTypeInAppNotification:
		s.Details, err = storeChannelDetailsFromPublic(p.Details, storeInAppNotificationChannelDetailsFromPublic)
	case store.ChannelTypePhoneCall:
		s.Details, err = storeChannelDetailsFromPublic(p.Details, storePhoneCallChannelDetailsFromPublic)
	case store.ChannelTypeRadio:
		s.Details, err = storeChannelDetailsFromPublic(p.Details, storeRadioChannelDetailsFromPublic)
	default:
		err = meh.NewBadInputErr("unsupported channel type", meh.Details{"type": s.Type})
	}
	if err != nil {
		return store.Channel{}, meh.Wrap(err, "parse details", meh.Details{
			"chan_type":    s.Type,
			"chan_details": p.Details,
		})
	}
	return s, nil
}

// publicDirectChannelDetails is the public representation of
// store.DirectChannelDetails.
type publicDirectChannelDetails struct {
	Info string `json:"info"`
}

// publicDirectChannelDetailsFromStore converts store.DirectChannelDetails to
// publicDirectChannelDetails.
func publicDirectChannelDetailsFromStore(s store.DirectChannelDetails) publicDirectChannelDetails {
	return publicDirectChannelDetails{
		Info: s.Info,
	}
}

// storeDirectChannelDetailsFromPublic converts publicDirectChannelDetails to
// store.DirectChannelDetails.
func storeDirectChannelDetailsFromPublic(p publicDirectChannelDetails) store.ChannelDetails {
	return store.DirectChannelDetails{
		Info: p.Info,
	}
}

// publicEmailChannelDetails is the public representation of
// store.EmailChannelDetails.
type publicEmailChannelDetails struct {
	Email string `json:"email"`
}

// publicEmailChannelDetailsFromStore converts store.EmailChannelDetails to
// publicEmailChannelDetails.
func publicEmailChannelDetailsFromStore(s store.EmailChannelDetails) publicEmailChannelDetails {
	return publicEmailChannelDetails{
		Email: s.Email,
	}
}

// storeEmailChannelDetailsFromPublic converts publicEmailChannelDetails to
// store.EmailChannelDetails.
func storeEmailChannelDetailsFromPublic(p publicEmailChannelDetails) store.ChannelDetails {
	return store.EmailChannelDetails{
		Email: p.Email,
	}
}

// publicForwardToGroupChannelDetails is the public representation of
// store.ForwardToGroupChannelDetails.
type publicForwardToGroupChannelDetails struct {
	ForwardToGroup []uuid.UUID `json:"forward_to_group"`
}

// publicForwardToGroupChannelDetailsFromStore converts
// store.ForwardToGroupChannelDetails to publicForwardToGroupChannelDetails.
func publicForwardToGroupChannelDetailsFromStore(s store.ForwardToGroupChannelDetails) publicForwardToGroupChannelDetails {
	return publicForwardToGroupChannelDetails{
		ForwardToGroup: s.ForwardToGroup,
	}
}

// storeForwardToGroupChannelDetailsFromPublic converts
// publicForwardToGroupChannelDetails to store.ForwardToGroupChannelDetails.
func storeForwardToGroupChannelDetailsFromPublic(p publicForwardToGroupChannelDetails) store.ChannelDetails {
	return store.ForwardToGroupChannelDetails{
		ForwardToGroup: p.ForwardToGroup,
	}
}

// publicForwardToUserChannelDetails is the public representation of
// store.ForwardToUserChannelDetails.
type publicForwardToUserChannelDetails struct {
	ForwardToUser []uuid.UUID `json:"forward_to_user"`
}

// publicForwardToUserChannelDetailsFromStore converts
// store.ForwardToUserChannelDetails to publicForwardToUserChannelDetails.
func publicForwardToUserChannelDetailsFromStore(s store.ForwardToUserChannelDetails) publicForwardToUserChannelDetails {
	return publicForwardToUserChannelDetails{
		ForwardToUser: s.ForwardToUser,
	}
}

// storeForwardToUserChannelDetailsFromPublic converts
// publicForwardToUserChannelDetails to store.ForwardToUserChannelDetails.
func storeForwardToUserChannelDetailsFromPublic(p publicForwardToUserChannelDetails) store.ChannelDetails {
	return store.ForwardToUserChannelDetails{
		ForwardToUser: p.ForwardToUser,
	}
}

// publicPhoneCallChannelDetails is the public representation of
// store.PhoneCallChannelDetails.
type publicPhoneCallChannelDetails struct {
	Phone string `json:"phone"`
}

// publicPhoneCallChannelDetailsFromStore converts
// store.PhoneCallChannelDetails to publicPhoneCallChannelDetails.
func publicPhoneCallChannelDetailsFromStore(s store.PhoneCallChannelDetails) publicPhoneCallChannelDetails {
	return publicPhoneCallChannelDetails{
		Phone: s.Phone,
	}
}

// storePhoneCallChannelDetailsFromPublic converts
// publicPhoneCallChannelDetails to store.PhoneCallChannelDetails.
func storePhoneCallChannelDetailsFromPublic(p publicPhoneCallChannelDetails) store.ChannelDetails {
	return store.PhoneCallChannelDetails{
		Phone: p.Phone,
	}
}

// publicInAppNotificationChannelDetails is the public representation of
// store.InAppNotificationChannelDetails.
type publicInAppNotificationChannelDetails struct {
}

// publicInAppNotificationChannelDetailsFromStore converts
// store.InAppNotificationChannelDetails to
// publicInAppNotificationChannelDetails.
func publicInAppNotificationChannelDetailsFromStore(_ store.InAppNotificationChannelDetails) publicInAppNotificationChannelDetails {
	return publicInAppNotificationChannelDetails{}
}

// storeInAppNotificationChannelDetailsFromPublic converts
// publicInAppNotificationChannelDetails to store.InAppNotificationChannelDetails.
func storeInAppNotificationChannelDetailsFromPublic(_ publicInAppNotificationChannelDetails) store.ChannelDetails {
	return store.InAppNotificationChannelDetails{}
}

// publicRadioChannelDetails is the public representation of
// store.RadioChannelDetails.
type publicRadioChannelDetails struct {
	Info string `json:"info"`
}

// publicRadioChannelDetailsFromStore converts
// store.RadioChannelDetails to publicRadioChannelDetails.
func publicRadioChannelDetailsFromStore(s store.RadioChannelDetails) publicRadioChannelDetails {
	return publicRadioChannelDetails{
		Info: s.Info,
	}
}

// storeRadioChannelDetailsFromPublic converts
// publicRadioChannelDetails to store.RadioChannelDetails.
func storeRadioChannelDetailsFromPublic(p publicRadioChannelDetails) store.ChannelDetails {
	return store.RadioChannelDetails{
		Info: p.Info,
	}
}

// handleUpdateChannelsByAddressBookEntryStore are the dependencies needed for
// handleUpdateChannelsByAddressBookEntry.
type handleUpdateChannelsByAddressBookEntryStore interface {
	UpdateChannelsByAddressBookEntry(ctx context.Context, entryID uuid.UUID, newChannels []store.Channel, limitToUser uuid.NullUUID) error
}

// handleUpdateChannelsByAddressBookEntry updates the channels for the address
// book entry with the given id. If the client does not have
// permission.UpdateAnyAddressBookEntry, updates are limited to entries,
// associated with the client.
func handleUpdateChannelsByAddressBookEntry(s handleUpdateChannelsByAddressBookEntryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Only check if authorized. Then, we check the permission to update any address
		// book entry. If this is not given, we will pass a flag to the controller for
		// updating.
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authorized", nil)
		}
		// Extract entry id.
		entryIDStr := c.Param("entryID")
		entryID, err := uuid.FromString(entryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse entry id", meh.Details{"was": entryIDStr})
		}
		// Parse body.
		var publicUpdate []publicChannel
		err = json.NewDecoder(c.Request.Body).Decode(&publicUpdate)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse body", nil)
		}
		update := make([]store.Channel, 0, len(publicUpdate))
		for _, pChannel := range publicUpdate {
			sChannel, err := storeChannelFromPublic(pChannel)
			if err != nil {
				return meh.Wrap(err, "store channel from public", nil)
			}
			update = append(update, sChannel)
		}
		// Remove user limit if update-permission granted.
		limitToUser := nulls.NewUUID(token.UserID)
		updateAnyGranted, err := auth.HasPermission(token, permission.UpdateAnyAddressBookEntry())
		if err != nil {
			return meh.Wrap(err, "permission check", nil)
		}
		if updateAnyGranted {
			limitToUser = uuid.NullUUID{}
		}
		// Update.
		err = s.UpdateChannelsByAddressBookEntry(c.Request.Context(), entryID, update, limitToUser)
		if err != nil {
			return meh.Wrap(err, "update channels by address book entry in store", meh.Details{
				"entry_id": entryID,
				"update":   update,
			})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// handleGetChannelsByAddressBookEntryStore are the dependencies needed for
// handleGetChannelsByAddressBookEntry.
type handleGetChannelsByAddressBookEntryStore interface {
	ChannelsByAddressBookEntry(ctx context.Context, entryID uuid.UUID, limitToUser uuid.NullUUID) ([]store.Channel, error)
}

// handleGetChannelsByAddressBookEntry retrieves channels for the address book
// entry with the given id.
func handleGetChannelsByAddressBookEntry(s handleGetChannelsByAddressBookEntryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Only check if authorized. Then, we check the permission to retrieve channels
		// for any address book entry. If this is not given, we will pass a flag to the
		// controller for retrieval.
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authorized", nil)
		}
		// Extract entry id.
		entryIDStr := c.Param("entryID")
		entryID, err := uuid.FromString(entryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse entry id", meh.Details{"was": entryIDStr})
		}
		// Remove user limit if view-permission granted.
		limitToUser := nulls.NewUUID(token.UserID)
		granted, err := auth.HasPermission(token, permission.ViewAnyAddressBookEntry())
		if err != nil {
			return meh.Wrap(err, "permission check", nil)
		}
		if granted {
			limitToUser = uuid.NullUUID{}
		}
		// Update.
		sChannels, err := s.ChannelsByAddressBookEntry(c.Request.Context(), entryID, limitToUser)
		if err != nil {
			return meh.Wrap(err, "channels by address book entry from store", meh.Details{
				"entry_id":      entryID,
				"limit_to_user": limitToUser,
			})
		}
		pChannels := make([]publicChannel, 0, len(sChannels))
		for _, sChannel := range sChannels {
			pChannel, err := publicChannelFromStore(sChannel)
			if err != nil {
				return meh.Wrap(err, "convert store to public channel", meh.Details{"store_channel": sChannel})
			}
			pChannels = append(pChannels, pChannel)
		}
		c.JSON(http.StatusOK, pChannels)
		return nil
	}
}
