package event

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"time"
)

// TypeIntelCreated for created intel.
const TypeIntelCreated Type = "intel-created"

// IntelType of Intel. Also describes the content.
type IntelType string

// IntelCreated for TypeIntelCreated.
type IntelCreated struct {
	// ID identifies the intel.
	ID uuid.UUID `json:"id"`
	// CreatedAt is the timestamp, the intel was created.
	CreatedAt time.Time `json:"created_at"`
	// CreatedBy is the id of the user, who created the intent.
	CreatedBy uuid.UUID `json:"created_by"`
	// Operation is the id of the associated operation.
	Operation uuid.UUID `json:"operation"`
	// Type of the intel.
	Type IntelType `json:"type"`
	// Content is the actual information.
	Content json.RawMessage `json:"content"`
	// SearchText for better searching. Used with higher priority than Content.
	SearchText nulls.String `json:"search_text"`
	// Importance of the intel. Used for example for prioritizing delivery methods.
	Importance int `json:"importance"`
	// IsValid describes whether the intel is still valid or marked as invalid
	// (equals deletion).
	IsValid bool `json:"is_valid"`
	// Assignments for the intel.
	Assignments []IntelAssignment `json:"assignments"`
}

// IntelAssignment is used in IntelCreated.
type IntelAssignment struct {
	// ID identifies the assignment.
	ID uuid.UUID `json:"id"`
	// To is the id of the target address book entry.
	To uuid.UUID `json:"to"`
}

// TypeIntelInvalidated for intel, that has been invalidated.
const TypeIntelInvalidated Type = "intel-invalidated"

// IntelInvalidated for TypeIntelInvalidated.
type IntelInvalidated struct {
	// ID identifies the intel.
	ID uuid.UUID `json:"id"`
	// By is the id of the user that invalidated the intel.
	By uuid.UUID `json:"by"`
}
