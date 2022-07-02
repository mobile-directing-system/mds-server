package endpoints

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"net/http"
)

// publicPermission is the public representation of store.Permission.
type publicPermission struct {
	Name    string               `json:"name"`
	Options nulls.JSONRawMessage `json:"options"`
}

// publicPermissionFromPermission converts a store.Permission to
// publicPermission.
func publicPermissionFromPermission(p store.Permission) publicPermission {
	return publicPermission{
		Name:    string(p.Name),
		Options: p.Options,
	}
}

// permissionFromPublic converts a publicPermission to permission.Permission.
func permissionFromPublic(public publicPermission) store.Permission {
	return store.Permission{
		Name:    permission.Name(public.Name),
		Options: public.Options,
	}
}

// handleGetPermissionsByUserStore are the dependencies needed for
// handleGetPermissionsByUser.
type handleGetPermissionsByUserStore interface {
	PermissionsByUser(ctx context.Context, userID uuid.UUID) ([]store.Permission, error)
}

// handleGetPermissionsByUser retrieves the permissions for a user by its id.
func handleGetPermissionsByUser(s handleGetPermissionsByUserStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract user id.
		userIDToViewStr := c.Param("userID")
		userIDToView, err := uuid.FromString(userIDToViewStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse user id", meh.Details{"was": userIDToViewStr})
		}
		// Check permission.
		if token.UserID != userIDToView {
			ok, err := auth.HasPermission(token, permission.ViewPermissions())
			if err != nil {
				return meh.Wrap(err, "check permission", nil)
			}
			if !ok {
				return meh.NewForbiddenErr("no permission to view permissions of other users", nil)
			}
		}
		// Retrieve.
		retrievedPermissions, err := s.PermissionsByUser(c.Request.Context(), userIDToView)
		if err != nil {
			return meh.Wrap(err, "permissions by user", meh.Details{"user_id": userIDToView})
		}
		publicPermissions := make([]publicPermission, 0, len(retrievedPermissions))
		for _, p := range retrievedPermissions {
			publicPermissions = append(publicPermissions, publicPermissionFromPermission(p))
		}
		c.JSON(http.StatusOK, publicPermissions)
		return nil
	}
}

// handleUpdatePermissionsByUserStore are the dependencies needed for
// handleUpdatePermissionsByUser.
type handleUpdatePermissionsByUserStore interface {
	UpdatePermissionsByUser(ctx context.Context, userID uuid.UUID, permissions []store.Permission) error
}

// handleUpdatePermissionsByUser updates the permissions for a user identified
// by its id.
func handleUpdatePermissionsByUser(s handleUpdatePermissionsByUserStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract user id.
		userIDToUpdateStr := c.Param("userID")
		userIDToUpdate, err := uuid.FromString(userIDToUpdateStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse user id", meh.Details{"was": userIDToUpdateStr})
		}
		// Parse body.
		var publicUpdatedPermissions []publicPermission
		err = json.NewDecoder(c.Request.Body).Decode(&publicUpdatedPermissions)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "decode body", nil)
		}
		updatedPermissions := make([]store.Permission, 0, len(publicUpdatedPermissions))
		for _, p := range publicUpdatedPermissions {
			updatedPermissions = append(updatedPermissions, permissionFromPublic(p))
		}
		// Check permission.
		ok, err := auth.HasPermission(token, permission.UpdatePermissions())
		if err != nil {
			return meh.Wrap(err, "check permission", nil)
		}
		if !ok {
			return meh.NewForbiddenErr("no permission to update permissions of users", nil)
		}
		// Update.
		err = s.UpdatePermissionsByUser(c.Request.Context(), userIDToUpdate, updatedPermissions)
		if err != nil {
			return meh.Wrap(err, "update permissions by user", meh.Details{
				"user_id":             userIDToUpdate,
				"updated_permissions": updatedPermissions,
			})
		}
		c.Status(http.StatusOK)
		return nil
	}
}
