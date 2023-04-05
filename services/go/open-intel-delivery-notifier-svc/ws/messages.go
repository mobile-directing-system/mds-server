package ws

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"time"
)

const (
	// messageTypeSubscribeOpenIntelDeliveries is expected from the client when
	// subscribing to open intel deliveries for a certain operation is desired.
	messageTypeSubscribeOpenIntelDeliveries wsutil.MessageType = "subscribe-open-intel-deliveries"
	// messageTypeUnsubscribeOpenIntelDeliveries is expected from the client when
	// unsubscribing from open intel deliveries for a certain operation is desired.
	messageTypeUnsubscribeOpenIntelDeliveries wsutil.MessageType = "unsubscribe-open-intel-deliveries"
	// messageTypeOpenIntelDeliveries informs about updated open intel deliveries for
	// a certain operation.
	messageTypeOpenIntelDeliveries wsutil.MessageType = "open-intel-deliveries"
	// messageTypeSubscribedOpenIntelDeliveries informs the client about subscribed
	// open intel deliveries.
	messageTypeSubscribedOpenIntelDeliveries wsutil.MessageType = "subscribed-open-intel-deliveries"
)

// messageTypeSubscribeOpenIntelDeliveries is the payload for messages with type
// messageTypeSubscribeOpenIntelDeliveries.
type messageSubscribeOpenIntelDeliveries struct {
	Operation uuid.UUID `json:"operation"`
}

// messageUnsubscribeOpenIntelDeliveries is the payload for messages with type
// messageTypeUnsubscribeOpenIntelDeliveries.
type messageUnsubscribeOpenIntelDeliveries struct {
	Operation uuid.UUID `json:"operation"`
}

// messageOpenIntelDeliveries is the payload for messages with type
// messageTypeOpenIntelDeliveries.
type messageOpenIntelDeliveries struct {
	Operation           uuid.UUID                        `json:"operation"`
	OpenIntelDeliveries []publicOpenIntelDeliverySummary `json:"open_intel_deliveries"`
}

// publicOpenIntelDeliverySummary is the public representation of
// store.OpenIntelDeliverySummary.
type publicOpenIntelDeliverySummary struct {
	Delivery publicActiveIntelDelivery `json:"delivery"`
	Intel    publicIntel               `json:"intel"`
}

// publicActiveIntelDelivery is the public representation of
// store.ActiveIntelDelivery.
type publicActiveIntelDelivery struct {
	ID    uuid.UUID    `json:"id"`
	Intel uuid.UUID    `json:"intel"`
	To    uuid.UUID    `json:"to"`
	Note  nulls.String `json:"note"`
}

// publicIntel is the public representation of store.Intel.
type publicIntel struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  uuid.UUID `json:"created_by"`
	Operation  uuid.UUID `json:"operation"`
	Importance int       `json:"importance"`
	IsValid    bool      `json:"is_valid"`
}

// publicOpenIntelDeliverySummaryFromStore converts
// store.OpenIntelDeliverySummary to publicOpenIntelDeliverySummary.
func publicOpenIntelDeliverySummaryFromStore(s store.OpenIntelDeliverySummary) publicOpenIntelDeliverySummary {
	return publicOpenIntelDeliverySummary{
		Delivery: publicActiveIntelDelivery{
			ID:    s.Delivery.ID,
			Intel: s.Delivery.Intel,
			To:    s.Delivery.To,
			Note:  s.Delivery.Note,
		},
		Intel: publicIntel{
			ID:         s.Intel.ID,
			CreatedAt:  s.Intel.CreatedAt,
			CreatedBy:  s.Intel.CreatedBy,
			Operation:  s.Intel.Operation,
			Importance: s.Intel.Importance,
			IsValid:    s.Intel.IsValid,
		},
	}
}

// messageSubscribedOpenIntelDeliveries is the message content for messages with
// type messageTypeSubscribedOpenIntelDeliveries.
type messageSubscribedOpenIntelDeliveries struct {
	Operations []uuid.UUID `json:"operations"`
}
