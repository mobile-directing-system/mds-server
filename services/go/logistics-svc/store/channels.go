package store

import (
	"context"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"time"
)

// ChannelType in Channel.Type that specifies the type as well as the content of
// Channel.Details.
type ChannelType string

// ChannelDetails in Channel.Details, that can also be validated.
type ChannelDetails interface {
	entityvalidation.Validatable
}

// Channel is an abstraction of a way to communicate with a user. It has a
// unique priority, label and uses a specific ChannelType which's details can be
// found in Details.
type Channel struct {
	// ID identifies the channel.
	ID uuid.UUID
	// Entry is the id of the entry the channel is assigned to.
	Entry uuid.UUID
	// Label of the channel for better human-readability.
	Label string
	// Type of the channel.
	Type ChannelType
	// Priority is a unique priority of the channel.
	Priority int32
	// MinImportance of information in order to use this channel.
	MinImportance float64
	// Details for the channel, for example PhoneCallChannelDetails, based on Type.
	Details ChannelDetails
	// Timeout is the timeout when delivery over this channel timed out.
	Timeout time.Duration
}

// Validate that Type is known as well as Details.
func (c Channel) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	// Validate type.
	if _, ok := ChannelTypeSupplier.ChannelTypes[c.Type]; !ok {
		report.AddError(fmt.Sprintf("unknown channel type: %v", c.Type))
	}
	// Validate details.
	detailsReport, err := c.Details.Validate()
	if err != nil {
		return entityvalidation.Report{}, meh.Wrap(err, "validate details", nil)
	}
	report.Include(detailsReport)
	// Validate set timeout.
	if c.Timeout <= 0 {
		report.AddError("timeout must be greater zero")
	}
	return report, nil
}

// channelOperator is used in Mall in order to map common operations for each
// ChannelType.
type channelOperator interface {
	// deleteDetailsByChannel deletes all details associated with the channel with
	// the given id.
	deleteDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) error
	// setChannelDetailsByChannel sets and associates the given details with the
	// given with the given id.
	setChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID, detailsRaw ChannelDetails) error
	// getChannelDetailsByChannel retrieves the details associated with the channel
	// with the given id.
	getChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (ChannelDetails, error)
}

// TODO: maybe adjust channel operators to provide some way of querying with multiple ids?

// TODO: If entry is a user-entry, forbid setting label manually (or changing)
//  as this should always be the user's first- and lastname.

// TODO: addressbuch sollte später auch routing machen.

// ChannelsByAddressBookEntry retrieves all channels for the address book entry
// with the given id.
func (m *Mall) ChannelsByAddressBookEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) ([]Channel, error) {
	err := m.AssureAddressBookEntryExists(ctx, tx, entryID)
	if err != nil {
		return nil, meh.Wrap(err, "assure address book entry exists", meh.Details{"entry_id": entryID})
	}
	// Retrieve channel metadata.
	channels, err := m.channelMetadataByAddressBookEntry(ctx, tx, entryID)
	if err != nil {
		return nil, meh.Wrap(err, "channel metadata by entry", meh.Details{"entry_id": entryID})
	}
	// Retrieve channel details.
	channelDetails, err := m.channelDetailsForChannels(ctx, tx, channels)
	if err != nil {
		return nil, meh.Wrap(err, "channel details for channels", nil)
	}
	for i, channel := range channels {
		details, ok := channelDetails[channel.ID]
		if !ok {
			return nil, meh.NewInternalErr("missing channel details", meh.Details{"channel_id": channel.ID})
		}
		channel.Details = details
		channels[i] = channel
	}
	return channels, nil
}

