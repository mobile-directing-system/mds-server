package event

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"time"
)

// TypeOperationCreated is used when an operation was created.
const TypeOperationCreated Type = "operation-created"

// OperationCreated is the value for TypeOperationCreated.
type OperationCreated struct {
	// ID identifies the operation.
	ID uuid.UUID `json:"id"`
	// Title of the operation.
	Title string `json:"title"`
	// Optional description. We treat a non-existent description as empty string.
	Description string `json:"description"`
	// Start timestamp of the operation.
	Start time.Time `json:"start"`
	// End is the optional timestamp, when the operation has finished. If End is not
	// set or set to a moment in the past, the operation is considered finished.
	End nulls.Time `json:"end"`
	// IsArchived describes whether the operation was archived. This is used instead
	// of deleting the operation in order to avoid unintended data loss.
	IsArchived bool `json:"is_archived"`
}

// TypeOperationUpdated is used when an operation was updated.
const TypeOperationUpdated Type = "operation-updated"

// OperationUpdated is the value for TypeOperationUpdated.
type OperationUpdated struct {
	// ID identifies the operation.
	ID uuid.UUID `json:"id"`
	// Title of the operation.
	Title string `json:"title"`
	// Optional description. We treat a non-existent description as empty string.
	Description string `json:"description"`
	// Start timestamp of the operation.
	Start time.Time `json:"start"`
	// End is the optional timestamp, when the operation has finished. If End is not
	// set or set to a moment in the past, the operation is considered finished.
	End nulls.Time `json:"end"`
	// IsArchived describes whether the operation was archived. This is used instead
	// of deleting the operation in order to avoid unintended data loss.
	IsArchived bool `json:"is_archived"`
}

// TypeOperationMembersUpdated is used when the member list for an operation was
// updated.
const TypeOperationMembersUpdated Type = "operation-members-updated"

// OperationMembersUpdated is the value for TypeOperationMembersUpdated.
type OperationMembersUpdated struct {
	// Operation is id of the operation, which's members have been updated.
	Operation uuid.UUID `json:"operation"`
	// Members is the list of ids of users, that are member of the operation.
	Members []uuid.UUID `json:"members"`
}
