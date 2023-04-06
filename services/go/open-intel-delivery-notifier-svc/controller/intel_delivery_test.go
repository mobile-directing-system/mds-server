package controller

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/lefinal/zaprec"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"math/rand"
	"pgregory.net/rapid"
	"strconv"
	"sync"
	"testing"
	"time"
)

func newWatcher(initialDoNotify bool) *openIntelDeliveryWatcher {
	return &openIntelDeliveryWatcher{
		doNotify:     initialDoNotify,
		doNotifyCond: sync.NewCond(&sync.Mutex{}),
	}
}

type intelDeliveriesListenerMock struct {
	notified chan []store.OpenIntelDeliverySummary
}

func (l *intelDeliveriesListenerMock) NotifyOpenIntelDeliveries(ctx context.Context, openDeliveries []store.OpenIntelDeliverySummary) bool {
	select {
	case <-ctx.Done():
		return false
	case l.notified <- openDeliveries:
	}
	return true
}

func newIntelDeliveriesListenerMock() *intelDeliveriesListenerMock {
	return &intelDeliveriesListenerMock{
		notified: make(chan []store.OpenIntelDeliverySummary),
	}
}

// openIntelDeliveriesNotifyHubSuite tests openIntelDeliveriesNotifyHub.
type openIntelDeliveriesNotifyHubSuite struct {
	suite.Suite
}

func (suite *openIntelDeliveriesNotifyHubSuite) TestMulti() {
	type listenerContainer struct {
		listener   *intelDeliveriesListenerMock
		startRound int
		endRound   int
		unregister func()
	}

	type roundActions struct {
		register   []*listenerContainer
		unregister []*listenerContainer
	}

	rapid.Check(suite.T(), func(t *rapid.T) {
		h := newOpenIntelDeliveriesNotifierHub()
		listenerCount := rapid.IntRange(0, 64).Draw(t, "listener_count")
		updates := rapid.IntRange(0, 128).Draw(t, "updates")
		possibleStartListen := rapid.SliceOfN(rapid.Float64Range(0, 1), 8, 16).Draw(t, "possible_start_listen")
		possibleEndListen := rapid.SliceOfN(rapid.Float64Range(0, 1), len(possibleStartListen), len(possibleStartListen)).Draw(t, "possible_end_listen")
		roundActions := make([]roundActions, updates+1)
		// Setup listeners and actions.
		for listenerNum := 0; listenerNum < listenerCount; listenerNum++ {
			c := &listenerContainer{
				listener: &intelDeliveriesListenerMock{
					notified: make(chan []store.OpenIntelDeliverySummary),
				},
				unregister: nil,
			}
			c.startRound = int(float64(len(roundActions)-1) * possibleStartListen[listenerNum%len(possibleStartListen)])
			c.endRound = c.startRound + int(float64(len(roundActions)-c.startRound-1)*possibleEndListen[listenerNum%len(possibleEndListen)])
			ra := roundActions[c.startRound]
			ra.register = append(ra.register, c)
			roundActions[c.startRound] = ra
			ra = roundActions[c.endRound]
			ra.unregister = append(ra.unregister, c)
			roundActions[c.endRound] = ra
		}
		// Run rounds.
		allCurrentlyRegistered := make([]*listenerContainer, 0)
		for round := 0; round < updates; round++ {
			timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromRapidT(t), 200*time.Millisecond)

			go func() {
				defer cancel()
				actionsToPerform := roundActions[round]
				// Register.
				for _, toRegister := range actionsToPerform.register {
					toRegister.unregister = h.registerListener(toRegister.listener)
					allCurrentlyRegistered = append(allCurrentlyRegistered, toRegister)
					select {
					case <-timeout.Done():
						t.Error("timeout while waiting for register-notify")
						return
					case got := <-toRegister.listener.notified:
						assert.Equal(t, h.currentOpenDeliveries, got, "should notify correct deliveries on register")
					}
				}
				// Unregister.
				for _, toUnregister := range actionsToPerform.unregister {
					toUnregister.unregister()
					newAllCurrentlyRegistered := make([]*listenerContainer, 0, len(allCurrentlyRegistered))
					for _, registered := range allCurrentlyRegistered {
						if registered != toUnregister {
							newAllCurrentlyRegistered = append(newAllCurrentlyRegistered, registered)
						}
					}
					allCurrentlyRegistered = newAllCurrentlyRegistered
				}
				// Notify registered.
				notify := []store.OpenIntelDeliverySummary{
					{
						Delivery: store.ActiveIntelDelivery{Note: nulls.NewString(strconv.Itoa(round))},
					},
				}
				var allListenersNotified sync.WaitGroup
				// Assure all registered received the update.
				for _, registered := range allCurrentlyRegistered {
					registered := registered
					allListenersNotified.Add(1)
					go func() {
						defer allListenersNotified.Done()
						select {
						case <-timeout.Done():
							assert.Fail(t, "timeout while waiting for listener to be notified")
						case gotNotficiation := <-registered.listener.notified:
							assert.Equal(t, notify, gotNotficiation, "listener should contain notified update")
						}
					}()
				}
				h.feed(notify)
				allListenersNotified.Wait()
			}()

			wait()
		}
	})
}

