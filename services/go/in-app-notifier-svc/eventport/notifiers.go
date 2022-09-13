package eventport

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"time"
)

// NotifyIntelDeliveryNotificationSent emits an
// event.TypeAddressBookEntryChannelsUpdated event.
func (p *Port) NotifyIntelDeliveryNotificationSent(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, sentTS time.Time) error {
	message := kafkautil.OutboundMessage{
		Topic:     event.InAppNotificationsTopic,
		Key:       attemptID.String(),
		EventType: event.TypeInAppNotificationForIntelSent,
		Value: event.InAppNotificationForIntelSent{
			Attempt: attemptID,
			SentAt:  sentTS,
		},
	}
	err := p.writer.AddOutboxMessages(ctx, tx, message)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": message})
	}
	return nil
}

// NotifyIntelDeliveryNotificationPending emits an
// event.TypeInAppNotificationForIntelPending event.
func (p *Port) NotifyIntelDeliveryNotificationPending(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, since time.Time) error {
	message := kafkautil.OutboundMessage{
		Topic:     event.InAppNotificationsTopic,
		Key:       attemptID.String(),
		EventType: event.TypeInAppNotificationForIntelPending,
		Value: event.InAppNotificationForIntelPending{
			Attempt: attemptID,
			Since:   since,
		},
	}
	err := p.writer.AddOutboxMessages(ctx, tx, message)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": message})
	}
	return nil
}
