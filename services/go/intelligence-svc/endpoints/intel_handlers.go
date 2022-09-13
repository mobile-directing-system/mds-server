package endpoints

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"net/http"
	"time"
)

// publicIntelType is the public representation of store.IntelType.
type publicIntelType string

const (
	intelTypePlainTextMessage publicIntelType = "plaintext-message"
)

// publicIntelTypeFromStore converts store.IntelType to publicIntelType. If no
// mapping was found, a meh.ErrInternal is returned.
func publicIntelTypeFromStore(s store.IntelType) (publicIntelType, error) {
	switch s {
	case store.IntelTypePlaintextMessage:
		return intelTypePlainTextMessage, nil
	default:
		return "", meh.NewInternalErr("unsupported type", meh.Details{"type": s})
	}
}

// storeIntelTypeFromPublic converts publicIntelType to store.IntelType. If no
// mapping was found, a meh.ErrBadInput is returned.
func storeIntelTypeFromPublic(p publicIntelType) (store.IntelType, error) {
	switch p {
	case intelTypePlainTextMessage:
		return store.IntelTypePlaintextMessage, nil
	default:
		return "", meh.NewBadInputErr("unsupported type", meh.Details{"type": p})
	}
}

// publicCreateIntel is the public representation of store.CreateIntel.
type publicCreateIntel struct {
	Operation   uuid.UUID               `json:"operation"`
	Type        publicIntelType         `json:"type"`
	Content     json.RawMessage         `json:"content"`
	Importance  int                     `json:"importance"`
	Assignments []publicIntelAssignment `json:"assignments"`
}

// publicIntelAssignment is the public representation of store.IntelAssignment.
type publicIntelAssignment struct {
	ID uuid.UUID `json:"id"`
	To uuid.UUID `json:"to"`
}

// storeCreateIntelFromPublic maps publicCreateIntel to store.CreateIntel.
func storeCreateIntelFromPublic(createdBy uuid.UUID, p publicCreateIntel) (store.CreateIntel, error) {
	intelType, err := storeIntelTypeFromPublic(p.Type)
	if err != nil {
		return store.CreateIntel{}, meh.Wrap(err, "store intel type from public", meh.Details{"type": p.Type})
	}
	return store.CreateIntel{
		CreatedBy:   createdBy,
		Operation:   p.Operation,
		Type:        intelType,
		Content:     p.Content,
		Importance:  p.Importance,
		Assignments: storeIntelAssignmentsFromPublic(uuid.UUID{}, p.Assignments),
	}, nil
}

// storeIntelAssignmentsFromPublic converts a publicIntelAssignment list to
// store.IntelAssignment list.
func storeIntelAssignmentsFromPublic(intelID uuid.UUID, ps []publicIntelAssignment) []store.IntelAssignment {
	s := make([]store.IntelAssignment, 0, len(ps))
	for _, p := range ps {
		s = append(s, store.IntelAssignment{
			ID:    p.ID,
			Intel: intelID,
			To:    p.To,
		})
	}
	return s
}

// publicIntel is the public representation of store.Intel.
type publicIntel struct {
	ID          uuid.UUID               `json:"id"`
	CreatedAt   time.Time               `json:"created_at"`
	CreatedBy   uuid.UUID               `json:"created_by"`
	Operation   uuid.UUID               `json:"operation"`
	Type        publicIntelType         `json:"type"`
	Content     json.RawMessage         `json:"content"`
	SearchText  nulls.String            `json:"search_text"`
	Importance  int                     `json:"importance"`
	IsValid     bool                    `json:"is_valid"`
	Assignments []publicIntelAssignment `json:"assignments"`
}

// publicIntelFromStore converts a store.Intel list to publicIntel list.
func publicIntelFromStore(s store.Intel) (publicIntel, error) {
	intelType, err := publicIntelTypeFromStore(s.Type)
	if err != nil {
		return publicIntel{}, meh.Wrap(err, "public intel type from store", meh.Details{"type": s.Type})
	}
	return publicIntel{
		ID:          s.ID,
		CreatedAt:   s.CreatedAt,
		CreatedBy:   s.CreatedBy,
		Operation:   s.Operation,
		Type:        intelType,
		Content:     s.Content,
		SearchText:  s.SearchText,
		Importance:  s.Importance,
		IsValid:     s.IsValid,
		Assignments: publicIntelAssignmentsFromStore(s.Assignments),
	}, nil
}

