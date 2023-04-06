package ws

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"time"
)

const (
	// messageTypeIntelNotification is used in connection.Notify for notifying about
	// an intel.
	messageTypeIntelNotification wsutil.MessageType = "intel-notification"
)

// publicIntelToDeliver is the public representation of store.IntelToDeliver.
type publicIntelToDeliver struct {
	Attempt    uuid.UUID       `json:"attempt"`
	ID         uuid.UUID       `json:"id"`
	CreatedAt  time.Time       `json:"created_at"`
	CreatedBy  uuid.UUID       `json:"created_by"`
	Operation  uuid.UUID       `json:"operation"`
	Type       string          `json:"type"`
	Content    json.RawMessage `json:"content"`
	Importance int             `json:"importance"`
}

// mapStoreIntelToDeliveryToPublic maps store.IntelToDeliver to
// publicIntelToDeliver.
func mapStoreIntelToDeliveryToPublic(s store.IntelToDeliver) publicIntelToDeliver {
	return publicIntelToDeliver{
		Attempt:    s.Attempt,
		ID:         s.ID,
		CreatedAt:  s.CreatedAt,
		CreatedBy:  s.CreatedBy,
		Operation:  s.Operation,
		Type:       string(s.Type),
		Content:    s.Content,
		Importance: s.Importance,
	}
}

// publicIntelDeliveryAttempt is the public representation of
// store.AcceptedIntelDeliveryAttempt.
type publicIntelDeliveryAttempt struct {
	ID              uuid.UUID     `json:"id"`
	AssignedTo      uuid.UUID     `json:"assigned_to"`
	AssignedToLabel string        `json:"assigned_to_label"`
	AssignedToUser  uuid.NullUUID `json:"assigned_to_user"`
	Delivery        uuid.UUID     `json:"delivery"`
	Channel         uuid.UUID     `json:"channel"`
	CreatedAt       time.Time     `json:"created_at"`
	IsActive        bool          `json:"is_active"`
	StatusTS        time.Time     `json:"status_ts"`
	Note            nulls.String  `json:"note"`
	AcceptedAt      time.Time     `json:"accepted_at"`
}

// mapStoreIntelDeliveryAttemptToPublic maps store.AcceptedIntelDeliveryAttempt
// to publicIntelDeliveryAttempt.
func mapStoreIntelDeliveryAttemptToPublic(s store.AcceptedIntelDeliveryAttempt) publicIntelDeliveryAttempt {
	return publicIntelDeliveryAttempt{
		ID:              s.ID,
		AssignedTo:      s.AssignedTo,
		AssignedToLabel: s.AssignedToLabel,
		AssignedToUser:  s.AssignedToUser,
		Delivery:        s.Delivery,
		Channel:         s.Channel,
		CreatedAt:       s.CreatedAt,
		IsActive:        s.IsActive,
		StatusTS:        s.StatusTS,
		Note:            s.Note,
		AcceptedAt:      s.AcceptedAt,
	}
}

// publicNotificationChannel is the public representation of
// store.NotificationChannel.
type publicNotificationChannel struct {
	ID      uuid.UUID     `json:"id"`
	Entry   uuid.UUID     `json:"entry"`
	Label   string        `json:"label"`
	Timeout time.Duration `json:"timeout"`
}

// mapStoreNotificationChannelToPublic maps store.NotificationChannel to
// publicNotificationChannel.
func mapStoreNotificationChannelToPublic(s store.NotificationChannel) publicNotificationChannel {
	return publicNotificationChannel{
		ID:      s.ID,
		Entry:   s.Entry,
		Label:   s.Label,
		Timeout: s.Timeout,
	}
}

// publicUser is the public representation of store.User.
type publicUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsActive  bool      `json:"is_active"`
}

// mapStoreToPublic maps store.User to publicUser.
func mapStoreUserToPublic(s store.User) publicUser {
	return publicUser{
		ID:        s.ID,
		Username:  s.Username,
		FirstName: s.FirstName,
		LastName:  s.LastName,
		IsActive:  s.IsActive,
	}
}

// publicIntelDeliveryNotification is the public representation of
// store.OutgoingIntelDeliveryNotification.
type publicIntelDeliveryNotification struct {
	IntelToDeliver   publicIntelToDeliver           `json:"intel_to_deliver"`
	DeliveryAttempt  publicIntelDeliveryAttempt     `json:"delivery_attempt"`
	Channel          publicNotificationChannel      `json:"channel"`
	CreatorDetails   publicUser                     `json:"creator_details"`
	RecipientDetails nulls.JSONNullable[publicUser] `json:"recipient_details"`
}

// mapStoreOutgoingIntelDeliveryNotificationToPublic maps
// store.OutgoingIntelDeliveryNotification to publicIntelDeliveryNotification.
func mapStoreOutgoingIntelDeliveryNotificationToPublic(s store.OutgoingIntelDeliveryNotification) publicIntelDeliveryNotification {
	n := publicIntelDeliveryNotification{
		IntelToDeliver:  mapStoreIntelToDeliveryToPublic(s.IntelToDeliver),
		DeliveryAttempt: mapStoreIntelDeliveryAttemptToPublic(s.DeliveryAttempt),
		Channel:         mapStoreNotificationChannelToPublic(s.Channel),
		CreatorDetails:  mapStoreUserToPublic(s.CreatorDetails),
	}
	if s.RecipientDetails.Valid {
		n.RecipientDetails = nulls.NewJSONNullable(mapStoreUserToPublic(s.RecipientDetails.V))
	}
	return n
}

// connection implements controller.Connection and maps notifications from the
// store-representation to the public one.
type connection struct {
	conn wsutil.AutoParserConnection
}

// newConnection returns a new connection.
func newConnection(conn wsutil.AutoParserConnection) *connection {
	return &connection{conn: conn}
}

// UserID returns the user id from the connection's auth-token.
func (conn *connection) UserID() uuid.UUID {
	return conn.conn.AuthToken().UserID
}

// Notify amps the given store.OutgoingIntelDeliveryNotification to it's public
// representation and sends it over the connection.
func (conn *connection) Notify(ctx context.Context, notification store.OutgoingIntelDeliveryNotification) error {
	publicNotif := mapStoreOutgoingIntelDeliveryNotificationToPublic(notification)
	err := conn.conn.Send(ctx, messageTypeIntelNotification, publicNotif)
	if err != nil {
		return meh.Wrap(err, "send over connection", meh.Details{"payload": publicNotif})
	}
	return nil
}

// Done returns the done-channel from the connection.
func (conn *connection) Done() <-chan struct{} {
	return conn.conn.Lifetime().Done()
}
