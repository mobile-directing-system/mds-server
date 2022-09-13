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

// NotificationChannel that is used for delivery-attempts.
type NotificationChannel struct {
	// ID identifies the channel.
	ID uuid.UUID
	// Entry is the id of the referenced address book entry.
	Entry uuid.UUID
	// Label is a human-readable label for the channel.
	Label string
	// Timeout until delivery-attempts time out for this channel.
	Timeout time.Duration
}

// CreateNotificationChannel creates the given NotificationChannel in the
// database.
func (m *Mall) CreateNotificationChannel(ctx context.Context, tx pgx.Tx, create NotificationChannel) error {
	q, _, err := m.dialect.Insert(goqu.T("notification_channels")).Rows(goqu.Record{
		"id":      create.ID,
		"entry":   create.Entry,
		"label":   create.Label,
		"timeout": create.Timeout,
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

// DeleteNotificationChannelsByEntry deletes the notification channels in the
// database associated with the address book entry with the given id.
func (m *Mall) DeleteNotificationChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	q, _, err := m.dialect.Delete(goqu.T("notification_channels")).
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

// NotificationChannelByID retrieves the NotificationChannel from the database
// with the given id.
func (m *Mall) NotificationChannelByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (NotificationChannel, error) {
	q, _, err := m.dialect.From(goqu.T("notification_channels")).
		Select(goqu.C("id"),
			goqu.C("entry"),
			goqu.C("label"),
			goqu.C("timeout")).
		Where(goqu.C("id").Eq(channelID)).ToSQL()
	if err != nil {
		return NotificationChannel{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return NotificationChannel{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return NotificationChannel{}, meh.NewNotFoundErr("not found", meh.Details{"query": q})
	}
	var channel NotificationChannel
	err = rows.Scan(&channel.ID,
		&channel.Entry,
		&channel.Label,
		&channel.Timeout)
	if err != nil {
		return NotificationChannel{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	return channel, nil
}
