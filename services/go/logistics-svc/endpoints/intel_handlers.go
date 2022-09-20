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
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// publicIntelType is the public representation of store.IntelType.
type publicIntelType string

// Public representations of store.IntelType.
const (
	intelTypeAnalogRadioMessage publicIntelType = "analog-radio-message"
	intelTypePlaintextMessage   publicIntelType = "plaintext-message"
)

// pITAnalogRadioMessageContent is the public representation of
// store.IntelTypeAnalogRadioMessageContent for intelTypeAnalogRadioMessage.
type pITAnalogRadioMessageContent struct {
	Channel  string `json:"channel"`
	Callsign string `json:"callsign"`
	Head     string `json:"head"`
	Content  string `json:"content"`
}

// pITPlaintextMessageContent is the public representation of
// store.IntelTypePlaintextMessageContent for intelTypePlaintextMessage.
type pITPlaintextMessageContent struct {
	Text string `json:"text"`
}

// publicIntelTypeFromStore converts store.IntelType to publicIntelType. If no
// mapping was found, a meh.ErrInternal is returned.
func publicIntelTypeFromStore(s store.IntelType) (publicIntelType, error) {
	switch s {
	case store.IntelTypeAnalogRadioMessage:
		return intelTypeAnalogRadioMessage, nil
	case store.IntelTypePlaintextMessage:
		return intelTypePlaintextMessage, nil
	default:
		return "", meh.NewInternalErr("unsupported type", meh.Details{"type": s})
	}
}

// storeIntelTypeFromPublic converts publicIntelType to store.IntelType. If no
// mapping was found, a meh.ErrBadInput is returned.
func storeIntelTypeFromPublic(p publicIntelType) (store.IntelType, error) {
	switch p {
	case intelTypeAnalogRadioMessage:
		return store.IntelTypeAnalogRadioMessage, nil
	case intelTypePlaintextMessage:
		return store.IntelTypePlaintextMessage, nil
	default:
		return "", meh.NewBadInputErr("unsupported type", meh.Details{"type": p})
	}
}

// intelContentMapper unmarshals the given raw message, calls the mapper
// function and marshals back as JSON.
func intelContentMapper[From any, To any](mapFn func(From) (To, error)) func(s json.RawMessage) (json.RawMessage, error) {
	return func(s json.RawMessage) (json.RawMessage, error) {
		var f From
		err := json.Unmarshal(s, &f)
		if err != nil {
			return nil, meh.NewInternalErrFromErr(err, "unmarshal message content", meh.Details{"raw": string(s)})
		}
		mapped, err := mapFn(f)
		if err != nil {
			return nil, meh.Wrap(err, "map fn", meh.Details{"from": f})
		}
		toRaw, err := json.Marshal(mapped)
		if err != nil {
			return nil, meh.NewInternalErrFromErr(err, "marshal mapped", nil)
		}
		return toRaw, nil
	}
}

// publicIntelContentFromStore maps the content from store.Intel to its
// event-representation.
func publicIntelContentFromStore(sType store.IntelType, sContentRaw json.RawMessage) (json.RawMessage, error) {
	var mapper func(s json.RawMessage) (json.RawMessage, error)
	switch sType {
	case store.IntelTypeAnalogRadioMessage:
		mapper = intelContentMapper(pITAnalogRadioMessageContentFromStore)
	case store.IntelTypePlaintextMessage:
		mapper = intelContentMapper(pITPlaintextMessageContentFromStore)
	}
	if mapper == nil {
		return nil, meh.NewInternalErr("no intel-content-mapper", meh.Details{"intel_type": sType})
	}
	mappedRaw, err := mapper(sContentRaw)
	if err != nil {
		return nil, meh.Wrap(err, "mapper fn", nil)
	}
	return mappedRaw, nil
}

// storeIntelContentFromPublic maps the content from publicIntel to its
// store-representation.
func storeIntelContentFromPublic(pType publicIntelType, pContentRaw json.RawMessage) (json.RawMessage, error) {
	var mapper func(s json.RawMessage) (json.RawMessage, error)
	switch pType {
	case intelTypeAnalogRadioMessage:
		mapper = intelContentMapper(sITAnalogRadioMessageContentFromPublic)
	case intelTypePlaintextMessage:
		mapper = intelContentMapper(sITPlaintextMessageContentFromPublic)
	}
	if mapper == nil {
		return nil, meh.NewInternalErr("no intel-content-mapper", meh.Details{"intel_type": pType})
	}
	mappedRaw, err := mapper(pContentRaw)
	if err != nil {
		return nil, meh.Wrap(err, "mapper fn", nil)
	}
	return mappedRaw, nil
}

