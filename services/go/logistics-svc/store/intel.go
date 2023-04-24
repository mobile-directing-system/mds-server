package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"time"
)

const intelSearchIndex search.Index = "intel"

const intelSearchAttrID search.Attribute = "id"
const intelSearchAttrCreatedAt search.Attribute = "created_at"
const intelSearchAttrCreatedBy search.Attribute = "created_by"
const intelSearchAttrOperation search.Attribute = "operation"
const intelSearchAttrType search.Attribute = "intel_type"
const intelSearchAttrContent search.Attribute = "content"
const intelSearchAttrSearchText search.Attribute = "search_text"
const intelSearchAttrImportance search.Attribute = "importance"
const intelSearchAttrIsValid search.Attribute = "is_valid"
const intelSearchAttrDeliveryForEntries search.Attribute = "delivery_for_entries"
const intelSearchAttrDeliveredToEntries search.Attribute = "delivered_to_entries"

// intelSearchIndexConfig is the index-config for the intelSearchIndex.
var intelSearchIndexConfig = search.IndexConfig{
	PrimaryKey: intelSearchAttrID,
	Searchable: []search.Attribute{
		intelSearchAttrSearchText,
		intelSearchAttrContent,
		intelSearchAttrType,
		intelSearchAttrOperation,
	},
	Filterable: []search.Attribute{
		intelSearchAttrCreatedAt,
		intelSearchAttrCreatedBy,
		intelSearchAttrOperation,
		intelSearchAttrType,
		intelSearchAttrContent,
		intelSearchAttrSearchText,
		intelSearchAttrImportance,
		intelSearchAttrIsValid,
		intelSearchAttrDeliveryForEntries,
		intelSearchAttrDeliveredToEntries,
	},
	Sortable: []search.Attribute{
		intelSearchAttrCreatedAt,
	},
}

// dcoumentFromIntel generates a search.Document for the given Intel.
func documentFromIntel(intel Intel, deliveryForEntry, deliveredToEntry []uuid.UUID) search.Document {
	return search.Document{
		intelSearchAttrID:                 intel.ID,
		intelSearchAttrCreatedAt:          intel.CreatedAt,
		intelSearchAttrCreatedBy:          intel.CreatedBy,
		intelSearchAttrOperation:          intel.Operation,
		intelSearchAttrType:               intel.Type,
		intelSearchAttrContent:            intel.Content,
		intelSearchAttrSearchText:         intel.SearchText,
		intelSearchAttrImportance:         intel.Importance,
		intelSearchAttrIsValid:            intel.IsValid,
		intelSearchAttrDeliveryForEntries: deliveryForEntry,
		intelSearchAttrDeliveredToEntries: deliveredToEntry,
	}
}

// IntelType is the type of intel. Also describes the content.
type IntelType string

// CreateIntel for creating intel.
type CreateIntel struct {
	// CreatedBy is the id of the user, who created the intent.
	CreatedBy uuid.UUID
	// Operation is the id of the associated operation.
	Operation uuid.UUID
	// Type of the intel.
	Type IntelType
	// Content is the actual information.
	Content json.RawMessage
	// SearchText for better searching. Used with higher priority than Content.
	SearchText nulls.String
	// Importance of the intel. Used for example for prioritizing delivery methods.
	Importance int
	// InitialDeliverTo contains the recipient address book entries to initially
	// deliver the intel to.
	InitialDeliverTo []uuid.UUID
}

// Validate the CreateIntel for Type, Content and Assignments.
func (i CreateIntel) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	// Validate intel-type and content.
	subReport, err := validateCreateIntelTypeAndContent(i.Type, i.Content)
	if err != nil {
		return entityvalidation.Report{}, meh.Wrap(err, "validate create-intel-type and content", nil)
	}
	report.Include(subReport)
	// Assure no duplicate delivery-entries.
	assignedTo := make(map[uuid.UUID]struct{}, len(i.InitialDeliverTo))
	for _, to := range i.InitialDeliverTo {
		if _, ok := assignedTo[to]; ok {
			report.AddError(fmt.Sprintf("duplicate entry in deliver-to %s", to.String()))
			continue
		}
		assignedTo[to] = struct{}{}
	}
	return report, nil
}

