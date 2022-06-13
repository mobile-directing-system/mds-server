package eventport

import (
	"github.com/google/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
)

// NotifyUserCreated creates an event.TypeUserCreated event.
func (port *Port) NotifyUserCreated(user store.UserWithPass) error {
	err := kafkautil.WriteMessages(port.kafkaWriter, kafkautil.Message{
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
func (port *Port) NotifyUserUpdated(user store.User) error {
	err := kafkautil.WriteMessages(port.kafkaWriter, kafkautil.Message{
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
func (port *Port) NotifyUserPassUpdated(userID uuid.UUID, newPass []byte) error {
	err := kafkautil.WriteMessages(port.kafkaWriter, kafkautil.Message{
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
func (port *Port) NotifyUserDeleted(userID uuid.UUID) error {
	err := kafkautil.WriteMessages(port.kafkaWriter, kafkautil.Message{
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