func Test_openIntelDeliveriesNotifyHub(t *testing.T) {
	suite.Run(t, new(openIntelDeliveriesNotifyHubSuite))
}

// runNewOpenIntelDeliveryWatcherSuite tests runNewOpenIntelDeliveryWatcher.
type runNewOpenIntelDeliveryWatcherSuite struct {
	suite.Suite
	listener *intelDeliveriesListenerMock
}

func (suite *runNewOpenIntelDeliveryWatcherSuite) SetupTest() {
	suite.listener = &intelDeliveriesListenerMock{notified: make(chan []store.OpenIntelDeliverySummary)}
}

func (suite *runNewOpenIntelDeliveryWatcherSuite) RetrieveFail() {
	l, recorder := zaprec.NewRecorder(zap.ErrorLevel)
	w := runNewOpenIntelDeliveryWatcher(l, time.Minute, time.Minute, func() ([]store.OpenIntelDeliverySummary, error) {
		return nil, errors.New("sad life")
	})
	w.shutdown()
	suite.NotEmpty(recorder.RecordsByLevel(zap.ErrorLevel), "should have logged error")
}

func (suite *runNewOpenIntelDeliveryWatcherSuite) TestPeriodicNotify() {
	const interval = 5 * time.Millisecond
	const count = 5
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()

	w := runNewOpenIntelDeliveryWatcher(zap.NewNop(), interval, 0, func() ([]store.OpenIntelDeliverySummary, error) {
		return nil, nil
	})
	defer w.shutdown()
	w.notifierHub.registerListener(suite.listener)

	var last time.Time
	for messageNum := 0; messageNum < count; messageNum++ {
		select {
		case <-timeout.Done():
			suite.Failf("timeout", "timeout while waiting for message %d", messageNum)
			break
		case <-suite.listener.notified:
		}
		if messageNum > 1 {
			suite.WithinDuration(last.Add(interval), time.Now(), interval/5, "should apply correct interval")
		}
		last = time.Now()
	}

	cancel()
	wait()
}