// pITAnalogRadioMessageContentFromStore maps
// store.IntelTypeAnalogRadioMessageContent to pITAnalogRadioMessageContent.
func pITAnalogRadioMessageContentFromStore(s store.IntelTypeAnalogRadioMessageContent) (pITAnalogRadioMessageContent, error) {
	return pITAnalogRadioMessageContent{
		Channel:  s.Channel,
		Callsign: s.Callsign,
		Head:     s.Head,
		Content:  s.Content,
	}, nil
}

// sITAnalogRadioMessageContentFromPublic maps pITAnalogRadioMessageContent to
// store.IntelTypeAnalogRadioMessageContent.
func sITAnalogRadioMessageContentFromPublic(s pITAnalogRadioMessageContent) (store.IntelTypeAnalogRadioMessageContent, error) {
	return store.IntelTypeAnalogRadioMessageContent{
		Channel:  s.Channel,
		Callsign: s.Callsign,
		Head:     s.Head,
		Content:  s.Content,
	}, nil
}

// pITPlaintextMessageContentFromStore maps
// store.IntelTypePlaintextMessageContent to pITPlaintextMessageContent.
func pITPlaintextMessageContentFromStore(s store.IntelTypePlaintextMessageContent) (pITPlaintextMessageContent, error) {
	return pITPlaintextMessageContent{
		Text: s.Text,
	}, nil
}

// sITPlaintextMessageContentFromPublic maps pITPlaintextMessageContent to
// store.IntelTypePlaintextMessageContent.
func sITPlaintextMessageContentFromPublic(p pITPlaintextMessageContent) (store.IntelTypePlaintextMessageContent, error) {
	return store.IntelTypePlaintextMessageContent{
		Text: p.Text,
	}, nil
}

// publicCreateIntel is the public representation of store.CreateIntel.
type publicCreateIntel struct {
	Operation        uuid.UUID       `json:"operation"`
	Type             publicIntelType `json:"type"`
	Content          json.RawMessage `json:"content"`
	Importance       int             `json:"importance"`
	InitialDeliverTo []uuid.UUID     `json:"initial_deliver_to"`
}

// storeCreateIntelFromPublic maps publicCreateIntel to store.CreateIntel.
func storeCreateIntelFromPublic(createdBy uuid.UUID, p publicCreateIntel) (store.CreateIntel, error) {
	intelType, err := storeIntelTypeFromPublic(p.Type)
	if err != nil {
		return store.CreateIntel{}, meh.Wrap(err, "store intel type from public", meh.Details{"type": p.Type})
	}
	intelContent, err := storeIntelContentFromPublic(p.Type, p.Content)
	if err != nil {
		return store.CreateIntel{}, meh.Wrap(err, "map intel-content", meh.Details{
			"intel_type":    p.Type,
			"intel_content": string(p.Content),
		})
	}
	return store.CreateIntel{
		CreatedBy:        createdBy,
		Operation:        p.Operation,
		Type:             intelType,
		Content:          intelContent,
		Importance:       p.Importance,
		InitialDeliverTo: p.InitialDeliverTo,
	}, nil
}

// publicIntel is the public representation of store.Intel.
type publicIntel struct {
	ID         uuid.UUID       `json:"id"`
	CreatedAt  time.Time       `json:"created_at"`
	CreatedBy  uuid.UUID       `json:"created_by"`
	Operation  uuid.UUID       `json:"operation"`
	Type       publicIntelType `json:"type"`
	Content    json.RawMessage `json:"content"`
	SearchText nulls.String    `json:"search_text"`
	Importance int             `json:"importance"`
	IsValid    bool            `json:"is_valid"`
}

