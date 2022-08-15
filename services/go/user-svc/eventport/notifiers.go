package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
	"golang.org/x/net/context"
)

// NotifyUserCreated creates an event.TypeUserCreated event.
func (port *Port) NotifyUserCreated(ctx context.Context, tx pgx.Tx, user store.UserWithPass) error {
	err := port.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.UsersTopic,
		Key:       user.ID.String(),
		EventType: event.TypeUserCreated,
		Value: event.UserCreated{
			ID:        user.ID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			IsAdmin:   user.IsAdmin,
			Pass:      user.Pass,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyUserUpdated creates an event.TypeUserUpdated event.
func (port *Port) NotifyUserUpdated(ctx context.Context, tx pgx.Tx, user store.User) error {
	err := port.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.UsersTopic,
		Key:       user.ID.String(),
		EventType: event.TypeUserUpdated,
		Value: event.UserUpdated{
			ID:        user.ID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			IsAdmin:   user.IsAdmin,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyUserPassUpdated creates an event.TypeUserPassUpdated event.
func (port *Port) NotifyUserPassUpdated(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newPass []byte) error {
	err := port.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.UsersTopic,
		Key:       userID.String(),
		EventType: event.TypeUserPassUpdated,
		Value: event.UserPassUpdated{
			User:    userID,
			NewPass: newPass,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}

// NotifyUserDeleted creates an event.TypeUserDeleted event.
func (port *Port) NotifyUserDeleted(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	err := port.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
		Topic:     event.UsersTopic,
		Key:       userID.String(),
		EventType: event.TypeUserDeleted,
		Value: event.UserDeleted{
			ID: userID,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}
