package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
)

// ChannelTypeInAppNotification is used for sending in-app notifications (like
// push-notifications).
const ChannelTypeInAppNotification ChannelType = "in-app-notification"

// InAppNotificationChannelDetails holds Channel.Details for
// ChannelTypeInAppNotification.
type InAppNotificationChannelDetails struct{}

// Validate has nothing to do.
func (d InAppNotificationChannelDetails) Validate() (entityvalidation.Report, error) {
	return entityvalidation.NewReport(), nil
}

// inAppNotificationChannelOperator is the channelOperator for channels with
// ChannelTypeInAppNotification.
type inAppNotificationChannelOperator struct {
	m *Mall
}

func (op *inAppNotificationChannelOperator) deleteDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) error {
	q, _, err := op.m.dialect.Delete(goqu.T("in_app_notification_channels")).
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

func (op *inAppNotificationChannelOperator) setChannelDetailsByChannel(ctx context.Context, tx pgx.Tx, channelID uuid.UUID, _ ChannelDetails) error {
	// Delete.
	err := op.deleteDetailsByChannel(ctx, tx, channelID)
	if err != nil && meh.ErrorCode(err) != meh.ErrNotFound {
		return meh.Wrap(err, "delete details by channel", nil)
	}
	// Insert.
	q, _, err := op.m.dialect.Insert(goqu.T("in_app_notification_channels")).Rows(goqu.Record{
		"channel": channelID,
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

func (op *inAppNotificationChannelOperator) getChannelDetailsByChannel(_ context.Context, _ pgx.Tx, _ uuid.UUID) (ChannelDetails, error) {
	return InAppNotificationChannelDetails{}, nil
}