func (suite *runNewOpenIntelDeliveryWatcherSuite) TestNotifyDelay() {
	const delay = 5 * time.Millisecond
	const segmentCount = 8
	const perSegmentCount = 32
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()

	messageNum := atomic.NewInt64(0)
	w := runNewOpenIntelDeliveryWatcher(zap.NewNop(), time.Minute, delay, func() ([]store.OpenIntelDeliverySummary, error) {
		return make([]store.OpenIntelDeliverySummary, messageNum.Load()), nil
	})
	w.notifierHub.registerListener(suite.listener)
	// Receive initial message from register.
	select {
	case <-timeout.Done():
		suite.Fail("timeout", "timeout while waiting for register-message")
		return
	case <-suite.listener.notified:
	}
	// Receive all messages.
	for messageNum.Load() < segmentCount*perSegmentCount {
		w.notifyIntelDeliveryChanged()
		if (messageNum.Load()+1)%perSegmentCount == 0 {
			// Receive one from segment.
			<-time.After(delay)
			select {
			case <-timeout.Done():
				suite.Fail("timeout", "timeout while waiting for message from segment")
			case got := <-suite.listener.notified:
				suite.Len(got, int(messageNum.Load()), "should have notified correct message")
			}
		}
		messageNum.Inc()
	}

	cancel()
	wait()
}

func (suite *runNewOpenIntelDeliveryWatcherSuite) TestShutdown() {
	const delay = 5 * time.Millisecond
	const messagesUntilShutdown = 4
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()

	w := runNewOpenIntelDeliveryWatcher(zap.NewNop(), time.Minute, delay, func() ([]store.OpenIntelDeliverySummary, error) {
		return make([]store.OpenIntelDeliverySummary, 1), nil
	})
	w.notifierHub.registerListener(suite.listener)
	// Receive initial message from register.
	select {
	case <-timeout.Done():
		suite.Fail("timeout", "timeout while waiting for register-message")
		return
	case <-suite.listener.notified:
	}

	for messageNum := 0; messageNum < messagesUntilShutdown; messageNum++ {
		for i := rand.Intn(16); i >= 0; i-- {
			w.notifyIntelDeliveryChanged()
		}
		<-time.After(delay)
		select {
		case <-timeout.Done():
			suite.Fail("timeout", "timeout while waiting for message")
			return
		case <-suite.listener.notified:
		}
	}
	w.shutdown()
	w.notifyIntelDeliveryChanged()
	select {
	case <-timeout.Done():
		suite.Fail("timeout", "timeout while waiting for final message")
	case <-time.After(2 * delay):
		// OK.
	case <-suite.listener.notified:
		suite.Fail("fail", "should not send message after being shut down")
	}

	cancel()
	wait()
}

func Test_runNewOpenIntelDeliveryWatcher(t *testing.T) {
	suite.Run(t, new(runNewOpenIntelDeliveryWatcherSuite))
}

