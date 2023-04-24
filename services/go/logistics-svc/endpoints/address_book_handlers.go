package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"net/http"
	"net/url"
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
	IsActive  bool      `json:"is_active"`
}

// publicUserFromStore converts store.User to publicUser.
func publicUserFromStore(s store.User) publicUser {
	return publicUser{
		ID:        s.ID,
		Username:  s.Username,
		FirstName: s.FirstName,
		LastName:  s.LastName,
		IsActive:  s.IsActive,
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

// addressBookEntryFiltersFromQuery parses store.AddressBookEntryFilters from
// the given url.Values.
func addressBookEntryFiltersFromQuery(q url.Values) (store.AddressBookEntryFilters, error) {
	var filters store.AddressBookEntryFilters
	if byUserStr := q.Get("by_user"); byUserStr != "" {
		byUser, err := uuid.FromString(byUserStr)
		if err != nil {
			return store.AddressBookEntryFilters{}, meh.NewBadInputErrFromErr(err, "by-user from string", meh.Details{"was": byUserStr})
		}
		filters.ByUser = nulls.NewUUID(byUser)
	}
	if forOperationStr := q.Get("for_operation"); forOperationStr != "" {
		forOperation, err := uuid.FromString(forOperationStr)
		if err != nil {
			return store.AddressBookEntryFilters{}, meh.NewBadInputErrFromErr(err, "for-operation from string", meh.Details{"was": forOperationStr})
		}
		filters.ForOperation = nulls.NewUUID(forOperation)
	}
	if excludeGlobalStr := q.Get("exclude_global"); excludeGlobalStr != "" {
		excludeGlobal, err := strconv.ParseBool(excludeGlobalStr)
		if err != nil {
			return store.AddressBookEntryFilters{}, meh.NewBadInputErrFromErr(err, "parse bool", meh.Details{"was": excludeGlobalStr})
		}
		filters.ExcludeGlobal = excludeGlobal
	}
	if visibleByStr := q.Get("visible_by"); visibleByStr != "" {
		visibleBy, err := uuid.FromString(visibleByStr)
		if err != nil {
			return store.AddressBookEntryFilters{}, meh.NewBadInputErrFromErr(err, "visible-by from string", meh.Details{"was": visibleByStr})
		}
		filters.VisibleBy = nulls.NewUUID(visibleBy)
	}
	if includeForInactiveUsersStr := q.Get("include_for_inactive_users"); includeForInactiveUsersStr != "" {
		includeForInactiveUsers, err := strconv.ParseBool(includeForInactiveUsersStr)
		if err != nil {
			return store.AddressBookEntryFilters{}, meh.NewBadInputErrFromErr(err, "include-for-inactive-users from string",
				meh.Details{"was": includeForInactiveUsersStr})
		}
		filters.IncludeForInactiveUsers = includeForInactiveUsers
	}
	if autoDeliveryEnabledStr := q.Get("auto_delivery_enabled"); autoDeliveryEnabledStr != "" {
		autoDeliveryEnabled, err := strconv.ParseBool(autoDeliveryEnabledStr)
		if err != nil {
			return store.AddressBookEntryFilters{}, meh.NewBadInputErrFromErr(err, "parse auto-delivery-enabled from string",
				meh.Details{"was": autoDeliveryEnabledStr})
		}
		filters.AutoDeliveryEnabled = nulls.NewBool(autoDeliveryEnabled)
	}
	return filters, nil
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
		// Parse entry filters.
		filters, err := addressBookEntryFiltersFromQuery(c.Request.URL.Query())
		if err != nil {
			return meh.Wrap(err, "address book entry filters from query", nil)
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
			filters.VisibleBy = nulls.NewUUID(token.UserID)
		}
		// Retrieve.
		sEntries, err := s.AddressBookEntries(c.Request.Context(), filters, paginationParams)
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

// handleSearchAddressBookEntriesStore are the dependencies needed for
// handleSearchAddressBookEntries.
type handleSearchAddressBookEntriesStore interface {
	SearchAddressBookEntries(ctx context.Context, filters store.AddressBookEntryFilters,
		searchParams search.Params) (search.Result[store.AddressBookEntryDetailed], error)
}

// handleSearchAddressBookEntries performs search on address book entries.
func handleSearchAddressBookEntries(s handleSearchAddressBookEntriesStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authorized", nil)
		}
		// Extract filters.
		entryFilters, err := addressBookEntryFiltersFromQuery(c.Request.URL.Query())
		if err != nil {
			return meh.Wrap(err, "address book entry filters from query", nil)
		}
		// Extract search params.
		searchParams, err := search.ParamsFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "search params from request", nil)
		}
		// Check permisions.
		viewAnyGranted, err := auth.HasPermission(token, permission.ViewAnyAddressBookEntry())
		if err != nil {
			return meh.Wrap(err, "check permission", nil)
		}
		if !viewAnyGranted {
			entryFilters.VisibleBy = nulls.NewUUID(token.UserID)
		}
		// Search.
		result, err := s.SearchAddressBookEntries(c.Request.Context(), entryFilters, searchParams)
		if err != nil {
			return meh.Wrap(err, "search address book entries", meh.Details{
				"filters":       entryFilters,
				"search_params": searchParams,
			})
		}
		c.JSON(http.StatusOK, search.MapResult(result, publicAddressBookEntryDetailedFromStore))
		return nil
	}
}

