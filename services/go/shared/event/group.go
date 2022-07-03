package event

import "github.com/gofrs/uuid"

// TypeGroupCreated is used when a group was created.
const TypeGroupCreated Type = "group-created"

// GroupCreated is the value for TypeGroupCreated.
type GroupCreated struct {
	// ID identifies the group.
	ID uuid.UUID `json:"id"`
	// Title of the group.
	Title string `json:"title"`
	// Description of the group.
	Description string `json:"description"`
	// Operation is the id an optional operation.
	Operation uuid.NullUUID `json:"operation"`
	// Members of the group represented by user ids.
	Members []uuid.UUID `json:"members"`
}

// TypeGroupUpdated is used when a group was updated.
const TypeGroupUpdated Type = "group-updated"

// GroupUpdated is the value for TypeGroupUpdated.
type GroupUpdated struct {
	// ID identifies the group.
	ID uuid.UUID `json:"id"`
	// Title of the group.
	Title string `json:"title"`
	// Description of the group.
	Description string `json:"description"`
	// Operation is the id an optional operation.
	Operation uuid.NullUUID `json:"operation"`
	// Members of the group represented by user ids.
	Members []uuid.UUID `json:"members"`
}

// TypeGroupDeleted is used when a group was deleted.
const TypeGroupDeleted Type = "group-deleted"

// GroupDeleted is the value for TypeGroupDeleted.
type GroupDeleted struct {
	// ID identifies the group.
	ID uuid.UUID `json:"id"`
}
