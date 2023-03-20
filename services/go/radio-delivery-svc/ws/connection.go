package ws

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
)

const (
	// messageTypeNewRadioDeliveriesAvailable is used in connection.Notify for
	// notifying about new availabe radio deliveries.
	messageTypeNewRadioDeliveriesAvailable wsutil.MessageType = "new-radio-deliveries-available"
)

// publicNewRadioDeliveriesAvailable holds the operation which might have new
// radio deliveries available.
type publicNewRadioDeliveriesAvailable struct {
	Operation uuid.UUID `json:"operation"`
}

// connection implements controller.Connection and maps notifications from the
// store-representation to the public one.
type connection struct {
	conn wsutil.AutoParserConnection
}

// newConnection returns a new connection.
func newConnection(conn wsutil.AutoParserConnection) *connection {
	return &connection{conn: conn}
}

// UserID returns the user id from the connection's auth-token.
func (conn *connection) UserID() uuid.UUID {
	return conn.conn.AuthToken().UserID
}

// NotifyNewAvailable sends a message with
// messageTypeNewRadioDeliveriesAvailable for the operation with the given id.
func (conn *connection) NotifyNewAvailable(ctx context.Context, operationID uuid.UUID) error {
	deliverGranted, err := auth.HasPermission(conn.conn.AuthToken(), permission.DeliverAnyRadioDelivery())
	if err != nil {
		return meh.Wrap(err, "check permission", nil)
	}
	manageGranted, err := auth.HasPermission(conn.conn.AuthToken(), permission.ManageAnyRadioDelivery())
	if err != nil {
		return meh.Wrap(err, "check permission", nil)
	}
	if !deliverGranted && !manageGranted {
		return nil
	}
	payload := publicNewRadioDeliveriesAvailable{
		Operation: operationID,
	}
	err = conn.conn.Send(ctx, messageTypeNewRadioDeliveriesAvailable, payload)
	if err != nil {
		return meh.Wrap(err, "send over connection", meh.Details{"payload": payload})
	}
	return nil
}

// Done returns the done-channel from the connection.
func (conn *connection) Done() <-chan struct{} {
	return conn.conn.Done()
}
