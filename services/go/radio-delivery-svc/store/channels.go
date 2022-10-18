package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"time"
)

// RadioChannel that is used for delivery-attempts.
type RadioChannel struct {
	// ID identifies the channel.
	ID uuid.UUID
	// Entry is the id of the referenced address book entry.
	Entry uuid.UUID
	// Label is a human-readable label for the channel.
	Label string
	// Timeout until delivery-attempts time out for this channel.
	Timeout time.Duration
	// Info for how to reach the recipient.
	Info string
}

// CreateRadioChannel creates the given RadioChannel in the database.
func (m *Mall) CreateRadioChannel(ctx context.Context, tx pgx.Tx, create RadioChannel) error {
	q, _, err := m.dialect.Insert(goqu.T("radio_channels")).Rows(goqu.Record{
		"id":      create.ID,
		"entry":   create.Entry,
		"label":   create.Label,
		"timeout": create.Timeout,
		"info":    create.Info,
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

// DeleteRadioChannelsByEntry deletes the radio channels in the database
// associated with the address book entry with the given id.
func (m *Mall) DeleteRadioChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	q, _, err := m.dialect.Delete(goqu.T("radio_channels")).
		Where(goqu.C("entry").Eq(entryID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// RadioChannelByID retrieves the RadioChannel from the database with the given
// id.
func (m *Mall) RadioChannelByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (RadioChannel, error) {
	q, _, err := m.dialect.From(goqu.T("radio_channels")).
		Select(goqu.C("id"),
			goqu.C("entry"),
			goqu.C("label"),
			goqu.C("timeout"),
			goqu.C("info")).
		Where(goqu.C("id").Eq(channelID)).ToSQL()
	if err != nil {
		return RadioChannel{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return RadioChannel{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return RadioChannel{}, meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	var channel RadioChannel
	err = rows.Scan(&channel.ID,
		&channel.Entry,
		&channel.Label,
		&channel.Timeout,
		&channel.Info)
	if err != nil {
		return RadioChannel{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	return channel, nil
}
