package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
)

// User contains all stored user information.
type User struct {
	// ID identifies the user.
	ID uuid.UUID
	// Username for the user.
	Username string
	// FirstName of the user.
	FirstName string
	// LastName of the user.
	LastName string
	// IsActive describes whether the user is active (not deleted).
	IsActive bool
}

// TODO: For controller: on user update: Update labels of associated entries.

// TODO: for controller: on user delete: Delete associated entries as well as forward_to_user channels.

// UserByID retrieves the User with the given id.
func (m *Mall) UserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (User, error) {
	q, _, err := m.dialect.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("first_name"),
			goqu.C("last_name"),
			goqu.C("is_active")).
		Where(goqu.C("id").Eq(userID)).ToSQL()
	if err != nil {
		return User{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return User{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return User{}, meh.NewNotFoundErr("not found", nil)
	}
	var user User
	err = rows.Scan(&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsActive)
	if err != nil {
		return User{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return user, nil
}

// usersByIDs retrieves user details for the users with the given ids.
func (m *Mall) usersByIDs(ctx context.Context, tx pgx.Tx, userIDs []uuid.UUID) (map[uuid.UUID]User, error) {
	users := make(map[uuid.UUID]User, len(userIDs))
	if len(userIDs) == 0 {
		return users, nil
	}
	q, _, err := m.dialect.From(goqu.T("users")).
		Select(goqu.C("id"),
			goqu.C("username"),
			goqu.C("first_name"),
			goqu.C("last_name"),
			goqu.C("is_active")).
		Where(goqu.C("id").In(userIDs)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.IsActive)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		users[user.ID] = user
	}
	return users, nil
}

// CreateUser creates the given User.
func (m *Mall) CreateUser(ctx context.Context, tx pgx.Tx, user User) error {
	q, _, err := m.dialect.Insert(goqu.T("users")).Rows(goqu.Record{
		"id":         user.ID,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"is_active":  user.IsActive,
	}).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// UpdateUser updates the user details of the given User, identified by its
// User.ID.
func (m *Mall) UpdateUser(ctx context.Context, tx pgx.Tx, user User) error {
	q, _, err := m.dialect.Update(goqu.T("users")).Set(goqu.Record{
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"is_active":  user.IsActive,
	}).Where(goqu.C("id").Eq(user.ID)).ToSQL()
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
	// Update associated address book entries.
	entries, err := m.AddressBookEntries(ctx, tx, AddressBookEntryFilters{
		ByUser:                  nulls.NewUUID(user.ID),
		ExcludeGlobal:           true,
		IncludeForInactiveUsers: true,
	}, pagination.Params{Limit: 0})
	if err != nil {
		return meh.Wrap(err, "address book entries by user for search updated", nil)
	}
	for _, entry := range entries.Entries {
		err = m.addOrUpdateAddressBookEntryInSearch(ctx, tx, entry.ID)
		if err != nil {
			return meh.Wrap(err, "addor update address book entry in search", meh.Details{"entry_id": entry.ID})
		}
	}
	return nil
}
