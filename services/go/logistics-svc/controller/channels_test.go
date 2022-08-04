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
	"math/rand"
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
	ctrl             *ControllerMock
	sampleEntryID    uuid.UUID
	entry            store.AddressBookEntryDetailed
	channelsToDelete []store.Channel
	channelsToUpdate []store.Channel
	channelsToCreate []store.Channel
	// oldChannels holds channels from channelsToDelete and channelsToUpdate.
	oldChannels []store.Channel
	// newChannels holds channels from channelsToUpdate and channelsToCreate.
	newChannels []store.Channel
	// finalChannels is the final list of channels after update has completed.
	finalChannels []store.Channel
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
	suite.channelsToDelete = make([]store.Channel, 16)
	for i := range suite.channelsToDelete {
		suite.channelsToDelete[i] = suite.genChannel(testutil.NewUUIDV4())
	}
	suite.channelsToUpdate = make([]store.Channel, 16)
	for i := range suite.channelsToUpdate {
		suite.channelsToUpdate[i] = suite.genChannel(testutil.NewUUIDV4())
	}
	suite.channelsToCreate = make([]store.Channel, 16)
	for i := range suite.channelsToCreate {
		suite.channelsToCreate[i] = suite.genChannel(uuid.UUID{})
	}
	suite.oldChannels = make([]store.Channel, 0, len(suite.channelsToDelete)+len(suite.channelsToUpdate))
	suite.oldChannels = append(suite.oldChannels, suite.channelsToDelete...)
	suite.oldChannels = append(suite.oldChannels, suite.channelsToUpdate...)
	rand.Shuffle(len(suite.oldChannels), func(i, j int) {
		suite.oldChannels[i], suite.oldChannels[j] = suite.oldChannels[j], suite.oldChannels[i]
	})
	suite.newChannels = make([]store.Channel, 0, len(suite.channelsToUpdate)+len(suite.channelsToCreate))
	suite.newChannels = append(suite.newChannels, suite.channelsToUpdate...)
	suite.newChannels = append(suite.newChannels, suite.channelsToCreate...)
	rand.Shuffle(len(suite.newChannels), func(i, j int) {
		suite.newChannels[i], suite.newChannels[j] = suite.newChannels[j], suite.newChannels[i]
	})
	suite.finalChannels = make([]store.Channel, 0, len(suite.newChannels))
	suite.finalChannels = append(suite.finalChannels, suite.newChannels...)
	for i := range suite.finalChannels {
		suite.finalChannels[i].ID = testutil.NewUUIDV4()
	}
	rand.Shuffle(len(suite.finalChannels), func(i, j int) {
		suite.finalChannels[i], suite.finalChannels[j] = suite.finalChannels[j], suite.finalChannels[i]
	})
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
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestDuplicateChannelIDs() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.newChannels = append(suite.newChannels, suite.newChannels...)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestUnknownChannel() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.newChannels[0].ID = testutil.NewUUIDV4()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestDeleteChannelFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	suite.ctrl.Store.On("DeleteChannelWithDetailsByID", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestCreateChannelFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	suite.ctrl.Store.On("DeleteChannelWithDetailsByID", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("CreateChannelWithDetails", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestUpdateChannelFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	suite.ctrl.Store.On("DeleteChannelWithDetailsByID", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("CreateChannelWithDetails", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("UpdateChannelWithDetails", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestRetrieveFinalUpdatedChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil).Once()
	suite.ctrl.Store.On("DeleteChannelWithDetailsByID", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("CreateChannelWithDetails", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("UpdateChannelWithDetails", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil)
	suite.ctrl.Store.On("DeleteChannelWithDetailsByID", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("CreateChannelWithDetails", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("UpdateChannelWithDetails", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], mock.Anything).
		Return(suite.newChannels, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyAddressBookEntryChannelsUpdated", mock.Anything, mock.Anything).
		Return(errors.New("sad lifei"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateChannelsByAddressBookEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID, uuid.NullUUID{}).
		Return(suite.entry, nil).Once()
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.oldChannels, nil).Once()
	// Channels to delete.
	for i := range suite.channelsToDelete {
		channel := suite.channelsToDelete[i]
		suite.ctrl.Store.On("DeleteChannelWithDetailsByID", timeout, suite.ctrl.DB.Tx[0], channel.ID, channel.Type).
			Return(nil).Once()
	}
	// Channels to create.
	for i := range suite.channelsToCreate {
		channel := suite.channelsToCreate[i]
		suite.ctrl.Store.On("CreateChannelWithDetails", timeout, suite.ctrl.DB.Tx[0], channel).
			Return(nil).Once()
	}
	// Channels to update.
	for i := range suite.channelsToUpdate {
		channel := suite.channelsToUpdate[i]
		suite.ctrl.Store.On("UpdateChannelWithDetails", timeout, suite.ctrl.DB.Tx[0], channel).
			Return(nil).Once()
	}
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntryID).
		Return(suite.finalChannels, nil).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyAddressBookEntryChannelsUpdated", suite.sampleEntryID, suite.finalChannels).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateChannelsByAddressBookEntry(timeout, suite.sampleEntryID, suite.newChannels,
			nulls.NewUUID(suite.entry.User.UUID))
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit")
	}()

	wait()
}

func TestController_UpdateChannelsByAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerUpdateChannelsByAddressBookEntrySuite))
}
