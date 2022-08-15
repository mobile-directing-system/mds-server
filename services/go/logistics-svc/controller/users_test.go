package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateUserSuite tests Controller.CreateUser.
type ControllerCreateUserSuite struct {
	suite.Suite
	ctrl       *ControllerMock
	sampleUser store.User
}

func (suite *ControllerCreateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUser = store.User{
		ID:        testutil.NewUUIDV4(),
		Username:  "arrow",
		FirstName: "tire",
		LastName:  "provide",
	}
}

func (suite *ControllerCreateUserSuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateUser", timeout, tx, suite.sampleUser).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, tx, suite.sampleUser)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateUserSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateUser", timeout, tx, suite.sampleUser).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, tx, suite.sampleUser)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_CreateUser(t *testing.T) {
	suite.Run(t, new(ControllerCreateUserSuite))
}

// ControllerUpdateUserSuite tests Controller.UpdateUser.
type ControllerUpdateUserSuite struct {
	suite.Suite
	ctrl       *ControllerMock
	sampleUser store.User
}

func (suite *ControllerUpdateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUser = store.User{
		ID:        testutil.NewUUIDV4(),
		Username:  "arrow",
		FirstName: "tire",
		LastName:  "provide",
	}
}

func (suite *ControllerUpdateUserSuite) TestUpdateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateUser", timeout, tx, suite.sampleUser).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, tx, suite.sampleUser)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateUserSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateUser", timeout, tx, suite.sampleUser).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, tx, suite.sampleUser)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateUser(t *testing.T) {
	suite.Run(t, new(ControllerUpdateUserSuite))
}

// ControllerDeleteUserByIDSuite tests Controller.DeleteUserByID.
type ControllerDeleteUserByIDSuite struct {
	suite.Suite
	ctrl                  *ControllerMock
	sampleUserID          uuid.UUID
	sampleAffectedEntries []uuid.UUID
}

func (suite *ControllerDeleteUserByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUserID = testutil.NewUUIDV4()
	suite.sampleAffectedEntries = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
}

func (suite *ControllerDeleteUserByIDSuite) TestDeleteForwardToUserChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToUserChannelsByUser", timeout, tx, suite.sampleUserID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestDeleteUserFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToUserChannelsByUser", timeout, tx, suite.sampleUserID).
		Return(suite.sampleAffectedEntries, nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, tx, suite.sampleUserID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestRetrieveUpdatedChannelsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToUserChannelsByUser", timeout, tx, suite.sampleUserID).
		Return(suite.sampleAffectedEntries, nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, tx, suite.sampleUserID).
		Return(nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, tx, mock.Anything).
		Return(store.Channel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToUserChannelsByUser", timeout, tx, suite.sampleUserID).
		Return(suite.sampleAffectedEntries, nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, tx, suite.sampleUserID).
		Return(nil)
	suite.ctrl.Store.On("ChannelsByAddressBookEntry", timeout, tx, mock.Anything).
		Return(store.Channel{}, nil)
	suite.ctrl.Notifier.On("NotifyAddressBookEntryChannelsUpdated", timeout, tx, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestOKWithoutAffectedEntries() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToUserChannelsByUser", timeout, tx, suite.sampleUserID).
		Return(nil, nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, tx, suite.sampleUserID).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestOKWithAffectedEntries() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteForwardToUserChannelsByUser", timeout, tx, suite.sampleUserID).
		Return(suite.sampleAffectedEntries, nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, tx, suite.sampleUserID).
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
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_DeleteUserByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteUserByIDSuite))
}
