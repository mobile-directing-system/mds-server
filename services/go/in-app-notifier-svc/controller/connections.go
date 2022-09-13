package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"time"
)

// Connection represents a connection to the user with the UserID, allowing
// sending notifications. It also notifies, when the connection is closed.
type Connection interface {
	// UserID returns the id of the user, the Connection is established to.
	UserID() uuid.UUID
	// Notify sends the given store.OutgoingIntelDeliveryNotification.
	Notify(ctx context.Context, notification store.OutgoingIntelDeliveryNotification) error
	// Done receives, when the connection is closed.
	//
	// Warning: It MUST close eventually!
	Done() <-chan struct{}
}

// AcceptNewConnection handles the new given Connection. Keep in mind, that in
// the background (without blocking), we will wait for Connection.Done to
// receive in order to remove it from the Controller.
func (c *Controller) AcceptNewConnection(conn Connection) {
	c.connectionsByUserMutex.Lock()
	defer c.connectionsByUserMutex.Unlock()
	// Add to connections.
	conns, ok := c.connectionsByUser[conn.UserID()]
	if !ok {
		conns = make([]Connection, 0, 1)
	}
	conns = append(conns, conn)
	c.connectionsByUser[conn.UserID()] = conns
	// On connection done, remove from list.
	go c.removeConnWhenDone(conn)
	// Schedule notification check concurrently. We do not need to wait for this or
	// rely on it to not fail as in case of failure, periodic checks would handle it
	// anyways.
	go func() {
		timeout, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		err := c.scheduleLookAfterUserNotifications(timeout, conn.UserID())
		if err != nil {
			mehlog.Log(c.logger, meh.Wrap(err, "schedule notification check for new connection", meh.Details{"user_id": conn.UserID()}))
			return
		}
	}()
}

// removeConnWhenDone waits until the given Connection is done via
// Connection.Done and removes it from connectionsByUser. Errors are logged to
// the logger.
func (c *Controller) removeConnWhenDone(conn Connection) {
	<-conn.Done()
	c.connectionsByUserMutex.Lock()
	defer c.connectionsByUserMutex.Unlock()
	conns, ok := c.connectionsByUser[conn.UserID()]
	if !ok || len(conns) == 0 {
		mehlog.Log(c.logger, meh.NewInternalErr("connection done but no connections for user", meh.Details{"user_id": conn.UserID()}))
		return
	}
	newConns := make([]Connection, 0, len(conns)-1)
	for _, userConn := range conns {
		if userConn == conn {
			continue
		}
		newConns = append(newConns, userConn)
	}
	if len(newConns) == len(conns) {
		mehlog.Log(c.logger, meh.NewInternalErr("connection done but not found in user conns", meh.Details{"user_id": conn.UserID()}))
		return
	}
	c.connectionsByUser[conn.UserID()] = newConns
}
