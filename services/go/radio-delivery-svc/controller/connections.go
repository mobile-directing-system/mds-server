package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
	"sync"
)

// Connection represents a connection to the user with the UserID, allowing
// sending notifications. It also notifies, when the connection is closed.
type Connection interface {
	// UserID returns the id of the user, the Connection is established to.
	UserID() uuid.UUID
	// NotifyNewAvailable notifies that new offers might be available for the
	// operation with the given id.
	NotifyNewAvailable(ctx context.Context, operationID uuid.UUID) error
	// Done receives, when the connection is closed.
	//
	// Warning: It MUST close eventually!
	Done() <-chan struct{}
}

// updateNotifier for notifying about updates over Connection.
type updateNotifier struct {
	logger *zap.Logger
	db     pgutil.DBTxSupplier
	store  Store
	// notifyRequestsForOperation holds requests for notifying about updates for the
	// passed operations.
	notifyRequestsForOperation chan uuid.UUID
	// connections holds all active connections.
	connections []Connection
	// connectionsByOperation holds a Connection list for the operation with the
	// given id. This is used for fast access when new offers are available.
	connectionsByOperation map[uuid.UUID][]Connection
	// connectionsMutex locks connections and connectionsByOperation.
	connectionsMutex sync.RWMutex
}

func newUpdateNotifier(logger *zap.Logger, db pgutil.DBTxSupplier, s Store) *updateNotifier {
	return &updateNotifier{
		logger:                     logger,
		db:                         db,
		store:                      s,
		notifyRequestsForOperation: make(chan uuid.UUID, 256),
		connections:                make([]Connection, 0),
		connectionsByOperation:     make(map[uuid.UUID][]Connection),
	}
}

// run reads notify-requests from notifyRequestsForOperation and calls
// notifyUpdatesForOperations concurrently for the requests.
func (n *updateNotifier) run(lifetime context.Context) {
	var wg sync.WaitGroup
	for {
		select {
		case <-lifetime.Done():
			return
		case operation := <-n.notifyRequestsForOperation:
			wg.Add(1)
			go func(operation uuid.UUID) {
				defer wg.Done()
				n.notifyUpdatesForOperations(lifetime, operation)
			}(operation)
		}
	}
}

func (n *updateNotifier) scheduleNotifyUpdatesForOperations(ctx context.Context, operations ...uuid.UUID) {
	for _, operation := range operations {
		select {
		case <-ctx.Done():
			n.logger.Debug("dropping notify-update-request for operation", zap.Any("operation", operation))
			return
		case n.notifyRequestsForOperation <- operation:
		}
	}
}

// AcceptNewConnection handles the new given Connection. Keep in mind, that in
// the background (without blocking), we will wait for Connection.Done to
// receive in order to remove it from the Controller.
func (n *updateNotifier) AcceptNewConnection(conn Connection) {
	n.connectionsMutex.Lock()
	defer n.connectionsMutex.Unlock()
	// Add to connections.
	n.connections = append(n.connections, conn)
	err := n.reassignConnectionsToOperations(context.Background())
	if err != nil {
		mehlog.Log(n.logger, meh.Wrap(err, "reassign connections to operations", nil))
	}
	// On connection done, remove from list.
	go n.removeConnWhenDone(conn)
}

// removeConnWhenDone waits until the given Connection is done via
// Connection.Done and removes it from connectionsByUser. Errors are logged to
// the logger.
func (n *updateNotifier) removeConnWhenDone(conn Connection) {
	<-conn.Done()
	n.connectionsMutex.Lock()
	defer n.connectionsMutex.Unlock()
	newConns := make([]Connection, 0, len(n.connections)-1)
	for _, oldConn := range n.connections {
		if oldConn == conn {
			continue
		}
		newConns = append(newConns, oldConn)
	}
	if len(newConns) == len(n.connections) {
		mehlog.Log(n.logger, meh.NewInternalErr("connection done but not found in conns", meh.Details{"user_id": conn.UserID()}))
		return
	}
	n.connections = newConns
	err := n.reassignConnectionsToOperations(context.Background())
	if err != nil {
		mehlog.Log(n.logger, meh.Wrap(err, "reassign connections to operations", nil))
	}
}

// reassignConnectionsToOperations clear and refills connectionsByOperation
// using all active connections.
//
// Warning: reassignConnectionsToOperations does NOT lock connectionsMutex!
func (n *updateNotifier) reassignConnectionsToOperations(ctx context.Context) error {
	operationsByConnection := make(map[Connection][]uuid.UUID)
	err := pgutil.RunInTx(ctx, n.db, func(ctx context.Context, tx pgx.Tx) error {
		for _, connection := range n.connections {
			operations, err := n.store.OperationsByMember(ctx, tx, connection.UserID())
			if err != nil {
				return meh.Wrap(err, "operations by member from store", meh.Details{"user_id": connection.UserID()})
			}
			operationsByConnection[connection] = operations
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	// Set.
	n.connectionsByOperation = make(map[uuid.UUID][]Connection)
	for conn, operations := range operationsByConnection {
		n.assignConnectionToOperations(ctx, conn, operations)
	}
	return nil
}

// assignConnectionToOperations assigns the given Connection to the given
// operations.
//
// Warning: assignConnectionToOperations does NOT lock connectionsMutex!
func (n *updateNotifier) assignConnectionToOperations(_ context.Context, conn Connection, operations []uuid.UUID) {
	for _, operation := range operations {
		cc := n.connectionsByOperation[operation]
		cc = append(cc, conn)
		n.connectionsByOperation[operation] = cc
	}
}

// notifyUpdatesForOperations calls Connection.NotifyNewAvailable for all
// connections for the given operations. Errors are logged to the logger.
//
// Blocks until all notifications have been sent.
func (n *updateNotifier) notifyUpdatesForOperations(ctx context.Context, operations ...uuid.UUID) {
	var wg sync.WaitGroup
	notify := func(conn Connection, operationID uuid.UUID) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := conn.NotifyNewAvailable(ctx, operationID)
			if err != nil {
				mehlog.LogToLevel(n.logger, zap.DebugLevel, meh.Wrap(err, "notify new available", meh.Details{
					"user":      conn.UserID(),
					"operation": operationID,
				}))
				return
			}
		}()
	}
	n.connectionsMutex.RLock()
	for _, operation := range operations {
		connectionsForOperation := n.connectionsByOperation[operation]
		for _, conn := range connectionsForOperation {
			notify(conn, operation)
		}
	}
	n.connectionsMutex.RUnlock()
	wg.Wait()
}
