package controller

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// ConnectionMock mocks Connection.
type ConnectionMock struct {
	userID      uuid.UUID
	lifetime    context.Context
	cancel      context.CancelFunc
	notifyFail  bool
	outbox      []uuid.UUID
	outboxMutex sync.Mutex
}

func NewConnectionMock() *ConnectionMock {
	lifetime, cancel := context.WithCancel(context.Background())
	return &ConnectionMock{
		userID:   testutil.NewUUIDV4(),
		lifetime: lifetime,
		cancel:   cancel,
	}
}

func (m *ConnectionMock) UserID() uuid.UUID {
	return m.userID
}

func (m *ConnectionMock) NotifyNewAvailable(_ context.Context, operationID uuid.UUID) error {
	if m.notifyFail {
		return errors.New("sad life")
	}
	m.outboxMutex.Lock()
	defer m.outboxMutex.Unlock()
	m.outbox = append(m.outbox, operationID)
	return nil
}

func (m *ConnectionMock) Done() <-chan struct{} {
	return m.lifetime.Done()
}

// updateNotifierAcceptNewConnectionSuite tests
// updateNotifier.AcceptNewConnection.
type updateNotifierAcceptNewConnectionSuite struct {
	suite.Suite
	ctrl    *ControllerMock
	newConn func() *ConnectionMock
}

func (suite *updateNotifierAcceptNewConnectionSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.newConn = func() *ConnectionMock {
		return NewConnectionMock()
	}
}

func (suite *updateNotifierAcceptNewConnectionSuite) waitAllConnectionsRemoved(ctx context.Context) {
	connsExist := true
	for connsExist {
		select {
		case <-ctx.Done():
			return
		case <-time.After(200 * time.Microsecond):
			connsExist = false
			suite.ctrl.Ctrl.connUpdateNotifier.connectionsMutex.RLock()
			connsExist = len(suite.ctrl.Ctrl.connUpdateNotifier.connections) != 0
			suite.ctrl.Ctrl.connUpdateNotifier.connectionsMutex.RUnlock()
		}
	}
}

// TestReassignFail lets reassining fail but should still listen for
// connection-close.
func (suite *updateNotifierAcceptNewConnectionSuite) TestReassignFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}, {}}
	conn := suite.newConn()
	suite.ctrl.Store.On("OperationsByMember", mock.Anything, suite.ctrl.DB.Tx[0], conn.UserID()).
		Return(nil, errors.New("sad life"))

	go func() {
		defer cancel()
		defer conn.cancel()
		suite.ctrl.Ctrl.AcceptNewConnection(conn)
		suite.Len(suite.ctrl.Ctrl.connUpdateNotifier.connections, 1, "should have added new connection")
		suite.Contains(suite.ctrl.Ctrl.connUpdateNotifier.connections, conn, "should have added new connection")
	}()

	suite.waitAllConnectionsRemoved(timeout)
	wait()
}

func (suite *updateNotifierAcceptNewConnectionSuite) TestFirstForUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}, {}}
	conn := suite.newConn()
	sampleOperationID1 := testutil.NewUUIDV4()
	sampleOperationID2 := testutil.NewUUIDV4()
	suite.ctrl.Store.On("OperationsByMember", mock.Anything, suite.ctrl.DB.Tx[0], conn.UserID()).
		Return([]uuid.UUID{sampleOperationID1, sampleOperationID2}, nil)

	go func() {
		defer cancel()
		defer conn.cancel()
		suite.ctrl.Ctrl.AcceptNewConnection(conn)
		suite.Len(suite.ctrl.Ctrl.connUpdateNotifier.connections, 1, "should have added new connection")
		suite.Contains(suite.ctrl.Ctrl.connUpdateNotifier.connections, conn, "should have added new connection")
		suite.Contains(suite.ctrl.Ctrl.connUpdateNotifier.connectionsByOperation[sampleOperationID1], conn,
			"should have added new connection for operation 1")
		suite.Contains(suite.ctrl.Ctrl.connUpdateNotifier.connectionsByOperation[sampleOperationID2], conn,
			"should have added new connection for operation 2")
	}()

	suite.waitAllConnectionsRemoved(timeout)
	wait()
}

