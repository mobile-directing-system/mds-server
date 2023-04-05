package store

import (
	"context"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
)

const abEntrySearchIndex search.Index = "address-book-entries"

const abEntrySearchAttrID search.Attribute = "id"
const abEntrySearchAttrLabel search.Attribute = "label"
const abEntrySearchAttrDescription search.Attribute = "description"
const abEntrySearchAttrOperationID search.Attribute = "operation_id"
const abEntrySearchAttrOperationTitle search.Attribute = "operation_title"
const abEntrySearchAttrOperationIsArchived search.Attribute = "operation_is_archived"
const abEntrySearchAttrUserID search.Attribute = "user_id"
const abEntrySearchAttrUserUsername search.Attribute = "user_username"
const abEntrySearchAttrUserFirstName search.Attribute = "user_first_name"
const abEntrySearchAttrUserLastName search.Attribute = "user_last_name"
const abEntrySearchAttrVisibleBy search.Attribute = "visible_by"
const abEntrySearchAttrUserIsActive search.Attribute = "use_is_active"

var abEntrySearchIndexConfig = search.IndexConfig{
	PrimaryKey: abEntrySearchAttrID,
	Searchable: []search.Attribute{
		abEntrySearchAttrLabel,
		abEntrySearchAttrDescription,
		abEntrySearchAttrUserLastName,
		abEntrySearchAttrUserFirstName,
		abEntrySearchAttrUserUsername,
		abEntrySearchAttrOperationTitle,
	},
	Filterable: []search.Attribute{
		abEntrySearchAttrUserID,
		abEntrySearchAttrOperationID,
		abEntrySearchAttrOperationIsArchived,
		abEntrySearchAttrUserIsActive,
		abEntrySearchAttrVisibleBy,
	},
	Sortable: nil,
}

// AddressBookEntryDetailed extends AddressBookEntry with user details if
// AddressBookEntry.User is set.
type AddressBookEntryDetailed struct {
	AddressBookEntry
	UserDetails nulls.JSONNullable[User]
}

// AddressBookEntry for a User that may optionally be assigned to an operation.
type AddressBookEntry struct {
	// ID identifies the entry.
	ID uuid.UUID
	// Label for better human-readability.
	Label string
	// Description for better human-readability.
	//
	// Example use-case: Multiple entries for high-rank groups are created. However,
	// each one targets slightly different people. This can be used in order to pick
	// the right one.
	Description string
	// Operation holds the id of an optionally assigned operation.
	Operation uuid.NullUUID
	// User is the id of an optionally assigned user.
	User uuid.NullUUID
}

// Validate that Label is not empty.
func (entry AddressBookEntry) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	if entry.Label == "" {
		report.AddError("label must not be empty")
	}
	return report, nil
}

// documentFromAddressBookEntry generates a search.Document for an address book
// entry from given details.
func documentFromAddressBookEntry(entry AddressBookEntry, user nulls.JSONNullable[User], operation nulls.JSONNullable[Operation],
	visibleBy []uuid.UUID) search.Document {
	d := search.Document{
		abEntrySearchAttrID:          entry.ID,
		abEntrySearchAttrLabel:       entry.Label,
		abEntrySearchAttrDescription: entry.Description,
	}
	// Fill user details.
	if user.Valid {
		user := user.V
		d[abEntrySearchAttrUserID] = user.ID
		d[abEntrySearchAttrUserUsername] = user.Username
		d[abEntrySearchAttrUserFirstName] = user.FirstName
		d[abEntrySearchAttrUserLastName] = user.LastName
		d[abEntrySearchAttrUserIsActive] = user.IsActive
		d[abEntrySearchAttrVisibleBy] = visibleBy
	}
	// Fill operation details.
	if operation.Valid {
		operation := operation.V
		d[abEntrySearchAttrOperationID] = operation.ID
		d[abEntrySearchAttrOperationTitle] = operation.Title
		d[abEntrySearchAttrOperationIsArchived] = operation.IsArchived
	}
	return d
}

