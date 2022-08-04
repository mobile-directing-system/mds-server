package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/ttacon/libphonenumber"
	"reflect"
)

const phoneNumberDefaultRegion = "DE"

// ChannelTypePhoneCall is used for communicating via phone calls.
const ChannelTypePhoneCall ChannelType = "phone-call"

// PhoneCallChannelDetails holds Channel.Details for ChannelTypePhoneCall.
type PhoneCallChannelDetails struct {
	// Phone is the phone number.
	Phone string
}

// Validate the Phone number.
func (d PhoneCallChannelDetails) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	parsedPhone, err := libphonenumber.Parse(d.Phone, phoneNumberDefaultRegion)
	if err != nil || !libphonenumber.IsValidNumber(parsedPhone) {
		report.AddError("invalid phone number (expected in E.164-format)")
	}
	return report, nil
}

// phoneCallChannelOperator is the channelOperator for channels with
// ChannelTypePhoneCall.
type phoneCallChannelOperator struct {
	m *Mall
}

func (op *phoneCallChannelOperator) deleteDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) error {
	q, _, err := op.m.dialect.Delete(goqu.T("phone_call_channels")).
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

func (op *phoneCallChannelOperator) setChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID, detailsRaw ChannelDetails) error {
	details, ok := detailsRaw.(PhoneCallChannelDetails)
	if !ok {
		return meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	// Delete.
	err := op.deleteDetailsByChannel(ctx, tx, channelID)
	if err != nil && meh.ErrorCode(err) != meh.ErrNotFound {
		return meh.Wrap(err, "delete details by channel", nil)
	}
	// Insert.
	q, _, err := op.m.dialect.Insert(goqu.T("phone_call_channels")).Rows(goqu.Record{
		"channel": channelID,
		"phone":   details.Phone,
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

func (op *phoneCallChannelOperator) getChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (ChannelDetails, error) {
	// Build query.
	q, _, err := op.m.dialect.From(goqu.T("phone_call_channels")).
		Select(goqu.C("phone")).
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
	var details PhoneCallChannelDetails
	err = rows.Scan(&details.Phone)
	if err != nil {
		return nil, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return details, nil
}
