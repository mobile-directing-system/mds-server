package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/atomic"
	"sync"
	"testing"
)

func TestController_scheduleLookAfterAllUserNotifications(t *testing.T) {
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	ctrl := NewMockController()
	// Create connections.
	conns := make([]*ConnectionMock, 64)
	remainingRequests := make(map[uuid.UUID]struct{}, len(conns))
	for i := range conns {
		conn := NewConnectionMock()
		conns[i] = conn
		ctrl.Ctrl.connectionsByUser[conn.UserID()] = []Connection{conn}
		remainingRequests[conn.UserID()] = struct{}{}
	}
	defer func() {
		for _, conn := range conns {
			conn.cancel()
		}
	}()

	go func() {
		err := ctrl.Ctrl.scheduleLookAfterAllUserNotifications(timeout)
		assert.NoError(t, err, "should not fail")
	}()

	for {
		if len(remainingRequests) == 0 {
			cancel()
			break
		}
		select {
		case <-timeout.Done():
			break
		case request := <-ctrl.Ctrl.lookAfterUserNotificationRequests:
			delete(remainingRequests, request)
		}
	}

	wait()
	assert.Empty(t, remainingRequests, "should not have any remaining requests")
}

func TestController_scheduleLookAfterUserNotifications(t *testing.T) {
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	ctrl := NewMockController()
	userID := testutil.NewUUIDV4()

	go func() {
		err := ctrl.Ctrl.scheduleLookAfterUserNotifications(timeout, userID)
		assert.NoError(t, err, "should not fail")
	}()

	select {
	case <-timeout.Done():
		break
	case request := <-ctrl.Ctrl.lookAfterUserNotificationRequests:
		assert.Equal(t, request, userID, "should receive request for correct user id")
		cancel()
	}

	wait()
}

func TestController_runLookAfterUserNotificationsScheduler(t *testing.T) {
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	ctrl := NewMockController()
	requests := notificationSchedulerWorkers * 20
	ctrl.DB.Tx = make([]*testutil.DBTx, 0, requests)
	conns := make([]*ConnectionMock, 0, requests)
	for i := 0; i < requests; i++ {
		ctrl.DB.Tx = append(ctrl.DB.Tx, &testutil.DBTx{})
		conn := NewConnectionMock()
		conns = append(conns, conn)
		connsByUser := ctrl.Ctrl.connectionsByUser[conn.UserID()]
		connsByUser = append(connsByUser, conn)
		ctrl.Ctrl.connectionsByUser[conn.UserID()] = connsByUser
	}
	defer func() {
		for _, conn := range conns {
			conn.cancel()
		}
	}()
	// Expect correct amount of calls.
	remaining := atomic.NewInt64(int64(requests))
	ctrl.Store.On("OldestPendingAttemptToNotifyByUser", mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) {
			remaining.Dec()
			if remaining.Load() == 0 {
				cancel()
			}
		}).
		Return(uuid.Nil, false, nil).Times(requests)
	defer ctrl.Store.AssertExpectations(t)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := ctrl.Ctrl.runLookAfterUserNotificationsScheduler(timeout)
		assert.NoError(t, err, "scheduler should not fail")
	}()

scheduleAllRequests:
	for _, conn := range conns {
		select {
		case <-timeout.Done():
			break scheduleAllRequests
		case ctrl.Ctrl.lookAfterUserNotificationRequests <- conn.UserID():
		}
	}

	wait()
	wg.Wait()
	assert.EqualValues(t, remaining.Load(), 0, "should not have any remaining calls")
}

// controllerLookAfterUserNotificationsSuite tests
// Controller.lookAfterUserNotifications.
type controllerLookAfterUserNotificationsSuite struct {
	suite.Suite
	ctrl               *ControllerMock
	tx                 *testutil.DBTx
	conn               *ConnectionMock
	sampleUserID       uuid.UUID
	sampleAttemptID    uuid.UUID
	sampleNotification store.OutgoingIntelDeliveryNotification
}

func (suite *controllerLookAfterUserNotificationsSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx, suite.tx}
	suite.conn = NewConnectionMock()
	suite.ctrl.Ctrl.connectionsByUser[suite.conn.UserID()] = []Connection{suite.conn}
	suite.sampleUserID = suite.conn.UserID()
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.sampleNotification = store.OutgoingIntelDeliveryNotification{
		IntelToDeliver:   store.IntelToDeliver{},
		DeliveryAttempt:  store.AcceptedIntelDeliveryAttempt{ID: suite.sampleAttemptID},
		Channel:          store.NotificationChannel{},
		CreatorDetails:   store.User{},
		RecipientDetails: nulls.NewJSONNullable(store.User{}),
	}
}

func (suite *controllerLookAfterUserNotificationsSuite) TearDownTest() {
	suite.conn.cancel()
}