func (suite *updateNotifierAcceptNewConnectionSuite) TestConnectionDone() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	conns := make([]*ConnectionMock, 32)
	suite.ctrl.DB.Tx = make([]*testutil.DBTx, len(conns)*2)
	for i := range suite.ctrl.DB.Tx {
		suite.ctrl.DB.Tx[i] = &testutil.DBTx{}
	}
	var wg sync.WaitGroup
	waitAllConnected, allConnected := context.WithCancel(timeout)
	suite.ctrl.Store.On("OperationsByMember", mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				suite.ctrl.Ctrl.connUpdateNotifier.connectionsMutex.RLock()
				defer suite.ctrl.Ctrl.connUpdateNotifier.connectionsMutex.RUnlock()
				if len(suite.ctrl.Ctrl.connUpdateNotifier.connections) == len(conns) {
					allConnected()
				}
			}()
		}).
		Return([]uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		}, nil)
	// Create and accept connections.
	for i := range conns {
		conn := NewConnectionMock()
		conn.userID = testutil.NewUUIDV4()
		conns[i] = conn
		wg.Add(1)
		go func() {
			defer wg.Done()
			suite.ctrl.Ctrl.AcceptNewConnection(conn)
		}()
	}
	<-waitAllConnected.Done()
	// Close all connections.
	for _, conn := range conns {
		conn.cancel()
	}

	suite.waitAllConnectionsRemoved(timeout)
	cancel()
	wg.Wait()
	wait()
}

func TestUpdateNotifier_AcceptNewConnection(t *testing.T) {
	suite.Run(t, new(updateNotifierAcceptNewConnectionSuite))
}

func TestUpdateNotifier_reassignConnectionsToOperations(t *testing.T) {
	// Generate sample data.
	availableOperations := make([]uuid.UUID, 8)
	for i := range availableOperations {
		availableOperations[i] = testutil.NewUUIDV4()
	}
	userCount := 64
	operationsByUsers := make(map[uuid.UUID][]uuid.UUID, userCount)
	for i := 0; i < userCount; i++ {
		max := rand.Intn(9)
		operationsForUser := make(map[uuid.UUID]struct{}, max)
		for i := 0; i < max; i++ {
			operationsForUser[availableOperations[rand.Intn(len(availableOperations))]] = struct{}{}
		}
		user := testutil.NewUUIDV4()
		operationsByUsers[user] = make([]uuid.UUID, len(operationsForUser))
		for operationID := range operationsForUser {
			operationsByUsers[user] = append(operationsByUsers[user], operationID)
		}
	}
	// Setup mocks.
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	ctrl := NewMockController()
	ctrl.DB.Tx = []*testutil.DBTx{{}}
	conns := make([]*ConnectionMock, 0, userCount)
	for userID, operations := range operationsByUsers {
		conn := NewConnectionMock()
		conn.userID = userID
		conns = append(conns, conn)
		ctrl.Ctrl.connUpdateNotifier.connections = append(ctrl.Ctrl.connUpdateNotifier.connections, conn)
		ctrl.Store.On("OperationsByMember", timeout, ctrl.DB.Tx[0], userID).
			Return(operations, nil).Once()
	}
	defer ctrl.Store.AssertExpectations(t)

	go func() {
		defer cancel()
		err := ctrl.Ctrl.connUpdateNotifier.reassignConnectionsToOperations(timeout)
		require.NoError(t, err, "should not fail")
		for userID, operations := range operationsByUsers {
			for _, operation := range operations {
				found := false
				for _, conn := range ctrl.Ctrl.connUpdateNotifier.connectionsByOperation[operation] {
					if conn.UserID() == userID {
						found = true
						break
					}
				}
				assert.True(t, found, "should have found user in connections by operation from notifier")
			}
		}
	}()

	wait()
	for _, conn := range conns {
		conn.cancel()
	}
}
