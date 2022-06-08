package eventport

import (
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
)

// NotifyUserCreated creates an event.TypeUserCreated event.
func (port *Port) NotifyUserCreated(user store.User) error {
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