// AddressBookEntryFilters are used in AddressBookEntries for advanced filtering
// besides pagination.
type AddressBookEntryFilters struct {
	// ByUser retrieves only entries that are assigned to the user with this id.
	ByUser uuid.NullUUID
	// ForOperation excludes entries that are assigned to an operation not being the
	// one with this id.
	ForOperation uuid.NullUUID
	// ExcludeGlobal excludes entries that are assigned to no operation.
	ExcludeGlobal bool
	// VisibleBy limits all entries to members that are colleagues of the user with
	// this id. This means, that if entries have an associated user and this user is
	// not part of any operation, this client is part of, it will be hidden.
	VisibleBy uuid.NullUUID
	// IncludeForInactiveUsers includes entries, associated with inactive users.
	IncludeForInactiveUsers bool
	// AutoDeliveryEnabled excludes entries with auto-delivery being
	// enabled/disabled.
	AutoDeliveryEnabled nulls.Bool
}

// documentFromAddressBookEntryByID generates the search.Document from the
// database for the address book entry with the given id.
func (m *Mall) documentFromAddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) (search.Document, error) {
	// Retrieve entry.
	entry, err := m.AddressBookEntryByID(ctx, tx, entryID, uuid.NullUUID{})
	if err != nil {
		return nil, meh.Wrap(err, "address book entry by id", meh.Details{"entry_id": entryID})
	}
	// Retrieve user details.
	var user nulls.JSONNullable[User]
	var visibleBy []uuid.UUID
	if entry.User.Valid {
		userDetails, err := m.UserByID(ctx, tx, entry.User.UUID)
		if err != nil {
			return nil, meh.NewInternalErrFromErr(err, "user by id", meh.Details{"user_id": entry.User.UUID})
		}
		user = nulls.NewJSONNullable(userDetails)
		visibleBy, err = m.visibleByForUser(ctx, tx, entry.User.UUID)
		if err != nil {
			return nil, meh.NewInternalErrFromErr(err, "visible-by for user", meh.Details{"user_id": entry.User.UUID})
		}
	}
	// Retrieve operation details.
	var operation nulls.JSONNullable[Operation]
	if entry.Operation.Valid {
		operationDetails, err := m.OperationByID(ctx, tx, entry.Operation.UUID)
		if err != nil {
			return nil, meh.NewInternalErrFromErr(err, "operation by id", meh.Details{"operation_id": entry.Operation.UUID})
		}
		operation = nulls.NewJSONNullable(operationDetails)
	}
	return documentFromAddressBookEntry(entry.AddressBookEntry, user, operation, visibleBy), nil
}

// addOrUpdateAddressBookEntryInSearch adds or updates the address book entry
// with the given id in the search.
func (m *Mall) addOrUpdateAddressBookEntryInSearch(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	d, err := m.documentFromAddressBookEntryByID(ctx, tx, entryID)
	if err != nil {
		return meh.Wrap(err, "document from address book entry by id", meh.Details{"entry_id": entryID})
	}
	err = m.searchClient.SafeAddOrUpdateDocument(ctx, tx, abEntrySearchIndex, d)
	if err != nil {
		return meh.Wrap(err, "safe add or update document in search", meh.Details{
			"index":        abEntrySearchIndex,
			"new_document": d,
		})
	}
	return nil
}

// visibleByForUser retrieves the id list of users who are
// member of any operation, the associated user of the entry is member of.
func (m *Mall) visibleByForUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]uuid.UUID, error) {
	q, _, err := m.dialect.From(goqu.T("operation_members").As("others")).
		Select(goqu.I("others.user")).
		Where(goqu.I("others.user").Neq(userID),
			goqu.I("others.operation").In(
				// Retrieve ids of operations, the user is member of.
				m.dialect.From(goqu.T("operation_members").As("entry_member")).
					Select(goqu.I("entry_member.operation")).
					Where(goqu.I("entry_member.user").Eq(userID)))).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	visibleBy := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		err = rows.Scan(&id)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		visibleBy = append(visibleBy, id)
	}
	rows.Close()
	return visibleBy, nil
}

