package endpoints

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
	"net/http"
	"net/url"
	"strconv"
)

// userFiltersFromQuery extracts store.UserFilters from the given url.Values.
func userFiltersFromQuery(q url.Values) (store.UserFilters, error) {
	var err error
	var filters store.UserFilters
	includeInactiveStr := q.Get("include_inactive")
	if includeInactiveStr != "" {
		filters.IncludeInactive, err = strconv.ParseBool(includeInactiveStr)
		if err != nil {
			return store.UserFilters{}, meh.NewBadInputErrFromErr(err, "parse include-inactive",
				meh.Details{"was": includeInactiveStr})
		}
	}
	return filters, nil
}

// createUserRequest contains all information for creating a user.
type createUserRequest struct {
	// Username for the user.
	Username string `json:"username"`
	// FirstName of the user.
	FirstName string `json:"first_name"`
	// LastName of the user.
	LastName string `json:"last_name"`
	// IsAdmin describes whether the user is an administrator.
	IsAdmin bool `json:"is_admin"`
	// Pass is the plaintext password.
	Pass string `json:"pass"`
}

// createUserResponse is a container for all information regarding a created
// user.
type createUserResponse struct {
	// ID identifies the user.
	ID uuid.UUID `json:"id"`
	// Username for the user.
	Username string `json:"username"`
	// FirstName of the user.
	FirstName string `json:"first_name"`
	// LastName of the user.
	LastName string `json:"last_name"`
	// IsAdmin describes whether the user is an administrator.
	IsAdmin bool `json:"is_admin"`
	// IsActive describes whether the user is active (not deleted).
	IsActive bool `json:"is_active"`
}

// handleCreateUserStore are the dependencies needed for handleCreateUser.
type handleCreateUserStore interface {
	// CreateUser creates the given store.UserWithPass.
	CreateUser(ctx context.Context, user store.UserWithPass) (store.UserWithPass, error)
}

// handleCreatedUser handles user creation.
func handleCreateUser(s handleCreateUserStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// First permission check for creating users.
		ok, err := auth.HasPermission(token, permission.CreateUser())
		if err != nil {
			return meh.Wrap(err, "has create-user permission", nil)
		}
		if !ok {
			return meh.NewForbiddenErr("no permission to create users", nil)
		}
		// Parse body.
		var userToCreate createUserRequest
		err = json.NewDecoder(c.Request.Body).Decode(&userToCreate)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "decode body", nil)
		}
		// Creating an admin user requires extra permission.
		if userToCreate.IsAdmin {
			ok, err = auth.HasPermission(token, permission.SetAdminUser())
			if err != nil {
				return meh.Wrap(err, "has set-admin permission", nil)
			}
			if !ok {
				return meh.NewForbiddenErr("no permission for creating admin user", nil)
			}
		}
		// Hash password.
		hashedPass, err := auth.HashPassword(userToCreate.Pass)
		if err != nil {
			return meh.Wrap(err, "hash pass", nil)
		}
		storeUserToCreate := store.UserWithPass{
			User: store.User{
				Username:  userToCreate.Username,
				FirstName: userToCreate.FirstName,
				LastName:  userToCreate.LastName,
				IsAdmin:   userToCreate.IsAdmin,
				IsActive:  true,
			},
			Pass: hashedPass,
		}
		// Validate.
		if ok, err := entityvalidation.ValidateInRequest(c, storeUserToCreate); err != nil {
			return meh.Wrap(err, "validate", meh.Details{"user": storeUserToCreate})
		} else if !ok {
			// Handled.
			return nil
		}
		// Create.
		createdUser, err := s.CreateUser(c.Request.Context(), storeUserToCreate)
		if err != nil {
			return meh.Wrap(err, "create user", meh.Details{"user_to_create": storeUserToCreate.User})
		}
		// Respond with created user.
		c.JSON(http.StatusCreated, createUserResponse{
			ID:        createdUser.ID,
			Username:  createdUser.Username,
			FirstName: createdUser.FirstName,
			LastName:  createdUser.LastName,
			IsAdmin:   createdUser.IsAdmin,
			IsActive:  createdUser.IsActive,
		})
		return nil
	}
}

