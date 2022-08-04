package endpoints

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"net/http"
	"strconv"
)

// publicAddressBookEntryDetailed is the public representation of
// store.AddressBookEntryDetailed.
type publicAddressBookEntryDetailed struct {
	publicAddressBookEntry
	UserDetails nulls.JSONNullable[publicUser] `json:"user_details"`
}

// publicAddressBookEntryDetailedFromStore converts
// store.AddressBookEntryDetailed to publicAddressBookEntryDetailed.
func publicAddressBookEntryDetailedFromStore(s store.AddressBookEntryDetailed) publicAddressBookEntryDetailed {
	e := publicAddressBookEntryDetailed{
		publicAddressBookEntry: publicAddressBookEntryFromStore(s.AddressBookEntry),
	}
	if s.UserDetails.Valid {
		e.UserDetails = nulls.NewJSONNullable(publicUserFromStore(s.UserDetails.V))
	}
	return e
}

// publicUser is the public representation of store.User.
type publicUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

// publicUserFromStore converts store.User to publicUser.
func publicUserFromStore(s store.User) publicUser {
	return publicUser{
		ID:        s.ID,
		Username:  s.Username,
		FirstName: s.FirstName,
		LastName:  s.LastName,
	}
}

// publicAddressBookEntry is the public representation of
// store.AddressBookEntry.
type publicAddressBookEntry struct {
	ID          uuid.UUID     `json:"id"`
	Label       string        `json:"label"`
	Description string        `json:"description"`
	Operation   uuid.NullUUID `json:"operation"`
	User        uuid.NullUUID `json:"user"`
}

// publicAddressBookEntryFromStore converts a store.AddressBookEntry to
// publicAddressBookEntry.
func publicAddressBookEntryFromStore(s store.AddressBookEntry) publicAddressBookEntry {
	return publicAddressBookEntry{
		ID:          s.ID,
		Label:       s.Label,
		Description: s.Description,
		Operation:   s.Operation,
		User:        s.User,
	}
}

// storeAddressBookEntryFromPublic converts publicAddressBookEntry to
// store.AddressBookEntry.
func storeAddressBookEntryFromPublic(p publicAddressBookEntry) store.AddressBookEntry {
	return store.AddressBookEntry{
		ID:          p.ID,
		Label:       p.Label,
		Description: p.Description,
		Operation:   p.Operation,
		User:        p.User,
	}
}

// handleGetAddressBookEntryByIDStore are the dependencies needed for
// handleGetAddressBookEntryByID.
type handleGetAddressBookEntryByIDStore interface {
	AddressBookEntryByID(ctx context.Context, entryID uuid.UUID, visibleBy uuid.NullUUID) (store.AddressBookEntryDetailed, error)
}

// handleGetAddressBookEntryByID retrieves an address book entry by its id.
func handleGetAddressBookEntryByID(s handleGetAddressBookEntryByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		// Permission check.
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authorized", nil)
		}
		var visibleBy uuid.NullUUID
		viewAnyGranted, err := auth.HasPermission(token, permission.ViewAnyAddressBookEntry())
		if err != nil {
			return meh.Wrap(err, "permission check", nil)
		}
		if !viewAnyGranted {
			visibleBy = nulls.NewUUID(token.UserID)
		}
		// Extract entry id.
		entryIDStr := c.Param("entryID")
		entryID, err := uuid.FromString(entryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse entry id", meh.Details{"was": entryIDStr})
		}
		// Retrieve.
		entry, err := s.AddressBookEntryByID(c.Request.Context(), entryID, visibleBy)
		if err != nil {
			return meh.Wrap(err, "entry from store", meh.Details{"entry_id": entryID})
		}
		c.JSON(http.StatusOK, publicAddressBookEntryDetailedFromStore(entry))
		return nil
	}
}

// handleGetAllAddressBookEntriesStore are the dependencies needed for
// handleGetAllAddressBookEntries.
type handleGetAllAddressBookEntriesStore interface {
	AddressBookEntries(ctx context.Context, filters store.AddressBookEntryFilters,
		paginationParams pagination.Params) (pagination.Paginated[store.AddressBookEntryDetailed], error)
}

// handleGetAllAddressBookEntries retrieves a paginated address book entry list.
// Without the permission.ViewAnyAddressBookEntry, only global ones and entries,
// associated with the client or participating operations is allowed.
func handleGetAllAddressBookEntries(s handleGetAllAddressBookEntriesStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authorized", nil)
		}
		// Parse filter params.
		var filter store.AddressBookEntryFilters
		if byUserStr := c.Query("by_user"); byUserStr != "" {
			byUser, err := uuid.FromString(byUserStr)
			if err != nil {
				return meh.NewBadInputErrFromErr(err, "by-user from string", meh.Details{"was": byUserStr})
			}
			filter.ByUser = nulls.NewUUID(byUser)
		}
		if forOperationStr := c.Query("for_operation"); forOperationStr != "" {
			forOperation, err := uuid.FromString(forOperationStr)
			if err != nil {
				return meh.NewBadInputErrFromErr(err, "for-operation from string", meh.Details{"was": forOperationStr})
			}
			filter.ForOperation = nulls.NewUUID(forOperation)
		}
		if excludeGlobalStr := c.Query("exclude_global"); excludeGlobalStr != "" {
			excludeGlobal, err := strconv.ParseBool(excludeGlobalStr)
			if err != nil {
				return meh.NewBadInputErrFromErr(err, "parse bool", meh.Details{"was": excludeGlobalStr})
			}
			filter.ExcludeGlobal = excludeGlobal
		}
		if visibleByStr := c.Query("visible_by"); visibleByStr != "" {
			visibleBy, err := uuid.FromString(visibleByStr)
			if err != nil {
				return meh.NewBadInputErrFromErr(err, "visible-by from string", meh.Details{"was": visibleByStr})
			}
			filter.VisibleBy = nulls.NewUUID(visibleBy)
		}
		// Parse pagination params.
		paginationParams, err := pagination.ParamsFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "pagination params from request", nil)
		}
		// Check permisions.
		viewAnyGranted, err := auth.HasPermission(token, permission.ViewAnyAddressBookEntry())
		if err != nil {
			return meh.Wrap(err, "check permission", nil)
		}
		if !viewAnyGranted {
			filter.VisibleBy = nulls.NewUUID(token.UserID)
		}
		// Retrieve.
		sEntries, err := s.AddressBookEntries(c.Request.Context(), filter, paginationParams)
		if err != nil {
			return meh.Wrap(err, "entries from store", nil)
		}
		c.JSON(http.StatusOK, pagination.MapPaginated(sEntries, publicAddressBookEntryDetailedFromStore))
		return nil
	}
}

