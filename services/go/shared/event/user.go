package event

// TypeUserCreated is used when a user was created.
const TypeUserCreated = "user-created"

// UserCreated is the value for TypeUserCreated.
type UserCreated struct {
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