// updateUserRequest is a container for all information regarding a user to be
// updated.
type updateUserRequest struct {
	// ID identifies the user.
	ID uuid.UUID `json:"id"`
	// Username for the user.
	Username string `json:"username"`
	// FirstName of the user.
	FirstName string `json:"first_name"`
	// LastName of the user.
	LastName string `json:"last_name"`
	// IsAdmin describes whether the user is an administrator.
	IsAdmin bool `json:"is_admin"`
	// IsActive describes whether the user is active (not deleted).
	IsActive bool `json:"is_active"`
}

// handleUpdateUserByIDStore are the dependencies needed for
// handleUpdateUserByID.
type handleUpdateUserByIDStore interface {
	// UpdateUser updates the given store.User and makes sure that changing the
	// admin state is only allowed if the flag is set as well as active-state
	// changing.
	UpdateUser(ctx context.Context, user store.User, allowAdminChange bool, allowActiveStateChange bool) error
}

// handleUpdateUserByID updates a user.
func handleUpdateUserByID(s handleUpdateUserByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Parse body.
		var user updateUserRequest
		err := json.NewDecoder(c.Request.Body).Decode(&user)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "decode body", nil)
		}
		userIDFromPath := c.Param("userID")
		if user.ID.String() != userIDFromPath {
			return meh.NewBadInputErr("id mismatch", meh.Details{
				"from_path": userIDFromPath,
				"from_body": user.ID.String(),
			})
		}
		// Permission check if not updating self.
		if token.UserID != user.ID {
			ok, err := auth.HasPermission(token, permission.UpdateUser())
			if err != nil {
				return meh.Wrap(err, "check permission for updating user", nil)
			}
			if !ok {
				return meh.NewForbiddenErr("missing permission for updating other user", nil)
			}
		}
		// Check if allowed to change admin-state.
		allowAdminChange, err := auth.HasPermission(token, permission.SetAdminUser())
		if err != nil {
			return meh.Wrap(err, "check permission for allowing admin change", nil)
		}
		// Check if allowed to change active-state.
		allowActiveStateChange, err := auth.HasPermission(token, permission.SetUserActiveState())
		if err != nil {
			return meh.Wrap(err, "check permission for allowing active-state change", nil)
		}
		updatedUser := store.User{
			ID:        user.ID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			IsAdmin:   user.IsAdmin,
			IsActive:  user.IsActive,
		}
		// Validate.
		if ok, err := entityvalidation.ValidateInRequest(c, updatedUser); err != nil {
			return meh.Wrap(err, "validate", meh.Details{"user": updatedUser})
		} else if !ok {
			// Handled.
			return nil
		}
		// Update.
		err = s.UpdateUser(c.Request.Context(), updatedUser, allowAdminChange, allowActiveStateChange)
		if err != nil {
			return meh.Wrap(err, "update user", meh.Details{
				"updated_user":       updatedUser,
				"allow_admin_change": allowAdminChange,
			})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// updateUserPassByUserIDRequest is the request body for
// handleUpdateUserPassByUserID.
type updateUserPassByUserIDRequest struct {
	// UserID is the id of the user to update the password for.
	UserID uuid.UUID `json:"user_id"`
	// NewPass is the new password in plaintext.
	NewPass string `json:"new_pass"`
}

// handleUpdateUserPassByUserIDStore are the dependencies needed for
// handleUpdateUserPassByUserID.
type handleUpdateUserPassByUserIDStore interface {
	// UpdateUserPassByUserID updates the password for the user with the given id.
	UpdateUserPassByUserID(ctx context.Context, userID uuid.UUID, newPass []byte) error
}

// handleUpdateUserPassByUserID updates the password for a user.
func handleUpdateUserPassByUserID(s handleUpdateUserPassByUserIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Parse body.
		var reqBody updateUserPassByUserIDRequest
		err := json.NewDecoder(c.Request.Body).Decode(&reqBody)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "decode body", nil)
		}
		userIDFromPath := c.Param("userID")
		if reqBody.UserID.String() != userIDFromPath {
			return meh.NewBadInputErr("id mismatch", meh.Details{
				"from_path": userIDFromPath,
				"from_body": reqBody.UserID.String(),
			})
		}
		// If foreign, check permission.
		if reqBody.UserID != token.UserID {
			ok, err := auth.HasPermission(token, permission.UpdateUserPass())
			if err != nil {
				return meh.Wrap(err, "check permission for updating foreign user pass", nil)
			}
			if !ok {
				return meh.NewForbiddenErr("no permission to update foreign user pass", nil)
			}
		}
		// Hash password.
		hashedPass, err := auth.HashPassword(reqBody.NewPass)
		if err != nil {
			return meh.Wrap(err, "hash password", nil)
		}
		// Update.
		err = s.UpdateUserPassByUserID(c.Request.Context(), reqBody.UserID, hashedPass)
		if err != nil {
			return meh.Wrap(err, "update user pass", meh.Details{"user_id": reqBody.UserID})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// handleDeleteUserByIDStore are the dependencies needed for
// handleDeleteUserByID.
type handleDeleteUserByIDStore interface {
	// SetUserInactiveByID sets the user with the given id to inactive.
	SetUserInactiveByID(ctx context.Context, userID uuid.UUID) error
}

// handleDeleteUserByID sets a user inactive. Setting self inactive is still
// only allowed with the required permission.
func handleDeleteUserByID(s handleDeleteUserByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract id.
		userIDToSetInactiveStr := c.Param("userID")
		userIDToSetInactive, err := uuid.FromString(userIDToSetInactiveStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse user id", meh.Details{"was": userIDToSetInactiveStr})
		}
		// Check permission.
		ok, err := auth.HasPermission(token, permission.SetUserActiveState())
		if err != nil {
			return meh.Wrap(err, "check permission for setting active-state for users", nil)
		}
		if !ok {
			return meh.NewForbiddenErr("no permission to set active-state for users", nil)
		}
		// Delete.
		err = s.SetUserInactiveByID(c.Request.Context(), userIDToSetInactive)
		if err != nil {
			return meh.Wrap(err, "set user inactive by id", meh.Details{"user_id": userIDToSetInactive})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// getUserResponse is a container for all information regarding a user.
type getUserResponse struct {
	// ID identifies the user.
	ID uuid.UUID `json:"id"`
	// Username for the user.
	Username string `json:"username"`
	// FirstName of the user.
	FirstName string `json:"first_name"`
	// LastName of the user.
	LastName string `json:"last_name"`
	// IsAdmin describes whether the user is an administrator.
	IsAdmin bool `json:"is_admin"`
	// IsActive describes whether the user is active (not deleted).
	IsActive bool `json:"is_active"`
}

// handleGetUserByIDStore are the dependencies needed for handleGetUserByID.
type handleGetUserByIDStore interface {
	// UserByID retrieves a store.User by its id.
	UserByID(ctx context.Context, userID uuid.UUID) (store.User, error)
}

// handleGetUserByID retrieves a user by its id.
func handleGetUserByID(s handleGetUserByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract user id to view.
		userIDToViewStr := c.Param("userID")
		userIDToView, err := uuid.FromString(userIDToViewStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse user id", meh.Details{"was": userIDToViewStr})
		}
		// Check permission.
		if token.UserID != userIDToView {
			ok, err := auth.HasPermission(token, permission.ViewUser())
			if err != nil {
				return meh.Wrap(err, "check permission", nil)
			}
			if !ok {
				return meh.NewForbiddenErr("no permission to view other users", nil)
			}
		}
		// Retrieve.
		retrievedUser, err := s.UserByID(c.Request.Context(), userIDToView)
		if err != nil {
			return meh.Wrap(err, "user by id", meh.Details{"user_id": userIDToView})
		}
		c.JSON(http.StatusOK, getUserResponse{
			ID:        retrievedUser.ID,
			Username:  retrievedUser.Username,
			FirstName: retrievedUser.FirstName,
			LastName:  retrievedUser.LastName,
			IsAdmin:   retrievedUser.IsAdmin,
			IsActive:  retrievedUser.IsActive,
		})
		return nil
	}
}

// getUsersResponseUser is a container for all information regarding users in a
// list.
type getUsersResponseUser struct {
	// ID identifies the user.
	ID uuid.UUID `json:"id"`
	// Username for the user.
	Username string `json:"username"`
	// FirstName of the user.
	FirstName string `json:"first_name"`
	// LastName of the user.
	LastName string `json:"last_name"`
	// IsAdmin describes whether the user is an administrator.
	IsAdmin bool `json:"is_admin"`
	// IsActive describes whether the user is active (not deleted).
	IsActive bool `json:"is_active"`
}

// handleGetUsersStore are the dependencies needed for handleGetUsers.
type handleGetUsersStore interface {
	// Users retrieves a paginated store.User list.
	Users(ctx context.Context, filters store.UserFilters, params pagination.Params) (pagination.Paginated[store.User], error)
}

// handleGetUsers retrieves a user list.
func handleGetUsers(s handleGetUsersStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Check permission.
		ok, err := auth.HasPermission(token, permission.ViewUser())
		if err != nil {
			return meh.Wrap(err, "check permission", nil)
		}
		if !ok {
			return meh.NewForbiddenErr("no permission to view users", nil)
		}
		// Extract pagination params.
		params, err := pagination.ParamsFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "params from request", nil)
		}
		// Extract filters.
		filters, err := userFiltersFromQuery(c.Request.URL.Query())
		if err != nil {
			return meh.Wrap(err, "user filters from query", nil)
		}
		// Retrieve.
		retrievedUsers, err := s.Users(c.Request.Context(), filters, params)
		if err != nil {
			return meh.Wrap(err, "retrieve users", meh.Details{"params": params})
		}
		c.JSON(http.StatusOK, pagination.MapPaginated(retrievedUsers, func(from store.User) getUsersResponseUser {
			return getUsersResponseUser{
				ID:        from.ID,
				Username:  from.Username,
				FirstName: from.FirstName,
				LastName:  from.LastName,
				IsAdmin:   from.IsAdmin,
				IsActive:  from.IsActive,
			}
		}))
		return nil
	}
}

