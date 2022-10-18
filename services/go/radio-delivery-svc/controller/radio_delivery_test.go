package controller

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
	"time"
)

// ControllerPickUpNextRadioDeliverySuite tests
// Controller.PickUpNextRadioDelivery.
type ControllerPickUpNextRadioDeliverySuite struct {
	suite.Suite
	ctrl                       *ControllerMock
	tx                         *testutil.DBTx
	sampleOperationID          uuid.UUID
	sampleBy                   uuid.UUID
	sampleOperationsOfBy       []uuid.UUID
	sampleActiveDeliveries     []store.ActiveRadioDelivery
	sampleIntelDeliveryAttempt store.AcceptedIntelDeliveryAttempt
	sampleRadioDelivery        store.RadioDelivery
}

func (suite *ControllerPickUpNextRadioDeliverySuite) genActiveDelivery(importance int, pickedUp bool) store.ActiveRadioDelivery {
	var pickedUpAt nulls.Time
	if pickedUp {
		pickedUpAt = nulls.NewTime(testutil.NewRandomTime())
	}
	return store.ActiveRadioDelivery{
		Attempt:          testutil.NewUUIDV4(),
		PickedUpAt:       pickedUpAt,
		IntelOperation:   testutil.NewUUIDV4(),
		IntelImportance:  importance,
		AttemptCreatedAt: testutil.NewRandomTime(),
	}
}