// AssureAddressBookEntryExists makes sure that the address book entry with the
// given id exists.
func (m *Mall) AssureAddressBookEntryExists(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	q, _, err := m.dialect.From(goqu.T("address_book_entries")).
		Select(goqu.C("id")).
		Where(goqu.C("id").Eq(entryID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	if !rows.Next() {
		return meh.NewNotFoundErr("not found", nil)
	}
	return nil
}

// channelMetadataByAddressBookEntry retrieves a Channel list for the entry with the given
// id, but only containing metadata, so no Channel.Details.
//
// Warning: channelMetadataByAddressBookEntry does NOT perform entry existence checks!
func (m *Mall) channelMetadataByAddressBookEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) ([]Channel, error) {
	q, _, err := m.dialect.From(goqu.T("channels")).
		Select(goqu.C("id"),
			goqu.C("entry"),
			goqu.C("label"),
			goqu.C("type"),
			goqu.C("priority"),
			goqu.C("min_importance"),
			goqu.C("timeout")).
		Where(goqu.C("entry").Eq(entryID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	channels := make([]Channel, 0)
	for rows.Next() {
		var channel Channel
		err = rows.Scan(&channel.ID,
			&channel.Entry,
			&channel.Label,
			&channel.Type,
			&channel.Priority,
			&channel.MinImportance,
			&channel.Timeout)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

// channelDetailsForChannels retrieves details for the given channels. They are
// returned in a map by channel id.
func (m *Mall) channelDetailsForChannels(ctx context.Context, tx pgx.Tx, channels []Channel) (map[uuid.UUID]ChannelDetails, error) {
	channelDetails := make(map[uuid.UUID]ChannelDetails, len(channels))
	if len(channels) == 0 {
		return channelDetails, nil
	}
	// For each channel, query details by the corresponding type.
	for _, channel := range channels {
		channelOperator, ok := m.channelOperators[channel.Type]
		if !ok {
			return nil, meh.NewInternalErr("missing channel operator for channel type",
				meh.Details{"channel_type": channel.Type})
		}
		details, err := channelOperator.getChannelDetailsByChannel(ctx, tx, channel.ID)
		if err != nil {
			return nil, meh.Wrap(err, "channel details by channel", meh.Details{
				"channel_id":   channel.ID,
				"channel_type": channel.Type,
			})
		}
		channelDetails[channel.ID] = details
	}
	return channelDetails, nil
}

// DeleteChannelWithDetailsByID deletes the channel with the given id and type.
// This is meant to be used as a "shortcut" for clearing channel details as
// well. This is why we expect the ChannelType as well without querying it
// again.
func (m *Mall) DeleteChannelWithDetailsByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID, channelType ChannelType) error {
	// Clear channel details.
	operator, ok := m.channelOperators[channelType]
	if !ok {
		return meh.NewInternalErr("missing channel operator", meh.Details{"channel_type": channelType})
	}
	err := operator.deleteDetailsByChannel(ctx, tx, channelID)
	if err != nil {
		return meh.Wrap(err, "delete channel details", meh.Details{
			"channel_id":   channelID,
			"channel_type": channelType,
		})
	}
	// Clear channel itself.
	q, _, err := m.dialect.Delete(goqu.T("channels")).
		Where(goqu.C("id").Eq(channelID)).ToSQL()
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

// CreateChannelWithDetails creates the given Channel with its details.
//
// Warning: No entry existence checks are performed!
func (m *Mall) CreateChannelWithDetails(ctx context.Context, tx pgx.Tx, channel Channel) error {
	// Create channel itself.
	q, _, err := m.dialect.Insert(goqu.T("channels")).Rows(goqu.Record{
		"entry":          channel.Entry,
		"label":          channel.Label,
		"type":           channel.Type,
		"priority":       channel.Priority,
		"min_importance": channel.MinImportance,
		"timeout":        channel.Timeout,
	}).Returning(goqu.C("id")).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return mehpg.NewQueryDBErr(err, "exec query", q)
		}
		return meh.NewInternalErr("no rows returned", meh.Details{"query": q})
	}
	err = rows.Scan(&channel.ID)
	if err != nil {
		return mehpg.NewScanRowsErr(err, "scan row", q)
	}
	rows.Close()
	// Set details.
	operator, ok := m.channelOperators[channel.Type]
	if !ok {
		return meh.NewInternalErr("missing channel operator", meh.Details{"channel_type": channel.Type})
	}
	err = operator.setChannelDetailsByChannel(ctx, tx, channel.ID, channel.Details)
	if err != nil {
		return meh.Wrap(err, "set channel details", meh.Details{
			"channel_id":      channel.ID,
			"channel_details": channel.Details,
		})
	}
	return nil
}

// UpdateChannelWithDetails updates the given Channel with its details.
//
// Warning: No entry existence checks are performed!
func (m *Mall) UpdateChannelWithDetails(ctx context.Context, tx pgx.Tx, channel Channel) error {
	// Set details.
	operator, ok := m.channelOperators[channel.Type]
	if !ok {
		return meh.NewInternalErr("missing channel operator", meh.Details{"channel_type": channel.Type})
	}
	err := operator.setChannelDetailsByChannel(ctx, tx, channel.ID, channel.Details)
	if err != nil {
		return meh.Wrap(err, "set channel details", meh.Details{
			"channel_id":      channel.ID,
			"channel_details": channel.Details,
		})
	}
	// Update channel itself.
	q, _, err := m.dialect.Update(goqu.T("channels")).Set(goqu.Record{
		"entry":          channel.Entry,
		"label":          channel.Label,
		"type":           channel.Type,
		"priority":       channel.Priority,
		"min_importance": channel.MinImportance,
		"timeout":        channel.Timeout,
	}).Where(goqu.C("id").Eq(channel.ID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewInternalErr("channel must exist but did not", nil)
	}
	return nil
}

// ChannelMetadataByID retrieves a Channel by its id without details.
func (m *Mall) ChannelMetadataByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (Channel, error) {
	q, _, err := m.dialect.From(goqu.T("channels")).
		Select(goqu.C("id"),
			goqu.C("entry"),
			goqu.C("label"),
			goqu.C("type"),
			goqu.C("priority"),
			goqu.C("min_importance"),
			goqu.C("timeout")).
		Where(goqu.C("id").Eq(channelID)).ToSQL()
	if err != nil {
		return Channel{}, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return Channel{}, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	var channel Channel
	if !rows.Next() {
		return Channel{}, meh.NewNotFoundErr("not found", nil)
	}
	err = rows.Scan(&channel.ID,
		&channel.Entry,
		&channel.Label,
		&channel.Type,
		&channel.Priority,
		&channel.MinImportance,
		&channel.Timeout)
	if err != nil {
		return Channel{}, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return channel, nil
}
