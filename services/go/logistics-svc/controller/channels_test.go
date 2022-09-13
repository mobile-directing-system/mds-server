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

// ControllerChannelsByAddressBookEntrySuite tests
// Controller.ChannelsByAddressBookEntry.
type ControllerChannelsByAddressBookEntrySuite struct {
	suite.Suite
	ctrl                *ControllerMock
	sampleEntryID       uuid.UUID
	sampleChannels      []store.Channel
	sampleEntryDetailed store.AddressBookEntryDetailed
}

func (suite *ControllerChannelsByAddressBookEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntryID = testutil.NewUUIDV4()
	suite.sampleChannels = []store.Channel{
		{
			ID:            testutil.NewUUIDV4(),
			Entry:         suite.sampleEntryID,
			Label:         "imitate",
			Type:          store.ChannelTypeDirect,
			Priority:      1,
			MinImportance: 2,
			Details: store.DirectChannelDetails{
				Info: "flag",
			},
			Timeout: 10 * time.Second,
		},
		{
			ID:            testutil.NewUUIDV4(),
			Entry:         suite.sampleEntryID,
			Label:         "popular",
			Type:          store.ChannelTypePhoneCall,
			Priority:      3,
			MinImportance: 1,
			Details: store.PhoneCallChannelDetails{
				Phone: "004915233335",
			},
			Timeout: 23 * time.Millisecond,
		},
	}
	suite.sampleEntryDetailed = store.AddressBookEntryDetailed{
		AddressBookEntry: store.AddressBookEntry{
			ID:          suite.sampleEntryID,
			Label:       "correct",
			Description: "around",
			Operation:   uuid.NullUUID{},
			User:        uuid.NullUUID{},
		},
		UserDetails: nulls.JSONNullable[store.User]{},
	}
}

func (suite *ControllerChannelsByAddressBookEntrySuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.ChannelsByAddressBookEntry(timeout, suite.sampleEntryID, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerChannelsByAddressBookEntrySuite) TestRetrieveEntryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(store.AddressBookEntryDetailed{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.ChannelsByAddressBookEntry(timeout, suite.sampleEntryID, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerChannelsByAddressBookEntrySuite) TestLimitViolationForGlobal() {
	limitToUser := testutil.NewUUIDV4()
	suite.sampleEntryDetailed.User = uuid.NullUUID{}
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.ChannelsByAddressBookEntry(timeout, suite.sampleEntryID, nulls.NewUUID(limitToUser))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerChannelsByAddressBookEntrySuite) TestLimitViolationForForeignUsersEntry() {
	limitToUser := testutil.NewUUIDV4()
	suite.sampleEntryDetailed.User = nulls.NewUUID(testutil.NewUUIDV4())
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.ChannelsByAddressBookEntry(timeout, suite.sampleEntryID, nulls.NewUUID(limitToUser))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerChannelsByAddressBookEntrySuite) TestRetrieveChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.ChannelsByAddressBookEntry(timeout, suite.sampleEntryID, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerChannelsByAddressBookEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.sampleChannels, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.ChannelsByAddressBookEntry(timeout, suite.sampleEntryID, uuid.NullUUID{})
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleChannels, got, "should return correct value")
	}()

	wait()
}

func TestController_ChannelsByAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerChannelsByAddressBookEntrySuite))
}

