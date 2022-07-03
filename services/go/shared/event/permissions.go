package event

import (
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
)

// TypePermissionsUpdated is used when permissions for a certain user have
// changed.
const TypePermissionsUpdated Type = "permissions-updated"

// PermissionsUpdated is the value for TypePermissionsUpdated.
type PermissionsUpdated struct {
	// User that identifies the user which permissions have changed.
	User uuid.UUID `json:"user"`
	// Permissions are the new updated permissions for the user.
	Permissions []permission.Permission `json:"permissions"`
}
