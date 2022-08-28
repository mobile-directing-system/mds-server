package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
)

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
	return nil
}