// handleSearchUsersStore are the dependencies needed for handleSearchUsers.
type handleSearchUsersStore interface {
	SearchUsers(ctx context.Context, filters store.UserFilters, searchParams search.Params) (search.Result[store.User], error)
}

// handleSearchUsers searches for users.
func handleSearchUsers(s handleSearchUsersStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Check permissions.
		err := auth.AssurePermission(token, permission.ViewUser())
		if err != nil {
			return meh.Wrap(err, "check permission", nil)
		}
		// Extract search params.
		searchParams, err := search.ParamsFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "search params from request", nil)
		}
		// Extract filter params.
		filters, err := userFiltersFromQuery(c.Request.URL.Query())
		if err != nil {
			return meh.Wrap(err, "user filters from query", nil)
		}
		// Search.
		result, err := s.SearchUsers(c.Request.Context(), filters, searchParams)
		if err != nil {
			return meh.Wrap(err, "search users", meh.Details{"search_params": searchParams})
		}
		c.JSON(http.StatusOK, search.MapResult(result, func(from store.User) getUsersResponseUser {
			return getUsersResponseUser{
				ID:        from.ID,
				Username:  from.Username,
				FirstName: from.FirstName,
				LastName:  from.LastName,
				IsAdmin:   from.IsAdmin,
				IsActive:  from.IsActive,
			}
		}))
		return nil
	}
}

// handleRebuildUserSearchStore are the dependencies needed for
// handleRebuildUserSearch.
type handleRebuildUserSearchStore interface {
	RebuildUserSearch(ctx context.Context)
}

// handleRebuildUserSearch rebuilds the user search.
func handleRebuildUserSearch(s handleRebuildUserSearchStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		err := auth.AssurePermission(token, permission.RebuildSearchIndex())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		go s.RebuildUserSearch(context.Background())
		c.Status(http.StatusOK)
		return nil
	}
}
