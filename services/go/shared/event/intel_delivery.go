package event

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"time"
)

// IntelDeliveryStatus for status in intel-deliveries.
type IntelDeliveryStatus string

const (
	// IntelDeliveryStatusOpen for deliveries, not being picked up, yet.
	IntelDeliveryStatusOpen IntelDeliveryStatus = "open"
	// IntelDeliveryStatusAwaitingDelivery for deliveries that have been picked up
	// by a mail carrier and are now awaiting delivery.
	IntelDeliveryStatusAwaitingDelivery IntelDeliveryStatus = "awaiting-delivery"
	// IntelDeliveryStatusDelivering for deliveries that are currently delivering
	// (for example ongoing phone calls).
	IntelDeliveryStatusDelivering IntelDeliveryStatus = "delivering"
	// IntelDeliveryStatusAwaitingAck for deliveries that are currently awaiting ACK
	// from the recipient (for example push notifications or email).
	IntelDeliveryStatusAwaitingAck IntelDeliveryStatus = "awaiting-ack"
	// IntelDeliveryStatusDelivered for deliveries that are successfully delivered
	// and acknowledged by the recipient.
	IntelDeliveryStatusDelivered IntelDeliveryStatus = "delivered"
	// IntelDeliveryStatusTimeout for deliveries that timed out (based on timeout
	// specified in channel properties).
	IntelDeliveryStatusTimeout IntelDeliveryStatus = "timeout"
	// IntelDeliveryStatusCanceled for manually cancelled deliveries.
	IntelDeliveryStatusCanceled IntelDeliveryStatus = "canceled"
	// IntelDeliveryStatusFailed is used for failed deliveries. An example might be
	// an invalid phone number that cannot be called.
	IntelDeliveryStatusFailed IntelDeliveryStatus = "failed"
)

// TypeIntelDeliveryCreated when an intel-delivery is created.
const TypeIntelDeliveryCreated Type = "intel-delivery-created"

// IntelDeliveryCreated for TypeIntelDeliveryCreated.
type IntelDeliveryCreated struct {
	// ID identifies the delivery.
	ID uuid.UUID `json:"id"`
	// Assignment is the id of the referenced assignment, holding further
	// information.
	Assignment uuid.UUID `json:"assignment"`
	// IsActive describes, whether the delivery is still active and should be
	// checked by the scheduler/controller.
	IsActive bool `json:"is_active"`
	// Success when delivery was successful.
	Success bool `json:"success"`
	// Note contains optional human-readable information regarding the delivery.
	Note nulls.String `json:"note"`
}

// TypeIntelDeliveryAttemptCreated when an attempt for intel-delivery for a
// specific channel was created.
const TypeIntelDeliveryAttemptCreated Type = "intel-delivery-attempt-created"

// IntelDeliveryAttemptCreated for TypeIntelDeliveryAttemptCreated.
type IntelDeliveryAttemptCreated struct {
	// ID identifies the attempt.
	ID uuid.UUID `json:"id"`
	// Delivery is the id of the referenced delivery.
	Delivery uuid.UUID `json:"delivery"`
	// Channel is the id of the channel to use for this attempt.
	Channel uuid.UUID `json:"channel"`
	// CreatedAt is the timestamp when the attempt was started.
	CreatedAt time.Time `json:"created_at"`
	// IsActive describes whether the attempt is still ongoing.
	IsActive bool `json:"is_active"`
	// Status is the current/last status of the attempt.
	Status IntelDeliveryStatus `json:"status"`
	// StatusTS is the timestamp when the Status was last updated.
	StatusTS time.Time `json:"status_ts"`
	// Note contains optional human-readable information regarding the attempt.
	Note nulls.String `json:"note"`
}

// TypeIntelDeliveryAttemptStatusUpdated for updated status of an
// intel-delivery-attempt.
const TypeIntelDeliveryAttemptStatusUpdated Type = "intel-delivery-attempt-status-updated"

// IntelDeliveryAttemptStatusUpdated for TypeIntelDeliveryAttemptStatusUpdated.
type IntelDeliveryAttemptStatusUpdated struct {
	// ID identifies the attempt.
	ID uuid.UUID `json:"id"`
	// IsActive describes whether the attempt is still ongoing.
	IsActive bool `json:"is_active"`
	// Status is the current/last status of the attempt.
	Status IntelDeliveryStatus `json:"status"`
	// StatusTS is the timestamp when the Status was last updated.
	StatusTS time.Time `json:"status_ts"`
	// Note contains optional human-readable information regarding the attempt.
	Note nulls.String `json:"note"`
}

// TypeIntelDeliveryStatusUpdated for updated status of an intel-delivery. In
// most cases used when the delivery is marked as finished.
const TypeIntelDeliveryStatusUpdated Type = "intel-delivery-status-updated"

// IntelDeliveryStatusUpdated for TypeIntelDeliveryStatusUpdated.
type IntelDeliveryStatusUpdated struct {
	// ID identifies the delivery.
	ID uuid.UUID `json:"id"`
	// IsActive describes whether the delivery is still active.
	IsActive bool `json:"is_active"`
	// Succes when delivery was successful.
	Success bool `json:"success"`
	// Note contains optional human-readable information regarding the delivery.
	Note nulls.String `json:"note"`
}