// handleCreateAddressBookEntryStore are the dependencies needed for
// handleCreateAddressBookEntry.
type handleCreateAddressBookEntryStore interface {
	CreateAddressBookEntry(ctx context.Context, entry store.AddressBookEntry) (store.AddressBookEntryDetailed, error)
}

// handleCreateAddressBookEntry creates an address book entry.
func handleCreateAddressBookEntry(s handleCreateAddressBookEntryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Parse body.
		var pCreate publicAddressBookEntry
		err := json.NewDecoder(c.Request.Body).Decode(&pCreate)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "decode body", nil)
		}
		sCreate := storeAddressBookEntryFromPublic(pCreate)
		// If trying to create global entry or associate with user, not being requesting
		// client, check permissions.
		if !sCreate.User.Valid || sCreate.User.UUID != token.UserID {
			createAnyGranted, err := auth.HasPermission(token, permission.CreateAnyAddressBookEntry())
			if err != nil {
				return meh.Wrap(err, "permission check", nil)
			}
			if !createAnyGranted {
				return meh.NewForbiddenErr("missing permission", nil)
			}
		}
		// Create.
		created, err := s.CreateAddressBookEntry(c.Request.Context(), sCreate)
		if err != nil {
			return meh.Wrap(err, "create address book entry", meh.Details{"create": sCreate})
		}
		c.JSON(http.StatusCreated, publicAddressBookEntryDetailedFromStore(created))
		return nil
	}
}

// handleUpdateAddressBookEntryStore are the dependencies needed for
// handleUpdateAddressBookEntry.
type handleUpdateAddressBookEntryStore interface {
	UpdateAddressBookEntry(ctx context.Context, update store.AddressBookEntry, limitToUser uuid.NullUUID) error
}

// handleUpdateAddressBookEntry updates the given address book entry.
func handleUpdateAddressBookEntry(s handleUpdateAddressBookEntryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract entry id.
		entryIDStr := c.Param("entryID")
		entryID, err := uuid.FromString(entryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse entry id", meh.Details{"was": entryIDStr})
		}
		// Parse body.
		var pUpdate publicAddressBookEntry
		err = json.NewDecoder(c.Request.Body).Decode(&pUpdate)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "decode body", nil)
		}
		if pUpdate.ID != entryID {
			return meh.NewBadInputErr("id mismatch", meh.Details{
				"id_from_param": entryID,
				"id_from_body":  pUpdate.ID,
			})
		}
		sUpdate := storeAddressBookEntryFromPublic(pUpdate)
		// Check permissions if trying to update global entry or one associated with a user, not being the
		// requesting client.
		limitToUser := nulls.NewUUID(token.UserID)
		updateAnyGranted, err := auth.HasPermission(token, permission.UpdateAnyAddressBookEntry())
		if err != nil {
			return meh.Wrap(err, "permission check", nil)
		}
		if updateAnyGranted {
			limitToUser = uuid.NullUUID{}
		}
		// Update.
		err = s.UpdateAddressBookEntry(c.Request.Context(), sUpdate, limitToUser)
		if err != nil {
			return meh.Wrap(err, "update address book entry", meh.Details{"update": sUpdate})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// handleDeleteAddressBookEntryByIDStore are the dependencies needed for
// handleDeleteAddressBookEntryByID.
type handleDeleteAddressBookEntryByIDStore interface {
	DeleteAddressBookEntryByID(ctx context.Context, entryID uuid.UUID, limitToUser uuid.NullUUID) error
}

// handleDeleteAddressBookEntryByID deletes the address book entry with the
// given id.
func handleDeleteAddressBookEntryByID(s handleDeleteAddressBookEntryByIDStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		// Extract entry id.
		entryIDStr := c.Param("entryID")
		entryID, err := uuid.FromString(entryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse entry id", meh.Details{"was": entryIDStr})
		}
		// Only allow deleting global entries or ones associated with other users than
		// the requester, if permission is granted.
		limitToUser := nulls.NewUUID(token.UserID)
		deleteAnyGranted, err := auth.HasPermission(token, permission.DeleteAnyAddressBookEntry())
		if err != nil {
			return meh.Wrap(err, "permission check", nil)
		}
		if deleteAnyGranted {
			limitToUser = uuid.NullUUID{}
		}
		// Delete.
		err = s.DeleteAddressBookEntryByID(c.Request.Context(), entryID, limitToUser)
		if err != nil {
			return meh.Wrap(err, "delete address book entry", meh.Details{
				"entry_id":      entryID,
				"limit_to_user": limitToUser,
			})
		}
		c.Status(http.StatusOK)
		return nil
	}
}
