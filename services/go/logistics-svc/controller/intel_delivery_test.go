package controller

import (
	"context"
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
	sampleAssignedEntry    store.AddressBookEntryDetailed
	sampleDelivery         store.IntelDelivery
	sampleDeliveryAttempts []store.IntelDeliveryAttempt
	sampleChannel          store.Channel
}

func (suite *controllerLookAfterDeliverySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleID = testutil.NewUUIDV4()
	userID := testutil.NewUUIDV4()
	suite.sampleAssignedEntry = store.AddressBookEntryDetailed{
		AddressBookEntry: store.AddressBookEntry{
			ID:          testutil.NewUUIDV4(),
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
		ID:         testutil.NewUUIDV4(),
		CreatedAt:  time.Date(2022, 9, 5, 22, 56, 49, 0, time.UTC),
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       "except",
		Content:    []byte(`{"hello":"world"}`),
		SearchText: nulls.NewString("Hello World!"),
		Importance: 744,
		IsValid:    true,
	}
	suite.sampleDelivery = store.IntelDelivery{
		ID:       suite.sampleID,
		Intel:    suite.sampleIntel.ID,
		To:       suite.sampleAssignedEntry.ID,
		IsActive: true,
		Success:  false,
		Note:     nulls.NewString("club"),
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
		Entry:         suite.sampleDelivery.To,
		Label:         "urgent",
		Type:          "gallon",
		Priority:      12,
		MinImportance: 564,
		Details:       nil,
		Timeout:       59 * time.Second,
	}
	suite.ctrl.Store.On("IsAutoDeliveryEnabledForAddressBookEntry", mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil).Maybe()
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

func (suite *controllerLookAfterDeliverySuite) TestCheckAutoDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	testutil.UnsetCallByMethod(&suite.ctrl.Store.Mock, "IsAutoDeliveryEnabledForAddressBookEntry")
	suite.ctrl.Store.On("IsAutoDeliveryEnabledForAddressBookEntry", mock.Anything, mock.Anything, mock.Anything).
		Return(false, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.lookAfterDelivery(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLookAfterDeliverySuite) TestAutoDeliveryDisabled() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleDelivery, nil)
	suite.ctrl.Store.On("TimedOutIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleID).
		Return(nil, nil)
	testutil.UnsetCallByMethod(&suite.ctrl.Store.Mock, "IsAutoDeliveryEnabledForAddressBookEntry")
	suite.ctrl.Store.On("IsAutoDeliveryEnabledForAddressBookEntry", mock.Anything, suite.tx, suite.sampleDelivery.To).
		Return(false, nil).Once()
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
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleDelivery.Intel).
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
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleDelivery.Intel).
		Return(suite.sampleIntel, nil).Once()
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, suite.sampleDelivery.To, uuid.NullUUID{}).
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
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleDelivery.Intel).
		Return(suite.sampleIntel, nil).Once()
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, suite.sampleDelivery.To, uuid.NullUUID{}).
		Return(suite.sampleAssignedEntry, nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptCreated", timeout, suite.tx, suite.sampleDeliveryAttempts[0],
		suite.sampleDelivery, suite.sampleAssignedEntry, suite.sampleIntel).
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
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleDelivery.Intel).
		Return(suite.sampleIntel, nil).Once()
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, suite.sampleDelivery.To, uuid.NullUUID{}).
		Return(suite.sampleAssignedEntry, nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptCreated", timeout, suite.tx, suite.sampleDeliveryAttempts[0],
		suite.sampleDelivery, suite.sampleAssignedEntry, suite.sampleIntel).
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

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestRetrieveUpdatedFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Twice()
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID, true,
		suite.sampleNewStatus, suite.sampleNewNote).
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

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Twice()
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID, true,
		suite.sampleNewStatus, suite.sampleNewNote).
		Return(nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.IntelDeliveryAttempt{}, nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateIntelDeliveryAttemptStatusForActive(timeout, suite.tx, suite.sampleAttemptID,
			suite.sampleNewStatus, suite.sampleNewNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateIntelDeliveryAttemptStatusForActiveSuite) TestOK() {
	updated := suite.sampleAttempt
	updated.Status = suite.sampleNewStatus
	updated.Note = suite.sampleNewNote
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Twice()
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID, true,
		suite.sampleNewStatus, suite.sampleNewNote).
		Return(nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(updated, nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, updated).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

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

// ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite tests
// Controller.MarkIntelDeliveryAndAttemptAsDelivered.
type ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite struct {
	suite.Suite
	ctrl                 *ControllerMock
	tx                   *testutil.DBTx
	sampleDeliveryID     uuid.UUID
	sampleAttemptID      uuid.UUID
	sampleDelivery       store.IntelDelivery
	sampleActiveAttempts []store.IntelDeliveryAttempt
}

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) SetupTest() {
	suite.tx = &testutil.DBTx{}
	suite.ctrl = NewMockController()
	suite.sampleDeliveryID = testutil.NewUUIDV4()
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.sampleDelivery = store.IntelDelivery{
		ID:       suite.sampleDeliveryID,
		Intel:    testutil.NewUUIDV4(),
		To:       testutil.NewUUIDV4(),
		IsActive: true,
		Success:  false,
		Note:     nulls.NewString("gap"),
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
}

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) TestRetrieveDeliveryFail() {
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

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) TestNotAssigned() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
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

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) TestRetrieveActiveDeliveryAttemptsFail() {
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

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) TestUpdateIntelDeliveryAttemptStatusFail() {
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

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) TestRetrieveUpdatedIntelDeliveryAttemptFail() {
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

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) TestNotifyAboutUpdatedDeliveryAttemptFail() {
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

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) TestUpdateDeliveryStatusFail() {
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

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) TestNotifyAboutUpdatedDeliveryStatusFail() {
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

func (suite *ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite) TestOK() {
	by := suite.sampleDelivery.To
	attemptID := suite.sampleActiveAttempts[1].ID
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
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
	suite.Run(t, new(ControllerMarkIntelDeliveryAndAttemptAsDeliveredSuite))
}

// ControllerCancelIntelDeliveryByIDSuite tests
// Controller.CancelIntelDeliveryByID.
type ControllerCancelIntelDeliveryByIDSuite struct {
	suite.Suite
	ctrl                 *ControllerMock
	tx                   *testutil.DBTx
	sampleDeliveryID     uuid.UUID
	sampleDelivery       store.IntelDelivery
	sampleActiveAttempts []store.IntelDeliveryAttempt
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) SetupTest() {
	suite.tx = &testutil.DBTx{}
	suite.ctrl = NewMockController()
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleDeliveryID = testutil.NewUUIDV4()
	suite.sampleDelivery = store.IntelDelivery{
		ID:       suite.sampleDeliveryID,
		Intel:    testutil.NewUUIDV4(),
		To:       testutil.NewUUIDV4(),
		IsActive: true,
		Success:  false,
		Note:     nulls.NewString("gap"),
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
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestRetrieveDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(store.IntelDelivery{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CancelIntelDeliveryByID(timeout, suite.sampleDeliveryID, false, nulls.String{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestDeliveryInactive() {
	suite.sampleDelivery.IsActive = false
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CancelIntelDeliveryByID(timeout, suite.sampleDeliveryID, false, nulls.String{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestRetrieveActiveDeliveryAttemptsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(nil, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CancelIntelDeliveryByID(timeout, suite.sampleDeliveryID, false, nulls.String{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestUpdateIntelDeliveryAttemptStatusFail() {
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
		err := suite.ctrl.Ctrl.CancelIntelDeliveryByID(timeout, suite.sampleDeliveryID, false, nulls.String{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestRetrieveUpdatedIntelDeliveryAttemptFail() {
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
		err := suite.ctrl.Ctrl.CancelIntelDeliveryByID(timeout, suite.sampleDeliveryID, false, nulls.String{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestNotifyAboutUpdatedDeliveryAttemptFail() {
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
		err := suite.ctrl.Ctrl.CancelIntelDeliveryByID(timeout, suite.sampleDeliveryID, false, nulls.String{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestUpdateDeliveryStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(nil, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryStatusByDelivery", timeout, suite.tx, suite.sampleDeliveryID, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CancelIntelDeliveryByID(timeout, suite.sampleDeliveryID, false, nulls.String{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestNotifyAboutUpdatedDeliveryStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(nil, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryStatusByDelivery", timeout, suite.tx, suite.sampleDeliveryID, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryStatusUpdated", timeout, suite.tx, suite.sampleDeliveryID, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CancelIntelDeliveryByID(timeout, suite.sampleDeliveryID, false, nulls.String{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) testOK(success bool, note nulls.String) {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", timeout, suite.tx, suite.sampleDeliveryID).
		Return(suite.sampleActiveAttempts, nil).Once()
	// Expect attempt updates for active ones.
	for _, sampleActiveAttempt := range suite.sampleActiveAttempts {
		newAttempt := sampleActiveAttempt
		newAttempt.IsActive = false
		suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, sampleActiveAttempt.ID,
			false, store.IntelDeliveryStatusCanceled, mock.Anything).
			Return(nil).Once()
		suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, sampleActiveAttempt.ID).
			Return(newAttempt, nil).Once()
		suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, newAttempt).
			Return(nil).Once()
	}
	// Expect intel delivery updates.
	matchNote := mock.MatchedBy(func(n nulls.String) bool {
		suite.True(n.Valid, "should set note")
		suite.NotEmpty(n.Valid, "should set note")
		if note.Valid {
			suite.Equal(note.String, n.String, "should set passed note")
		}
		return true
	})
	suite.ctrl.Store.On("UpdateIntelDeliveryStatusByDelivery", timeout, suite.tx, suite.sampleDeliveryID, false, success, matchNote).
		Return(nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryStatusUpdated", timeout, suite.tx, suite.sampleDeliveryID, false, success, matchNote).
		Return(nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CancelIntelDeliveryByID(timeout, suite.sampleDeliveryID, success, note)
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestOKWithSuccessWithoutNote() {
	suite.testOK(true, nulls.String{})
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestOKWithoutSuccessWithoutNote() {
	suite.testOK(false, nulls.String{})
}

func (suite *ControllerCancelIntelDeliveryByIDSuite) TestOKWithSuccessWithNote() {
	suite.testOK(true, nulls.NewString("wait"))
}

func TestController_CancelIntelDeliveryByID(t *testing.T) {
	suite.Run(t, new(ControllerCancelIntelDeliveryByIDSuite))
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
		ID:       suite.sampleAttempt.Delivery,
		Intel:    testutil.NewUUIDV4(),
		To:       testutil.NewUUIDV4(),
		IsActive: false,
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
		ID:       suite.sampleDeliveryID,
		Intel:    testutil.NewUUIDV4(),
		To:       testutil.NewUUIDV4(),
		IsActive: false,
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

// ControllerMarkIntelDeliveryAttemptAsFailedSuite tests
// Controller.MarkIntelDeliveryAttemptAsFailed.
type ControllerMarkIntelDeliveryAttemptAsFailedSuite struct {
	suite.Suite
	ctrl            *ControllerMock
	tx              *testutil.DBTx
	sampleAttemptID uuid.UUID
	sampleNote      nulls.String
	sampleDelivery  store.IntelDelivery
	sampleAttempt   store.IntelDeliveryAttempt
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.sampleNote = nulls.NewString("list")
	suite.sampleDelivery = store.IntelDelivery{
		ID:       testutil.NewUUIDV4(),
		Intel:    testutil.NewUUIDV4(),
		To:       testutil.NewUUIDV4(),
		IsActive: false,
	}
	suite.sampleAttempt = store.IntelDeliveryAttempt{
		ID:        suite.sampleAttemptID,
		Delivery:  suite.sampleDelivery.ID,
		Channel:   testutil.NewUUIDV4(),
		CreatedAt: time.Date(2022, 10, 13, 13, 53, 7, 0, time.UTC),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusAwaitingAck,
		StatusTS:  time.Date(2022, 10, 13, 13, 53, 37, 0, time.UTC),
		Note:      nulls.NewString("paper"),
	}
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) TestRetrieveAttemptForDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsFailed(timeout, suite.tx, suite.sampleAttemptID, suite.sampleNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) TestRetrieveAndLockDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(store.IntelDelivery{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsFailed(timeout, suite.tx, suite.sampleAttemptID, suite.sampleNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) TestRetrieveLockedAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsFailed(timeout, suite.tx, suite.sampleAttemptID, suite.sampleNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) TestAttemptInactive() {
	suite.sampleAttempt.IsActive = false
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsFailed(timeout, suite.tx, suite.sampleAttemptID, suite.sampleNote)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) TestUpdateAttemptStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID,
		mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsFailed(timeout, suite.tx, suite.sampleAttemptID, suite.sampleNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) TestRetrieveUpdatedAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID,
		mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsFailed(timeout, suite.tx, suite.sampleAttemptID, suite.sampleNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID,
		mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsFailed(timeout, suite.tx, suite.sampleAttemptID, suite.sampleNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) TestLookAfterDeliveryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID,
		mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleDelivery.ID).
		Return(store.IntelDelivery{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsFailed(timeout, suite.tx, suite.sampleAttemptID, suite.sampleNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerMarkIntelDeliveryAttemptAsFailedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryByIDAndLockOrWait", timeout, suite.tx, suite.sampleAttempt.Delivery).
		Return(suite.sampleDelivery, nil).Once()
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(suite.sampleAttempt, nil).Once()
	suite.ctrl.Store.On("UpdateIntelDeliveryAttemptStatusByID", timeout, suite.tx, suite.sampleAttemptID,
		false, store.IntelDeliveryStatusFailed, suite.sampleNote).
		Return(nil)
	updatedAttempt := suite.sampleAttempt
	updatedAttempt.IsActive = false
	updatedAttempt.Status = store.IntelDeliveryStatusFailed
	suite.ctrl.Store.On("IntelDeliveryAttemptByID", timeout, suite.tx, suite.sampleAttemptID).
		Return(updatedAttempt, nil).Once()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.tx, updatedAttempt).
		Return(nil)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.tx, suite.sampleDelivery.ID).
		Return(store.IntelDelivery{IsActive: false}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.MarkIntelDeliveryAttemptAsFailed(timeout, suite.tx, suite.sampleAttemptID, suite.sampleNote)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_MarkIntelDeliveryAttemptAsFailed(t *testing.T) {
	suite.Run(t, new(ControllerMarkIntelDeliveryAttemptAsFailedSuite))
}

// ControllerCreateIntelDeliveryAttemptSuite tests
// Controller.CreateIntelDeliveryAttempt.
type ControllerCreateIntelDeliveryAttemptSuite struct {
	suite.Suite
	ctrl       *ControllerMock
	tx         *testutil.DBTx
	deliveryID uuid.UUID
	channelID  uuid.UUID
	created    store.IntelDeliveryAttempt
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.deliveryID = suite.deliveryID
	suite.channelID = suite.channelID
	suite.created = store.IntelDeliveryAttempt{
		ID:        testutil.NewUUIDV4(),
		Delivery:  suite.deliveryID,
		Channel:   suite.channelID,
		CreatedAt: time.Now(),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusOpen,
		StatusTS:  time.Now(),
		Note:      nulls.String{},
	}
	delivery := store.IntelDelivery{IsActive: true}

	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", mock.Anything, suite.tx, suite.deliveryID).
		Return(nil).Maybe()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", mock.Anything, suite.tx, suite.deliveryID).
		Return([]store.IntelDeliveryAttempt{}, nil).Maybe()
	suite.ctrl.Store.On("IntelDeliveryByID", mock.Anything, suite.tx, suite.deliveryID).
		Return(delivery, nil).Maybe()
	suite.ctrl.Store.On("CreateIntelDeliveryAttempt", mock.Anything, suite.tx, mock.Anything).
		Return(suite.created, nil).Maybe()
	suite.ctrl.Store.On("IntelByID", mock.Anything, suite.tx, mock.Anything).
		Return(store.Intel{}, nil).Maybe()
	suite.ctrl.Store.On("AddressBookEntryByID", mock.Anything, suite.tx, mock.Anything, mock.Anything).
		Return(store.AddressBookEntryDetailed{}, nil).Maybe()
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptCreated",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()
	suite.T().Cleanup(func() {
		suite.ctrl.Store.AssertExpectations(suite.T())
		suite.ctrl.Notifier.AssertExpectations(suite.T())
	})
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestBeginTxFail() {
	suite.ctrl.DB.BeginFail = true

	_, err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(context.Background(), suite.deliveryID, suite.channelID)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestLockDeliveryFail() {
	testutil.UnsetCallByMethod(&suite.ctrl.Store.Mock, "LockIntelDeliveryByIDOrWait")
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life")).Once()

	_, err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(context.Background(), suite.deliveryID, suite.channelID)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestRetrieveActiveAttemptsFail() {
	testutil.UnsetCallByMethod(&suite.ctrl.Store.Mock, "ActiveIntelDeliveryAttemptsByDelivery")
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("sad life")).Once()

	_, err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(context.Background(), suite.deliveryID, suite.channelID)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestCreateFail() {
	testutil.UnsetCallByMethod(&suite.ctrl.Store.Mock, "CreateIntelDeliveryAttempt")
	suite.ctrl.Store.On("CreateIntelDeliveryAttempt", mock.Anything, mock.Anything, mock.Anything).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life")).Once()

	_, err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(context.Background(), suite.deliveryID, suite.channelID)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateIntelDeliveryAttemptSuite) TestOK() {
	testutil.UnsetCallByMethod(&suite.ctrl.Store.Mock, "LockIntelDeliveryByIDOrWait")
	suite.ctrl.Store.On("LockIntelDeliveryByIDOrWait", mock.Anything, suite.tx, suite.deliveryID).
		Return(nil).Once()
	testutil.UnsetCallByMethod(&suite.ctrl.Store.Mock, "ActiveIntelDeliveryAttemptsByDelivery")
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByDelivery", mock.Anything, suite.tx, suite.deliveryID).
		Return([]store.IntelDeliveryAttempt{}, nil).Once()
	testutil.UnsetCallByMethod(&suite.ctrl.Store.Mock, "CreateIntelDeliveryAttempt")
	suite.ctrl.Store.On("CreateIntelDeliveryAttempt", mock.Anything, suite.tx, mock.MatchedBy(func(a store.IntelDeliveryAttempt) bool {
		if !suite.Equal(suite.deliveryID, a.Delivery, "should create attempt with correct delivery") {
			return false
		}
		if !suite.Equal(suite.channelID, a.Channel, "should create attempt with correct channel") {
			return false
		}
		return true
	})).Return(suite.created, nil).Once()

	created, err := suite.ctrl.Ctrl.CreateIntelDeliveryAttempt(context.Background(), suite.deliveryID, suite.channelID)
	suite.Require().NoError(err, "should not fail")
	suite.Equal(suite.created, created, "should return correct value")
}

func TestController_CreateIntelDeliveryAttempt(t *testing.T) {
	suite.Run(t, new(ControllerCreateIntelDeliveryAttemptSuite))
}

// ControllerIntelDeliveryAttemptsByDeliverySuite tests
// Controller.IntelDeliveryAttemptsByDelivery.
type ControllerIntelDeliveryAttemptsByDeliverySuite struct {
	suite.Suite
	ctrl       *ControllerMock
	tx         *testutil.DBTx
	deliveryID uuid.UUID
	attempts   []store.IntelDeliveryAttempt
}

func (suite *ControllerIntelDeliveryAttemptsByDeliverySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.deliveryID = testutil.NewUUIDV4()
	suite.attempts = []store.IntelDeliveryAttempt{
		{
			ID:       testutil.NewUUIDV4(),
			Delivery: suite.deliveryID,
		},
		{
			ID:       testutil.NewUUIDV4(),
			Delivery: suite.deliveryID,
		},
	}
}

func (suite *ControllerIntelDeliveryAttemptsByDeliverySuite) TestBeginTxFail() {
	suite.ctrl.DB.BeginFail = true

	_, err := suite.ctrl.Ctrl.IntelDeliveryAttemptsByDelivery(context.Background(), suite.deliveryID)
	suite.Error(err, "should fail")
}

func (suite *ControllerIntelDeliveryAttemptsByDeliverySuite) TestRetrieveFail() {
	suite.ctrl.Store.On("IntelDeliveryAttemptsByDelivery", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	_, err := suite.ctrl.Ctrl.IntelDeliveryAttemptsByDelivery(context.Background(), suite.deliveryID)
	suite.Error(err, "should fail")
}

func (suite *ControllerIntelDeliveryAttemptsByDeliverySuite) TestOK() {
	suite.ctrl.Store.On("IntelDeliveryAttemptsByDelivery", mock.Anything, suite.tx, suite.deliveryID).
		Return(suite.attempts, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	got, err := suite.ctrl.Ctrl.IntelDeliveryAttemptsByDelivery(context.Background(), suite.deliveryID)
	suite.Require().NoError(err, "should not fail")
	suite.Equal(suite.attempts, got, "should return correct value")
}

func TestController_IntelDeliveryAttemptsByDelivery(t *testing.T) {
	suite.Run(t, new(ControllerIntelDeliveryAttemptsByDeliverySuite))
}
