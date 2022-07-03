package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
)

// NotifyPermissionsUpdated emits an event.PermissionsUpdated event.
func (p *Port) NotifyPermissionsUpdated(userID uuid.UUID, permissions []store.Permission) error {
	permissionsOut := make([]permission.Permission, 0, len(permissions))
	for _, p := range permissions {
		permissionsOut = append(permissionsOut, permission.Permission(p))
	}
	err := kafkautil.WriteMessages(p.writer, kafkautil.Message{
		Topic:     event.PermissionsTopic,
		Key:       userID.String(),
		EventType: event.TypePermissionsUpdated,
		Value: event.PermissionsUpdated{
			User:        userID,
			Permissions: permissionsOut,
		},
	})
	if err != nil {
		return meh.Wrap(err, "write kafka messages", nil)
	}
	return nil
}
