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

// ChannelTypeRadio is used for communicating via radio.
const ChannelTypeRadio ChannelType = "radio"

// RadioChannelDetails holds Channel.Details for ChannelTypeRadio.
type RadioChannelDetails struct {
	// Info holds any free-text information until radio communication is further
	// specified.
	Info string
}

// Validate has nothing to do.
func (d RadioChannelDetails) Validate() (entityvalidation.Report, error) {
	return entityvalidation.NewReport(), nil
}

// radioChannelOperator is the channelOperator for channels with
// ChannelTypeRadio.
type radioChannelOperator struct {
	m *Mall
}

func (op *radioChannelOperator) deleteDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) error {
	q, _, err := op.m.dialect.Delete(goqu.T("radio_channels")).
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

func (op *radioChannelOperator) setChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID, detailsRaw ChannelDetails) error {
	details, ok := detailsRaw.(RadioChannelDetails)
	if !ok {
		return meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	// Delete.
	err := op.deleteDetailsByChannel(ctx, tx, channelID)
	if err != nil && meh.ErrorCode(err) != meh.ErrNotFound {
		return meh.Wrap(err, "delete details by channel", nil)
	}
	// Insert.
	q, _, err := op.m.dialect.Insert(goqu.T("radio_channels")).Rows(goqu.Record{
		"channel": channelID,
		"info":    details.Info,
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

func (op *radioChannelOperator) getChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (ChannelDetails, error) {
	// Build query.
	q, _, err := op.m.dialect.From(goqu.T("radio_channels")).
		Select(goqu.C("info")).
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
	if !rows.Next() {
		return nil, meh.NewNotFoundErr("not found", nil)
	}
	var details RadioChannelDetails
	err = rows.Scan(&details.Info)
	if err != nil {
		return nil, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return details, nil
}