// Intel holds intel information.
type Intel struct {
	// ID identifies the intel.
	ID uuid.UUID
	// CreatedAt is the timestamp, the intel was created.
	CreatedAt time.Time
	// CreatedBy is the id of the user, who created the intent.
	CreatedBy uuid.UUID
	// Operation is the id of the associated operation.
	Operation uuid.UUID
	// Type of the intel.
	Type IntelType
	// Content is the actual information.
	Content json.RawMessage
	// SearchText for searching with higher priority than Content.
	SearchText nulls.String
	// Importance of the intel.
	Importance int
	// IsValid describes whether the intel is still valid or marked as invalid
	// (equals deletion).
	IsValid bool
}

// CreateIntel creates the given intel with its assignments.
func (m *Mall) CreateIntel(ctx context.Context, tx pgx.Tx, create CreateIntel) (Intel, error) {
	// Create intel.
	q, _, err := m.dialect.Insert(goqu.T("intel")).Rows(goqu.Record{
		"created_at":  time.Now().UTC(),
		"created_by":  create.CreatedBy,
		"operation":   create.Operation,
		"type":        create.Type,
		"content":     []byte(create.Content),
		"search_text": create.SearchText,
		"importance":  create.Importance,
		"is_valid":    true,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return Intel{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Intel{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return Intel{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return Intel{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	var intelID uuid.UUID
	err = rows.Scan(&intelID)
	if err != nil {
		return Intel{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	created, err := m.IntelByID(ctx, tx, intelID)
	if err != nil {
		return Intel{}, meh.Wrap(err, "created intel by id", meh.Details{"intel_id": intelID})
	}
	// Create in search.
	err = m.addOrUpdateIntelInSearch(ctx, tx, intelID)
	if err != nil {
		return Intel{}, meh.Wrap(err, "add or update intel in search", meh.Details{"intel_id": intelID})
	}
	return created, nil
}

// InvalidateIntelByID sets the valid-field of the intel with the given id to
// false.
func (m *Mall) InvalidateIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error {
	q, _, err := m.dialect.Update(goqu.T("intel")).Set(goqu.Record{
		"is_valid": false,
	}).Where(goqu.C("id").Eq(intelID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("not found", nil)
	}
	// Update in search.
	err = m.addOrUpdateIntelInSearch(ctx, tx, intelID)
	if err != nil {
		return meh.Wrap(err, "update intel in search", meh.Details{"intel_id": intelID})
	}
	return nil
}

// IntelByID retrieves an Intel by its id.
func (m *Mall) IntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) (Intel, error) {
	// Retrieve intel information.
	q, _, err := m.dialect.From(goqu.T("intel")).
		Select(goqu.C("id"),
			goqu.C("created_at"),
			goqu.C("created_by"),
			goqu.C("operation"),
			goqu.C("type"),
			goqu.C("content"),
			goqu.C("search_text"),
			goqu.C("importance"),
			goqu.C("is_valid")).
		Where(goqu.C("id").Eq(intelID)).ToSQL()
	if err != nil {
		return Intel{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Intel{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return Intel{}, meh.NewNotFoundErr("not found", nil)
	}
	var intel Intel
	err = rows.Scan(&intel.ID,
		&intel.CreatedAt,
		&intel.CreatedBy,
		&intel.Operation,
		&intel.Type,
		&intel.Content,
		&intel.SearchText,
		&intel.Importance,
		&intel.IsValid)
	if err != nil {
		return Intel{}, mehpg.NewScanRowsErr(err, "scan rows", q)
	}
	rows.Close()
	return intel, nil
}

// IntelFilters for batch-intel-retrieval.
type IntelFilters struct {
	// CreatedBy is the id of the user that created the intel.
	CreatedBy uuid.NullUUID
	// Operation, intel needs to be part of.
	Operation uuid.NullUUID
	// IntelType of the intel.
	IntelType nulls.JSONNullable[IntelType]
	// MinImportance is the minimum importance of the intel.
	MinImportance nulls.Int
	// IncludeInvalid includes invalid intel.
	IncludeInvalid nulls.Bool
	// OneOfDeliveryForEntries only includes intel having deliveries for one of the
	// given entries.
	OneOfDeliveryForEntries []uuid.UUID
	// OneOfDeliveredToEntries only includes intel having successful deliveries for
	// one of the given entries.
	OneOfDeliveredToEntries []uuid.UUID
}

// SearchIntel using the given IntelFilters and search.Params.
func (m *Mall) SearchIntel(ctx context.Context, tx pgx.Tx, filters IntelFilters, searchParams search.Params) (search.Result[Intel], error) {
	// Search.
	var searchFilters [][]string
	if filters.CreatedBy.Valid {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%v = '%s'", intelSearchAttrCreatedBy, filters.CreatedBy.UUID.String()),
		})
	}
	if filters.Operation.Valid {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%v = '%s'", intelSearchAttrOperation, filters.Operation.UUID.String()),
		})
	}
	if filters.IntelType.Valid {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%v = '%v'", intelSearchAttrType, filters.IntelType.V),
		})
	}
	if filters.MinImportance.Valid {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%v >= %d", intelSearchAttrImportance, filters.MinImportance.Int),
		})
	}
	if !filters.IncludeInvalid.Valid || !filters.IncludeInvalid.Bool {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%v = true", intelSearchAttrIsValid),
		})
	}
	if len(filters.OneOfDeliveryForEntries) > 0 {
		ors := make([]string, 0, len(filters.OneOfDeliveryForEntries))
		for _, entry := range filters.OneOfDeliveryForEntries {
			ors = append(ors, fmt.Sprintf("%v = '%s'", intelSearchAttrDeliveryForEntries, entry.String()))
		}
		searchFilters = append(searchFilters, ors)
	}
	if len(filters.OneOfDeliveredToEntries) > 0 {
		ors := make([]string, 0, len(filters.OneOfDeliveredToEntries))
		for _, entry := range filters.OneOfDeliveredToEntries {
			ors = append(ors, fmt.Sprintf("%v = '%s'", intelSearchAttrDeliveredToEntries, entry.String()))
		}
		searchFilters = append(searchFilters, ors)
	}
	searchSort := []string{fmt.Sprintf("%v:desc", intelSearchAttrCreatedAt)}
	resultUUIDs, err := search.UUIDSearch(m.searchClient, intelSearchIndex, searchParams, search.Request{
		Filter: searchFilters,
		Sort:   searchSort,
	})
	if err != nil {
		return search.Result[Intel]{}, meh.Wrap(err, "search uuids", meh.Details{
			"index":  intelSearchIndex,
			"params": searchParams,
			"filter": searchFilters,
			"sort":   searchSort,
		})
	}
	// Query.
	qb := m.dialect.From(goqu.T("intel")).
		Select(goqu.C("id"),
			goqu.C("created_at"),
			goqu.C("created_by"),
			goqu.C("operation"),
			goqu.C("type"),
			goqu.C("content"),
			goqu.C("search_text"),
			goqu.C("importance"),
			goqu.C("is_valid"))
	// For safety in order to hide intel not having deliveries for the optionally
	// set user.
	if len(filters.OneOfDeliveryForEntries) > 0 {
		qb = qb.Where(goqu.C("id").In(m.dialect.From(goqu.T("intel_deliveries")).
			Select(goqu.I("intel_deliveries.intel")).
			Where(goqu.I("intel_deliveries.to").In(filters.OneOfDeliveryForEntries))))
	}
	q, _, err := pgutil.QueryWithOrdinalityUUID(qb, goqu.C("id"), resultUUIDs.Hits).ToSQL()
	if err != nil {
		return search.Result[Intel]{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return search.Result[Intel]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	result := make([]Intel, 0, len(resultUUIDs.Hits))
	for rows.Next() {
		var intel Intel
		err = rows.Scan(&intel.ID,
			&intel.CreatedAt,
			&intel.CreatedBy,
			&intel.Operation,
			&intel.Type,
			&intel.Content,
			&intel.SearchText,
			&intel.Importance,
			&intel.IsValid)
		if err != nil {
			return search.Result[Intel]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		result = append(result, intel)
	}
	rows.Close()
	return search.ResultFromResult(resultUUIDs, result), nil
}

// documentFromIntelByID retrieves the search.Document for the intel with the
// given id, using documentFromIntel.
func (m *Mall) documentFromIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) (search.Document, error) {
	// Retrieve intel.
	intel, err := m.IntelByID(ctx, tx, intelID)
	if err != nil {
		return nil, meh.Wrap(err, "intel by id", meh.Details{"intel_id": intelID})
	}
	// Retrieve deliveries.
	deliveries, err := m.IntelDeliveriesByIntel(ctx, tx, intelID)
	if err != nil {
		return nil, meh.Wrap(err, "intel-deliveries by intel", meh.Details{"intel_id": intelID})
	}
	deliveryForEntries := make([]uuid.UUID, 0, len(deliveries))
	for _, delivery := range deliveries {
		deliveryForEntries = append(deliveryForEntries, delivery.To)
	}
	deliveredToEntries := make([]uuid.UUID, 0, len(deliveryForEntries))
	for _, delivery := range deliveries {
		if !delivery.Success {
			continue
		}
		deliveredToEntries = append(deliveredToEntries, delivery.To)
	}
	return documentFromIntel(intel, deliveryForEntries, deliveredToEntries), nil
}

// addOrUpdateIntelInSearch adds or updates the intel with the given id in the
// search. This should be called everytime, the intel, intel-deliveries or
// intel-assignments change.
func (m *Mall) addOrUpdateIntelInSearch(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error {
	d, err := m.documentFromIntelByID(ctx, tx, intelID)
	if err != nil {
		return meh.Wrap(err, "document from intel by id", meh.Details{"intel_id": intelID})
	}
	err = m.searchClient.SafeAddOrUpdateDocument(ctx, tx, intelSearchIndex, d)
	if err != nil {
		return meh.Wrap(err, "safe add or update document in search", meh.Details{
			"index":        intelSearchIndex,
			"new_document": d,
		})
	}
	return nil
}

// RebuildIntelSearch rebuilds the intel-search.
func (m *Mall) RebuildIntelSearch(ctx context.Context, tx pgx.Tx) error {
	err := m.searchClient.SafeRebuildIndex(ctx, tx, intelSearchIndex)
	if err != nil {
		return meh.Wrap(err, "safe rebuild index", nil)
	}
	return nil
}

// Intel retrieves a paginated Intel list using the given IntelFilters and
// pagination.Params, sorted descending by creation date.
//
// Warning: Sorting via pagination.Params is discarded!
func (m *Mall) Intel(ctx context.Context, tx pgx.Tx, filters IntelFilters,
	paginationParams pagination.Params) (pagination.Paginated[Intel], error) {
	// Retrieve intel information. If in the future retrieval gets too slow when not
	// applying any filters, the query might be adjusted in terms of aggregation
	// (over by), indexes, timeframes, etc.
	qb := m.dialect.From(goqu.T("intel")).
		Select(goqu.I("intel.id"),
			goqu.I("intel.created_at"),
			goqu.I("intel.created_by"),
			goqu.I("intel.operation"),
			goqu.I("intel.type"),
			goqu.I("intel.content"),
			goqu.I("intel.search_text"),
			goqu.I("intel.importance"),
			goqu.I("intel.is_valid")).
		Order(goqu.I("intel.created_at").Desc())
	if filters.CreatedBy.Valid {
		qb = qb.Where(goqu.I("intel.created_by").Eq(filters.CreatedBy.UUID))
	}
	if filters.Operation.Valid {
		qb = qb.Where(goqu.I("intel.operation").Eq(filters.Operation.UUID))
	}
	if filters.IntelType.Valid {
		qb = qb.Where(goqu.I("intel.type").Eq(filters.IntelType.V))
	}
	if filters.MinImportance.Valid {
		qb = qb.Where(goqu.I("intel.importance").Gte(filters.MinImportance.Int))
	}
	if !filters.IncludeInvalid.Valid || !filters.IncludeInvalid.Bool {
		qb = qb.Where(goqu.I("intel.is_valid").IsTrue())
	}
	if len(filters.OneOfDeliveryForEntries) > 0 {
		qb = qb.Where(goqu.I("intel.id").In(m.dialect.From(goqu.T("intel_deliveries").As("one_of_delivery")).
			Select(goqu.I("one_of_delivery.intel")).
			Where(goqu.I("one_of_delivery.to").In(filters.OneOfDeliveryForEntries))))
	}
	if len(filters.OneOfDeliveredToEntries) > 0 {
		qb = qb.Where(goqu.I("intel.id").In(m.dialect.From(goqu.T("intel_deliveries").As("one_of_delivered")).
			Select(goqu.I("one_of_delivered.intel")).
			Where(goqu.I("one_of_delivered.success").IsTrue(),
				goqu.I("one_of_delivered.to").In(filters.OneOfDeliveredToEntries))))
	}
	paginationParams.OrderBy = nulls.String{}
	q, _, err := pagination.QueryToSQLWithPagination(qb, paginationParams, pagination.FieldMap{})
	if err != nil {
		return pagination.Paginated[Intel]{}, meh.NewInternalErrFromErr(err, "query with pagination to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return pagination.Paginated[Intel]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	intelList := make([]Intel, 0)
	var total int
	for rows.Next() {
		var intel Intel
		err = rows.Scan(&intel.ID,
			&intel.CreatedAt,
			&intel.CreatedBy,
			&intel.Operation,
			&intel.Type,
			&intel.Content,
			&intel.SearchText,
			&intel.Importance,
			&intel.IsValid,
			&total)
		if err != nil {
			return pagination.Paginated[Intel]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		intelList = append(intelList, intel)
	}
	rows.Close()
	return pagination.NewPaginated(paginationParams, intelList, total), nil
}

// IntelDeliveryAttemptFilters are filters for delivery attempt retrieval.
type IntelDeliveryAttemptFilters struct {
	// ByOperation only includes attempts for deliveries for intel for the operation
	// with the given id.
	ByOperation uuid.NullUUID
	// ByDelivery only includes attempts for the delivery with the given id.
	ByDelivery uuid.NullUUID
	// ByChannel only includes attempts for the channel with the given id.
	ByChannel uuid.NullUUID
	// ByActive only includes attempts being (in)active.
	ByActive nulls.Bool
}

// IntelDeliveryAttempts retrieves a paginated IntelDeliveryAttempt list using
// the given IntelDeliveryAttemptFilters and pagination.Params, sorted descending
// by creation date.
//
// Warning: Sorting via pagination.Params is discarded!
func (m *Mall) IntelDeliveryAttempts(ctx context.Context, tx pgx.Tx, filters IntelDeliveryAttemptFilters,
	page pagination.Params) (pagination.Paginated[IntelDeliveryAttempt], error) {
	qb := m.dialect.From(goqu.T("intel_delivery_attempts")).
		InnerJoin(goqu.T("intel_deliveries"),
			goqu.On(goqu.I("intel_deliveries.id").Eq(goqu.I("intel_delivery_attempts.delivery")))).
		InnerJoin(goqu.T("intel"),
			goqu.On(goqu.I("intel.id").Eq(goqu.I("intel_deliveries.intel")))).
		Select(goqu.I("intel_delivery_attempts.id"),
			goqu.I("intel_delivery_attempts.delivery"),
			goqu.I("intel_delivery_attempts.channel"),
			goqu.I("intel_delivery_attempts.created_at"),
			goqu.I("intel_delivery_attempts.is_active"),
			goqu.I("intel_delivery_attempts.status"),
			goqu.I("intel_delivery_attempts.status_ts"),
			goqu.I("intel_delivery_attempts.note")).
		Order(goqu.I("intel_delivery_attempts.created_at").Desc())
	if filters.ByOperation.Valid {
		qb = qb.Where(goqu.I("intel.operation").Eq(filters.ByOperation.UUID))
	}
	if filters.ByDelivery.Valid {
		qb = qb.Where(goqu.I("intel_delivery_attempts.delivery").Eq(filters.ByDelivery.UUID))
	}
	if filters.ByChannel.Valid {
		qb = qb.Where(goqu.I("intel_delivery_attempts.channel").Eq(filters.ByChannel.UUID))
	}
	if filters.ByActive.Valid {
		qb = qb.Where(goqu.I("intel_delivery_attempts.is_active").Eq(filters.ByActive.Bool))
	}
	page.OrderBy = nulls.String{}
	q, _, err := pagination.QueryToSQLWithPagination(qb, page, pagination.FieldMap{})
	if err != nil {
		return pagination.Paginated[IntelDeliveryAttempt]{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return pagination.Paginated[IntelDeliveryAttempt]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	attempts := make([]IntelDeliveryAttempt, 0)
	var total int
	for rows.Next() {
		var attempt IntelDeliveryAttempt
		err = rows.Scan(&attempt.ID,
			&attempt.Delivery,
			&attempt.Channel,
			&attempt.CreatedAt,
			&attempt.IsActive,
			&attempt.Status,
			&attempt.StatusTS,
			&attempt.Note,
			&total)
		if err != nil {
			return pagination.Paginated[IntelDeliveryAttempt]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		attempts = append(attempts, attempt)
	}
	rows.Close()
	return pagination.NewPaginated(page, attempts, total), nil
}
