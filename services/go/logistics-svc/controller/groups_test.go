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
)

// ControllerCreateGroupSuite tests Controller.CreateGroup.
type ControllerCreateGroupSuite struct {
	suite.Suite
	ctrl                     *ControllerMock
	sampleGroup              store.Group
	sampleOperationMembersOK []uuid.UUID
}

func (suite *ControllerCreateGroupSuite) SetupTest() {
	suite.ctrl = NewMockController()
	members := make([]uuid.UUID, 16)
	suite.sampleOperationMembersOK = make([]uuid.UUID, 0, len(members))
	for i := range members {
		members[i] = testutil.NewUUIDV4()
		suite.sampleOperationMembersOK = append(suite.sampleOperationMembersOK, members[i])
		// Random user.
		suite.sampleOperationMembersOK = append(suite.sampleOperationMembersOK, testutil.NewUUIDV4())
	}
	suite.sampleGroup = store.Group{
		Title:       "command",
		Description: "ready",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members:     members,
	}
}

func (suite *ControllerCreateGroupSuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateGroup", timeout, tx, suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateGroup(timeout, tx, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateGroupSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateGroup", timeout, tx, suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateGroup(timeout, tx, suite.sampleGroup)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_CreateGroup(t *testing.T) {
	suite.Run(t, new(ControllerCreateGroupSuite))
}

// ControllerUpdateGroupSuite tests Controller.UpdateGroup.
type ControllerUpdateGroupSuite struct {
	suite.Suite
	ctrl                     *ControllerMock
	sampleGroup              store.Group
	sampleOperationMembersOK []uuid.UUID
}

func (suite *ControllerUpdateGroupSuite) SetupTest() {
	suite.ctrl = NewMockController()
	members := make([]uuid.UUID, 16)
	suite.sampleOperationMembersOK = make([]uuid.UUID, 0, len(members))
	for i := range members {
		members[i] = testutil.NewUUIDV4()
		suite.sampleOperationMembersOK = append(suite.sampleOperationMembersOK, members[i])
		// Random user.
		suite.sampleOperationMembersOK = append(suite.sampleOperationMembersOK, testutil.NewUUIDV4())
	}
	suite.sampleGroup = store.Group{
		ID:          testutil.NewUUIDV4(),
		Title:       "command",
		Description: "ready",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members:     members,
	}
}

func (suite *ControllerUpdateGroupSuite) TestUpdateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateGroup", timeout, tx, suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, tx, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateGroupSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateGroup", timeout, tx, suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, tx, suite.sampleGroup)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateGroup(t *testing.T) {
	suite.Run(t, new(ControllerUpdateGroupSuite))
}

// ControllerDeleteGroupByIDSuite tests Controller.DeleteGroupByID.
type ControllerDeleteGroupByIDSuite struct {
	suite.Suite
	ctrl                  *ControllerMock
	sampleGroupID         uuid.UUID
	sampleAffectedEntries []uuid.UUID
}

func (suite *ControllerDeleteGroupByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleGroupID = testutil.NewUUIDV4()
	suite.sampleAffectedEntries = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
}

func (suite *ControllerDeleteGroupByIDSuite) TestDeleteForwardToGroupChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToGroupChannelsByGroup", timeout, tx, suite.sampleGroupID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, tx, suite.sampleGroupID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestDeleteInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToGroupChannelsByGroup", timeout, tx, suite.sampleGroupID).
		Return(suite.sampleAffectedEntries, nil)
	suite.ctrl.Store.On("DeleteGroupByID", timeout, tx, suite.sampleGroupID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, tx, suite.sampleGroupID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestRetrieveUpdatedChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToGroupChannelsByGroup", timeout, tx, suite.sampleGroupID).
		Return(suite.sampleAffectedEntries, nil)
	suite.ctrl.Store.On("DeleteGroupByID", timeout, tx, suite.sampleGroupID).
		Return(nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, tx, mock.Anything).
		Return(store.Channel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, tx, suite.sampleGroupID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToGroupChannelsByGroup", timeout, tx, suite.sampleGroupID).
		Return(suite.sampleAffectedEntries, nil)
	suite.ctrl.Store.On("DeleteGroupByID", timeout, tx, suite.sampleGroupID).
		Return(nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, tx, mock.Anything).
		Return(store.Channel{}, nil)
	suite.ctrl.Notifier.On("NotifyAddressBookEntryChannelsUpdated", timeout, tx, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, tx, suite.sampleGroupID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestOKWithoutAffectedEntries() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToGroupChannelsByGroup", timeout, tx, suite.sampleGroupID).
		Return(nil, nil)
	suite.ctrl.Store.On("DeleteGroupByID", timeout, tx, suite.sampleGroupID).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, tx, suite.sampleGroupID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestOKWithAffectedEntries() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToGroupChannelsByGroup", timeout, tx, suite.sampleGroupID).
		Return(suite.sampleAffectedEntries, nil)
	suite.ctrl.Store.On("DeleteGroupByID", timeout, tx, suite.sampleGroupID).
		Return(nil)
	for _, entryID := range suite.sampleAffectedEntries {
		channels := make([]store.Channel, 8)
		suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, tx, entryID).
			Return(channels, nil).Once()
		suite.ctrl.Notifier.On("NotifyAddressBookEntryChannelsUpdated", timeout, tx, entryID, channels).
			Return(nil).Once()
	}
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, tx, suite.sampleGroupID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_DeleteGroupByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteGroupByIDSuite))
}
