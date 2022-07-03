package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// NotifyUserLoggedIn notifies that a user has logged in via an
// event.TypeUserLoggedIn event.
func (p *Port) NotifyUserLoggedIn(userID uuid.UUID, username string, requestMetadata controller.AuthRequestMetadata) error {
	err := kafkautil.WriteMessages(p.writer, kafkautil.Message{
		Topic:     event.AuthTopic,
		Key:       userID.String(),
		EventType: event.TypeUserLoggedIn,
		Value: event.UserLoggedIn{
			User:       userID,
			Username:   username,
			Host:       requestMetadata.Host,
			UserAgent:  requestMetadata.UserAgent,
			RemoteAddr: requestMetadata.RemoteAddr,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka message", nil)
	}
	return nil
}

// NotifyUserLoggedOut notifies that a user has logged out via an
// event.TypeUserLoggedOut event.
func (p *Port) NotifyUserLoggedOut(userID uuid.UUID, username string, requestMetadata controller.AuthRequestMetadata) error {
	err := kafkautil.WriteMessages(p.writer, kafkautil.Message{
		Topic:     event.AuthTopic,
		Key:       userID.String(),
		EventType: event.TypeUserLoggedOut,
		Value: event.UserLoggedOut{
			User:       userID,
			Username:   username,
			Host:       requestMetadata.Host,
			UserAgent:  requestMetadata.UserAgent,
			RemoteAddr: requestMetadata.RemoteAddr,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka message", nil)
	}
	return nil
}