func TestController_ServeOpenIntelDeliveriesListener(t *testing.T) {
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	defer cancel()
	c := NewMockController()
	c.DB.GenTx = true
	operationID := testutil.NewUUIDV4()
	openIntelDeliveries := []store.OpenIntelDeliverySummary{
		{
			Delivery: store.ActiveIntelDelivery{
				ID: testutil.NewUUIDV4(),
			},
			Intel: store.Intel{
				ID:        testutil.NewUUIDV4(),
				Operation: operationID,
			},
		},
		{
			Delivery: store.ActiveIntelDelivery{
				ID: testutil.NewUUIDV4(),
			},
			Intel: store.Intel{
				ID:        testutil.NewUUIDV4(),
				Operation: operationID,
			},
		},
	}
	c.Store.On("OpenIntelDeliveriesByOperation", mock.Anything, mock.Anything, operationID).
		Return(openIntelDeliveries, nil).Once()
	defer c.Store.AssertExpectations(t)
	var listenerWG sync.WaitGroup
	// Start serving.
	const listenerCount = 128
	listenerLifetime, cancelListeners := context.WithCancel(timeout)
	defer cancelListeners()
	listeners := make([]*intelDeliveriesListenerMock, listenerCount)
	for listenerNum := 0; listenerNum < listenerCount; listenerNum++ {
		listenerNum := listenerNum
		listeners[listenerNum] = newIntelDeliveriesListenerMock()
		listenerWG.Add(1)
		go func() {
			defer listenerWG.Done()
			c.Ctrl.ServeOpenIntelDeliveriesListener(listenerLifetime, operationID, listeners[listenerNum])
		}()
	}
	// Wait for watcher up.
	var watcher *openIntelDeliveryWatcher
	var ok bool
	for {
		c.Ctrl.openIntelDeliveryWatchersByOperationMutex.RLock()
		watcher, ok = c.Ctrl.openIntelDeliveryWatchersByOperation[operationID]
		c.Ctrl.openIntelDeliveryWatchersByOperationMutex.RUnlock()
		if !ok {
			continue
		}
		watcher.notifierHub.listenersMutex.Lock()
		listeners := watcher.notifierHub.listeners
		watcher.notifierHub.listenersMutex.Unlock()
		if listeners == listenerCount {
			break
		}
		<-time.After(time.Millisecond)
	}
	// Receive first from all.
	for _, listener := range listeners {
		got := <-listener.notified
		assert.Emptyf(t, got, "should have notified empty list")
	}
	// Notify.
	<-time.After(10 * time.Millisecond)
	watcher.notifyIntelDeliveryChanged()
	// Assure all notified.
	for _, listener := range listeners {
		got := <-listener.notified
		assert.Equal(t, openIntelDeliveries, got, "should have notified correct list")
	}
	// Notify-part done.
	cancelListeners()

	// Assure watcher removed.
	go func() {
		defer cancel()
		listenerWG.Wait()
		for {
			c.Ctrl.openIntelDeliveryWatchersByOperationMutex.RLock()
			_, ok := c.Ctrl.openIntelDeliveryWatchersByOperation[operationID]
			c.Ctrl.openIntelDeliveryWatchersByOperationMutex.RUnlock()
			if !ok {
				break
			}
			select {
			case <-timeout.Done():
				assert.Fail(t, "timeout", "timeout while waiting for watcher to be removed")
				return
			case <-time.After(time.Millisecond):
			}
		}
	}()

	wait()
}

// controllerNotifyIntelDeliveryChangedSuite tests
// Controller.notifyIntelDeliveryChanged.
type controllerNotifyIntelDeliveryChangedSuite struct {
	suite.Suite
	c           *ControllerMock
	operationID uuid.UUID
	watcher     *openIntelDeliveryWatcher
}

func (suite *controllerNotifyIntelDeliveryChangedSuite) SetupTest() {
	suite.c = NewMockController()
	suite.operationID = testutil.NewUUIDV4()
	suite.watcher = runNewOpenIntelDeliveryWatcher(zap.NewNop(), time.Minute, 50*time.Millisecond, func() ([]store.OpenIntelDeliverySummary, error) {
		return []store.OpenIntelDeliverySummary{}, nil
	})
	suite.c.Ctrl.openIntelDeliveryWatchersByOperation[suite.operationID] = suite.watcher
	suite.T().Cleanup(func() {
		suite.watcher.shutdown()
	})
	// Wait until notify-flag unset.
	for {
		suite.watcher.doNotifyCond.L.Lock()
		doNotify := suite.watcher.doNotify
		suite.watcher.doNotifyCond.L.Unlock()
		if !doNotify {
			break
		}
		<-time.After(time.Millisecond)
	}
}

func (suite *controllerNotifyIntelDeliveryChangedSuite) TestOperationNotFound() {
	suite.NotPanics(func() {
		suite.c.Ctrl.notifyIntelDeliveryChanged(testutil.NewUUIDV4())
	})
}

func (suite *controllerNotifyIntelDeliveryChangedSuite) TestOK() {
	suite.c.Ctrl.notifyIntelDeliveryChanged(suite.operationID)
	suite.watcher.doNotifyCond.L.Lock()
	defer suite.watcher.doNotifyCond.L.Unlock()
	suite.True(suite.watcher.doNotify, "should set notify-flag")
}