// publicIntelFromStore converts a store.Intel list to publicIntel list.
func publicIntelFromStore(s store.Intel) (publicIntel, error) {
	intelType, err := publicIntelTypeFromStore(s.Type)
	if err != nil {
		return publicIntel{}, meh.Wrap(err, "public intel type from store", meh.Details{"type": s.Type})
	}
	intelContent, err := publicIntelContentFromStore(s.Type, s.Content)
	if err != nil {
		return publicIntel{}, meh.Wrap(err, "map intel-content", meh.Details{
			"intel_type":    s.Type,
			"intel_content": string(s.Content),
		})
	}
	return publicIntel{
		ID:         s.ID,
		CreatedAt:  s.CreatedAt,
		CreatedBy:  s.CreatedBy,
		Operation:  s.Operation,
		Type:       intelType,
		Content:    intelContent,
		SearchText: s.SearchText,
		Importance: s.Importance,
		IsValid:    s.IsValid,
	}, nil
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
		c.JSON(http.StatusCreated, pCreated)
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

// intelFiltersFromRequest parses store.IntelFilters from the given query
// url.Values.
func intelFiltersFromRequest(q url.Values) (store.IntelFilters, error) {
	var filters store.IntelFilters
	// Created-by.
	if v := q.Get("created_by"); v != "" {
		id, err := uuid.FromString(v)
		if err != nil {
			return store.IntelFilters{}, meh.NewBadInputErrFromErr(err, "parse created-by", meh.Details{"was": v})
		}
		filters.CreatedBy = nulls.NewUUID(id)
	}
	// Operation.
	if v := q.Get("operation"); v != "" {
		id, err := uuid.FromString(v)
		if err != nil {
			return store.IntelFilters{}, meh.NewBadInputErrFromErr(err, "parse operation", meh.Details{"was": v})
		}
		filters.Operation = nulls.NewUUID(id)
	}
	// Intel-type.
	if v := q.Get("intel_type"); v != "" {
		sIntelType, err := storeIntelTypeFromPublic(publicIntelType(v))
		if err != nil {
			return store.IntelFilters{}, meh.Wrap(err, "store intel-type from public", meh.Details{"was": v})
		}
		filters.IntelType = nulls.NewJSONNullable(sIntelType)
	}
	// Minimum importance.
	if v := q.Get("min_importance"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return store.IntelFilters{}, meh.NewBadInputErrFromErr(err, "parse minimum importance", meh.Details{"was": v})
		}
		filters.MinImportance = nulls.NewInt(n)
	}
	// Include invalid.
	if v := q.Get("include_invalid"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return store.IntelFilters{}, meh.NewBadInputErrFromErr(err, "parse include-invalid", meh.Details{"was": v})
		}
		filters.IncludeInvalid = nulls.NewBool(b)
	}
	// One of delivery for entries.
	if v := q.Get("one_of_delivery_for_entries"); v != "" {
		err := json.Unmarshal([]byte(v), &filters.OneOfDeliveryForEntries)
		if err != nil {
			return store.IntelFilters{}, meh.NewBadInputErrFromErr(err, "parse one-of-delivery-for-entries",
				meh.Details{"was": v})
		}
	}
	// One of delivered to entries.
	if v := q.Get("one_of_delivered_to_entries"); v != "" {
		err := json.Unmarshal([]byte(v), &filters.OneOfDeliveredToEntries)
		if err != nil {
			return store.IntelFilters{}, meh.NewBadInputErrFromErr(err, "parse one-of-delivered-to-entries",
				meh.Details{"was": v})
		}
	}
	return filters, nil
}

// handleSearchIntelStore are the dependencies needed for handleSearchIntel.
type handleSearchIntelStore interface {
	SearchIntel(ctx context.Context, intelFilters store.IntelFilters, searchParams search.Params,
		limitToUser uuid.NullUUID) (search.Result[store.Intel], error)
}

