package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"reflect"
)

// ChannelTypeForwardToGroup is used for forwarding to another group.
const ChannelTypeForwardToGroup ChannelType = "forward-to-group"

// ForwardToGroupChannelDetails holds Channel.Details for
// ChannelTypeForwardToGroup.
type ForwardToGroupChannelDetails struct {
	// ForwardToGroup is the id of the group that should be forwarded to.
	ForwardToGroup []uuid.UUID
}

// Validate assures no duplicate group ids in ForwardToGroup.
func (d ForwardToGroupChannelDetails) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	forwardToGroupMap := map[uuid.UUID]struct{}{}
	for _, groupID := range d.ForwardToGroup {
		if _, ok := forwardToGroupMap[groupID]; ok {
			report.AddError("duplicate group ids")
			break
		}
		forwardToGroupMap[groupID] = struct{}{}
	}
	return report, nil
}

// forwardToGroupChannelOperator is the channelOperator for channels with
// ChannelTypeForwardToGroup.
type forwardToGroupChannelOperator struct {
	m *Mall
}

func (op *forwardToGroupChannelOperator) deleteDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) error {
	q, _, err := op.m.dialect.Delete(goqu.T("forward_to_group_channel_entries")).
		Where(goqu.C("channel").Eq(channelID)).ToSQL()
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

func (op *forwardToGroupChannelOperator) setChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID,
	detailsRaw ChannelDetails) error {
	details, ok := detailsRaw.(ForwardToGroupChannelDetails)
	if !ok {
		return meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	// Delete.
	err := op.deleteDetailsByChannel(ctx, tx, channelID)
	if err != nil && meh.ErrorCode(err) != meh.ErrNotFound {
		return meh.Wrap(err, "delete details by channel", nil)
	}
	// Insert.
	records := make([]any, 0, len(details.ForwardToGroup))
	for _, groupID := range details.ForwardToGroup {
		records = append(records, goqu.Record{
			"channel":          channelID,
			"forward_to_group": groupID,
		})
	}
	if len(records) == 0 {
		return nil
	}
	q, _, err := op.m.dialect.Insert(goqu.T("forward_to_group_channel_entries")).
		Rows(records...).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

func (op *forwardToGroupChannelOperator) getChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (ChannelDetails, error) {
	// Build query.
	q, _, err := op.m.dialect.From(goqu.T("forward_to_group_channel_entries")).
		Where(goqu.C("channel").Eq(channelID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := tx.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	details := ForwardToGroupChannelDetails{
		ForwardToGroup: make([]uuid.UUID, 0),
	}
	for rows.Next() {
		var groupID uuid.UUID
		err = rows.Scan(&groupID)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		details.ForwardToGroup = append(details.ForwardToGroup, groupID)
	}
	return details, nil
}

// DeleteForwardToGroupChannelsByGroup deletes all channels with channel type
// ChannelTypeForwardToGroup, that forward to the group with the given id. It
// returns the list of affected address book entries.
func (m *Mall) DeleteForwardToGroupChannelsByGroup(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) ([]uuid.UUID, error) {
	// Find affected channels.
	affectedChannelsQuery, _, err := m.dialect.From(goqu.T("forward_to_group_channel_entries")).
		Select(goqu.DISTINCT(goqu.C("channel"))).
		Where(goqu.C("forward_to_group").Eq(groupID)).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "affected-channels-query to sql", nil)
	}
	rows, err := tx.Query(ctx, affectedChannelsQuery)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "exec affected-channels-query", affectedChannelsQuery)
	}
	defer rows.Close()
	affectedChannels := make([]uuid.UUID, 0)
	for rows.Next() {
		var c uuid.UUID
		err = rows.Scan(&c)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", affectedChannelsQuery)
		}
		affectedChannels = append(affectedChannels, c)
	}
	rows.Close()
	// Delete all of them.
	affectedEntries := make([]uuid.UUID, 0, len(affectedChannels))
	for _, affectedChannel := range affectedChannels {
		channelMetadata, err := m.ChannelMetadataByID(ctx, tx, affectedChannel)
		if err != nil {
			return nil, meh.Wrap(err, "channel metadata by id", meh.Details{"channel_id": affectedChannel})
		}
		affectedEntries = append(affectedEntries, channelMetadata.Entry)
		err = m.DeleteChannelWithDetailsByID(ctx, tx, affectedChannel, ChannelTypeForwardToGroup)
		if err != nil {
			return nil, meh.Wrap(err, "delete channel with details", meh.Details{"channel_id": affectedChannel})
		}
	}
	return affectedEntries, nil
}
