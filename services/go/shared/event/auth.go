package event

import "github.com/gofrs/uuid"

// TypeUserLoggedIn is used when a user logged in.
const TypeUserLoggedIn Type = "user-logged-in"

// UserLoggedIn is the value for TypeUserLoggedIn.
type UserLoggedIn struct {
	// User is the id of the user that logged in.
	User uuid.UUID
	// Username of the user that logged in.
	Username string
	// Host from http.Request.
	Host string
	// UserAgent from http.Request.
	UserAgent string
	// RemoteAddr from http.Request.
	RemoteAddr string
}

// TypeUserLoggedOut is used when a user logged out.
const TypeUserLoggedOut Type = "user-logged-out"

// UserLoggedOut is the value for TypeUserLoggedOut.
type UserLoggedOut struct {
	// User is the id of the user that logged out.
	User uuid.UUID
	// Username of the user that logged out.
	Username string
	// Host from http.Request.
	Host string
	// UserAgent from http.Request.
	UserAgent string
	// RemoteAddr from http.Request.
	RemoteAddr string
}
