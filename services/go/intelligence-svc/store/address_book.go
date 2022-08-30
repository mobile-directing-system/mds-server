package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
)

// AddressBookEntry for a User that may optionally be assigned to an operation.
type AddressBookEntry struct {
	// ID identifies the entry.
	ID uuid.UUID
	// Label for better human-readability.
	Label string
	// Description for better human-readability.
	Description string
	// Operation holds the id of an optionally assigned operation.
	Operation uuid.NullUUID
	// User is the id of an optionally assigned user.
	User uuid.NullUUID
}

// CreateAddressBookEntry creates the given AddressBookEntry.
func (m *Mall) CreateAddressBookEntry(ctx context.Context, tx pgx.Tx, entry AddressBookEntry) error {
	// Create.
	q, _, err := m.dialect.Insert(goqu.T("address_book_entries")).Rows(goqu.Record{
		"id":          entry.ID,
		"label":       entry.Label,
		"description": entry.Description,
		"operation":   entry.Operation,
		"user":        entry.User,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
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

// AddressBookEntryByID retrieves the store.AddressBookEntry with the given id.
func (m *Mall) AddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) (AddressBookEntry, error) {
	q, _, err := m.dialect.From(goqu.T("address_book_entries")).
		Select(goqu.C("id"),
			goqu.C("label"),
			goqu.C("description"),
			goqu.C("operation"),
			goqu.C("user")).
		Where(goqu.C("id").Eq(entryID)).ToSQL()
	if err != nil {
		return AddressBookEntry{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return AddressBookEntry{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return AddressBookEntry{}, meh.NewNotFoundErr("not found", nil)
	}
	var entry AddressBookEntry
	err = rows.Scan(&entry.ID,
		&entry.Label,
		&entry.Description,
		&entry.Operation,
		&entry.User)
	if err != nil {
		return AddressBookEntry{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	return entry, nil
}
