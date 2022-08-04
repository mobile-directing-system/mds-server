package endpoints

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
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

// publicUser is the public representation of store.User.
type publicUser struct {
	// ID identifies the user.
	ID uuid.UUID `json:"id"`
	// Username for logging in.
	Username string `json:"username"`
	// FirstName of the user.
	FirstName string `json:"first_name"`
	// LastName of the user.
	LastName string `json:"last_name"`
}

// publicUserFromStore converts a store.User to publicUser.
func publicUserFromStore(s store.User) publicUser {
	return publicUser{
		ID:        s.ID,
		Username:  s.Username,
		FirstName: s.FirstName,
		LastName:  s.LastName,
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
		operationID, err := uuid.FromString(operationIDStr)
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
		storeCreate := storeOperationFromPublic(create)
		// Validate.
		if ok, err := entityvalidation.ValidateInRequest(c, storeCreate); err != nil {
			return meh.Wrap(err, "validate", meh.Details{"operation": storeCreate})
		} else if !ok {
			// Handled.
			return nil
		}
		// Create.
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
		idFromQuery, err := uuid.FromString(idFromQueryStr)
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
		storeUpdate := storeOperationFromPublic(update)
		// Validate.
		if ok, err := entityvalidation.ValidateInRequest(c, storeUpdate); err != nil {
			return meh.Wrap(err, "validate", meh.Details{"group": storeUpdate})
		} else if !ok {
			// Handled.
			return nil
		}
		// Update.
		err = s.UpdateOperation(c.Request.Context(), storeUpdate)
		if err != nil {
			return meh.Wrap(err, "update operation", meh.Details{"update": storeUpdate})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// handleGetOperationMembersByOperationStore are the dependencies needed for
// handleGetOperationMembersByOperation.
type handleGetOperationMembersByOperationStore interface {
	OperationMembersByOperation(ctx context.Context, operationID uuid.UUID,
		paginationParams pagination.Params) (pagination.Paginated[store.User], error)
}

// handleGetOperationMembersByOperation retrieves a paginated member list for
// the given operation.
func handleGetOperationMembersByOperation(s handleGetOperationMembersByOperationStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permission.
		err := auth.AssurePermission(token, permission.ViewOperationMembers())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		// Extract id from params.
		idFromQueryStr := c.Param("operationID")
		idFromQuery, err := uuid.FromString(idFromQueryStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse operation id from query", meh.Details{"str": idFromQueryStr})
		}
		// Extract pagination params.
		paginationParams, err := pagination.ParamsFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "pagination params from request", nil)
		}
		// Retrieve.
		users, err := s.OperationMembersByOperation(c.Request.Context(), idFromQuery, paginationParams)
		if err != nil {
			return meh.Wrap(err, "operation members by operation", meh.Details{
				"operation_id":      idFromQuery,
				"pagination_params": paginationParams,
			})
		}
		c.JSON(http.StatusOK, pagination.MapPaginated(users, publicUserFromStore))
		return nil
	}
}

// handleUpdateOperationMembersByOperationStore are the dependencies needed for
// handleUpdateOperationMembersByOperation.
type handleUpdateOperationMembersByOperationStore interface {
	UpdateOperationMembersByOperation(ctx context.Context, operationID uuid.UUID, members []uuid.UUID) error
}

// handleUpdateOperationMembersByOperation retrieves a paginated member list for
// the given operation.
func handleUpdateOperationMembersByOperation(s handleUpdateOperationMembersByOperationStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permission.
		err := auth.AssurePermission(token, permission.UpdateOperationMembers())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		// Extract id from params.
		idFromQueryStr := c.Param("operationID")
		idFromQuery, err := uuid.FromString(idFromQueryStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse operation id from query", meh.Details{"str": idFromQueryStr})
		}
		// Parse body.
		var members []uuid.UUID
		err = json.NewDecoder(c.Request.Body).Decode(&members)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "decode body", nil)
		}
		// Update.
		err = s.UpdateOperationMembersByOperation(c.Request.Context(), idFromQuery, members)
		if err != nil {
			return meh.Wrap(err, "update operation members by operation", meh.Details{
				"operation_id": idFromQuery,
				"members":      members,
			})
		}
		c.Status(http.StatusOK)
		return nil
	}
}
