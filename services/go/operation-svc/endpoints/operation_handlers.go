package endpoints

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"net/http"
	"time"
)

// publicOperation is the public representation of store.Operation.
type publicOperation struct {
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

// publicOperationFromStore converts a store.Operation to publicOperation.
func publicOperationFromStore(s store.Operation) publicOperation {
	return publicOperation{
		ID:          s.ID,
		Title:       s.Title,
		Description: s.Description,
		Start:       s.Start,
		End:         s.End,
		IsArchived:  s.IsArchived,
	}
}

// publicOperationFromStore converts a publicOperation to store.Operation.
func storeOperationFromPublic(public publicOperation) store.Operation {
	return store.Operation{
		ID:          public.ID,
		Title:       public.Title,
		Description: public.Description,
		Start:       public.Start,
		End:         public.End,
		IsArchived:  public.IsArchived,
	}
}

// handleGetOperationsStore are the dependencies needed for handleGetOperations.
type handleGetOperationsStore interface {
	Operations(ctx context.Context, params pagination.Params) (pagination.Paginated[store.Operation], error)
}

// handleGetOperations retrieves a list of registered operations.
func handleGetOperations(s handleGetOperationsStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permissions.
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		ok, err := auth.HasPermission(token, permission.ViewAnyOperation())
		if err != nil {
			return meh.Wrap(err, "check permission", nil)
		}
		if !ok {
			return meh.NewForbiddenErr("no permission to view all operations", nil)
		}
		// Params.
		params, err := pagination.ParamsFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "params from request", nil)
		}
		// Retrieve.
		operations, err := s.Operations(c.Request.Context(), params)
		if err != nil {
			return meh.Wrap(err, "retrieve operations from store", meh.Details{"params": params})
		}
		public := pagination.MapPaginated(operations, publicOperationFromStore)
		c.JSON(http.StatusOK, public)
		return nil
	}
}

// handleGetOperationByIDStore are the dependencies needed for store.Operation.
type handleGetOperationByIDStore interface {
	OperationByID(ctx context.Context, operationID uuid.UUID) (store.Operation, error)
}

// handleGetOperationByID retrieves an operation by its id.
func handleGetOperationByID(s handleGetOperationByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract operation id.
		operationIDStr := c.Param("operationID")
		operationID, err := uuid.Parse(operationIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse operation id", meh.Details{"str": operationIDStr})
		}
		// Retrieve.
		operation, err := s.OperationByID(c.Request.Context(), operationID)
		if err != nil {
			return meh.Wrap(err, "operation by id from store", meh.Details{"operation_id": operationID})
		}
		c.JSON(http.StatusOK, publicOperationFromStore(operation))
		return nil
	}
}

// handleCreateOperationStore are the dependencies needed for
// handleCreateOperation.
type handleCreateOperationStore interface {
	CreateOperation(ctx context.Context, create store.Operation) (store.Operation, error)
}

// handleCreateOperation allows creating an operation.
func handleCreateOperation(s handleCreateOperationStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permissions.
		err := auth.AssurePermission(token, permission.CreateOperation())
		if err != nil {
			return meh.Wrap(err, "assure permission", nil)
		}
		// Parse body.
		var create publicOperation
		err = json.NewDecoder(c.Request.Body).Decode(&create)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "decode body", nil)
		}
		// Create.
		storeCreate := storeOperationFromPublic(create)
		created, err := s.CreateOperation(c.Request.Context(), storeCreate)
		if err != nil {
			return meh.Wrap(err, "create operation", meh.Details{"create": storeCreate})
		}
		c.JSON(http.StatusOK, publicOperationFromStore(created))
		return nil
	}
}

// handleUpdateOperationStore are the dependencies needed for
// handleUpdateOperation.
type handleUpdateOperationStore interface {
	UpdateOperation(ctx context.Context, update store.Operation) error
}

// handleUpdateOperation updates the operation with the given id.
func handleUpdateOperation(s handleUpdateOperationStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permissions.
		err := auth.AssurePermission(token, permission.UpdateOperation())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		// Extract id from params.
		idFromQueryStr := c.Param("operationID")
		idFromQuery, err := uuid.Parse(idFromQueryStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse operation id from query", meh.Details{"str": idFromQueryStr})
		}
		// Parse body.
		var update publicOperation
		err = json.NewDecoder(c.Request.Body).Decode(&update)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "decode body", nil)
		}
		if update.ID != idFromQuery {
			return meh.NewBadInputErr("id mismatch", meh.Details{
				"id_from_query_params": idFromQuery,
				"id_from_body":         update.ID,
			})
		}
		// Update.
		storeUpdate := storeOperationFromPublic(update)
		err = s.UpdateOperation(c.Request.Context(), storeUpdate)
		if err != nil {
			return meh.Wrap(err, "update operation", meh.Details{"update": storeUpdate})
		}
		c.Status(http.StatusOK)
		return nil
	}
}
