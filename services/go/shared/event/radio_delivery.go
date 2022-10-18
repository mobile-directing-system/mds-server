package event

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"time"
)

// TypeRadioDeliveryReadyForPickup notifies that a radio delivery for an
// intel-delivery-attempt is ready for pickup by a user for delivery.
const TypeRadioDeliveryReadyForPickup Type = "radio-delivery-ready-for-pickup"

// RadioDeliveryReadyForPickup is the content for
// TypeRadioDeliveryReadyForPickup.
type RadioDeliveryReadyForPickup struct {
	// Attempt identifies the intel-delivery-attempt.
	Attempt uuid.UUID
	// Intel is the id of the referenced intel.
	Intel uuid.UUID
	// IntelOperation is the operation the referenced intel is assigned to.
	IntelOperation uuid.UUID
	// IntelImportance is the importance of the referenced intel.
	IntelImportance int
	// AttemptAssignedTo is the id of the assigned address book entry.
	AttemptAssignedTo uuid.UUID
	// AttemptAssignedToLabel is the label of the assigned address book entry.
	AttemptAssignedToLabel string
	// Delivery is the id of the referenced delivery.
	Delivery uuid.UUID
	// Channel is the id of the channel to use for this attempt.
	Channel uuid.UUID
	// Note contains optional human-readable information regarding the radio
	// delivery.
	Note string
	// AttemptAcceptedAt is the timestamp when the intel-delivery-attempt was accepted by
	// the radio-delivery-service.
	AttemptAcceptedAt time.Time
}

// TypeRadioDeliveryPickedUp for a radio delivery being picked up.
const TypeRadioDeliveryPickedUp Type = "radio-delivery-picked-up"

// RadioDeliveryPickedUp is the content for RadioDeliveryPickedUp.
type RadioDeliveryPickedUp struct {
	// Attempt is the id of the referenced intel-delivery-attempt.
	Attempt uuid.UUID
	// PickedUpBy is the id of the user that picked up the radio delivery.
	PickedUpBy uuid.UUID
	// PickedUpAt is the timestamp when the radio delivery was picked up.
	PickedUpAt time.Time
}

// TypeRadioDeliveryReleased for picked up radio deliveries being released again
// and therefore ready for pickup.
const TypeRadioDeliveryReleased Type = "radio-delivery-released"

// RadioDeliveryReleased is the content for TypeRadioDeliveryReleased.
type RadioDeliveryReleased struct {
	// Attempt is the id of the referenced intel-delivery-attempt.
	Attempt uuid.UUID
	// ReleasedAt is the timestamp when the radio delivery was released.
	ReleasedAt time.Time
}

// TypeRadioDeliveryFinished when a radio delivery was finished/canceled.
const TypeRadioDeliveryFinished Type = "radio-delivery-finished"

// RadioDeliveryFinished is the content for TypeRadioDeliveryFinished.
type RadioDeliveryFinished struct {
	// Attempt is the id of the referenced intel-delivery-attempt.
	Attempt uuid.UUID
	// PickedUpBy is the id of the user that picked up the radio delivery.
	PickedUpBy uuid.NullUUID
	// PickedUpAt is the timestamp when the user from PickedUpBy picked up the radio
	// delivery.
	PickedUpAt nulls.Time
	// Success of the radio delivery.
	Success bool
	// FinishedAt is the timestamp when the radio delivery was marked as finished.
	FinishedAt time.Time
	// Note contains optional information.
	Note string
}
