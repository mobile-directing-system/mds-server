package store

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
)

// ChannelTypePush is used for only providing push messages via the MDS
// application.
const ChannelTypePush ChannelType = "push"

// PushChannelDetails holds Channel.Details for ChannelTypePush.
type PushChannelDetails struct {
}

// Validate has nothing to validate.
func (d PushChannelDetails) Validate() (entityvalidation.Report, error) {
	return entityvalidation.NewReport(), nil
}

type pushChannelOperator struct{}

func (op *pushChannelOperator) deleteDetailsByChannel(_ context.Context, _ pgx.Tx, _ uuid.UUID) error {
	return nil
}

func (op *pushChannelOperator) setChannelDetailsByChannel(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ ChannelDetails) error {
	return nil
}

func (op *pushChannelOperator) getChannelDetailsByChannel(_ context.Context, _ pgx.Tx, _ uuid.UUID) (ChannelDetails, error) {
	return PushChannelDetails{}, nil
}
