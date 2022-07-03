package controller

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateUserSuite tests Controller.CreateUser.
type ControllerCreateUserSuite struct {
	suite.Suite
	ctrl         *ControllerMock
	sampleUserID uuid.UUID
}

func (suite *ControllerCreateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUserID = testutil.NewUUIDV4()
}

func (suite *ControllerCreateUserSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerCreateUserSuite) TestStoreCreateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerCreateUserSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, suite.sampleUserID)
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestController_CreateUser(t *testing.T) {
	suite.Run(t, new(ControllerCreateUserSuite))
}

// ControllerDeleteUserByIDSuite tests Controller.DeleteUserByID.
type ControllerDeleteUserByIDSuite struct {
	suite.Suite
	ctrl         *ControllerMock
	sampleUserID uuid.UUID
}

func (suite *ControllerDeleteUserByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUserID = testutil.NewUUIDV4()
}

func (suite *ControllerDeleteUserByIDSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerDeleteUserByIDSuite) TestPermissionUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID, []store.Permission{}).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerDeleteUserByIDSuite) TestNotifyFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID, []store.Permission{}).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyPermissionsUpdated", suite.sampleUserID, []store.Permission{}).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerDeleteUserByIDSuite) TestDeleteUserFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID, []store.Permission{}).
		Return(nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyPermissionsUpdated", suite.sampleUserID, []store.Permission{}).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerDeleteUserByIDSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID, []store.Permission{}).
		Return(nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyPermissionsUpdated", suite.sampleUserID, []store.Permission{}).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUserID)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestController_DeleteUserByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteUserByIDSuite))
}