func (suite *ControllerPickUpNextRadioDeliverySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleOperationID = testutil.NewUUIDV4()
	suite.sampleBy = testutil.NewUUIDV4()
	suite.sampleOperationsOfBy = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		suite.sampleOperationID,
		testutil.NewUUIDV4(),
	}
	suite.sampleActiveDeliveries = []store.ActiveRadioDelivery{
		suite.genActiveDelivery(267, true),
		suite.genActiveDelivery(755, false),
		suite.genActiveDelivery(445, false),
		suite.genActiveDelivery(516, true),
		suite.genActiveDelivery(445, false),
		suite.genActiveDelivery(516, true),
		suite.genActiveDelivery(516, true),
	}
	suite.sampleIntelDeliveryAttempt = store.AcceptedIntelDeliveryAttempt{
		ID:              suite.sampleActiveDeliveries[0].Attempt,
		Intel:           testutil.NewUUIDV4(),
		IntelOperation:  suite.sampleOperationID,
		IntelImportance: suite.sampleActiveDeliveries[0].IntelImportance,
		AssignedTo:      testutil.NewUUIDV4(),
		AssignedToLabel: "defend",
		AssignedToUser:  nulls.NewUUID(testutil.NewUUIDV4()),
		Delivery:        testutil.NewUUIDV4(),
		Channel:         testutil.NewUUIDV4(),
		CreatedAt:       testutil.NewRandomTime(),
		IsActive:        true,
		StatusTS:        testutil.NewRandomTime(),
		Note:            nulls.NewString("young"),
		AcceptedAt:      testutil.NewRandomTime(),
	}
	suite.sampleRadioDelivery = store.RadioDelivery{
		Attempt:    suite.sampleIntelDeliveryAttempt.ID,
		PickedUpBy: nulls.NewUUID(suite.sampleBy),
		PickedUpAt: nulls.NewTime(testutil.NewRandomTime()),
		Success:    nulls.Bool{},
		SuccessTS:  testutil.NewRandomTime(),
		Note:       "ease",
	}
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestBeginTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestRetrieveOperationsForUserFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(nil, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestNoMemberOfOperation1() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return([]uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		}, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return error with correct code")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestNoMemberOfOperation2() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return([]uuid.UUID{}, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return error with correct code")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestRetrieveActiveRadioDeliveriesFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(nil, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestNoOpenRadioDeliveries1() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return([]store.ActiveRadioDelivery{
			{PickedUpAt: nulls.NewTime(testutil.NewRandomTime())},
			{PickedUpAt: nulls.NewTime(testutil.NewRandomTime())},
			{PickedUpAt: nulls.NewTime(testutil.NewRandomTime())},
		}, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, ok, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.False(ok, "should return false")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestNoOpenRadioDeliveries2() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return([]store.ActiveRadioDelivery{}, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, ok, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.False(ok, "should return false")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestMarkRadioDeliveryAsPickedUpFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(suite.sampleActiveDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestRetrieveIntelDeliveryAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(suite.sampleActiveDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, mock.Anything).
		Return(store.AcceptedIntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestRetrieveRadioDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(suite.sampleActiveDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, mock.Anything).
		Return(suite.sampleIntelDeliveryAttempt, nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, mock.Anything).
		Return(store.RadioDelivery{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(suite.sampleActiveDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, mock.Anything).
		Return(suite.sampleIntelDeliveryAttempt, nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, mock.Anything).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryPickedUp", timeout, suite.tx, mock.Anything, suite.sampleBy,
		suite.sampleRadioDelivery.PickedUpAt.Time).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(suite.sampleActiveDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, mock.Anything).
		Return(suite.sampleIntelDeliveryAttempt, nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, mock.Anything).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryPickedUp", timeout, suite.tx, mock.Anything, suite.sampleBy,
		suite.sampleRadioDelivery.PickedUpAt.Time).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, ok, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.True(ok, "should return true")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestPickSortByOldestFirst1() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	activeRadioDeliveries := []store.ActiveRadioDelivery{
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(200, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.NewTime(testutil.NewRandomTime()),
			AttemptCreatedAt: time.Date(300, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(400, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  300,
		},
	}
	shouldPick := activeRadioDeliveries[0].Attempt
	pickedAttempt := suite.sampleIntelDeliveryAttempt
	pickedAttempt.ID = shouldPick
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(activeRadioDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, shouldPick).
		Return(pickedAttempt, nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, shouldPick).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryPickedUp", timeout, suite.tx, shouldPick, suite.sampleBy,
		suite.sampleRadioDelivery.PickedUpAt.Time).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		attempt, ok, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.True(ok, "should return true")
		suite.Equal(shouldPick, attempt.ID, "should pick correct attempt")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestPickSortByOldestFirst2() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	activeRadioDeliveries := []store.ActiveRadioDelivery{
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(700, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.NewTime(testutil.NewRandomTime()),
			AttemptCreatedAt: time.Date(300, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(500, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  300,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.NewTime(testutil.NewRandomTime()),
			AttemptCreatedAt: time.Date(600, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.NewTime(testutil.NewRandomTime()),
			AttemptCreatedAt: time.Date(800, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
	}
	shouldPick := activeRadioDeliveries[2].Attempt
	pickedAttempt := suite.sampleIntelDeliveryAttempt
	pickedAttempt.ID = shouldPick
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(activeRadioDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, shouldPick).
		Return(pickedAttempt, nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, shouldPick).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryPickedUp", timeout, suite.tx, shouldPick, suite.sampleBy,
		suite.sampleRadioDelivery.PickedUpAt.Time).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		attempt, ok, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.True(ok, "should return true")
		suite.Equal(shouldPick, attempt.ID, "should pick correct attempt")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestPickSortByImportanceFirst1() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	activeRadioDeliveries := []store.ActiveRadioDelivery{
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(200, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  300,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.NewTime(testutil.NewRandomTime()),
			AttemptCreatedAt: time.Date(300, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(400, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  900,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.NewTime(testutil.NewRandomTime()),
			AttemptCreatedAt: time.Date(300, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
	}
	shouldPick := activeRadioDeliveries[2].Attempt
	pickedAttempt := suite.sampleIntelDeliveryAttempt
	pickedAttempt.ID = shouldPick
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(activeRadioDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, shouldPick).
		Return(pickedAttempt, nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, shouldPick).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryPickedUp", timeout, suite.tx, shouldPick, suite.sampleBy,
		suite.sampleRadioDelivery.PickedUpAt.Time).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		attempt, ok, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.True(ok, "should return true")
		suite.Equal(shouldPick, attempt.ID, "should pick correct attempt")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestPickSortByImportanceFirst2() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	activeRadioDeliveries := []store.ActiveRadioDelivery{
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(100, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.NewTime(testutil.NewRandomTime()),
			AttemptCreatedAt: time.Date(300, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(500, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  300,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.NewTime(testutil.NewRandomTime()),
			AttemptCreatedAt: time.Date(800, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
	}
	shouldPick := activeRadioDeliveries[2].Attempt
	pickedAttempt := suite.sampleIntelDeliveryAttempt
	pickedAttempt.ID = shouldPick
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(activeRadioDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, shouldPick).
		Return(pickedAttempt, nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, shouldPick).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryPickedUp", timeout, suite.tx, shouldPick, suite.sampleBy,
		suite.sampleRadioDelivery.PickedUpAt.Time).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		attempt, ok, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.True(ok, "should return true")
		suite.Equal(shouldPick, attempt.ID, "should pick correct attempt")
	}()

	wait()
}

func (suite *ControllerPickUpNextRadioDeliverySuite) TestPickSortByImportanceFirst3() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	activeRadioDeliveries := []store.ActiveRadioDelivery{
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(100, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  200,
		},
		{
			Attempt:          testutil.NewUUIDV4(),
			PickedUpAt:       nulls.Time{},
			AttemptCreatedAt: time.Date(500, 1, 1, 1, 1, 1, 1, time.UTC),
			IntelImportance:  300,
		},
	}
	shouldPick := activeRadioDeliveries[1].Attempt
	pickedAttempt := suite.sampleIntelDeliveryAttempt
	pickedAttempt.ID = shouldPick
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.tx, suite.sampleBy).
		Return(suite.sampleOperationsOfBy, nil).Once()
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, nulls.NewUUID(suite.sampleOperationID)).
		Return(activeRadioDeliveries, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, mock.Anything,
		nulls.NewUUID(suite.sampleBy), mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, shouldPick).
		Return(pickedAttempt, nil).Once()
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, shouldPick).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryPickedUp", timeout, suite.tx, shouldPick, suite.sampleBy,
		suite.sampleRadioDelivery.PickedUpAt.Time).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		attempt, ok, err := suite.ctrl.Ctrl.PickUpNextRadioDelivery(timeout, suite.sampleOperationID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.True(ok, "should return true")
		suite.Equal(shouldPick, attempt.ID, "should pick correct attempt")
	}()

	wait()
}

func TestController_PickUpNextRadioDelivery(t *testing.T) {
	suite.Run(t, new(ControllerPickUpNextRadioDeliverySuite))
}

// ControllerReleasePickedUpRadioDeliverySuite tests
// Controller.ReleasePickedUpRadioDelivery.
type ControllerReleasePickedUpRadioDeliverySuite struct {
	suite.Suite
	ctrl                *ControllerMock
	tx                  *testutil.DBTx
	sampleAttemptID     uuid.UUID
	sampleRadioDelivery store.RadioDelivery
	sampleAttempt       store.AcceptedIntelDeliveryAttempt
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.sampleRadioDelivery = store.RadioDelivery{
		Attempt:    suite.sampleAttemptID,
		PickedUpBy: nulls.NewUUID(testutil.NewUUIDV4()),
		PickedUpAt: nulls.NewTime(testutil.NewRandomTime()),
		Success:    nulls.Bool{},
		SuccessTS:  testutil.NewRandomTime(),
		Note:       "upper",
	}
	suite.sampleAttempt = store.AcceptedIntelDeliveryAttempt{
		ID:             suite.sampleAttemptID,
		IntelOperation: testutil.NewUUIDV4(),
	}
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) TestBeginTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.ReleasePickedUpRadioDelivery(timeout, suite.sampleAttemptID, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) TestRetrieveRadioDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.RadioDelivery{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.ReleasePickedUpRadioDelivery(timeout, suite.sampleAttemptID, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) TestRetrieveIntelDeliveryAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.AcceptedIntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.ReleasePickedUpRadioDelivery(timeout, suite.sampleAttemptID, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) TestDeliveryNotActiveAnymore() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleRadioDelivery.Success = nulls.NewBool(false)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.ReleasePickedUpRadioDelivery(timeout, suite.sampleAttemptID, uuid.NullUUID{})
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrBadInput, meh.ErrorCode(err), "should return correct error code")
	}()

	wait()
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) TestDeliveryNotPickedUp() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleRadioDelivery.PickedUpBy = uuid.NullUUID{}
	suite.sampleRadioDelivery.PickedUpAt = nulls.Time{}
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.ReleasePickedUpRadioDelivery(timeout, suite.sampleAttemptID, uuid.NullUUID{})
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrBadInput, meh.ErrorCode(err), "should return correct error code")
	}()

	wait()
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) TestLimitToPickedUpByNotOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleRadioDelivery.PickedUpBy = nulls.NewUUID(testutil.NewUUIDV4())
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.ReleasePickedUpRadioDelivery(timeout, suite.sampleAttemptID, nulls.NewUUID(testutil.NewUUIDV4()))
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return correct error code")
	}()

	wait()
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) TestReleaseFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, suite.sampleAttemptID,
		uuid.NullUUID{}, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.ReleasePickedUpRadioDelivery(timeout, suite.sampleAttemptID, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, suite.sampleAttemptID,
		uuid.NullUUID{}, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryReleased", timeout, suite.tx, suite.sampleAttemptID, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.ReleasePickedUpRadioDelivery(timeout, suite.sampleAttemptID, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerReleasePickedUpRadioDeliverySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RadioDeliveryByAttempt", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleRadioDelivery, nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, suite.sampleAttemptID,
		uuid.NullUUID{}, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryReleased", timeout, suite.tx, suite.sampleAttemptID, mock.Anything).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.ReleasePickedUpRadioDelivery(timeout, suite.sampleAttemptID, uuid.NullUUID{})
		suite.Require().NoError(err, "should not fail")
		// Assure notification-update scheduled.
		select {
		case <-timeout.Done():
			suite.Fail("should schedule notification-update")
		case op := <-suite.ctrl.Ctrl.connUpdateNotifier.notifyRequestsForOperation:
			suite.Equal(suite.sampleAttempt.IntelOperation, op, "should schedule notification-update for correct operation")
		}
	}()

	wait()
}

func TestController_ReleasePickedUpRadioDelivery(t *testing.T) {
	suite.Run(t, new(ControllerReleasePickedUpRadioDeliverySuite))
}

// controllerReleaseTimedOutRadioDeliveriesSuite tests
// Controller.releaseTimedOutRadioDeliveries.
type controllerReleaseTimedOutRadioDeliveriesSuite struct {
	suite.Suite
	ctrl                               *ControllerMock
	tx                                 *testutil.DBTx
	sampleActiveDelivery               store.ActiveRadioDelivery
	sampleUpdatedActiveDeliveryAttempt store.AcceptedIntelDeliveryAttempt
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleActiveDelivery = store.ActiveRadioDelivery{
		Attempt:          testutil.NewUUIDV4(),
		PickedUpAt:       nulls.NewTime(time.Now().Add(-suite.ctrl.samplePickedUpTimeout - 10*time.Second)),
		IntelOperation:   testutil.NewUUIDV4(),
		IntelImportance:  469,
		AttemptCreatedAt: testutil.NewRandomTime(),
	}
	suite.sampleUpdatedActiveDeliveryAttempt = store.AcceptedIntelDeliveryAttempt{
		ID:             suite.sampleActiveDelivery.Attempt,
		IntelOperation: suite.sampleActiveDelivery.IntelOperation,
	}
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) expectNotifyUpdatesForOperations(ctx context.Context, forOperations ...uuid.UUID) {
	// Special case: none expected.
	if len(forOperations) == 0 {
		select {
		case op := <-suite.ctrl.Ctrl.connUpdateNotifier.notifyRequestsForOperation:
			suite.Failf("unexpected", "no notify-updates expected but got for %v", op)
		default:
			return
		}
	}
	// Regular behavior.
	operations := make(map[uuid.UUID]int, 0)
	for _, operation := range forOperations {
		operations[operation] = operations[operation] + 1
	}
	for len(operations) > 0 {
		select {
		case <-ctx.Done():
			suite.Failf("timeout", "expected still updates for %d operations", len(operations))
		case op := <-suite.ctrl.Ctrl.connUpdateNotifier.notifyRequestsForOperation:
			remaining := operations[op]
			if remaining <= 0 {
				suite.Failf("unexpected", "unexpected notify-update for operation %v", op)
			} else {
				remaining--
				if remaining <= 0 {
					delete(operations, op)
				}
			}
		}
	}
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestBeginTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.Error(err, "should fail")
		suite.expectNotifyUpdatesForOperations(timeout)
	}()

	wait()
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestRetrieveActiveDeliveriesFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, uuid.NullUUID{}).
		Return(nil, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.Error(err, "should fail")
		suite.expectNotifyUpdatesForOperations(timeout)
	}()

	wait()
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestNoActiveDeliveries() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, uuid.NullUUID{}).
		Return([]store.ActiveRadioDelivery{}, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.NoError(err, "should not fail")
		suite.expectNotifyUpdatesForOperations(timeout)
	}()

	wait()
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestNoPickedUpDeliveries() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, uuid.NullUUID{}).
		Return([]store.ActiveRadioDelivery{
			{},
			{},
		}, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.NoError(err, "should not fail")
		suite.expectNotifyUpdatesForOperations(timeout)
	}()

	wait()
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestNoTimedOutDeliveries() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, uuid.NullUUID{}).
		Return([]store.ActiveRadioDelivery{
			{PickedUpAt: nulls.NewTime(time.Now().Add(-suite.ctrl.samplePickedUpTimeout / 2))},
			{PickedUpAt: nulls.NewTime(time.Now().Add(-suite.ctrl.samplePickedUpTimeout / 2))},
		}, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.NoError(err, "should not fail")
		suite.expectNotifyUpdatesForOperations(timeout)
	}()

	wait()
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestReleaseFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, uuid.NullUUID{}).
		Return([]store.ActiveRadioDelivery{suite.sampleActiveDelivery}, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, suite.sampleActiveDelivery.Attempt,
		uuid.NullUUID{}, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.Error(err, "should fail")
		suite.expectNotifyUpdatesForOperations(timeout)
	}()

	wait()
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestRetrieveUpdatedAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, uuid.NullUUID{}).
		Return([]store.ActiveRadioDelivery{suite.sampleActiveDelivery}, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, suite.sampleActiveDelivery.Attempt,
		uuid.NullUUID{}, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleActiveDelivery.Attempt).
		Return(store.AcceptedIntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.Error(err, "should fail")
		suite.expectNotifyUpdatesForOperations(timeout)
	}()

	wait()
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, uuid.NullUUID{}).
		Return([]store.ActiveRadioDelivery{suite.sampleActiveDelivery}, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, suite.sampleActiveDelivery.Attempt,
		uuid.NullUUID{}, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleActiveDelivery.Attempt).
		Return(suite.sampleUpdatedActiveDeliveryAttempt, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryReadyForPickup", timeout, suite.tx,
		suite.sampleUpdatedActiveDeliveryAttempt, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.Error(err, "should fail")
		suite.expectNotifyUpdatesForOperations(timeout)
	}()

	wait()
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestOKSingle() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, uuid.NullUUID{}).
		Return([]store.ActiveRadioDelivery{suite.sampleActiveDelivery}, nil).Once()
	suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, suite.sampleActiveDelivery.Attempt,
		uuid.NullUUID{}, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleActiveDelivery.Attempt).
		Return(suite.sampleUpdatedActiveDeliveryAttempt, nil).Once()
	suite.ctrl.Notifier.On("NotifyRadioDeliveryReadyForPickup", timeout, suite.tx,
		suite.sampleUpdatedActiveDeliveryAttempt, mock.Anything).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.Require().NoError(err, "should not fail")
		suite.expectNotifyUpdatesForOperations(timeout, suite.sampleActiveDelivery.IntelOperation)
	}()

	wait()
}

