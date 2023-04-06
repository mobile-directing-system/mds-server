package controller

import (
	"errors"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
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
