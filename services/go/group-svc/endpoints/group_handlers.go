package endpoints

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"net/http"
	"strconv"
)

// publicGroup is the public representation of store.Group.
type publicGroup struct {
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

// publicGroupFromStore converts a store.Group to publicGroup.
func publicGroupFromStore(s store.Group) publicGroup {
	return publicGroup{
		ID:          s.ID,
		Title:       s.Title,
		Description: s.Description,
		Operation:   s.Operation,
		Members:     s.Members,
	}
}

// storeGroupFromPublic converts a publicGroup to store.Group.
func storeGroupFromPublic(public publicGroup) store.Group {
	return store.Group{
		ID:          public.ID,
		Title:       public.Title,
		Description: public.Description,
		Operation:   public.Operation,
		Members:     public.Members,
	}
}

// handleGetGroupsStore are the dependencies needed for handleGetGroups.
type handleGetGroupsStore interface {
	Groups(ctx context.Context, filters store.GroupFilters, params pagination.Params) (pagination.Paginated[store.Group], error)
}

// handleGetGroups retrieves a paginated group list.
func handleGetGroups(s handleGetGroupsStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permissions.
		err := auth.AssurePermission(token, permission.ViewGroup())
		if err != nil {
			return meh.Wrap(err, "assure permission", nil)
		}
		// Extract group filters and pagination params.
		paginationParams, err := pagination.ParamsFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "pagination params from request", nil)
		}
		groupFilters, err := groupFiltersFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "group filters from request", nil)
		}
		// Retrieve.
		groups, err := s.Groups(c.Request.Context(), groupFilters, paginationParams)
		if err != nil {
			return meh.Wrap(err, "retrieve groups", meh.Details{
				"group_filters":     groupFilters,
				"pagination_params": paginationParams,
			})
		}
		public := pagination.MapPaginated(groups, publicGroupFromStore)
		c.JSON(http.StatusOK, public)
		return nil
	}
}

// groupFiltersFromRequest extracts store.GroupFilters from the given gin
// request.
func groupFiltersFromRequest(c *gin.Context) (store.GroupFilters, error) {
	var groupFilters store.GroupFilters
	// For operation.
	forOperationStr := c.Query("for_operation")
	if forOperationStr != "" {
		forOperation, err := uuid.FromString(forOperationStr)
		if err != nil {
			return store.GroupFilters{}, meh.NewBadInputErrFromErr(err, "parse for-operation uuid",
				meh.Details{"was": forOperationStr})
		}
		groupFilters.ForOperation = nulls.NewUUID(forOperation)
	}
	// Exclude global.
	excludeGlobalStr := c.Query("exclude_global")
	if excludeGlobalStr != "" {
		excludeGlobal, err := strconv.ParseBool(excludeGlobalStr)
		if err != nil {
			return store.GroupFilters{}, meh.NewBadInputErrFromErr(err, "parse exclude-global",
				meh.Details{"was": excludeGlobalStr})
		}
		groupFilters.ExcludeGlobal = excludeGlobal
	}
	// By user.
	byUserStr := c.Query("by_user")
	if byUserStr != "" {
		byUser, err := uuid.FromString(byUserStr)
		if err != nil {
			return store.GroupFilters{}, meh.NewBadInputErrFromErr(err, "parse by-user",
				meh.Details{"was": byUserStr})
		}
		groupFilters.ByUser = nulls.NewUUID(byUser)
	}
	return groupFilters, nil
}

// handleGetGroupByIDStore are the dependencies needed for handleGetGroupByID.
type handleGetGroupByIDStore interface {
	GroupByID(ctx context.Context, groupID uuid.UUID) (store.Group, error)
}

// handleGetGroupByID retrieves a group by its id.
func handleGetGroupByID(s handleGetGroupByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permissions.
		err := auth.AssurePermission(token, permission.ViewGroup())
		if err != nil {
			return meh.Wrap(err, "assure permission", nil)
		}
		// Extract group id.
		groupIDStr := c.Param("groupID")
		groupID, err := uuid.FromString(groupIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse group id", meh.Details{"was": groupIDStr})
		}
		// Retrieve.
		group, err := s.GroupByID(c.Request.Context(), groupID)
		if err != nil {
			return meh.Wrap(err, "retrieve group", meh.Details{"group_id": groupID})
		}
		c.JSON(http.StatusOK, publicGroupFromStore(group))
		return nil
	}
}

// handleCreateGroupStore are the dependencies needed for handleCreateGroup.
type handleCreateGroupStore interface {
	CreateGroup(ctx context.Context, create store.Group) (store.Group, error)
}

// handleCreateGroup creates a group.
func handleCreateGroup(s handleCreateGroupStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permissions.
		err := auth.AssurePermission(token, permission.CreateGroup())
		if err != nil {
			return meh.Wrap(err, "assure permission", nil)
		}
		// Parse body.
		var toCreatePublic publicGroup
		err = json.NewDecoder(c.Request.Body).Decode(&toCreatePublic)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse body", nil)
		}
		// Create.
		toCreate := storeGroupFromPublic(toCreatePublic)
		created, err := s.CreateGroup(c.Request.Context(), toCreate)
		if err != nil {
			return meh.Wrap(err, "create group", meh.Details{"create": toCreate})
		}
		c.JSON(http.StatusOK, publicGroupFromStore(created))
		return nil
	}
}

// handleUpdateGroupStore are the dependencies needed for handleUpdateGroup.
type handleUpdateGroupStore interface {
	UpdateGroup(ctx context.Context, update store.Group) error
}

// handleUpdateGroup updates a group.
func handleUpdateGroup(s handleUpdateGroupStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permissions.
		err := auth.AssurePermission(token, permission.UpdateGroup())
		if err != nil {
			return meh.Wrap(err, "assure permission", nil)
		}
		// Extract group id.
		groupIDStr := c.Param("groupID")
		groupID, err := uuid.FromString(groupIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse group id", meh.Details{"was": groupIDStr})
		}
		// Parse body.
		var toUpdatePublic publicGroup
		err = json.NewDecoder(c.Request.Body).Decode(&toUpdatePublic)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse body", nil)
		}
		if groupID != toUpdatePublic.ID {
			return meh.NewBadInputErrFromErr(err, "id mismatch", meh.Details{
				"id_from_public": groupID,
				"id_from_body":   toUpdatePublic.ID,
			})
		}
		// Update.
		toUpdate := storeGroupFromPublic(toUpdatePublic)
		err = s.UpdateGroup(c.Request.Context(), toUpdate)
		if err != nil {
			return meh.Wrap(err, "update group", meh.Details{"update": toUpdate})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// handleDeleteGroupByIDStore are the dependencies needed for
// handleDeleteGroupByID.
type handleDeleteGroupByIDStore interface {
	DeleteGroupByID(ctx context.Context, groupID uuid.UUID) error
}

// handleDeleteGroupByID deletes a group.
func handleDeleteGroupByID(s handleDeleteGroupByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permissions.
		err := auth.AssurePermission(token, permission.DeleteGroup())
		if err != nil {
			return meh.Wrap(err, "assure permission", nil)
		}
		// Extract group id.
		groupIDStr := c.Param("groupID")
		groupID, err := uuid.FromString(groupIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse group id", meh.Details{"was": groupIDStr})
		}
		// Delete.
		err = s.DeleteGroupByID(c.Request.Context(), groupID)
		if err != nil {
			return meh.Wrap(err, "delete group", meh.Details{"group_id": groupID})
		}
		c.Status(http.StatusOK)
		return nil
	}
}
