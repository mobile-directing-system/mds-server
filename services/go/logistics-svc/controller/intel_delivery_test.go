package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// controllerLookAfterDeliverySuite tests Controller.lookAfterDelivery.
type controllerLookAfterDeliverySuite struct {
	suite.Suite
	ctrl                   *ControllerMock
	tx                     *testutil.DBTx
	sampleID               uuid.UUID
	sampleDelivery         store.IntelDelivery
	sampleDeliveryAttempts []store.IntelDeliveryAttempt
	sampleChannel          store.Channel
}

func (suite *controllerLookAfterDeliverySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleID = testutil.NewUUIDV4()
	suite.sampleDelivery = store.IntelDelivery{
		ID:         suite.sampleID,
		Assignment: testutil.NewUUIDV4(),
		IsActive:   true,
		Success:    false,
		Note:       nulls.NewString("club"),
	}
	suite.sampleDeliveryAttempts = []store.IntelDeliveryAttempt{
		{
			ID:        testutil.NewUUIDV4(),
			Delivery:  suite.sampleID,
			Channel:   testutil.NewUUIDV4(),
			CreatedAt: time.Date(2021, 9, 1, 11, 26, 10, 0, time.UTC),
			IsActive:  false,
			Status:    store.IntelDeliveryStatusTimeout,
			StatusTS:  time.Date(2021, 9, 1, 11, 26, 30, 0, time.UTC),
			Note:      nulls.NewString("last"),
		},
		{
			ID:        testutil.NewUUIDV4(),
			Delivery:  suite.sampleID,
			Channel:   testutil.NewUUIDV4(),
			CreatedAt: time.Date(2022, 9, 1, 11, 26, 10, 0, time.UTC),
			IsActive:  true,
			Status:    store.IntelDeliveryStatusAwaitingAck,
			StatusTS:  time.Date(2022, 9, 1, 11, 26, 30, 0, time.UTC),
			Note:      nulls.NewString("joe"),
		},
	}
	suite.sampleChannel = store.Channel{
		ID:            testutil.NewUUIDV4(),
		Entry:         suite.sampleDelivery.Assignment,
		Label:         "urgent",
		Type:          "gallon",
		Priority:      12,
		MinImportance: 564,
		Details:       nil,
		Timeout:       59 * time.Second,
	}
}

func (suite *controllerLookAfterDeliverySuite) TestRetrieveIntelFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(store.IntelDelivery{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestRetrieveTimedOutDeliveryAttemptsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestRetrieveChannelMetadataForTimedOutAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDeliveryAttempts, nil)
	suite.ctrl.Store.On("ChannelMetadataByID", timeout, suite.tx, mock.Anything).
		Return(store.Channel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestUpdateTimedOutIntelDeliveryAttemptStatusByIDFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDeliveryAttempts, nil)
	suite.ctrl.Store.On("ChannelMetadataByID", timeout, suite.tx, mock.Anything).
		Return(suite.sampleChannel, nil)
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, mock.Anything,
		false, store.IntelDeliveryStatusTimeout, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestRetrieveIntelDeliveryAttemptForTimedOutAttempt() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDeliveryAttempts, nil)
	suite.ctrl.Store.On("ChannelMetadataByID", timeout, suite.tx, mock.Anything).
		Return(suite.sampleChannel, nil)
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, mock.Anything,
		false, store.IntelDeliveryStatusTimeout, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, mock.Anything).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestNotifyAboutTimedOutDeliveryAttempt() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDeliveryAttempts, nil)
	suite.ctrl.Store.On("ChannelMetadataByID", timeout, suite.tx, mock.Anything).
		Return(suite.sampleChannel, nil)
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, mock.Anything,
		false, store.IntelDeliveryStatusTimeout, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, mock.Anything).
		Return(suite.sampleDeliveryAttempts[0], nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, suite.sampleDeliveryAttempts[0]).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestRetrieveActiveDeliveryAttemptsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestActiveDeliveryAttempts() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDeliveryAttempts, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestNextChannelForDeliveryAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("NextChannelForDeliveryAttempt", timeout, suite.tx, suite.sampleID).
		Return(store.Channel{}, false, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestUpdateDeliveryStatusBecauseOfNoMoreChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("NextChannelForDeliveryAttempt", timeout, suite.tx, suite.sampleID).
		Return(store.Channel{}, false, nil)
	suite.ctrl.Store.On("UpdateIntelDeliveryStatusByDelivery", timeout, suite.tx, suite.sampleID, false, false, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestNotifyDeliveryStatusUpdatedBecauseOfNoMoreChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("NextChannelForDeliveryAttempt", timeout, suite.tx, suite.sampleID).
		Return(store.Channel{}, false, nil)
	suite.ctrl.Store.On("UpdateIntelDeliveryStatusByDelivery", timeout, suite.tx, suite.sampleID, false, false, mock.Anything).
		Return(nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryStatusUpdated", timeout, suite.tx, suite.sampleID, false, false, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestOKNoMoreChannels() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("NextChannelForDeliveryAttempt", timeout, suite.tx, suite.sampleID).
		Return(store.Channel{}, false, nil)
	suite.ctrl.Store.On("UpdateIntelDeliveryStatusByDelivery", timeout, suite.tx, suite.sampleID, false, false, mock.Anything).
		Return(nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryStatusUpdated", timeout, suite.tx, suite.sampleID, false, false, mock.Anything).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestCreateNewAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("NextChannelForDeliveryAttempt", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleChannel, true, nil)
	suite.ctrl.Store.On("CreateIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestNotifyAttemptCreatedFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("NextChannelForDeliveryAttempt", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleChannel, true, nil)
	suite.ctrl.Store.On("CreateIntelDeliveryAttempt", timeout, suite.tx, mock.Anything).
		Return(suite.sampleDeliveryAttempts[0], nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptCreated", timeout, suite.tx, suite.sampleDeliveryAttempts[0]).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestOKWithNewAttempt() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("NextChannelForDeliveryAttempt", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleChannel, true, nil)
	suite.ctrl.Store.On("CreateIntelDeliveryAttempt", timeout, suite.tx, mock.MatchedBy(func(v any) bool {
		expect := store.IntelDeliveryAttempt{
			Delivery:  suite.sampleID,
			Channel:   suite.sampleChannel.ID,
			CreatedAt: time.Time{}, // Skip.
			IsActive:  true,
			Status:    store.IntelDeliveryStatusOpen,
			StatusTS:  time.Time{},
			Note:      nulls.String{},
		}
		vv, ok := v.(store.IntelDeliveryAttempt)
		if !ok {
			return false
		}
		vv.CreatedAt = time.Time{}
		vv.StatusTS = time.Time{}
		return vv == expect
	})).
		Return(suite.sampleDeliveryAttempts[0], nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptCreated", timeout, suite.tx, suite.sampleDeliveryAttempts[0]).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_lookAfterDelivery(t *testing.T) {
	suite.Run(t, new(controllerLookAfterDeliverySuite))
}
