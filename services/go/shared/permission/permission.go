package permission

import "github.com/lefinal/nulls"

// Name identifies a permission.
type Name string

// Permission that is granted to a user.
type Permission struct {
	// Name is the permission identifier.
	Name Name `json:"name"`
	// Options contains additional options for the permission.
	Options nulls.JSONRawMessage `json:"options,omitempty"`
}

// TODO: VALIDATION
