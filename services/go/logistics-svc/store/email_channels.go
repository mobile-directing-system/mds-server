package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"net/mail"
	"reflect"
)

// ChannelTypeEmail is used for communicating via email.
const ChannelTypeEmail ChannelType = "email"

// EmailChannelDetails holds Channel.Details for ChannelTypeEmail.
type EmailChannelDetails struct {
	// Email is the email address.
	Email string
}

// Validate the Email.
func (e EmailChannelDetails) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	// Validate email address.
	_, err := mail.ParseAddress(e.Email)
	if err != nil {
		report.AddError("invalid mail adress")
	}
	return report, nil
}

// emailChannelOperator is the channelOperator for channels with
// ChannelTypeEmail
type emailChannelOperator struct {
	m *Mall
}

func (op *emailChannelOperator) deleteDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) error {
	q, _, err := op.m.dialect.Delete(goqu.T("email_channels")).
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

func (op *emailChannelOperator) setChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID, detailsRaw ChannelDetails) error {
	details, ok := detailsRaw.(EmailChannelDetails)
	if !ok {
		return meh.NewInternalErr("cannot cast details", meh.Details{"was": reflect.TypeOf(detailsRaw)})
	}
	// Delete.
	err := op.deleteDetailsByChannel(ctx, tx, channelID)
	if err != nil && meh.ErrorCode(err) != meh.ErrNotFound {
		return meh.Wrap(err, "delete details by channel", nil)
	}
	// Insert.
	q, _, err := op.m.dialect.Insert(goqu.T("email_channels")).Rows(goqu.Record{
		"channel": channelID,
		"email":   details.Email,
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

func (op *emailChannelOperator) getChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (ChannelDetails, error) {
	// Build query.
	q, _, err := op.m.dialect.From(goqu.T("email_channels")).
		Select(goqu.C("email")).
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
	var details EmailChannelDetails
	err = rows.Scan(&details.Email)
	if err != nil {
		return nil, mehpg.NewScanRowsErr(err, "scan row", q)
	}
	return details, nil
}
