package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
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
	sampleIntel            store.Intel
	sampleIntelAssignment  store.IntelAssignment
	sampleAssignedEntry    store.AddressBookEntryDetailed
	sampleDelivery         store.IntelDelivery
	sampleDeliveryAttempts []store.IntelDeliveryAttempt
	sampleChannel          store.Channel
}

func (suite *controllerLookAfterDeliverySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleID = testutil.NewUUIDV4()
	suite.sampleIntelAssignment = store.IntelAssignment{
		ID:    testutil.NewUUIDV4(),
		Intel: testutil.NewUUIDV4(),
		To:    testutil.NewUUIDV4(),
	}
	userID := testutil.NewUUIDV4()
	suite.sampleAssignedEntry = store.AddressBookEntryDetailed{
		AddressBookEntry: store.AddressBookEntry{
			ID:          suite.sampleIntelAssignment.To,
			Label:       "bay",
			Description: "cage",
			Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
			User:        nulls.NewUUID(userID),
		},
		UserDetails: nulls.NewJSONNullable(store.User{
			ID:        userID,
			Username:  "that",
			FirstName: "tribe",
			LastName:  "late",
			IsActive:  true,
		}),
	}
	suite.sampleIntel = store.Intel{
		ID:          suite.sampleIntelAssignment.Intel,
		CreatedAt:   time.Date(2022, 9, 5, 22, 56, 49, 0, time.UTC),
		CreatedBy:   testutil.NewUUIDV4(),
		Operation:   testutil.NewUUIDV4(),
		Type:        "except",
		Content:     []byte(`{"hello":"world"}`),
		SearchText:  nulls.NewString("Hello World!"),
		Importance:  744,
		IsValid:     true,
		Assignments: []store.IntelAssignment{suite.sampleIntelAssignment},
	}
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

func (suite *controllerLookAfterDeliverySuite) TestRetrieveIntelAssignmentForNotifyFail() {
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
	suite.ctrl.Store.On("IntelAssignmentByID", timeout, suite.tx, suite.sampleDelivery.Assignment).
		Return(store.IntelAssignment{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestRetrieveIntelForNotifyFail() {
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
	suite.ctrl.Store.On("IntelAssignmentByID", timeout, suite.tx, suite.sampleDelivery.Assignment).
		Return(suite.sampleIntelAssignment, nil).Once()
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntelAssignment.Intel).
		Return(store.Intel{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestRetrieveAssignedEntryFail() {
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
	suite.ctrl.Store.On("IntelAssignmentByID", timeout, suite.tx, suite.sampleDelivery.Assignment).
		Return(suite.sampleIntelAssignment, nil).Once()
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntelAssignment.Intel).
		Return(suite.sampleIntel, nil).Once()
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, suite.sampleIntelAssignment.To, uuid.NullUUID{}).
		Return(store.AddressBookEntryDetailed{}, errors.New("sad life"))
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
	suite.ctrl.Store.On("IntelAssignmentByID", timeout, suite.tx, suite.sampleDelivery.Assignment).
		Return(suite.sampleIntelAssignment, nil).Once()
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntelAssignment.Intel).
		Return(suite.sampleIntel, nil).Once()
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, suite.sampleIntelAssignment.To, uuid.NullUUID{}).
		Return(suite.sampleAssignedEntry, nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptCreated", timeout, suite.tx, suite.sampleDeliveryAttempts[0],
		suite.sampleDelivery, suite.sampleIntelAssignment, suite.sampleAssignedEntry, suite.sampleIntel).
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
	suite.ctrl.Store.On("IntelAssignmentByID", timeout, suite.tx, suite.sampleDelivery.Assignment).
		Return(suite.sampleIntelAssignment, nil).Once()
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntelAssignment.Intel).
		Return(suite.sampleIntel, nil).Once()
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, suite.sampleIntelAssignment.To, uuid.NullUUID{}).
		Return(suite.sampleAssignedEntry, nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptCreated", timeout, suite.tx, suite.sampleDeliveryAttempts[0],
		suite.sampleDelivery, suite.sampleIntelAssignment, suite.sampleAssignedEntry, suite.sampleIntel).
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

// ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite tests
// Controller.UpdateIntelDeliveryAttemptStatusForActive.
type ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite struct {
	suite.Suite
	ctrl            *ControllerMock
	tx              *testutil.DBTx
	sampleAttemptID uuid.UUID
	sampleAttempt   store.IntelDeliveryAttempt
	sampleNewStatus store.IntelDeliveryStatus
	sampleNewNote   nulls.String
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.sampleAttempt = store.IntelDeliveryAttempt{
		ID:        suite.sampleAttemptID,
		Delivery:  testutil.NewUUIDV4(),
		Channel:   testutil.NewUUIDV4(),
		CreatedAt: time.Date(2022, 9, 8, 12, 52, 50, 0, time.UTC),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusDelivering,
		StatusTS:  time.Date(2022, 9, 8, 12, 53, 5, 0, time.UTC),
		Note:      nulls.NewString("manner"),
	}
	suite.sampleNewStatus = store.IntelDeliveryStatusAwaitingAck
	suite.sampleNewNote = nulls.NewString("ounce")
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestFirstAttemptRetrieveFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatusForActive(timeout, suite.tx, suite.sampleAttemptID,
			suite.sampleNewStatus, suite.sampleNewNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestLockDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatusForActive(timeout, suite.tx, suite.sampleAttemptID,
			suite.sampleNewStatus, suite.sampleNewNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestFinalIntelDeliveryAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatusForActive(timeout, suite.tx, suite.sampleAttemptID,
			suite.sampleNewStatus, suite.sampleNewNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestAttemptInactive() {
	suite.sampleAttempt.IsActive = false
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Twice()
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatusForActive(timeout, suite.tx, suite.sampleAttemptID,
			suite.sampleNewStatus, suite.sampleNewNote)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestAttemptCanceled() {
	suite.sampleAttempt.Status = store.IntelDeliveryStatusCanceled
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Twice()
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatusForActive(timeout, suite.tx, suite.sampleAttemptID,
			suite.sampleNewStatus, suite.sampleNewNote)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Twice()
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID, true,
		suite.sampleNewStatus, suite.sampleNewNote).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatusForActive(timeout, suite.tx, suite.sampleAttemptID,
			suite.sampleNewStatus, suite.sampleNewNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Twice()
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID, true,
		suite.sampleNewStatus, suite.sampleNewNote).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatusForActive(timeout, suite.tx, suite.sampleAttemptID,
			suite.sampleNewStatus, suite.sampleNewNote)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateIntelDeliveryAttemptStatusForActive(t *testing.T) {
	suite.Run(t, new(ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite))
}

// ControllerMarkIntelAsDeliveredSuite tests
// Controller.MarkIntelDeliveryAndAttemptAsDelivered.
type ControllerMarkIntelAsDeliveredSuite struct {
	suite.Suite
	ctrl                 *ControllerMock
	tx                   *testutil.DBTx
	sampleDeliveryID     uuid.UUID
	sampleAttemptID      uuid.UUID
	sampleDelivery       store.IntelDelivery
	sampleActiveAttempts []store.IntelDeliveryAttempt
	sampleAssignment     store.IntelAssignment
}

func (suite *ControllerMarkIntelAsDeliveredSuite) SetupTest() {
	suite.tx = &testutil.DBTx{}
	suite.ctrl = NewMockController()
	suite.sampleDeliveryID = testutil.NewUUIDV4()
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.sampleDelivery = store.IntelDelivery{
		ID:         suite.sampleDeliveryID,
		Assignment: testutil.NewUUIDV4(),
		IsActive:   true,
		Success:    false,
		Note:       nulls.NewString("gap"),
	}
	suite.sampleActiveAttempts = []store.IntelDeliveryAttempt{
		{
			ID:        testutil.NewUUIDV4(),
			Delivery:  suite.sampleDeliveryID,
			Channel:   testutil.NewUUIDV4(),
			CreatedAt: time.Date(2022, 9, 8, 14, 44, 3, 0, time.UTC),
			IsActive:  true,
			Status:    store.IntelDeliveryStatusAwaitingAck,
			StatusTS:  time.Date(2022, 9, 8, 14, 44, 19, 0, time.UTC),
			Note:      nulls.NewString("proposal"),
		},
		{
			ID:        testutil.NewUUIDV4(),
			Delivery:  suite.sampleDeliveryID,
			Channel:   testutil.NewUUIDV4(),
			CreatedAt: time.Date(2022, 9, 8, 14, 44, 52, 0, time.UTC),
			IsActive:  true,
			Status:    store.IntelDeliveryStatusDelivering,
			StatusTS:  time.Date(2022, 9, 8, 14, 44, 19, 0, time.UTC),
			Note:      nulls.NewString("lay"),
		},
	}
	suite.sampleAssignment = store.IntelAssignment{
		ID:    testutil.NewUUIDV4(),
		Intel: testutil.NewUUIDV4(),
		To:    testutil.NewUUIDV4(),
	}
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestRetrieveDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(store.IntelDelivery{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID, uuid.NullUUID{}, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestRetrieveIntelAssignmentForByCheckFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelAssignmentByID", timeout, suite.tx, suite.sampleDelivery.Assignment).
		Return(store.IntelAssignment{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID,
			uuid.NullUUID{}, nulls.NewUUID(testutil.NewUUIDV4()))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestNotAssigned() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelAssignmentByID", timeout, suite.tx, suite.sampleDelivery.Assignment).
		Return(suite.sampleAssignment, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID,
			uuid.NullUUID{}, nulls.NewUUID(testutil.NewUUIDV4()))
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return correct error code")
	}()

	wait()
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestRetrieveActiveDeliveryAttemptsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(nil, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID, uuid.NullUUID{}, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestUpdateIntelDeliveryAttemptStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleActiveAttempts, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID, uuid.NullUUID{}, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestRetrieveUpdatedIntelDeliveryAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleActiveAttempts, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, mock.Anything).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID, uuid.NullUUID{}, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestNotifyAboutUpdatedDeliveryAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleActiveAttempts, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, mock.Anything).
		Return(suite.sampleActiveAttempts[0], nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, suite.sampleActiveAttempts[0]).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID, uuid.NullUUID{}, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestUpdateDeliveryStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(nil, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryStatusByDelivery", timeout, suite.tx, suite.sampleDeliveryID, false, true, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID, uuid.NullUUID{}, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestNotifyAboutUpdatedDeliveryStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(nil, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryStatusByDelivery", timeout, suite.tx, suite.sampleDeliveryID, false, true, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryStatusUpdated", timeout, suite.tx, suite.sampleDeliveryID, false, true, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID, uuid.NullUUID{}, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelAsDeliveredSuite) TestOK() {
	by := suite.sampleAssignment.To
	attemptID := suite.sampleActiveAttempts[1].ID
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelAssignmentByID", timeout, suite.tx, suite.sampleDelivery.Assignment).
		Return(suite.sampleAssignment, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleActiveAttempts, nil).Once()
	// For other attempt:
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleActiveAttempts[0].ID,
		false, store.IntelDeliveryStatusCanceled, mock.Anything).
		Return(nil).Once()
	newAttempt0 := suite.sampleActiveAttempts[0]
	newAttempt0.Note = nulls.NewString("dot")
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleActiveAttempts[0].ID).
		Return(newAttempt0, nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, newAttempt0).
		Return(nil).Once()
	// For selected attempt:
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleActiveAttempts[1].ID,
		false, store.IntelDeliveryStatusDelivered, mock.Anything).
		Return(nil).Once()
	newAttempt1 := suite.sampleActiveAttempts[1]
	newAttempt1.Note = nulls.NewString("dot")
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleActiveAttempts[1].ID).
		Return(newAttempt1, nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, newAttempt1).
		Return(nil).Once()
	// Continue.
	suite.ctrl.Store.On("UpdateIntelDeliveryStatusByDelivery", timeout, suite.tx, suite.sampleDeliveryID, false, true, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryStatusUpdated", timeout, suite.tx, suite.sampleDeliveryID, false, true, mock.Anything).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAndAttemptAsDelivered(timeout, suite.tx, suite.sampleDeliveryID,
			nulls.NewUUID(attemptID), nulls.NewUUID(by))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_MarkIntelDeliveryAndAttemptAsDelivered(t *testing.T) {
	suite.Run(t, new(ControllerMarkIntelAsDeliveredSuite))
}

// ControllerMarkIntelDeliveryAttemptAsDeliveredSuite tests
// Controller.MarkIntelDeliveryAttemptAsDelivered.
type ControllerMarkIntelDeliveryAttemptAsDeliveredSuite struct {
	suite.Suite
	ctrl            *ControllerMock
	tx              *testutil.DBTx
	sampleAttemptID uuid.UUID
	sampleBy        uuid.NullUUID
	sampleAttempt   store.IntelDeliveryAttempt
	sampleDelivery  store.IntelDelivery
}

func (suite *ControllerMarkIntelDeliveryAttemptAsDeliveredSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.sampleBy = nulls.NewUUID(testutil.NewUUIDV4())
	suite.sampleAttempt = store.IntelDeliveryAttempt{
		ID:       suite.sampleAttemptID,
		Delivery: testutil.NewUUIDV4(),
		Channel:  testutil.NewUUIDV4(),
		IsActive: true,
	}
	suite.sampleDelivery = store.IntelDelivery{
		ID:         suite.sampleAttempt.Delivery,
		Assignment: testutil.NewUUIDV4(),
		IsActive:   false,
	}
}

func (suite *ControllerMarkIntelDeliveryAttemptAsDeliveredSuite) TestBeginTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsDelivered(timeout, suite.sampleAttemptID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsDeliveredSuite) TestRetrieveDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsDelivered(timeout, suite.sampleAttemptID, suite.sampleBy)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsDeliveredSuite) TestMarkDeliveredFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDelivery.ID).
		Return(store.IntelDelivery{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsDelivered(timeout, suite.sampleAttemptID, suite.sampleBy)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsDeliveredSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDelivery.ID).
		Return(suite.sampleDelivery, meh.NewErr("done", "", nil)).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsDelivered(timeout, suite.sampleAttemptID, suite.sampleBy)
		suite.Require().Error(err, "should not fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
		suite.EqualValues("done", meh.ErrorCode(err))
	}()

	wait()
}

func TestController_MarkIntelDeliveryAttemptAsDelivered(t *testing.T) {
	suite.Run(t, new(ControllerMarkIntelDeliveryAttemptAsDeliveredSuite))
}

// ControllerMarkIntelDeliveryAsDeliveredSuite tests
// Controller.MarkIntelDeliveryAsDelivered.
type ControllerMarkIntelDeliveryAsDeliveredSuite struct {
	suite.Suite
	ctrl             *ControllerMock
	tx               *testutil.DBTx
	sampleDeliveryID uuid.UUID
	sampleBy         uuid.NullUUID
	sampleDelivery   store.IntelDelivery
}

func (suite *ControllerMarkIntelDeliveryAsDeliveredSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleDeliveryID = testutil.NewUUIDV4()
	suite.sampleBy = nulls.NewUUID(testutil.NewUUIDV4())
	suite.sampleDelivery = store.IntelDelivery{
		ID:         suite.sampleDeliveryID,
		Assignment: testutil.NewUUIDV4(),
		IsActive:   false,
	}
}

func (suite *ControllerMarkIntelDeliveryAsDeliveredSuite) TestBeginTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAsDelivered(timeout, suite.sampleDeliveryID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAsDeliveredSuite) TestMarkDeliveredFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDelivery.ID).
		Return(store.IntelDelivery{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAsDelivered(timeout, suite.sampleDeliveryID, suite.sampleBy)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAsDeliveredSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDelivery.ID).
		Return(store.IntelDelivery{}, meh.NewErr("done", "", nil)).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAsDelivered(timeout, suite.sampleDeliveryID, suite.sampleBy)
		suite.Require().Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
		suite.EqualValues("done", meh.ErrorCode(err))
	}()

	wait()
}

func TestController_MarkIntelDeliveryAsDelivered(t *testing.T) {
	suite.Run(t, new(ControllerMarkIntelDeliveryAsDeliveredSuite))
}