// publicIntelAssignmentsFromStore converts a store.IntelAssignment list to
// publicIntelAssignment list.
func publicIntelAssignmentsFromStore(s []store.IntelAssignment) []publicIntelAssignment {
	assignments := make([]publicIntelAssignment, 0, len(s))
	for _, a := range s {
		assignments = append(assignments, publicIntelAssignment{
			ID: a.ID,
			To: a.To,
		})
	}
	return assignments
}

// handleCreateIntelStore are the dependencies needed for handleCreateIntel.
type handleCreateIntelStore interface {
	CreateIntel(ctx context.Context, create store.CreateIntel) (store.Intel, error)
}

// handleCreateIntel creates the given intel.
func handleCreateIntel(s handleCreateIntelStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		err := auth.AssurePermission(token, permission.CreateIntel())
		if err != nil {
			return meh.Wrap(err, "assure permission", nil)
		}
		// Parse body.
		var pCreateIntel publicCreateIntel
		err = json.NewDecoder(c.Request.Body).Decode(&pCreateIntel)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse body", nil)
		}
		sCreateIntel, err := storeCreateIntelFromPublic(token.UserID, pCreateIntel)
		if err != nil {
			return meh.Wrap(err, "store create intel type from public", meh.Details{"public": pCreateIntel})
		}
		// Validate.
		if ok, err := entityvalidation.ValidateInRequest(c, sCreateIntel); err != nil {
			return meh.Wrap(err, "validate in request", meh.Details{"intel": sCreateIntel})
		} else if !ok {
			// Handled.
			return nil
		}
		// Create.
		sCreated, err := s.CreateIntel(c.Request.Context(), sCreateIntel)
		if err != nil {
			return meh.Wrap(err, "create intel", nil)
		}
		pCreated, err := publicIntelFromStore(sCreated)
		if err != nil {
			c.Status(http.StatusOK)
			return httpendpoints.NoResponse(meh.Wrap(err, "public intel from store", nil))
		}
		c.JSON(http.StatusOK, pCreated)
		return nil
	}
}

// handleInvalidateIntelByIDStore are the dependencies needed for
// handleInvalidateIntelByID.
type handleInvalidateIntelByIDStore interface {
	InvalidateIntelByID(ctx context.Context, intelID uuid.UUID, by uuid.UUID) error
}

// handleInvalidateIntelByID invalidates the given intel.
func handleInvalidateIntelByID(s handleInvalidateIntelByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		err := auth.AssurePermission(token, permission.InvalidateIntel())
		if err != nil {
			return meh.Wrap(err, "assure permission", nil)
		}
		// Extract intel id.
		intelIDStr := c.Param("intelID")
		intelID, err := uuid.FromString(intelIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse intel id", meh.Details{"was": intelIDStr})
		}
		// Invalidate.
		err = s.InvalidateIntelByID(c.Request.Context(), intelID, token.UserID)
		if err != nil {
			return meh.Wrap(err, "invalidate intel", meh.Details{"intel_id": intelID})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// handleGetIntelByIDStore are the dependencies needed for handleGetIntelByID.
type handleGetIntelByIDStore interface {
	IntelByID(ctx context.Context, intelID uuid.UUID, limitToAssignedUser uuid.NullUUID) (store.Intel, error)
}

// handleGetIntelByID retrieves the intel with the given id.
func handleGetIntelByID(s handleGetIntelByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract intel id.
		intelIDStr := c.Param("intelID")
		intelID, err := uuid.FromString(intelIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse intel id", meh.Details{"was": intelIDStr})
		}
		// Check view-any-permission.
		limitToAssignedUser := nulls.NewUUID(token.UserID)
		viewAnyGranted, err := auth.HasPermission(token, permission.ViewAnyIntel())
		if err != nil {
			return meh.Wrap(err, "check view-any-permission", nil)
		}
		if viewAnyGranted {
			limitToAssignedUser = uuid.NullUUID{}
		}
		// Retrieve.
		sIntel, err := s.IntelByID(c.Request.Context(), intelID, limitToAssignedUser)
		if err != nil {
			return meh.Wrap(err, "intel by id", meh.Details{
				"intel_id":               intelID,
				"limit_to_assigned_user": limitToAssignedUser,
			})
		}
		pIntel, err := publicIntelFromStore(sIntel)
		if err != nil {
			return meh.Wrap(err, "public intel from store", meh.Details{"store_intel": sIntel})
		}
		c.JSON(http.StatusOK, pIntel)
		return nil
	}
}