// AddressBookEntries retrieves a paginated AddressBookEntryDetailed list using
// the given AddressBookEntryFilters and pagination.Params.
func (m *Mall) AddressBookEntries(ctx context.Context, tx pgx.Tx, filters AddressBookEntryFilters,
	paginationParams pagination.Params) (pagination.Paginated[AddressBookEntryDetailed], error) {
	entries := make([]AddressBookEntryDetailed, 0)
	// Retrieve entries.
	qb := m.dialect.From(goqu.T("address_book_entries")).
		LeftJoin(goqu.T("users"), goqu.On(goqu.I("users.id").Eq(goqu.I("address_book_entries.user")))).
		Select(goqu.I("address_book_entries.id"),
			goqu.I("address_book_entries.label"),
			goqu.I("address_book_entries.description"),
			goqu.I("address_book_entries.operation"),
			goqu.I("address_book_entries.user"))
	whereAnd := make([]exp.Expression, 0)
	// Apply filters.
	if filters.ByUser.Valid {
		whereAnd = append(whereAnd, goqu.I("address_book_entries.user").Eq(filters.ByUser.UUID))
	}
	if filters.ForOperation.Valid {
		whereAnd = append(whereAnd, goqu.Or(goqu.I("address_book_entries.operation").Eq(filters.ForOperation.UUID)),
			goqu.I("address_book_entries.operation").IsNull())
	}
	if filters.ExcludeGlobal {
		whereAnd = append(whereAnd, goqu.I("address_book_entries.operation").IsNotNull())
	}
	if filters.VisibleBy.Valid {
		whereAnd = append(whereAnd, goqu.Or(goqu.I("address_book_entries.user").IsNull(),
			goqu.C("user").In(m.dialect.From(goqu.T("operation_members")).As("visible_by_op_members").
				Select(goqu.I("visible_by_op_members.user")).
				Where(goqu.I("visible_by_op_members.operation").
					In(m.dialect.From(goqu.T("operation_members")).As("visible_by_op_members_c_opm").
						Select(goqu.I("visible_by_op_members_c_opm.operation")).
						Where(goqu.I("visible_by_op_members_c_opm.user").Eq(filters.VisibleBy.UUID)))))))
	}
	if !filters.IncludeForInactiveUsers {
		whereAnd = append(whereAnd, goqu.Or(goqu.I("address_book_entries.user").IsNull(),
			goqu.I("users.is_active").IsTrue()))
	}
	if filters.AutoDeliveryEnabled.Valid {
		whereAnd = append(whereAnd, goqu.I("address_book_entries.id").
			In(m.dialect.From(goqu.T("auto_intel_delivery_address_book_entries")).
				Select(goqu.I("auto_intel_delivery_address_book_entries.entry"))))
	}
	if len(whereAnd) > 0 {
		qb = qb.Where(goqu.And(whereAnd...))
	}
	q, _, err := pagination.QueryToSQLWithPagination(qb, paginationParams, pagination.FieldMap{
		"label":       goqu.I("address_book_entries.label"),
		"description": goqu.I("address_book_entries.description"),
	})
	if err != nil {
		return pagination.Paginated[AddressBookEntryDetailed]{}, meh.NewInternalErrFromErr(err, "query with pagination to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return pagination.Paginated[AddressBookEntryDetailed]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	var total int
	for rows.Next() {
		var entry AddressBookEntry
		err = rows.Scan(&entry.ID,
			&entry.Label,
			&entry.Description,
			&entry.Operation,
			&entry.User,
			&total)
		if err != nil {
			return pagination.Paginated[AddressBookEntryDetailed]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		entries = append(entries, AddressBookEntryDetailed{AddressBookEntry: entry})
	}
	// Add user details.
	userIDs := make([]uuid.UUID, 0)
	for _, entry := range entries {
		if entry.User.Valid {
			userIDs = append(userIDs, entry.User.UUID)
		}
	}
	userDetailsForAll, err := m.usersByIDs(ctx, tx, userIDs)
	if err != nil {
		return pagination.Paginated[AddressBookEntryDetailed]{}, meh.Wrap(err, "users by ids",
			meh.Details{"user_ids": userIDs})
	}
	for i, entry := range entries {
		if !entry.User.Valid {
			continue
		}
		userDetails, ok := userDetailsForAll[entry.User.UUID]
		if !ok {
			return pagination.Paginated[AddressBookEntryDetailed]{}, meh.NewInternalErr("missing user details",
				meh.Details{
					"entry_id": entry.ID,
					"user_id":  entry.User.UUID,
				})
		}
		entry.UserDetails = nulls.NewJSONNullable(userDetails)
		entries[i] = entry
	}
	return pagination.NewPaginated(paginationParams, entries, total), nil
}

// AddressBookEntryByID retrieves the AddressBookEntryDetailed with the given
// id. If visible-by is given, an meh.ErrNotFound will be returned, if the entry
// is associated with a user, that is not part of any operation, the client
// (visible-by) is part of.
func (m *Mall) AddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID, visibleBy uuid.NullUUID) (AddressBookEntryDetailed, error) {
	var entry AddressBookEntryDetailed
	// Retrieve metadata.
	entryMetadata, err := m.addressBookEntryMetadataByID(ctx, tx, entryID)
	if err != nil {
		return AddressBookEntryDetailed{}, meh.Wrap(err, "entry metadata by id", meh.Details{"entry_id": entryID})
	}
	entry.AddressBookEntry = entryMetadata
	// Assure visible-by.
	if visibleBy.Valid && entryMetadata.User.Valid {
		isVisibleBy, err := m.isUserVisibleBy(ctx, tx, entryMetadata.User.UUID, visibleBy.UUID)
		if err != nil {
			return AddressBookEntryDetailed{}, meh.Wrap(err, "is user visible by", meh.Details{
				"user_id":    entryMetadata.User.UUID,
				"visible_by": visibleBy.UUID,
			})
		}
		if !isVisibleBy {
			return AddressBookEntryDetailed{}, meh.NewNotFoundErr("not visible", meh.Details{"visible_by": isVisibleBy})
		}
	}
	// Retrieve user details if user set.
	if entry.User.Valid {
		userDetails, err := m.UserByID(ctx, tx, entry.User.UUID)
		if err != nil {
			return AddressBookEntryDetailed{}, meh.NewInternalErrFromErr(err, "user by id", meh.Details{"user_id": entry.User.UUID})
		}
		entry.UserDetails = nulls.NewJSONNullable(userDetails)
	}
	return entry, nil
}

// addressBookEntryMetadataByID retrieves all AddressBookEntry-details without
// channels.
func (m *Mall) addressBookEntryMetadataByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) (AddressBookEntry, error) {
	entryQuery, _, err := m.dialect.From(goqu.T("address_book_entries")).
		Select(goqu.C("id"),
			goqu.C("label"),
			goqu.C("description"),
			goqu.C("operation"),
			goqu.C("user")).
		Where(goqu.C("id").Eq(entryID)).ToSQL()
	if err != nil {
		return AddressBookEntry{}, meh.NewInternalErrFromErr(err, "entry-query to sql", nil)
	}
	entryRows, err := tx.Query(ctx, entryQuery)
	if err != nil {
		return AddressBookEntry{}, mehpg.NewQueryDBErr(err, "exec entry-query", entryQuery)
	}
	defer entryRows.Close()
	if !entryRows.Next() {
		return AddressBookEntry{}, meh.NewNotFoundErr("entry not found", nil)
	}
	var entry AddressBookEntry
	err = entryRows.Scan(&entry.ID,
		&entry.Label,
		&entry.Description,
		&entry.Operation,
		&entry.User)
	if err != nil {
		return AddressBookEntry{}, mehpg.NewScanRowsErr(err, "scan entry-row", entryQuery)
	}
	return entry, nil
}

// isUserVisibleBy checks if the given user is part of any operation, the
// visible-by user is part of as well.
func (m *Mall) isUserVisibleBy(ctx context.Context, tx pgx.Tx, userID uuid.UUID, visibleBy uuid.UUID) (bool, error) {
	q, _, err := m.dialect.From(goqu.T("operation_members").As("opm1")).
		Select(goqu.COUNT("*")).
		Where(goqu.I("opm1.user").Eq(userID),
			goqu.I("opm1.operation").
				In(m.dialect.From(goqu.T("operation_members").As("opm2")).
					Select(goqu.I("opm2.operation")).
					Where(goqu.I("opm2.user").Eq(visibleBy)))).ToSQL()
	if err != nil {
		return false, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return false, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return false, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	var count int
	err = rows.Scan(&count)
	if err != nil {
		return false, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return count > 0, nil
}

// CreateAddressBookEntry creates the given AddressBookEntry.
func (m *Mall) CreateAddressBookEntry(ctx context.Context, tx pgx.Tx, entry AddressBookEntry) (AddressBookEntryDetailed, error) {
	// Create.
	q, _, err := m.dialect.Insert(goqu.T("address_book_entries")).Rows(goqu.Record{
		"label":       entry.Label,
		"description": entry.Description,
		"operation":   entry.Operation,
		"user":        entry.User,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return AddressBookEntryDetailed{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return AddressBookEntryDetailed{}, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return AddressBookEntryDetailed{}, mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return AddressBookEntryDetailed{}, meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	err = rows.Scan(&entry.ID)
	if err != nil {
		return AddressBookEntryDetailed{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	created := AddressBookEntryDetailed{
		AddressBookEntry: entry,
	}
	rows.Close()
	// Retrieve user details.
	if entry.User.Valid {
		userDetails, err := m.UserByID(ctx, tx, entry.User.UUID)
		if err != nil {
			return AddressBookEntryDetailed{}, meh.NewInternalErrFromErr(err, "user by id",
				meh.Details{"user_id": entry.User.UUID})
		}
		created.UserDetails = nulls.NewJSONNullable(userDetails)
	}
	// Add to search.
	err = m.addOrUpdateAddressBookEntryInSearch(ctx, tx, created.ID)
	if err != nil {
		return AddressBookEntryDetailed{}, meh.Wrap(err, "add or update in search", nil)
	}
	return created, nil
}

// UpdateAddressBookEntry updates the given AddressBookEntry, identified by its
// id.
func (m *Mall) UpdateAddressBookEntry(ctx context.Context, tx pgx.Tx, entry AddressBookEntry) error {
	q, _, err := m.dialect.Update(goqu.T("address_book_entries")).Set(goqu.Record{
		"label":       entry.Label,
		"description": entry.Description,
		"operation":   entry.Operation,
		"user":        entry.User,
	}).Where(goqu.C("id").Eq(entry.ID)).ToSQL()
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
	err = m.addOrUpdateAddressBookEntryInSearch(ctx, tx, entry.ID)
	if err != nil {
		return meh.Wrap(err, "add or update in search", nil)
	}
	return nil
}

// DeleteAddressBookEntryByID deletes the address book entry with the given id.
func (m *Mall) DeleteAddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	q, _, err := m.dialect.Delete(goqu.T("address_book_entries")).
		Where(goqu.C("id").Eq(entryID)).ToSQL()
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
	// Delete in search.
	err = m.searchClient.SafeDeleteDocumentByUUID(ctx, tx, abEntrySearchIndex, entryID)
	if err != nil {
		return meh.Wrap(err, "safe delete document in search", meh.Details{"id": entryID})
	}
	return nil
}

// SearchAddressBookEntries with the given AddressBookEntryFilters and
// search.Params.
func (m *Mall) SearchAddressBookEntries(ctx context.Context, tx pgx.Tx, filters AddressBookEntryFilters,
	searchParams search.Params) (search.Result[AddressBookEntryDetailed], error) {
	// Search.
	var searchFilters [][]string
	if filters.ByUser.Valid {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%s = '%s'", abEntrySearchAttrUserID, filters.ByUser.UUID.String()),
		})
	}
	if filters.ForOperation.Valid {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%s = null", abEntrySearchAttrOperationID),
			fmt.Sprintf("%s = '%s'", abEntrySearchAttrOperationID, filters.ForOperation.UUID.String()),
		})
	}
	if filters.ExcludeGlobal {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%s != null", abEntrySearchAttrOperationID),
		})
	}
	if filters.VisibleBy.Valid {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%s = '%s'", abEntrySearchAttrVisibleBy, filters.VisibleBy.UUID.String()),
		})
	}
	if !filters.IncludeForInactiveUsers {
		searchFilters = append(searchFilters, []string{
			fmt.Sprintf("%s = null", abEntrySearchAttrUserID),
			fmt.Sprintf("%s = true", abEntrySearchAttrUserIsActive),
		})
	}
	if filters.AutoDeliveryEnabled.Valid {
		return search.Result[AddressBookEntryDetailed]{},
			meh.NewBadInputErr("auto-delivery-enabled-filter is not available in search", nil)
	}
	resultUUIDs, err := search.UUIDSearch(m.searchClient, abEntrySearchIndex, searchParams, search.Request{
		Filter: searchFilters,
	})
	if err != nil {
		return search.Result[AddressBookEntryDetailed]{}, meh.Wrap(err, "uuid search", meh.Details{
			"index":  abEntrySearchIndex,
			"params": searchParams,
			"filter": searchFilters,
		})
	}
	// Query.
	q, _, err := pgutil.QueryWithOrdinalityUUID(m.dialect.From(goqu.T("address_book_entries").As("entries")).
		LeftJoin(goqu.T("users"), goqu.On(goqu.I("users.id").Eq(goqu.I("entries.user")))).
		Select(goqu.I("entries.id"),
			goqu.I("entries.label"),
			goqu.I("entries.description"),
			goqu.I("entries.operation"),
			goqu.I("entries.user"),
			goqu.I("users.username"),
			goqu.I("users.first_name"),
			goqu.I("users.last_name"),
			goqu.I("users.is_active")), goqu.I("entries.id"), resultUUIDs.Hits).ToSQL()
	if err != nil {
		return search.Result[AddressBookEntryDetailed]{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return search.Result[AddressBookEntryDetailed]{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	entries := make([]AddressBookEntryDetailed, 0, len(resultUUIDs.Hits))
	for rows.Next() {
		var entry AddressBookEntryDetailed
		var userUsername nulls.String
		var userFirstName nulls.String
		var userLastName nulls.String
		var userIsActive nulls.Bool
		err = rows.Scan(&entry.ID,
			&entry.Label,
			&entry.Description,
			&entry.Operation,
			&entry.User,
			&userUsername,
			&userFirstName,
			&userLastName,
			&userIsActive)
		if err != nil {
			return search.Result[AddressBookEntryDetailed]{}, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		// Set optional user.
		if entry.User.Valid {
			entry.UserDetails = nulls.NewJSONNullable(User{
				ID:        entry.User.UUID,
				Username:  userUsername.String,
				FirstName: userFirstName.String,
				LastName:  userLastName.String,
				IsActive:  userIsActive.Bool,
			})
		}
		entries = append(entries, entry)
	}
	rows.Close()
	return search.ResultFromResult(resultUUIDs, entries), nil
}

// RebuildAddressBookEntrySearch rebuilds the address-book-entry-search.
func (m *Mall) RebuildAddressBookEntrySearch(ctx context.Context, tx pgx.Tx) error {
	err := m.searchClient.SafeRebuildIndex(ctx, tx, abEntrySearchIndex)
	if err != nil {
		return meh.Wrap(err, "safe rebuild index", nil)
	}
	return nil
}

// UsersWithDeliveriesByIntel retrieves all users having associated address book
// entries with deliveries for the intel with the given id.
func (m *Mall) UsersWithDeliveriesByIntel(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) ([]uuid.UUID, error) {
	q, _, err := m.dialect.From(goqu.T("intel_deliveries")).
		InnerJoin(goqu.T("address_book_entries"),
			goqu.On(goqu.I("address_book_entries.id").Eq(goqu.I("intel_deliveries.to")))).
		Select(goqu.DISTINCT(goqu.I("address_book_entries.user"))).
		Where(goqu.I("intel_deliveries.intel").Eq(intelID),
			goqu.I("address_book_entries.user").IsNotNull()).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	userIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var userID uuid.UUID
		err = rows.Scan(&userID)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

// associatedUsersByAddressBookEntries retrieves a list of user ids, associated
// with the given address book entries.
func (m *Mall) associatedUsersByAddressBookEntries(ctx context.Context, tx pgx.Tx, entryIDs []uuid.UUID) ([]uuid.UUID, error) {
	q, _, err := m.dialect.From(goqu.T("address_book_entries")).
		Select(goqu.DISTINCT(goqu.C("user"))).
		Where(goqu.C("id").In(entryIDs)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	users := make([]uuid.UUID, 0, len(entryIDs))
	for rows.Next() {
		var id uuid.UUID
		err = rows.Scan(&id)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		users = append(users, id)
	}
	return users, nil
}

// IsAutoDeliveryEnabledForAddressBookEntry checks whether the address book entry
// with the given id is marked for auto-delivery.
func (m *Mall) IsAutoDeliveryEnabledForAddressBookEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) (bool, error) {
	q, _, err := m.dialect.From(goqu.T("auto_intel_delivery_address_book_entries")).
		Select(goqu.C("entry")).
		Where(goqu.C("entry").Eq(entryID)).ToSQL()
	if err != nil {
		return false, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return false, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	enabled := rows.Next()
	return enabled, nil
}

// clearAutoDeliveryEnabledForAllAddressBookEntries removes all address book
// entries from the list with auto-delivery enabled ones. The ids of all address
// book entries are returned that have been cleared of auto delivery.
func (m *Mall) clearAutoDeliveryEnabledForAllAddressBookEntries(ctx context.Context, tx pgx.Tx) ([]uuid.UUID, error) {
	q, _, err := m.dialect.Delete(goqu.T("auto_intel_delivery_address_book_entries")).
		Returning(goqu.C("entry")).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	cleared := make([]uuid.UUID, 0)
	for rows.Next() {
		var entryID uuid.UUID
		err = rows.Scan(&entryID)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		cleared = append(cleared, entryID)
	}
	return cleared, nil
}

// SetAddressBookEntriesWithAutoDeliveryEnabled sets the list of address book
// entries with auto-delivery being enabled to the given ones. The returned list
// of ids are of address book entries that had previously auto delivery enabled,
// but now disabled.
func (m *Mall) SetAddressBookEntriesWithAutoDeliveryEnabled(ctx context.Context, tx pgx.Tx, entryIDs []uuid.UUID) ([]uuid.UUID, error) {
	// Clear all and collect entry ids for removing the 'new' ones.
	cleared, err := m.clearAutoDeliveryEnabledForAllAddressBookEntries(ctx, tx)
	if err != nil {
		return nil, meh.Wrap(err, "clear auto-delivery-enabled for all address book entries", nil)
	}
	clearedWithoutReenabled := make(map[uuid.UUID]struct{}, len(cleared))
	for _, entryID := range cleared {
		clearedWithoutReenabled[entryID] = struct{}{}
	}
	// Enabled auto delivery for the given entries.
	rows := make([]any, 0, len(entryIDs))
	for _, entryID := range entryIDs {
		rows = append(rows, goqu.Record{
			"entry": entryID,
		})
		delete(clearedWithoutReenabled, entryID)
	}
	if len(rows) == 0 {
		return cleared, nil
	}
	q, _, err := m.dialect.Insert("auto_intel_delivery_address_book_entries").Rows(rows...).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "exec query", q)
	}
	clearedClean := make([]uuid.UUID, 0, len(clearedWithoutReenabled))
	for entryID := range clearedWithoutReenabled {
		clearedClean = append(clearedClean, entryID)
	}
	return cleared, nil
}
