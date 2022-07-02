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
	ID uuid.UUID
	// Title of the operation.
	Title string
	// Optional description. We treat a non-existent description as empty string.
	Description string
	// Start timestamp of the operation.
	Start time.Time
	// End is the optional timestamp, when the operation has finished. If End is not
	// set or set to a moment in the past, the operation is considered finished.
	End nulls.Time
	// IsArchived describes whether the operation was archived. This is used instead
	// of deleting the operation in order to avoid unintended data loss.
	IsArchived bool
}

// TypeOperationUpdated is used when an operation was updated.
const TypeOperationUpdated Type = "operation-updated"

// OperationUpdated is the value for TypeOperationUpdated.
type OperationUpdated struct {
	// ID identifies the operation.
	ID uuid.UUID
	// Title of the operation.
	Title string
	// Optional description. We treat a non-existent description as empty string.
	Description string
	// Start timestamp of the operation.
	Start time.Time
	// End is the optional timestamp, when the operation has finished. If End is not
	// set or set to a moment in the past, the operation is considered finished.
	End nulls.Time
	// IsArchived describes whether the operation was archived. This is used instead
	// of deleting the operation in order to avoid unintended data loss.
	IsArchived bool
}
