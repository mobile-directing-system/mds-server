package store

import (
	"context"
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

// Operation search.
const (
	operationSearchIndex           search.Index     = "operations"
	operationSearchAttrID          search.Attribute = "id"
	operationSearchAttrTitle       search.Attribute = "title"
	operationSearchAttrDescription search.Attribute = "description"
	operationSearchAttrStartTS     search.Attribute = "start_ts"
	operationSearchAttrEndTS       search.Attribute = "end_ts"
	operationSearchAttrIsArchived  search.Attribute = "is_archived"
	operationSearchAttrMembers     search.Attribute = "members"
)

var operationSearchIndexConfig = search.IndexConfig{
	PrimaryKey: operationSearchAttrID,
	Searchable: []search.Attribute{
		operationSearchAttrTitle,
		operationSearchAttrDescription,
		operationSearchAttrID,
	},
	Filterable: []search.Attribute{
		operationSearchAttrStartTS,
		operationSearchAttrEndTS,
		operationSearchAttrIsArchived,
		operationSearchAttrMembers,
	},
	Sortable: nil,
}

func documentFromOperation(o Operation, members []uuid.UUID) search.Document {
	membersMapped := make([]string, 0, len(members))
	for _, member := range members {
		membersMapped = append(membersMapped, member.String())
	}
	var endTS nulls.Int64
	if o.End.Valid {
		endTS = nulls.NewInt64(o.End.Time.UTC().UnixNano())
	}
	return search.Document{
		operationSearchAttrID:          o.ID,
		operationSearchAttrTitle:       o.Title,
		operationSearchAttrDescription: o.Description,
		operationSearchAttrStartTS:     o.Start.UTC().UnixNano(),
		operationSearchAttrEndTS:       endTS,
		operationSearchAttrIsArchived:  o.IsArchived,
		operationSearchAttrMembers:     membersMapped,
	}
}

// Operation is the store representation of an operation.
type Operation struct {
	// ID identifies the operation.
	ID uuid.UUID
	// Title of the operation.
	Title string
	// Optional description. We treat a non-existent description as empty string.
	Description string
	// Start timestamp of the operation.
	Start time.Time
	// End is the optional timestamp, when the operation has finished. If End is not
	// set or set to a moment in the past, the operation is considered finished.
	End nulls.Time
	// IsArchived describes whether the operation was archived. This is used instead
	// of deleting the operation in order to avoid unintended data loss.
	IsArchived bool
}

// Validate assures that the title is set and End is not before Start.
func (o Operation) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	if o.Title == "" {
		report.AddError("title must not be empty")
	}
	if o.End.Valid && o.End.Time.UTC().Before(o.Start.UTC()) {
		report.AddError("end-time must not be before start-time")
	}
	return report, nil
}

