package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"golang.org/x/net/context"
)

// NotifyPermissionsUpdated emits an event.PermissionsUpdated event.
func (p *Port) NotifyPermissionsUpdated(ctx context.Context, tx pgx.Tx, userID uuid.UUID, permissions []store.Permission) error {
	permissionsOut := make([]permission.Permission, 0, len(permissions))
	for _, p := range permissions {
		permissionsOut = append(permissionsOut, permission.Permission(p))
	}
	err := p.writer.AddOutboxMessages(ctx, tx, kafkautil.OutboundMessage{
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
