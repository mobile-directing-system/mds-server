package auth

import "github.com/mobile-directing-system/mds-server/services/go/shared/permission"

// UserDetails holds details regarding a user performing a request. This is
// needed in order to allow proper authentication.
type UserDetails struct {
	// Username of the user that performed the request.
	Username string `json:"username"`
	// IsAuthenticated describes whether the user is currently logged in.
	IsAuthenticated bool `json:"is_authenticated"`
	// IsAdmin describes whether the user is an admin.
	IsAdmin bool `json:"is_admin"`
	// Permissions the user was granted.
	Permissions []permission.Permission `json:"permissions"`
}