// OperationByID retrieves an Operation by its id.
func (m *Mall) OperationByID(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) (Operation, error) {
	// Build query.
	q, _, err := m.dialect.From(goqu.T("operations")).
		Select(goqu.C("id"),
			goqu.C("title"),
			goqu.C("description"),
			goqu.C("start_ts"),
			goqu.C("end_ts"),
			goqu.C("is_archived")).
		Where(goqu.C("id").Eq(operationID)).ToSQL()
	if err != nil {
		return Operation{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Operation{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	if !rows.Next() {
		return Operation{}, meh.NewNotFoundErr("operation not found", nil)
	}
	var operation Operation
	err = rows.Scan(&operation.ID,
		&operation.Title,
		&operation.Description,
		&operation.Start,
		&operation.End,
		&operation.IsArchived)
	if err != nil {
		return Operation{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return operation, nil
}

// OperationRetrievalFilters for retrieval.
type OperationRetrievalFilters struct {
	// OnlyOngoing only includes ongoing operations that have not ended, yet.
	OnlyOngoing bool
	// IncludeArchived includes operations being marked as archived.
	IncludeArchived bool
	// ForUser only includes operations, the user with the given id is member of.
	ForUser uuid.NullUUID
}

// Operations retrieves an Operation list.
func (m *Mall) Operations(ctx context.Context, tx pgx.Tx, operationFilters OperationRetrievalFilters,
	paginationParams pagination.Params) (pagination.Paginated[Operation], error) {
	// Build query.
	qb := m.dialect.From(goqu.T("operations")).
		Select(goqu.I("operations.id"),
			goqu.I("operations.title"),
			goqu.I("operations.description"),
			goqu.I("operations.start_ts"),
			goqu.I("operations.end_ts"),
			goqu.I("operations.is_archived")).
		Order(goqu.I("operations.end_ts").Desc())
	if operationFilters.OnlyOngoing {
		qb = qb.Where(goqu.Or(goqu.I("operations.end_ts").IsNull(),
			goqu.I("operations.end_ts").Gt(goqu.L("now()"))))
	}
	if !operationFilters.IncludeArchived {
		qb = qb.Where(goqu.I("operations.is_archived").IsFalse())
	}
	if operationFilters.ForUser.Valid {
		qb = qb.Where(goqu.I("operations.id").In(m.dialect.From(goqu.T("operation_members")).
			Select(goqu.I("operation_members.operation")).Where(goqu.I("operation_members.user").Eq(operationFilters.ForUser.UUID))))
	}
	q, _, err := pagination.QueryToSQLWithPagination(qb, paginationParams, pagination.FieldMap{
		"title":       goqu.I("operations.title"),
		"description": goqu.I("operations.description"),
		"start":       goqu.I("operations.start_ts"),
		"end":         goqu.I("operations.end_ts"),
		"is_archived": goqu.I("operations.is_archived"),
	})
	if err != nil {
		return pagination.Paginated[Operation]{}, meh.Wrap(err, "query to sql with pagination", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return pagination.Paginated[Operation]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	operations := make([]Operation, 0)
	var total int
	for rows.Next() {
		var operation Operation
		err = rows.Scan(&operation.ID,
			&operation.Title,
			&operation.Description,
			&operation.Start,
			&operation.End,
			&operation.IsArchived,
			&total)
		if err != nil {
			return pagination.Paginated[Operation]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		operations = append(operations, operation)
	}
	return pagination.NewPaginated(paginationParams, operations, total), nil
}

// CreateOperation creates the given Operation and returns it with its assigned
// id.
func (m *Mall) CreateOperation(ctx context.Context, tx pgx.Tx, operation Operation) (Operation, error) {
	// Build query.
	q, _, err := m.dialect.Insert(goqu.T("operations")).Rows(goqu.Record{
		"title":       operation.Title,
		"description": operation.Description,
		"start_ts":    operation.Start.UTC(),
		"end_ts":      operation.End.UTC(),
		"is_archived": operation.IsArchived,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return Operation{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Operation{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return Operation{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return Operation{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	err = rows.Scan(&operation.ID)
	if err != nil {
		return Operation{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	// Add to search.
	members, err := m.allOperationMembersByOperation(ctx, tx, operation.ID)
	if err != nil {
		return Operation{}, meh.Wrap(err, "all operation members by operation", nil)
	}
	err = m.searchClient.SafeAddOrUpdateDocument(ctx, tx, operationSearchIndex, documentFromOperation(operation, members))
	if err != nil {
		return Operation{}, meh.Wrap(err, "safe add or update in search", nil)
	}
	return operation, nil
}

// UpdateOperation updates the given Operation, identified by its Operation.ID.
func (m *Mall) UpdateOperation(ctx context.Context, tx pgx.Tx, operation Operation) error {
	// Build query.
	q, _, err := m.dialect.Update(goqu.T("operations")).Set(goqu.Record{
		"title":       operation.Title,
		"description": operation.Description,
		"start_ts":    operation.Start.UTC(),
		"end_ts":      operation.End.UTC(),
		"is_archived": operation.IsArchived,
	}).Where(goqu.C("id").Eq(operation.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("operation not found", nil)
	}
	// Update in search.
	members, err := m.allOperationMembersByOperation(ctx, tx, operation.ID)
	if err != nil {
		return meh.Wrap(err, "all operation members by operation", nil)
	}
	err = m.searchClient.SafeAddOrUpdateDocument(ctx, tx, operationSearchIndex, documentFromOperation(operation, members))
	if err != nil {
		return meh.Wrap(err, "safe add or update in search", nil)
	}
	return nil
}

// SearchOperations searches for operations with the given search.Params.
func (m *Mall) SearchOperations(ctx context.Context, tx pgx.Tx, operationFilters OperationRetrievalFilters,
	searchParams search.Params) (search.Result[Operation], error) {
	// Search.
	var filters [][]string
	if operationFilters.OnlyOngoing {
		filters = append(filters, []string{
			fmt.Sprintf("%s <= %d", operationSearchAttrStartTS, time.Now().UTC().UnixNano()),
			fmt.Sprintf("%s >= %d", operationSearchAttrEndTS, time.Now().UTC().UnixNano()),
		})
	}
	if !operationFilters.IncludeArchived {
		filters = append(filters, []string{
			fmt.Sprintf("%s = false", operationSearchAttrIsArchived),
		})
	}
	if operationFilters.ForUser.Valid {
		filters = append(filters, []string{
			fmt.Sprintf("%s = '%s'", operationSearchAttrMembers, operationFilters.ForUser.UUID.String()),
		})
	}
	resultUUIDs, err := search.UUIDSearch(m.searchClient, operationSearchIndex, searchParams, search.Request{
		Filter: filters,
	})
	if err != nil {
		return search.Result[Operation]{}, meh.Wrap(err, "search uuids", meh.Details{
			"index":  operationSearchIndex,
			"params": searchParams,
		})
	}
	// Query.
	qb := m.dialect.From(goqu.T("operations")).
		Select(goqu.I("operations.id"),
			goqu.I("operations.title"),
			goqu.I("operations.description"),
			goqu.I("operations.start_ts"),
			goqu.I("operations.end_ts"),
			goqu.I("operations.is_archived"))
	if operationFilters.ForUser.Valid { // Safety.
		qb = qb.Where(goqu.I("operations.id").In(m.dialect.From(goqu.T("operation_members")).
			Select(goqu.I("operation_members.operation")).
			Where(goqu.I("operation_members.user").Eq(operationFilters.ForUser.UUID))))
	}
	q, _, err := pgutil.QueryWithOrdinalityUUID(qb, goqu.I("operations.id"), resultUUIDs.Hits).ToSQL()
	if err != nil {
		return search.Result[Operation]{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return search.Result[Operation]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	operations := make([]Operation, 0, len(resultUUIDs.Hits))
	for rows.Next() {
		var operation Operation
		err = rows.Scan(&operation.ID,
			&operation.Title,
			&operation.Description,
			&operation.Start,
			&operation.End,
			&operation.IsArchived)
		if err != nil {
			return search.Result[Operation]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		operations = append(operations, operation)
	}
	return search.ResultFromResult(resultUUIDs, operations), nil
}

// RebuildOperationSearch rebuilds the operation-search.
func (m *Mall) RebuildOperationSearch(ctx context.Context, tx pgx.Tx) error {
	err := m.searchClient.SafeRebuildIndex(ctx, tx, operationSearchIndex)
	if err != nil {
		return meh.Wrap(err, "safe rebuild index", nil)
	}
	return nil
}