// handleRebuildAddressBookEntrySearchStore are the dependencies needed for
// handleRebuildAddressBookEntrySearch.
type handleRebuildAddressBookEntrySearchStore interface {
	RebuildAddressBookEntrySearch(ctx context.Context)
}

// handleRebuildAddressBookEntrySearch rebuilds the search for address book
// entries.
func handleRebuildAddressBookEntrySearch(s handleRebuildAddressBookEntrySearchStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		err := auth.AssurePermission(token, permission.RebuildSearchIndex())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		go s.RebuildAddressBookEntrySearch(context.Background())
		c.Status(http.StatusOK)
		return nil
	}
}

// handleSetAddressBookEntriesWithAutoDeliveryEnabledStore are the dependencies
// needed for handleSetAddressBookEntriesWithAutoDeliveryEnabled.
type handleSetAddressBookEntriesWithAutoDeliveryEnabledStore interface {
	SetAddressBookEntriesWithAutoDeliveryEnabled(ctx context.Context, entryIDs []uuid.UUID) error
}

// handleSetAddressBookEntriesWithAutoDeliveryEnabled sets auto intel-delivery
// enabled for the given address book entries and disabled for all other ones.
func handleSetAddressBookEntriesWithAutoDeliveryEnabled(s handleSetAddressBookEntriesWithAutoDeliveryEnabledStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		err := auth.AssurePermission(token, permission.ManageIntelDelivery())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		// Parse address book entries.
		var entries []uuid.UUID
		err = c.BindJSON(&entries)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse body", nil)
		}
		// Update.
		err = s.SetAddressBookEntriesWithAutoDeliveryEnabled(c.Request.Context(), entries)
		if err != nil {
			return meh.Wrap(err, "set address book entries with auto delivery enabled", meh.Details{"new_entries": entries})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// handleGetAutoIntelDeliveryEnabledForAddressBookEntryStore are the dependencies
// needed for handleGetAutoIntelDeliveryEnabledForAddressBookEntry.
type handleGetAutoIntelDeliveryEnabledForAddressBookEntryStore interface {
	IsAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID) (bool, error)
}

// handleGetAutoIntelDeliveryEnabledForAddressBookEntry checks whether auto intel
// delivery is enabled for the address book entry with the given id.
func handleGetAutoIntelDeliveryEnabledForAddressBookEntry(s handleGetAutoIntelDeliveryEnabledForAddressBookEntryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		err := auth.AssurePermission(token, permission.ManageIntelDelivery())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		// Extract entry id.
		entryIDStr := c.Param("entryID")
		entryID, err := uuid.FromString(entryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse entry id", meh.Details{"was": entryIDStr})
		}
		// Retrieve.
		isEnabled, err := s.IsAutoIntelDeliveryEnabledForAddressBookEntry(c.Request.Context(), entryID)
		if err != nil {
			return meh.Wrap(err, "is auto intel delivery enabled for address book entry", meh.Details{"entry_id": entryID})
		}
		c.String(http.StatusOK, fmt.Sprintf("%t", isEnabled))
		return nil
	}
}

// handleEnableAutoIntelDeliveryForAddressBookEntryStore are the dependencies
// needed for handleEnableAutoIntelDeliveryForAddressBookEntry.
type handleEnableAutoIntelDeliveryForAddressBookEntryStore interface {
	SetAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID, enabled bool) error
}

// handleEnableAutoIntelDeliveryForAddressBookEntry sets auto intel delivery to
// enabled for the address book entry with the given id.
func handleEnableAutoIntelDeliveryForAddressBookEntry(s handleEnableAutoIntelDeliveryForAddressBookEntryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		err := auth.AssurePermission(token, permission.ManageIntelDelivery())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		// Extract entry id.
		entryIDStr := c.Param("entryID")
		entryID, err := uuid.FromString(entryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse entry id", meh.Details{"was": entryIDStr})
		}
		// Set.
		err = s.SetAutoIntelDeliveryEnabledForAddressBookEntry(c.Request.Context(), entryID, true)
		if err != nil {
			return meh.Wrap(err, "set auto intel delivery enabled for address book entry", meh.Details{"entry_id": entryID})
		}
		c.Status(http.StatusOK)
		return nil
	}
}

// handleDisableAutoIntelDeliveryForAddressBookEntryStore are the dependencies
// needed for handleDisableAutoIntelDeliveryForAddressBookEntry.
type handleDisableAutoIntelDeliveryForAddressBookEntryStore interface {
	SetAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID, enabled bool) error
}

// handleDisableAutoIntelDeliveryForAddressBookEntry sets auto intel delivery to
// disabled for the address book entry with the given id.
func handleDisableAutoIntelDeliveryForAddressBookEntry(s handleDisableAutoIntelDeliveryForAddressBookEntryStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		err := auth.AssurePermission(token, permission.ManageIntelDelivery())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		// Extract entry id.
		entryIDStr := c.Param("entryID")
		entryID, err := uuid.FromString(entryIDStr)
		if err != nil {
			return meh.NewBadInputErrFromErr(err, "parse entry id", meh.Details{"was": entryIDStr})
		}
		// Set.
		err = s.SetAutoIntelDeliveryEnabledForAddressBookEntry(c.Request.Context(), entryID, false)
		if err != nil {
			return meh.Wrap(err, "set auto intel delivery disabled for address book entry", meh.Details{"entry_id": entryID})
		}
		c.Status(http.StatusOK)
		return nil
	}
}
