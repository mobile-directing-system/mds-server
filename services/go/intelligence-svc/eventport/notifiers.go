package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"golang.org/x/net/context"
)

// intelTypeFromStore converts store.IntelType to event.IntelType.
func intelTypeFromStore(s store.IntelType) (event.IntelType, error) {
	switch s {
	case store.IntelTypePlainTextMessage:
		return event.IntelTypePlainTextMessage, nil
	default:
		return "", meh.NewInternalErr("unsupported intel type", meh.Details{"intel_type": s})
	}
}

// intelAssignmentsFromStore converts a store.IntelAssignment list to
// event.IntelAssignment list.
func intelAssignmentsFromStore(s []store.IntelAssignment) []event.IntelAssignment {
	assignments := make([]event.IntelAssignment, 0, len(s))
	for _, assignment := range s {
		assignments = append(assignments, event.IntelAssignment{
			ID: assignment.ID,
			To: assignment.To,
		})
	}
	return assignments
}

// NotifyIntelCreated notifies about created intel.
func (p *Port) NotifyIntelCreated(ctx context.Context, tx pgx.Tx, created store.Intel) error {
	intelType, err := intelTypeFromStore(created.Type)
	if err != nil {
		return meh.Wrap(err, "intel type from store", nil)
	}
	intelCreatedMessage := kafkautil.OutboundMessage{
		Topic:     event.IntelTopic,
		Key:       created.ID.String(),
		EventType: event.TypeIntelCreated,
		Value: event.IntelCreated{
			ID:          created.ID,
			CreatedAt:   created.CreatedAt,
			CreatedBy:   created.CreatedBy,
			Operation:   created.Operation,
			Type:        intelType,
			Content:     created.Content,
			SearchText:  created.SearchText,
			Importance:  created.Importance,
			IsValid:     created.IsValid,
			Assignments: intelAssignmentsFromStore(created.Assignments),
		},
	}
	err = p.writer.AddOutboxMessages(ctx, tx, intelCreatedMessage)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": intelCreatedMessage})
	}
	return nil
}

// NotifyIntelInvalidated notifies about invalidated intel.
func (p *Port) NotifyIntelInvalidated(ctx context.Context, tx pgx.Tx, intelID uuid.UUID, by uuid.UUID) error {
	intelInvalidatedMessage := kafkautil.OutboundMessage{
		Topic:     event.IntelTopic,
		Key:       intelID.String(),
		EventType: event.TypeIntelInvalidated,
		Value: event.IntelInvalidated{
			ID: intelID,
			By: by,
		},
	}
	err := p.writer.AddOutboxMessages(ctx, tx, intelInvalidatedMessage)
	if err != nil {
		return meh.Wrap(err, "add outbox messages", meh.Details{"message": intelInvalidatedMessage})
	}
	return nil
}