func TestController_notifyIntelDeliveryChanged(t *testing.T) {
	suite.Run(t, new(controllerNotifyIntelDeliveryChangedSuite))
}

// ControllerCreateActiveIntelDeliverySuite tests
// Controller.CreateActiveIntelDelivery.
type ControllerCreateActiveIntelDeliverySuite struct {
	suite.Suite
	c      *ControllerMock
	create store.ActiveIntelDelivery
	intel  store.Intel
}

func (suite *ControllerCreateActiveIntelDeliverySuite) SetupTest() {
	suite.c = NewMockController()
	suite.c.DB.GenTx = true
	suite.create = store.ActiveIntelDelivery{
		ID:    testutil.NewUUIDV4(),
		Intel: testutil.NewUUIDV4(),
		To:    testutil.NewUUIDV4(),
		Note:  nulls.NewString("dear"),
	}
	suite.intel = store.Intel{
		ID:        suite.create.Intel,
		Operation: testutil.NewUUIDV4(),
	}
}

func (suite *ControllerCreateActiveIntelDeliverySuite) TestBeginTxFail() {
	suite.c.DB.BeginFail = true
	err := suite.c.Ctrl.CreateActiveIntelDelivery(context.Background(), suite.create)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateActiveIntelDeliverySuite) TestCreateFail() {
	suite.c.Store.On("CreateActiveIntelDelivery", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.CreateActiveIntelDelivery(context.Background(), suite.create)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateActiveIntelDeliverySuite) TestRetrieveIntelFail() {
	suite.c.Store.On("CreateActiveIntelDelivery", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()
	suite.c.Store.On("IntelByID", mock.Anything, mock.Anything, mock.Anything).
		Return(store.Intel{}, errors.New("sad life"))
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.CreateActiveIntelDelivery(context.Background(), suite.create)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateActiveIntelDeliverySuite) TestOK() {
	suite.c.Store.On("CreateActiveIntelDelivery", mock.Anything, mock.Anything, suite.create).
		Return(nil)
	suite.c.Store.On("IntelByID", mock.Anything, mock.Anything, suite.create.Intel).
		Return(suite.intel, nil)
	defer suite.c.Store.AssertExpectations(suite.T())
	w := newWatcher(false)
	otherWatcher := newWatcher(false)
	suite.c.Ctrl.openIntelDeliveryWatchersByOperation[suite.intel.Operation] = w
	suite.c.Ctrl.openIntelDeliveryWatchersByOperation[testutil.NewUUIDV4()] = otherWatcher

	err := suite.c.Ctrl.CreateActiveIntelDelivery(context.Background(), suite.create)
	suite.NoError(err, "should not fail")
	suite.True(w.doNotify, "should have notified correct watcher")
	suite.False(otherWatcher.doNotify, "should not have notified watcher for other operation")
}

func TestController_CreateActiveIntelDelivery(t *testing.T) {
	suite.Run(t, new(ControllerCreateActiveIntelDeliverySuite))
}

// ControllerDeleteActiveIntelDeliveryByIDSuite tests
// Controller.DeleteActiveIntelDeliveryByID.
type ControllerDeleteActiveIntelDeliveryByIDSuite struct {
	suite.Suite
	c           *ControllerMock
	deliveryID  uuid.UUID
	operationID uuid.UUID
}

func (suite *ControllerDeleteActiveIntelDeliveryByIDSuite) SetupTest() {
	suite.c = NewMockController()
	suite.c.DB.GenTx = true
	suite.deliveryID = testutil.NewUUIDV4()
	suite.operationID = testutil.NewUUIDV4()
}

func (suite *ControllerDeleteActiveIntelDeliveryByIDSuite) TestBeginTxFail() {
	suite.c.DB.BeginFail = true

	err := suite.c.Ctrl.DeleteActiveIntelDeliveryByID(context.Background(), suite.deliveryID)
	suite.Error(err, "should fail")
}

func (suite *ControllerDeleteActiveIntelDeliveryByIDSuite) TestRetrieveOperationIDFail() {
	suite.c.Store.On("IntelOperationByDelivery", mock.Anything, mock.Anything, mock.Anything).
		Return(uuid.UUID{}, errors.New("sad life"))
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.DeleteActiveIntelDeliveryByID(context.Background(), suite.deliveryID)
	suite.Error(err, "should fail")
}

func (suite *ControllerDeleteActiveIntelDeliveryByIDSuite) TestDeleteFail() {
	suite.c.Store.On("IntelOperationByDelivery", mock.Anything, mock.Anything, mock.Anything).
		Return(suite.operationID, nil).Maybe()
	suite.c.Store.On("DeleteActiveIntelDeliveryByID", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad ilfe"))
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.DeleteActiveIntelDeliveryByID(context.Background(), suite.deliveryID)
	suite.Error(err, "should fail")
}

func (suite *ControllerDeleteActiveIntelDeliveryByIDSuite) TestOK() {
	suite.c.Store.On("IntelOperationByDelivery", mock.Anything, mock.Anything, suite.deliveryID).
		Return(suite.operationID, nil).Maybe()
	suite.c.Store.On("DeleteActiveIntelDeliveryByID", mock.Anything, mock.Anything, suite.deliveryID).
		Return(nil)
	defer suite.c.Store.AssertExpectations(suite.T())
	w := newWatcher(false)
	otherW := newWatcher(false)
	suite.c.Ctrl.openIntelDeliveryWatchersByOperation[suite.operationID] = w
	suite.c.Ctrl.openIntelDeliveryWatchersByOperation[testutil.NewUUIDV4()] = otherW

	err := suite.c.Ctrl.DeleteActiveIntelDeliveryByID(context.Background(), suite.deliveryID)
	suite.NoError(err, "should not fail")
	suite.True(w.doNotify, "should notify correct watcher")
	suite.False(otherW.doNotify, "should not notify watchers for other operations")
}

func TestController_DeleteActiveIntelDeliveryByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteActiveIntelDeliveryByIDSuite))
}