// handleSearchIntel searches for intel.
func handleSearchIntel(s handleSearchIntelStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authorized", nil)
		}
		// Extract filters.
		IntelFilters, err := intelFiltersFromRequest(c.Request.URL.Query())
		if err != nil {
			return meh.Wrap(err, "intel filters from request", nil)
		}
		// Check permissions.
		limitToUser := nulls.NewUUID(token.UserID)
		ok, err := auth.HasPermission(token, permission.ViewAnyIntel())
		if err != nil {
			return meh.Wrap(err, "check permission", nil)
		}
		if ok {
			limitToUser = uuid.NullUUID{}
		}
		// Extract params.
		searchParams, err := search.ParamsFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "search params from request", nil)
		}
		// Search.
		result, err := s.SearchIntel(c.Request.Context(), IntelFilters, searchParams, limitToUser)
		if err != nil {
			return meh.Wrap(err, "search Intels", meh.Details{"search_params": searchParams})
		}
		publicIntelList := make([]publicIntel, 0, len(result.Hits))
		for _, sIntel := range result.Hits {
			pIntel, err := publicIntelFromStore(sIntel)
			if err != nil {
				return meh.NewInternalErrFromErr(err, "map store intel to public", meh.Details{"store_intel": sIntel})
			}
			publicIntelList = append(publicIntelList, pIntel)
		}
		c.JSON(http.StatusOK, search.ResultFromResult(result, publicIntelList))
		return nil
	}
}

// handleRebuildIntelSearchStore are the dependencies needed for
// handleRebuildIntelSearch.
type handleRebuildIntelSearchStore interface {
	RebuildIntelSearch(ctx context.Context)
}

// handleRebuildIntelSearch rebuilds the search for intel.
func handleRebuildIntelSearch(s handleRebuildIntelSearchStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		err := auth.AssurePermission(token, permission.RebuildSearchIndex())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		go s.RebuildIntelSearch(context.Background())
		c.Status(http.StatusOK)
		return nil
	}
}

// handleGetIntelByIDStore are the dependencies needed for handleGetIntelByID.
type handleGetIntelByIDStore interface {
	IntelByID(ctx context.Context, intelID uuid.UUID, limitToUser uuid.NullUUID) (store.Intel, error)
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
		// Check permissions for viewing any intel.
		limitToUser := nulls.NewUUID(token.UserID)
		ok, err := auth.HasPermission(token, permission.ViewAnyIntel())
		if err != nil {
			return meh.Wrap(err, "check permissions", nil)
		}
		if ok {
			limitToUser = uuid.NullUUID{}
		}
		// Retrieve.
		sIntel, err := s.IntelByID(c.Request.Context(), intelID, limitToUser)
		if err != nil {
			return meh.Wrap(err, "intel by id", meh.Details{
				"intel_id":      intelID,
				"limit_to_user": limitToUser,
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

// handleGetAllIntelStore are the dependencies needed for handleGetAllIntel.
type handleGetAllIntelStore interface {
	Intel(ctx context.Context, filters store.IntelFilters, paginationParams pagination.Params,
		limitToUser uuid.NullUUID) (pagination.Paginated[store.Intel], error)
}

// handleGetAllIntel retrieves a paginated intel list. Without the
// permission.ViewAnyIntel, the filter for deliveries for entries will be set
// automatically.
func handleGetAllIntel(s handleGetAllIntelStore) httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authorized", nil)
		}
		// Parse intel filters.
		filters, err := intelFiltersFromRequest(c.Request.URL.Query())
		if err != nil {
			return meh.Wrap(err, "intel filters from query", nil)
		}
		// Parse pagination params.
		paginationParams, err := pagination.ParamsFromRequest(c)
		if err != nil {
			return meh.Wrap(err, "pagination params from request", nil)
		}
		// Check permisions.
		limitToUser := nulls.NewUUID(token.UserID)
		viewAnyGranted, err := auth.HasPermission(token, permission.ViewAnyIntel())
		if err != nil {
			return meh.Wrap(err, "check permission", nil)
		}
		if viewAnyGranted {
			limitToUser = uuid.NullUUID{}
		}
		// Retrieve.
		sResult, err := s.Intel(c.Request.Context(), filters, paginationParams, limitToUser)
		if err != nil {
			return meh.Wrap(err, "entries from store", nil)
		}
		pIntelList := make([]publicIntel, 0, len(sResult.Entries))
		for _, sIntel := range sResult.Entries {
			pIntel, err := publicIntelFromStore(sIntel)
			if err != nil {
				return meh.Wrap(err, "public intel from store", meh.Details{"store_intel": sIntel})
			}
			pIntelList = append(pIntelList, pIntel)
		}
		c.JSON(http.StatusOK, pagination.PaginatedFromPaginated(sResult, pIntelList))
		return nil
	}
}