func (suite *controllerLookAfterUserNotificationsSuite) TestNoConnsForUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterUserNotifications(timeout, testutil.NewUUIDV4())
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *controllerLookAfterUserNotificationsSuite) TestRetrieveOldestPendingAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OldestPendingAttemptToNotifyByUser", timeout, suite.tx, suite.sampleUserID).
		Return(uuid.Nil, false, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterUserNotifications(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterUserNotificationsSuite) TestRetrieveOutgoingNotificationForAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OldestPendingAttemptToNotifyByUser", timeout, suite.tx, suite.sampleUserID).
		Return(suite.sampleAttemptID, true, nil).Once()
	suite.ctrl.Store.On("OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.OutgoingIntelDeliveryNotification{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterUserNotifications(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterUserNotificationsSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.conn.notifyFail = true
	suite.ctrl.Store.On("OldestPendingAttemptToNotifyByUser", timeout, suite.tx, suite.sampleUserID).
		Return(suite.sampleAttemptID, true, nil).Once()
	suite.ctrl.Store.On("OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleNotification, nil).Once()
	suite.ctrl.Store.On("CreateIntelNotificationHistoryEntry", timeout, suite.tx, suite.sampleAttemptID, mock.Anything).
		Return(meh.NewErr("done", "", nil))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterUserNotifications(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
		suite.EqualValues("done", meh.ErrorCode(err))
	}()

	wait()
}

func (suite *controllerLookAfterUserNotificationsSuite) TestNotifyFailForOneConnection() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	connectionCount := 256
	// Create all connections and make two of them fail.
	conns := make([]*ConnectionMock, connectionCount)
	suite.ctrl.Ctrl.connectionsByUser[suite.sampleUserID] = []Connection{}
	for i := range conns {
		conn := NewConnectionMock()
		conn.userID = suite.sampleUserID
		if i == connectionCount-2 || i == connectionCount/2 {
			conn.notifyFail = true
		}
		connsByUser := suite.ctrl.Ctrl.connectionsByUser[suite.sampleUserID]
		connsByUser = append(connsByUser, conn)
		suite.ctrl.Ctrl.connectionsByUser[suite.sampleUserID] = connsByUser
		conns[i] = conn
	}
	defer func() {
		for _, conn := range conns {
			conn.cancel()
		}
	}()
	suite.ctrl.Store.On("OldestPendingAttemptToNotifyByUser", timeout, suite.tx, suite.sampleUserID).
		Return(suite.sampleAttemptID, true, nil).Once()
	suite.ctrl.Store.On("OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleNotification, nil).Once()
	suite.ctrl.Store.On("CreateIntelNotificationHistoryEntry", timeout, suite.tx, suite.sampleAttemptID, mock.Anything).
		Return(meh.NewErr("done", "", nil))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterUserNotifications(timeout, suite.sampleUserID)
		suite.Require().Error(err, "should fail")
		suite.EqualValues("done", meh.ErrorCode(err))
	}()

	wait()
}

func (suite *controllerLookAfterUserNotificationsSuite) TestCreateNotificationHistoryEntryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OldestPendingAttemptToNotifyByUser", timeout, suite.tx, suite.sampleUserID).
		Return(suite.sampleAttemptID, true, nil).Once()
	suite.ctrl.Store.On("OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleNotification, nil).Once()
	suite.ctrl.Store.On("CreateIntelNotificationHistoryEntry", timeout, suite.tx, suite.sampleAttemptID, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterUserNotifications(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterUserNotificationsSuite) TestNotifyNotificationSentFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OldestPendingAttemptToNotifyByUser", timeout, suite.tx, suite.sampleUserID).
		Return(suite.sampleAttemptID, true, nil).Once()
	suite.ctrl.Store.On("OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleNotification, nil).Once()
	suite.ctrl.Store.On("CreateIntelNotificationHistoryEntry", timeout, suite.tx, suite.sampleAttemptID, mock.Anything).
		Return(nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryNotificationSent", timeout, suite.tx, suite.sampleAttemptID, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterUserNotifications(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterUserNotificationsSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	// Test with multiple connections and notifications.
	connCount := 32
	notificationCount := 8
	expectOutbox := make([]store.OutgoingIntelDeliveryNotification, 0, notificationCount)
	for i := 0; i < notificationCount; i++ {
		expectOutbox = append(expectOutbox, suite.sampleNotification)
		suite.ctrl.DB.Tx = append(suite.ctrl.DB.Tx, suite.tx)
	}
	suite.ctrl.Ctrl.connectionsByUser[suite.sampleUserID] = []Connection{}
	conns := make([]*ConnectionMock, connCount)
	for i := range conns {
		conn := NewConnectionMock()
		conn.userID = suite.sampleUserID
		conns[i] = conn
		connsByUser := suite.ctrl.Ctrl.connectionsByUser[conn.UserID()]
		connsByUser = append(connsByUser, conn)
		suite.ctrl.Ctrl.connectionsByUser[conn.UserID()] = connsByUser
	}
	defer func() {
		for _, conn := range conns {
			conn.cancel()
		}
	}()
	suite.ctrl.DB.Tx = append(suite.ctrl.DB.Tx, suite.tx) // The last check.
	suite.ctrl.Store.On("OldestPendingAttemptToNotifyByUser", timeout, suite.tx, suite.sampleUserID).
		Return(suite.sampleAttemptID, true, nil).Times(notificationCount)
	suite.ctrl.Store.On("OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleNotification, nil).Times(notificationCount)
	suite.ctrl.Store.On("CreateIntelNotificationHistoryEntry", timeout, suite.tx, suite.sampleAttemptID, mock.Anything).
		Return(nil).Times(notificationCount)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryNotificationSent", timeout, suite.tx, suite.sampleAttemptID, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("OldestPendingAttemptToNotifyByUser", timeout, suite.tx, suite.sampleUserID).
		Return(uuid.Nil, false, nil).Once() // The last check.
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterUserNotifications(timeout, suite.sampleUserID)
		suite.NoError(err, "should not fail")
	}()

	wait()
	// Check outbox of all connections.
	for _, conn := range conns {
		suite.Require().Equal(expectOutbox, conn.outbox, "should have sent messages")
	}
}

func TestController_lookAfterUserNotifications(t *testing.T) {
	suite.Run(t, new(controllerLookAfterUserNotificationsSuite))
}
