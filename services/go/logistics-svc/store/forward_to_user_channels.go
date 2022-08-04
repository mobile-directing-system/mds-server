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

// TODO: Message forwarding/routing to forward-to-user or -group with zero entries should always fail immediately!

// ChannelTypeForwardToUser is used for forwarding to another user.
const ChannelTypeForwardToUser ChannelType = "forward-to-user"

// ForwardToUserChannelDetails holds Channel.Details for
// ChannelTypeForwardToUser.
type ForwardToUserChannelDetails struct {
	// ForwardToUser is the id of the user that should be forwarded to.
	ForwardToUser []uuid.UUID
}

// Validate assures no duplicate user ids in ForwardToUser.
func (d ForwardToUserChannelDetails) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	forwardToUserMap := map[uuid.UUID]struct{}{}
	for _, userID := range d.ForwardToUser {
		if _, ok := forwardToUserMap[userID]; ok {
			report.AddError("duplicate user ids")
			break
		}
		forwardToUserMap[userID] = struct{}{}
	}
	return report, nil
}

// forwardToUserChannelOperator is the channelOperator for channels with
// ChannelTypeForwardToUser.
type forwardToUserChannelOperator struct {
	m *Mall
}

func (op *forwardToUserChannelOperator) deleteDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) error {
	q, _, err := op.m.dialect.Delete(goqu.T("forward_to_user_channel_entries")).
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

func (op *forwardToUserChannelOperator) setChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID,
	detailsRaw ChannelDetails) error {
	details, ok := detailsRaw.(ForwardToUserChannelDetails)
	if !ok {
		return meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	// Delete.
	err := op.deleteDetailsByChannel(ctx, tx, channelID)
	if err != nil && meh.ErrorCode(err) != meh.ErrNotFound {
		return meh.Wrap(err, "delete details by channel", nil)
	}
	// Insert.
	records := make([]any, 0, len(details.ForwardToUser))
	for _, userID := range details.ForwardToUser {
		records = append(records, goqu.Record{
			"channel":         channelID,
			"forward_to_user": userID,
		})
	}
	if len(records) == 0 {
		return nil
	}
	q, _, err := op.m.dialect.Insert(goqu.T("forward_to_user_channel_entries")).
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

func (op *forwardToUserChannelOperator) getChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (ChannelDetails, error) {
	// Build query.
	q, _, err := op.m.dialect.From(goqu.T("forward_to_user_channel_entries")).
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
	details := ForwardToUserChannelDetails{
		ForwardToUser: make([]uuid.UUID, 0),
	}
	for rows.Next() {
		var userID uuid.UUID
		err = rows.Scan(&userID)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		details.ForwardToUser = append(details.ForwardToUser, userID)
	}
	return details, nil
}

// DeleteForwardToUserChannelsByUser deletes all channels with channel type
// ChannelTypeForwardToUser, that forward to the user with the given id. It
// returns the list of affected address book entries.
func (m *Mall) DeleteForwardToUserChannelsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]uuid.UUID, error) {
	// Find affected channels.
	affectedChannelsQuery, _, err := m.dialect.From(goqu.T("forward_to_user_channel_entries")).
		Select(goqu.DISTINCT(goqu.C("channel"))).
		Where(goqu.C("forward_to_user").Eq(userID)).ToSQL()
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
		err = m.DeleteChannelWithDetailsByID(ctx, tx, affectedChannel, ChannelTypeForwardToUser)
		if err != nil {
			return nil, meh.Wrap(err, "delete channel with details", meh.Details{"channel_id": affectedChannel})
		}
	}
	return affectedEntries, nil
}
