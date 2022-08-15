package eventport

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// NotifyGroupCreated creates an event.TypeGroupCreated event.
func (p *Port) NotifyGroupCreated(ctx context.Context, tx pgx.Tx, group store.Group) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.GroupsTopic,
		Key:       group.ID.String(),
		EventType: event.TypeGroupCreated,
		Value: event.GroupCreated{
			ID:          group.ID,
			Title:       group.Title,
			Description: group.Description,
			Operation:   group.Operation,
			Members:     group.Members,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyGroupUpdated creates an event.TypeGroupUpdated event.
func (p *Port) NotifyGroupUpdated(ctx context.Context, tx pgx.Tx, group store.Group) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.GroupsTopic,
		Key:       group.ID.String(),
		EventType: event.TypeGroupUpdated,
		Value: event.GroupUpdated{
			ID:          group.ID,
			Title:       group.Title,
			Description: group.Description,
			Operation:   group.Operation,
			Members:     group.Members,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyGroupDeleted creates an event.TypeGroupDeleted event.
func (p *Port) NotifyGroupDeleted(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.GroupsTopic,
		Key:       groupID.String(),
		EventType: event.TypeGroupDeleted,
		Value: event.GroupDeleted{
			ID: groupID,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}