// ControllerUpdateChannelsByAddressBookEntrySuite test
// Controller.UpdateChannelsByAddressBookEntry.
type ControllerUpdateChannelsByAddressBookEntrySuite struct {
	suite.Suite
	ctrl            *ControllerMock
	sampleEntryID   uuid.UUID
	entry           store.AddressBookEntryDetailed
	oldChannels     []store.Channel
	oldChannelIDs   []uuid.UUID
	newChannels     []store.Channel
	updatedChannels []store.Channel
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) genChannel(id uuid.UUID) store.Channel {
	return store.Channel{
		ID:    id,
		Entry: suite.sampleEntryID,
	}
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntryID = testutil.NewUUIDV4()
	suite.entry = store.AddressBookEntryDetailed{
		AddressBookEntry: store.AddressBookEntry{
			ID:          suite.sampleEntryID,
			Label:       "arrest",
			Description: "stain",
			Operation:   uuid.NullUUID{},
			User:        nulls.NewUUID(testutil.NewUUIDV4()),
		},
		UserDetails: nulls.JSONNullable[store.User]{},
	}
	for i := 0; i < 32; i++ {
		suite.oldChannels = append(suite.oldChannels, suite.genChannel(testutil.NewUUIDV4()))
		suite.newChannels = append(suite.newChannels, suite.genChannel(uuid.Nil))
	}
	suite.oldChannelIDs = make([]uuid.UUID, 0, len(suite.oldChannels))
	for _, oldChannel := range suite.oldChannels {
		suite.oldChannelIDs = append(suite.oldChannelIDs, oldChannel.ID)
	}
	suite.updatedChannels = make([]store.Channel, 0, len(suite.newChannels))
	for _, newChannel := range suite.newChannels {
		newChannel.ID = testutil.NewUUIDV4()
		suite.updatedChannels = append(suite.updatedChannels, newChannel)
	}
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestRetrieveAddressBookEntryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(store.AddressBookEntryDetailed{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestLimitToUserGlobalEntry() {
	suite.entry.User = uuid.NullUUID{}
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels,
			nulls.NewUUID(testutil.NewUUIDV4()))
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return correct error code")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestLimitToUserForeignUsersEntry() {
	suite.entry.User = nulls.NewUUID(testutil.NewUUIDV4())
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels,
			nulls.NewUUID(testutil.NewUUIDV4()))
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return correct error code")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestRetrieveOldChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestRetrieveAssociatedDeliveryAttemptsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait", timeout, suite.ctrl.DB.Tx[0], suite.oldChannelIDs).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestNotifyAboutAffectedAttemptFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	activeAttempt := store.IntelDeliveryAttempt{
		ID:        testutil.NewUUIDV4(),
		Delivery:  testutil.NewUUIDV4(),
		Channel:   testutil.NewUUIDV4(),
		CreatedAt: time.Now().UTC(),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusDelivering,
		StatusTS:  time.Now().UTC(),
		Note:      nulls.NewString("material"),
	}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait", timeout, suite.ctrl.DB.Tx[0], suite.oldChannelIDs).
		Return([]store.IntelDeliveryAttempt{activeAttempt}, nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestDeleteDeliveryAttemptsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	activeAttempt := store.IntelDeliveryAttempt{
		ID:        testutil.NewUUIDV4(),
		Delivery:  testutil.NewUUIDV4(),
		Channel:   testutil.NewUUIDV4(),
		CreatedAt: time.Now().UTC(),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusDelivering,
		StatusTS:  time.Now().UTC(),
		Note:      nulls.NewString("material"),
	}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait", timeout, suite.ctrl.DB.Tx[0], suite.oldChannelIDs).
		Return([]store.IntelDeliveryAttempt{activeAttempt}, nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("DeleteIntelDeliveryAttemptsByChannel", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestUpdateChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait", timeout, suite.ctrl.DB.Tx[0], suite.oldChannelIDs).
		Return(nil, nil)
	suite.ctrl.Store.On("DeleteIntelDeliveryAttemptsByChannel", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("UpdateChannelsByEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, suite.newChannels).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestRetrieveFinalChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil).Once()
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait", timeout, suite.ctrl.DB.Tx[0], suite.oldChannelIDs).
		Return(nil, nil)
	suite.ctrl.Store.On("DeleteIntelDeliveryAttemptsByChannel", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("UpdateChannelsByEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, suite.newChannels).
		Return(nil).Once()
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(nil, errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestNotifyAboutUpdatedChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil).Once()
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait", timeout, suite.ctrl.DB.Tx[0], suite.oldChannelIDs).
		Return(nil, nil)
	suite.ctrl.Store.On("DeleteIntelDeliveryAttemptsByChannel", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("UpdateChannelsByEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, suite.newChannels).
		Return(nil).Once()
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.updatedChannels, nil).Once()
	suite.ctrl.Notifier.On("NotifyAddressBookEntryChannelsUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, suite.updatedChannels).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestLookAfterAffectedDeliveriesFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	affectedDelivery := testutil.NewUUIDV4()
	attempt := store.IntelDeliveryAttempt{Delivery: affectedDelivery}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil).Once()
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait", timeout, suite.ctrl.DB.Tx[0], suite.oldChannelIDs).
		Return([]store.IntelDeliveryAttempt{attempt}, nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("DeleteIntelDeliveryAttemptsByChannel", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("UpdateChannelsByEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, suite.newChannels).
		Return(nil).Once()
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.updatedChannels, nil).Once()
	suite.ctrl.Notifier.On("NotifyAddressBookEntryChannelsUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, suite.updatedChannels).
		Return(nil)
	suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.ctrl.DB.Tx[0], affectedDelivery).
		Return(store.IntelDelivery{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	affectedDeliveries := []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
	attempts := make([]store.IntelDeliveryAttempt, 0, len(affectedDeliveries))
	for _, delivery := range affectedDeliveries {
		attempts = append(attempts, store.IntelDeliveryAttempt{
			ID:       testutil.NewUUIDV4(),
			Delivery: delivery,
		})
	}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil).Once()
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil).Once()
	suite.ctrl.Store.On("ActiveIntelDeliveryAttemptsByChannelsAndLockOrWait", timeout, suite.ctrl.DB.Tx[0], suite.oldChannelIDs).
		Return(attempts, nil)
	suite.ctrl.Notifier.On("NotifyIntelDeliveryAttemptStatusUpdated", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil).Times(len(attempts))
	for _, oldChannelID := range suite.oldChannelIDs {
		suite.ctrl.Store.On("DeleteIntelDeliveryAttemptsByChannel", timeout, suite.ctrl.DB.Tx[0], oldChannelID).
			Return(nil).Once()
	}
	suite.ctrl.Store.On("UpdateChannelsByEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, suite.newChannels).
		Return(nil).Once()
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.updatedChannels, nil).Once()
	suite.ctrl.Notifier.On("NotifyAddressBookEntryChannelsUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, suite.updatedChannels).
		Return(nil).Once()
	for _, affectedDelivery := range affectedDeliveries {
		suite.ctrl.Store.On("IntelDeliveryByID", timeout, suite.ctrl.DB.Tx[0], affectedDelivery).
			Return(store.IntelDelivery{IsActive: false}, nil).Once()
	}
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_UpdateChannelsByAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerUpdateChannelsByAddressBookEntrySuite))
}
