package controller

import (
	"encoding/json"
	"errors"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// ControllerCreateIntelDeliveryAttemptSuite tests
// Controller.CreateIntelDeliveryAttempt.
type ControllerCreateIntelDeliveryAttemptSuite struct {
	suite.Suite
	ctrl                 *ControllerMock
	tx                   *testutil.DBTx
	sampleAttempt        store.AcceptedIntelDeliveryAttempt
	sampleIntelToDeliver store.IntelToDeliver
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleAttempt = store.AcceptedIntelDeliveryAttempt{
		ID:         testutil.NewUUIDV4(),
		AssignedTo: testutil.NewUUIDV4(),
		Delivery:   testutil.NewUUIDV4(),
		Channel:    testutil.NewUUIDV4(),
		CreatedAt:  time.Date(2022, 9, 8, 1, 45, 59, 0, time.UTC),
		IsActive:   true,
		StatusTS:   time.Date(2022, 9, 8, 1, 46, 10, 0, time.UTC),
		Note:       nulls.NewString("make"),
		AcceptedAt: time.Date(2022, 9, 8, 1, 46, 23, 0, time.UTC),
	}
	suite.sampleIntelToDeliver = store.IntelToDeliver{
		Attempt:    suite.sampleAttempt.ID,
		ID:         testutil.NewUUIDV4(),
		CreatedAt:  time.Date(2022, 9, 8, 1, 46, 44, 0, time.UTC),
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       "DWNdn",
		Content:    json.RawMessage(`{"hello":"world"}`),
		Importance: 834,
	}
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestRetrieveNotificationChannelFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("NotificationChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(store.NotificationChannel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt, suite.sampleIntelToDeliver)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestCreateAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("NotificationChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(store.NotificationChannel{}, nil)
	suite.ctrl.Store.On("CreateAcceptedIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt, suite.sampleIntelToDeliver)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestCreateIntelFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("NotificationChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(store.NotificationChannel{}, nil)
	suite.ctrl.Store.On("CreateAcceptedIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("CreateIntelToDeliver", timeout, suite.tx, suite.sampleIntelToDeliver).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt, suite.sampleIntelToDeliver)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("NotificationChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(store.NotificationChannel{}, nil)
	suite.ctrl.Store.On("CreateIntelToDeliver", timeout, suite.tx, suite.sampleIntelToDeliver).
		Return(nil)
	suite.ctrl.Store.On("CreateAcceptedIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryNotificationPending", timeout, suite.tx, suite.sampleAttempt.ID, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt, suite.sampleIntelToDeliver)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("NotificationChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(store.NotificationChannel{}, nil)
	suite.ctrl.Store.On("CreateIntelToDeliver", timeout, suite.tx, suite.sampleIntelToDeliver).
		Return(nil)
	suite.ctrl.Store.On("CreateAcceptedIntelDeliveryAttempt", timeout, suite.tx, mock.MatchedBy(func(attemptRaw any) bool {
		attempt, ok := attemptRaw.(store.AcceptedIntelDeliveryAttempt)
		if !ok {
			return false
		}
		attempt.AcceptedAt = suite.sampleAttempt.AcceptedAt
		return attempt == suite.sampleAttempt
	})).Return(nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryNotificationPending", timeout, suite.tx, suite.sampleAttempt.ID, mock.Anything).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt, suite.sampleIntelToDeliver)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_CreateIntelDeliveryAttempt(t *testing.T) {
	suite.Run(t, new(ControllerCreateIntelDeliveryAttemptSuite))
}

// ControllerUpdateIntelDeliveryAttemptStatusSuite tests
// Controller.UpdateIntelDeliveryAttemptStatus.
type ControllerUpdateIntelDeliveryAttemptStatusSuite struct {
	suite.Suite
	ctrl         *ControllerMock
	tx           *testutil.DBTx
	sampleStatus store.AcceptedIntelDeliveryAttemptStatus
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleStatus = store.AcceptedIntelDeliveryAttemptStatus{
		ID:       testutil.NewUUIDV4(),
		IsActive: true,
		StatusTS: time.Date(2022, 9, 8, 2, 5, 31, 0, time.UTC),
		Note:     nulls.NewString("deceit"),
	}
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestStoreUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleStatus).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleStatus)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestNotFound() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleStatus).
		Return(meh.NewNotFoundErr("not found", nil))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleStatus)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleStatus).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleStatus)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateIntelDeliveryAttemptStatus(t *testing.T) {
	suite.Run(t, new(ControllerUpdateIntelDeliveryAttemptStatusSuite))
}
