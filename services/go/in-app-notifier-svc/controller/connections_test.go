package controller

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
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
	outbox      []store.OutgoingIntelDeliveryNotification
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

func (m *ConnectionMock) Notify(_ context.Context, notification store.OutgoingIntelDeliveryNotification) error {
	if m.notifyFail {
		return errors.New("sad life")
	}
	m.outboxMutex.Lock()
	defer m.outboxMutex.Unlock()
	m.outbox = append(m.outbox, notification)
	return nil
}

func (m *ConnectionMock) Done() <-chan struct{} {
	return m.lifetime.Done()
}

// ControllerAcceptNewConnectionSuite tests Controller.AcceptNewConnection.
type ControllerAcceptNewConnectionSuite struct {
	suite.Suite
	ctrl    *ControllerMock
	newConn func() *ConnectionMock
}

func (suite *ControllerAcceptNewConnectionSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.newConn = func() *ConnectionMock {
		return NewConnectionMock()
	}
}

func (suite *ControllerAcceptNewConnectionSuite) waitForLookAfterRequest(ctx context.Context) {
	select {
	case <-ctx.Done():
		suite.Fail("context done while waiting for look-after-request")
	case <-suite.ctrl.Ctrl.lookAfterUserNotificationRequests:
		return
	}
}

func (suite *ControllerAcceptNewConnectionSuite) waitAllConnectionsRemoved(ctx context.Context) {
	connsExist := true
	for connsExist {
		select {
		case <-ctx.Done():
			return
		case <-time.After(200 * time.Microsecond):
			connsExist = false
			suite.ctrl.Ctrl.connectionsByUserMutex.RLock()
			for _, conns := range suite.ctrl.Ctrl.connectionsByUser {
				if len(conns) != 0 {
					connsExist = true
					break
				}
			}
			suite.ctrl.Ctrl.connectionsByUserMutex.RUnlock()
		}
	}
}

func (suite *ControllerAcceptNewConnectionSuite) TestFirstForUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn := suite.newConn()
		defer conn.cancel()
		suite.ctrl.Ctrl.AcceptNewConnection(conn)
		suite.Len(suite.ctrl.Ctrl.connectionsByUser[conn.UserID()], 1, "should have added new connection")
		suite.Contains(suite.ctrl.Ctrl.connectionsByUser[conn.UserID()], conn, "should have added new connection")
	}()

	suite.waitForLookAfterRequest(timeout)
	cancel()
	wg.Wait()
	wait()
}

func (suite *ControllerAcceptNewConnectionSuite) TestMultipleForUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	otherConns := make([]Connection, 8)
	newConn := suite.newConn()
	defer newConn.cancel()
	for i := range otherConns {
		conn := NewConnectionMock()
		conn.userID = newConn.UserID()
		//goland:noinspection ALL
		defer conn.cancel()
		otherConns[i] = conn
	}
	suite.ctrl.Ctrl.connectionsByUser[newConn.UserID()] = otherConns

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		suite.ctrl.Ctrl.AcceptNewConnection(newConn)
		newConns := suite.ctrl.Ctrl.connectionsByUser[newConn.UserID()]
		for _, otherConn := range otherConns {
			suite.Contains(newConns, otherConn, "should keep old connections")
		}
		// Manual check because of race conditions.
		for _, conn := range newConns {
			if conn == newConn {
				return
			}
		}
		suite.Fail("should have added new connection")
	}()

	suite.waitForLookAfterRequest(timeout)
	cancel()
	wg.Wait()
	wait()
}

func (suite *ControllerAcceptNewConnectionSuite) TestConnectionDone() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	newConn := suite.newConn()
	defer newConn.cancel()
	conns := make([]*ConnectionMock, 32)
	// Create and accept connections.
	var wg sync.WaitGroup
	for i := range conns {
		conn := NewConnectionMock()
		conn.userID = newConn.UserID()
		conns[i] = conn
		wg.Add(1)
		go func() {
			defer wg.Done()
			suite.ctrl.Ctrl.AcceptNewConnection(conn)
		}()
	}
	// Wait for requests.
	for range conns {
		suite.waitForLookAfterRequest(timeout)
	}
	// Close all connections.
	for _, conn := range conns {
		conn.cancel()
	}

	suite.waitAllConnectionsRemoved(timeout)
	cancel()
	wg.Wait()
	wait()
}

func (suite *ControllerAcceptNewConnectionSuite) TestConnectionDoneMessAround() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	newConn := suite.newConn()
	defer newConn.cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		suite.ctrl.Ctrl.AcceptNewConnection(newConn)
	}()
	suite.waitForLookAfterRequest(timeout)

	// Remove all connections for user manually (simulate bug).
	suite.ctrl.Ctrl.connectionsByUserMutex.Lock()
	delete(suite.ctrl.Ctrl.connectionsByUser, newConn.UserID())
	suite.ctrl.Ctrl.connectionsByUserMutex.Unlock()

	newConn.cancel()
	suite.waitAllConnectionsRemoved(timeout)
	cancel()
	wg.Wait()
	wait()
}

func TestController_AcceptNewConnection(t *testing.T) {
	suite.Run(t, new(ControllerAcceptNewConnectionSuite))
}
