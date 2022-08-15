package eventport

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// NotifyOperationCreated emits an event.TypeOperationCreated event.
func (p *Port) NotifyOperationCreated(ctx context.Context, tx pgx.Tx, operation store.Operation) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.OperationsTopic,
		Key:       operation.ID.String(),
		EventType: event.TypeOperationCreated,
		Value: event.OperationCreated{
			ID:          operation.ID,
			Title:       operation.Title,
			Description: operation.Description,
			Start:       operation.Start,
			End:         operation.End,
			IsArchived:  operation.IsArchived,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyOperationUpdated emits an event.TypeOperationUpdated event.
func (p *Port) NotifyOperationUpdated(ctx context.Context, tx pgx.Tx, operation store.Operation) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.OperationsTopic,
		Key:       operation.ID.String(),
		EventType: event.TypeOperationUpdated,
		Value: event.OperationUpdated{
			ID:          operation.ID,
			Title:       operation.Title,
			Description: operation.Description,
			Start:       operation.Start,
			End:         operation.End,
			IsArchived:  operation.IsArchived,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyOperationMembersUpdated emits an event.TypeOperationMembersUpdated.
func (p *Port) NotifyOperationMembersUpdated(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, members []uuid.UUID) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.OperationsTopic,
		Key:       operationID.String(),
		EventType: event.TypeOperationMembersUpdated,
		Value: event.OperationMembersUpdated{
			Operation: operationID,
			Members:   members,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}
