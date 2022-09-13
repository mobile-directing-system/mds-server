package event

import (
	"github.com/gofrs/uuid"
	"time"
)

// TypeInAppNotificationForIntelPending is used when an in-app-notification for
// an intel is pending.
const TypeInAppNotificationForIntelPending Type = "in-app-notification-for-intel-pending"

// InAppNotificationForIntelPending is the value for
// TypeInAppNotificationForIntelPending.
type InAppNotificationForIntelPending struct {
	// Attempt is the id of the associated intel-delivery-attempt.
	Attempt uuid.UUID `json:"attempt"`
	// Since is the timestamp since when the notification is pending.
	Since time.Time `json:"since"`
}

// TypeInAppNotificationForIntelSent is used when an in-app-notification for an
// intel was sent.
const TypeInAppNotificationForIntelSent Type = "in-app-notification-for-intel-sent"

// InAppNotificationForIntelSent is the value for
// TypeInAppNotificationForIntelSent.
type InAppNotificationForIntelSent struct {
	// Attempt is the id of the associated intel-delivery-attempt.
	Attempt uuid.UUID `json:"attempt"`
	// SentAt is the timestamp when the notification was sent.
	SentAt time.Time `json:"sent_at"`
}
