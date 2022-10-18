package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateIntelDeliveryAttemptSuite tests
// Controller.CreateIntelDeliveryAttempt.
type ControllerCreateIntelDeliveryAttemptSuite struct {
	suite.Suite
	ctrl                       *ControllerMock
	tx                         *testutil.DBTx
	sampleAttempt              store.AcceptedIntelDeliveryAttempt
	sampleRadioChannel         store.RadioChannel
	sampleCreatedRadioDelivery store.RadioDelivery
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleAttempt = store.AcceptedIntelDeliveryAttempt{
		ID:              testutil.NewUUIDV4(),
		Intel:           testutil.NewUUIDV4(),
		IntelOperation:  testutil.NewUUIDV4(),
		IntelImportance: 950,
		AssignedTo:      testutil.NewUUIDV4(),
		AssignedToLabel: "not",
		AssignedToUser:  nulls.NewUUID(testutil.NewUUIDV4()),
		Delivery:        testutil.NewUUIDV4(),
		Channel:         testutil.NewUUIDV4(),
		CreatedAt:       testutil.NewRandomTime(),
		IsActive:        true,
		StatusTS:        testutil.NewRandomTime(),
		Note:            nulls.NewString("these"),
		AcceptedAt:      testutil.NewRandomTime(),
	}
	suite.sampleRadioChannel = store.RadioChannel{
		ID:      suite.sampleAttempt.Channel,
		Entry:   testutil.NewUUIDV4(),
		Label:   "thicken",
		Timeout: 530,
		Info:    "young",
	}
	suite.sampleCreatedRadioDelivery = store.RadioDelivery{
		Attempt:    suite.sampleAttempt.ID,
		PickedUpBy: uuid.NullUUID{},
		PickedUpAt: nulls.Time{},
		Success:    nulls.Bool{},
		SuccessTS:  testutil.NewRandomTime(),
		Note:       "sun",
	}
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestRadioChannelByIDFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(store.RadioChannel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestRadioChannelNotFound() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(store.RadioChannel{}, meh.NewNotFoundErr("not found", nil))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestCreateAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(suite.sampleRadioChannel, nil)
	suite.ctrl.Store.On("CreateAcceptedIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestCreateRadioDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(suite.sampleRadioChannel, nil)
	suite.ctrl.Store.On("CreateAcceptedIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("CreateRadioDelivery", timeout, suite.tx, suite.sampleAttempt.ID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestRetrieveCreatedRadioDelivery() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(suite.sampleRadioChannel, nil)
	suite.ctrl.Store.On("CreateAcceptedIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("CreateRadioDelivery", timeout, suite.tx, suite.sampleAttempt.ID).
		Return(nil)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttempt.ID).
		Return(store.RadioDelivery{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(suite.sampleRadioChannel, nil)
	suite.ctrl.Store.On("CreateAcceptedIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("CreateRadioDelivery", timeout, suite.tx, suite.sampleAttempt.ID).
		Return(nil)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttempt.ID).
		Return(suite.sampleCreatedRadioDelivery, nil)
	suite.ctrl.Notifier.On("NotifyRadioDeliveryReadyForPickup", timeout, suite.tx, mock.Anything,
		suite.sampleCreatedRadioDelivery.Note).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioChannelByID", timeout, suite.tx, suite.sampleAttempt.Channel).
		Return(suite.sampleRadioChannel, nil)
	suite.ctrl.Store.On("CreateAcceptedIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("CreateRadioDelivery", timeout, suite.tx, suite.sampleAttempt.ID).
		Return(nil)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttempt.ID).
		Return(suite.sampleCreatedRadioDelivery, nil)
	suite.ctrl.Notifier.On("NotifyRadioDeliveryReadyForPickup", timeout, suite.tx, mock.Anything,
		suite.sampleCreatedRadioDelivery.Note).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(timeout, suite.tx, suite.sampleAttempt)
		suite.NoError(err, "should not fail")
		// Assure notify request scheduled.
		select {
		case <-timeout.Done():
			suite.Fail("no notify request for operation")
		case op := <-suite.ctrl.Ctrl.connUpdateNotifier.notifyRequestsForOperation:
			suite.Equal(op, suite.sampleAttempt.IntelOperation, "should have scheduled notify request for operation")
		}
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
	ctrl                       *ControllerMock
	tx                         *testutil.DBTx
	sampleNewStatus            store.AcceptedIntelDeliveryAttemptStatus
	sampleRadioDelivery        store.RadioDelivery
	sampleUpdatedRadioDelivery store.RadioDelivery
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleNewStatus = store.AcceptedIntelDeliveryAttemptStatus{
		ID:       testutil.NewUUIDV4(),
		IsActive: false,
		StatusTS: testutil.NewRandomTime(),
		Note:     nulls.NewString("lovely"),
	}
	suite.sampleRadioDelivery = store.RadioDelivery{
		Attempt:    suite.sampleNewStatus.ID,
		PickedUpBy: uuid.NullUUID{},
		PickedUpAt: nulls.Time{},
		Success:    nulls.Bool{},
		SuccessTS:  testutil.NewRandomTime(),
		Note:       "sun",
	}
	suite.sampleUpdatedRadioDelivery = suite.sampleRadioDelivery
	suite.sampleUpdatedRadioDelivery.Success = nulls.NewBool(false)
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleNewStatus).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleNewStatus)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestAttemptNotFound() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleNewStatus).
		Return(meh.NewNotFoundErr("not found", nil)).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleNewStatus)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestNewActive() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleNewStatus.IsActive = true
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleNewStatus).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleNewStatus)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestRetrieveRadioDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleNewStatus).
		Return(nil)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID).
		Return(store.RadioDelivery{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleNewStatus)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestRadioDeliveryAlreadyFinished() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleRadioDelivery.Success = nulls.NewBool(false)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleNewStatus).
		Return(nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID).
		Return(suite.sampleRadioDelivery, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleNewStatus)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestUpdateRadioDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleNewStatus).
		Return(nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("UpdateRadioDeliveryStatusByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID,
		nulls.NewBool(false), mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleNewStatus)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestRetrieveUpdatedRadioDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleNewStatus).
		Return(nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("UpdateRadioDeliveryStatusByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID,
		nulls.NewBool(false), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID).
		Return(store.RadioDelivery{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleNewStatus)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleNewStatus).
		Return(nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("UpdateRadioDeliveryStatusByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID,
		nulls.NewBool(false), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID).
		Return(suite.sampleUpdatedRadioDelivery, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryFinished", timeout, suite.tx, suite.sampleUpdatedRadioDelivery).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleNewStatus)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UpdateAcceptedIntelDeliveryAttemptStatus", timeout, suite.tx, suite.sampleNewStatus).
		Return(nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("UpdateRadioDeliveryStatusByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID,
		nulls.NewBool(false), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleNewStatus.ID).
		Return(suite.sampleUpdatedRadioDelivery, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryFinished", timeout, suite.tx, suite.sampleUpdatedRadioDelivery).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatus(timeout, suite.tx, suite.sampleNewStatus)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateIntelDeliveryAttemptStatus(t *testing.T) {
	suite.Run(t, new(ControllerUpdateIntelDeliveryAttemptStatusSuite))
}
