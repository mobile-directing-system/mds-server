package eventport

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"time"
)

// NotifyRadioDeliveryReadyForPickup emeits an
// event.TypeRadioDeliveryReadyForPickup event.
func (p *Port) NotifyRadioDeliveryReadyForPickup(ctx context.Context, tx pgx.Tx, intelDeliveryAttempt store.AcceptedIntelDeliveryAttempt,
	radioDeliveryNote string) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		Key:       intelDeliveryAttempt.ID.String(),
		EventType: event.TypeRadioDeliveryReadyForPickup,
		Value: event.RadioDeliveryReadyForPickup{
			Attempt:                intelDeliveryAttempt.ID,
			Intel:                  intelDeliveryAttempt.Intel,
			IntelOperation:         intelDeliveryAttempt.IntelOperation,
			IntelImportance:        intelDeliveryAttempt.IntelImportance,
			AttemptAssignedTo:      intelDeliveryAttempt.AssignedTo,
			AttemptAssignedToLabel: intelDeliveryAttempt.AssignedToLabel,
			Delivery:               intelDeliveryAttempt.Delivery,
			Channel:                intelDeliveryAttempt.Channel,
			Note:                   radioDeliveryNote,
			AttemptAcceptedAt:      intelDeliveryAttempt.AcceptedAt,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyRadioDeliveryPickedUp emits an event.NotifyRadioDeliveryPickedUp event.
func (p *Port) NotifyRadioDeliveryPickedUp(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, pickedUpBy uuid.UUID,
	pickedUpAt time.Time) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		Key:       attemptID.String(),
		EventType: event.TypeRadioDeliveryPickedUp,
		Value: event.RadioDeliveryPickedUp{
			Attempt:    attemptID,
			PickedUpBy: pickedUpBy,
			PickedUpAt: pickedUpAt,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyRadioDeliveryReleased emits an event.NotifyRadioDeliveryReleased event.
func (p *Port) NotifyRadioDeliveryReleased(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, releasedAt time.Time) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		Key:       attemptID.String(),
		EventType: event.TypeRadioDeliveryReleased,
		Value: event.RadioDeliveryReleased{
			Attempt:    attemptID,
			ReleasedAt: releasedAt,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyRadioDeliveryFinished emits an event.NotifyRadioDeliveryFinished event.
func (p *Port) NotifyRadioDeliveryFinished(ctx context.Context, tx pgx.Tx, radioDelivery store.RadioDelivery) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		Key:       radioDelivery.Attempt.String(),
		EventType: event.TypeRadioDeliveryFinished,
		Value: event.RadioDeliveryFinished{
			Attempt:    radioDelivery.Attempt,
			PickedUpBy: radioDelivery.PickedUpBy,
			PickedUpAt: radioDelivery.PickedUpAt,
			Success:    radioDelivery.Success.Bool,
			FinishedAt: radioDelivery.SuccessTS,
			Note:       radioDelivery.Note,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}