// ControllerCreateActiveIntelDeliveryAttemptSuite tests
// Controller.CreateActiveIntelDeliveryAttempt.
type ControllerCreateActiveIntelDeliveryAttemptSuite struct {
	suite.Suite
	c           *ControllerMock
	create      store.ActiveIntelDeliveryAttempt
	operationID uuid.UUID
}

func (suite *ControllerCreateActiveIntelDeliveryAttemptSuite) SetupTest() {
	suite.c = NewMockController()
	suite.c.DB.GenTx = true
	suite.create = store.ActiveIntelDeliveryAttempt{
		ID:       testutil.NewUUIDV4(),
		Delivery: testutil.NewUUIDV4(),
	}
	suite.operationID = testutil.NewUUIDV4()
}

func (suite *ControllerCreateActiveIntelDeliveryAttemptSuite) TestBeginTxFail() {
	suite.c.DB.BeginFail = true

	err := suite.c.Ctrl.CreateActiveIntelDeliveryAttempt(context.Background(), suite.create)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateActiveIntelDeliveryAttemptSuite) TestCreateFail() {
	suite.c.Store.On("CreateActiveIntelDeliveryAttempt", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.CreateActiveIntelDeliveryAttempt(context.Background(), suite.create)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateActiveIntelDeliveryAttemptSuite) TestRetrieveOperationIDFail() {
	tx := &testutil.DBTx{}
	suite.c.DB.Tx = []*testutil.DBTx{tx}
	suite.c.Store.On("CreateActiveIntelDeliveryAttempt", mock.Anything, tx, mock.Anything).
		Return(nil).Maybe()
	suite.c.Store.On("IntelOperationByDeliveryAttempt", mock.Anything, tx, mock.Anything).
		Return(uuid.Nil, errors.New("sad life"))
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.CreateActiveIntelDeliveryAttempt(context.Background(), suite.create)
	suite.NoError(err, "should not fail")
	suite.True(tx.IsCommitted, "should still commit tx")
}

func (suite *ControllerCreateActiveIntelDeliveryAttemptSuite) TestOK() {
	suite.c.Store.On("CreateActiveIntelDeliveryAttempt", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()
	suite.c.Store.On("IntelOperationByDeliveryAttempt", mock.Anything, mock.Anything, mock.Anything).
		Return(suite.operationID, nil)
	defer suite.c.Store.AssertExpectations(suite.T())
	w := newWatcher(false)
	otherW := newWatcher(false)
	suite.c.Ctrl.openIntelDeliveryWatchersByOperation[suite.operationID] = w
	suite.c.Ctrl.openIntelDeliveryWatchersByOperation[testutil.NewUUIDV4()] = otherW

	err := suite.c.Ctrl.CreateActiveIntelDeliveryAttempt(context.Background(), suite.create)
	suite.NoError(err, "should not fail")
	suite.True(w.doNotify, "should notify correct watcher")
	suite.False(otherW.doNotify, "should not notify watchers for other operations")
}

func TestController_CreateActiveIntelDeliveryAttempt(t *testing.T) {
	suite.Run(t, new(ControllerCreateActiveIntelDeliveryAttemptSuite))
}

// ControllerDeleteActiveIntelDeliveryAttemptByIDSuite tests
// Controller.DeleteActiveIntelDeliveryAttemptByID.
type ControllerDeleteActiveIntelDeliveryAttemptByIDSuite struct {
	suite.Suite
	c           *ControllerMock
	attemptID   uuid.UUID
	operationID uuid.UUID
}

func (suite *ControllerDeleteActiveIntelDeliveryAttemptByIDSuite) SetupTest() {
	suite.c = NewMockController()
	suite.c.DB.GenTx = true
	suite.attemptID = testutil.NewUUIDV4()
	suite.operationID = testutil.NewUUIDV4()
}

func (suite *ControllerDeleteActiveIntelDeliveryAttemptByIDSuite) TestBeginTxFail() {
	suite.c.DB.BeginFail = true

	err := suite.c.Ctrl.DeleteActiveIntelDeliveryAttemptByID(context.Background(), suite.attemptID)
	suite.Error(err, "should fail")
}

func (suite *ControllerDeleteActiveIntelDeliveryAttemptByIDSuite) TestRetrieveOperationFail() {
	suite.c.Store.On("IntelOperationByDeliveryAttempt", mock.Anything, mock.Anything, mock.Anything).
		Return(uuid.Nil, errors.New("sad life"))
	suite.c.Store.On("DeleteActiveIntelDeliveryAttemptByID", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.DeleteActiveIntelDeliveryAttemptByID(context.Background(), suite.attemptID)
	suite.NoError(err, "should not fail")
}

func (suite *ControllerDeleteActiveIntelDeliveryAttemptByIDSuite) TestDeleteFail() {
	suite.c.Store.On("IntelOperationByDeliveryAttempt", mock.Anything, mock.Anything, mock.Anything).
		Return(suite.operationID, nil).Maybe()
	suite.c.Store.On("DeleteActiveIntelDeliveryAttemptByID", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.DeleteActiveIntelDeliveryAttemptByID(context.Background(), suite.attemptID)
	suite.Error(err, "should fail")
}

func (suite *ControllerDeleteActiveIntelDeliveryAttemptByIDSuite) TestOK() {
	suite.c.Store.On("IntelOperationByDeliveryAttempt", mock.Anything, mock.Anything, mock.Anything).
		Return(suite.operationID, nil).Maybe()
	suite.c.Store.On("DeleteActiveIntelDeliveryAttemptByID", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	defer suite.c.Store.AssertExpectations(suite.T())
	w := newWatcher(false)
	otherW := newWatcher(false)
	suite.c.Ctrl.openIntelDeliveryWatchersByOperation[suite.operationID] = w
	suite.c.Ctrl.openIntelDeliveryWatchersByOperation[testutil.NewUUIDV4()] = otherW

	err := suite.c.Ctrl.DeleteActiveIntelDeliveryAttemptByID(context.Background(), suite.attemptID)
	suite.NoError(err, "should not fail")
	suite.True(w.doNotify, "should notify correct watcher")
	suite.False(otherW.doNotify, "should not notify watchers for other operations")
}

func TestController_DeleteActiveIntelDeliveryAttemptByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteActiveIntelDeliveryAttemptByIDSuite))
}

// ControllerSetAutoInteldDeliveryEnabldeForAddressBookEntrySuite tests
// Controller.SetAutoIntelDeliveryEnabledForAddressBookEntry.
type ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite struct {
	suite.Suite
	c                  *ControllerMock
	entryID            uuid.UUID
	enabled            bool
	affectedOperations []uuid.UUID
	affectedWatchers   []*openIntelDeliveryWatcher
	unaffectedWatchers []*openIntelDeliveryWatcher
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) SetupTest() {
	suite.c = NewMockController()
	suite.c.DB.GenTx = true
	suite.entryID = testutil.NewUUIDV4()
	suite.enabled = true
	suite.affectedOperations = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
	suite.affectedWatchers = make([]*openIntelDeliveryWatcher, 0)
	for _, affectedOperation := range suite.affectedOperations {
		w := newWatcher(false)
		suite.c.Ctrl.openIntelDeliveryWatchersByOperation[affectedOperation] = w
		suite.affectedWatchers = append(suite.affectedWatchers, w)
	}
	suite.unaffectedWatchers = make([]*openIntelDeliveryWatcher, 0)
	for i := 0; i < 16; i++ {
		w := newWatcher(false)
		suite.c.Ctrl.openIntelDeliveryWatchersByOperation[testutil.NewUUIDV4()] = w
		suite.unaffectedWatchers = append(suite.unaffectedWatchers, w)
	}
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestBeginTxFail() {
	suite.c.DB.BeginFail = true

	err := suite.c.Ctrl.SetAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID, suite.enabled)
	suite.Error(err, "should fail")
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestSetFail() {
	suite.c.Store.On("SetAutoIntelDeliveryEnabledForEntry", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	suite.c.Store.On("IntelOperationsByActiveIntelDeliveryRecipient", mock.Anything, mock.Anything, mock.Anything).
		Return(suite.affectedOperations, nil).Maybe()
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.SetAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID, suite.enabled)
	suite.Error(err, "should fail")
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestRetrieveIntelOperationsFail() {
	suite.c.Store.On("SetAutoIntelDeliveryEnabledForEntry", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()
	suite.c.Store.On("IntelOperationsByActiveIntelDeliveryRecipient", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("sad life"))
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.SetAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID, suite.enabled)
	suite.Error(err, "should fail")
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestOK() {
	suite.c.Store.On("SetAutoIntelDeliveryEnabledForEntry", mock.Anything, mock.Anything, suite.entryID, suite.enabled).
		Return(nil)
	suite.c.Store.On("IntelOperationsByActiveIntelDeliveryRecipient", mock.Anything, mock.Anything, suite.entryID).
		Return(suite.affectedOperations, nil)
	defer suite.c.Store.AssertExpectations(suite.T())

	err := suite.c.Ctrl.SetAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID, suite.enabled)
	suite.Require().NoError(err, "should not fail")
	for _, watcher := range suite.affectedWatchers {
		suite.True(watcher.doNotify, "should notify affected watchers")
	}
	for _, watcher := range suite.unaffectedWatchers {
		suite.False(watcher.doNotify, "should not notify unaffected watchers")
	}
}

func TestController_SetAutoIntelDeliveryEnabledForAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite))
}
