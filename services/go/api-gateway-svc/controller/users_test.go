package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateUserSuite tests Controller.CreateUser.
type ControllerCreateUserSuite struct {
	suite.Suite
	ctrl       *ControllerMock
	createUser store.UserWithPass
}

func (suite *ControllerCreateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.createUser = store.UserWithPass{
		User: store.User{
			ID:       testutil.NewUUIDV4(),
			Username: "duty",
			IsAdmin:  true,
		},
		Pass: []byte("meow"),
	}
}

func (suite *ControllerCreateUserSuite) TestStoreCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateUser", timeout, tx, suite.createUser).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, tx, suite.createUser)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateUserSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	created := suite.createUser
	created.ID = testutil.NewUUIDV4()
	suite.ctrl.Store.On("CreateUser", timeout, tx, suite.createUser).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, tx, suite.createUser)
		suite.NoError(err, "should fail")
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
	updateUser store.User
}

func (suite *ControllerUpdateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.updateUser = store.User{
		ID:       testutil.NewUUIDV4(),
		Username: "cook",
		IsAdmin:  false,
		IsActive: true,
	}
}

func (suite *ControllerUpdateUserSuite) TestSessionTokenDeleteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.updateUser.IsActive = false
	suite.ctrl.Store.On("DeleteSessionTokensByUser", timeout, tx, suite.updateUser.ID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, tx, suite.updateUser)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateUserSuite) TestStoreUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	originalUser := suite.updateUser
	originalUser.Username = "force"
	suite.ctrl.Store.On("UpdateUser", timeout, tx, suite.updateUser).Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, tx, suite.updateUser)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateUserSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	originalUser := suite.updateUser
	originalUser.Username = "faith"
	suite.ctrl.Store.On("UpdateUser", timeout, tx, suite.updateUser).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, tx, suite.updateUser)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateUser(t *testing.T) {
	suite.Run(t, new(ControllerUpdateUserSuite))
}

// ControllerUpdateUserPassByUserIDSuite tests
// Controller.UpdateUserPassByUserID.
type ControllerUpdateUserPassByUserIDSuite struct {
	suite.Suite
	ctrl    *ControllerMock
	userID  uuid.UUID
	newPass []byte
}

func (suite *ControllerUpdateUserPassByUserIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.userID = testutil.NewUUIDV4()
	suite.newPass = []byte("meow")
}

func (suite *ControllerUpdateUserPassByUserIDSuite) TestSessionTokenDeleteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteSessionTokensByUser", timeout, tx, suite.userID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUserPassByUserID(timeout, tx, suite.userID, suite.newPass)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateUserPassByUserIDSuite) TestStoreUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteSessionTokensByUser", timeout, tx, suite.userID).Return(nil)
	suite.ctrl.Store.On("UpdateUserPassByUserID", timeout, tx, suite.userID, suite.newPass).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUserPassByUserID(timeout, tx, suite.userID, suite.newPass)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateUserPassByUserIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteSessionTokensByUser", timeout, tx, suite.userID).Return(nil)
	suite.ctrl.Store.On("UpdateUserPassByUserID", timeout, tx, suite.userID, suite.newPass).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUserPassByUserID(timeout, tx, suite.userID, suite.newPass)
		suite.NoError(err, "should fail")
	}()

	wait()
}

func TestController_UpdateUserPassByUserID(t *testing.T) {
	suite.Run(t, new(ControllerUpdateUserPassByUserIDSuite))
}