func (suite *controllerReleaseTimedOutRadioDeliveriesSuite) TestOKMulti() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	// Gen data.
	timedOutCount := 16
	okCount := 64
	activeDeliveries := make([]store.ActiveRadioDelivery, 0, timedOutCount+okCount)
	expectNotifyForOperations := []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
	for timedOutCount > 0 || okCount > 0 {
		genTimeout := timedOutCount > 0
		if genTimeout && okCount > 0 {
			genTimeout = rand.Intn(10) > 7
		}
		if genTimeout {
			timedOutCount--
			d := store.ActiveRadioDelivery{
				Attempt:          testutil.NewUUIDV4(),
				PickedUpAt:       nulls.NewTime(time.Now().Add(-suite.ctrl.samplePickedUpTimeout - 10*time.Second)),
				IntelOperation:   expectNotifyForOperations[timedOutCount%len(expectNotifyForOperations)],
				IntelImportance:  574,
				AttemptCreatedAt: testutil.NewRandomTime(),
			}
			activeDeliveries = append(activeDeliveries, d)
			suite.ctrl.Store.On("MarkRadioDeliveryAsPickedUpByAttempt", timeout, suite.tx, d.Attempt, uuid.NullUUID{}, mock.Anything).
				Return(nil).Once()
			updatedAttempt := store.AcceptedIntelDeliveryAttempt{
				ID:             d.Attempt,
				IntelOperation: d.IntelOperation,
			}
			suite.ctrl.Store.On("AcceptedIntelDeliveryAttemptByID", timeout, suite.tx, d.Attempt).
				Return(updatedAttempt, nil).Once()
			suite.ctrl.Notifier.On("NotifyRadioDeliveryReadyForPickup", timeout, suite.tx, updatedAttempt, mock.Anything).
				Return(nil).Once()
		} else {
			okCount--
			activeDeliveries = append(activeDeliveries, store.ActiveRadioDelivery{
				Attempt:          testutil.NewUUIDV4(),
				PickedUpAt:       nulls.NewTime(time.Now().Add(10 * time.Second)),
				IntelOperation:   testutil.NewUUIDV4(),
				IntelImportance:  153,
				AttemptCreatedAt: testutil.NewRandomTime(),
			})
		}
	}
	suite.ctrl.Store.On("ActiveRadioDeliveriesAndLockOrWait", timeout, suite.tx, uuid.NullUUID{}).
		Return(activeDeliveries, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.releaseTimedOutRadioDeliveries(timeout)
		suite.Require().NoError(err, "should not fail")
		suite.expectNotifyUpdatesForOperations(timeout, expectNotifyForOperations...)
	}()

	wait()
}

func TestController_releaseTimedOutRadioDeliveries(t *testing.T) {
	suite.Run(t, new(controllerReleaseTimedOutRadioDeliveriesSuite))
}
