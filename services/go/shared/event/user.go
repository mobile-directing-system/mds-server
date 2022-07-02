package event

import "github.com/gofrs/uuid"

// TypeUserCreated is used when a user was created.
const TypeUserCreated Type = "user-created"

// UserCreated is the value for TypeUserCreated.
type UserCreated struct {
	// ID that identifies the created user.
	ID uuid.UUID `json:"id"`
	// Username of the created user.
	Username string `json:"username"`
	// FirstName of the created user.
	FirstName string `json:"first_name"`
	// LastName of the created user.
	LastName string `json:"last_name"`
	// IsAdmin describes whether the user is an admin.
	IsAdmin bool `json:"is_admin"`
	// Pass is the hashed password for the user.
	Pass []byte `json:"pass"`
}

// TypeUserUpdated is used when a user was updated.
const TypeUserUpdated Type = "user-updated"

// UserUpdated is the value for TypeUserUpdated.
type UserUpdated struct {
	// ID that identifies the updated user.
	ID uuid.UUID `json:"id"`
	// Username of the updated user.
	Username string `json:"username"`
	// FirstName of the updated user.
	FirstName string `json:"first_name"`
	// LastName of the updated user.
	LastName string `json:"last_name"`
	// IsAdmin describes whether the user is an admin.
	IsAdmin bool `json:"is_admin"`
}

// TypeUserPassUpdated is used when a user updates its password.
const TypeUserPassUpdated Type = "user-password-updated"

// UserPassUpdated is the value for TypeUserPassUpdated.
type UserPassUpdated struct {
	// User is the id of the user that updated its password.
	User uuid.UUID `json:"user"`
	// NewPass is the new password.
	NewPass []byte `json:"new_pass"`
}

// TypeUserDeleted is used when a user was deleted.
const TypeUserDeleted Type = "user-deleted"

// UserDeleted is the value for TypeUserDeleted.
type UserDeleted struct {
	// ID of the user that was deleted.
	ID uuid.UUID `json:"id"`
}
