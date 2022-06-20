package controller

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerPermissionsByUserSuite tests Controller.PermissionsByUser.
type ControllerPermissionsByUserSuite struct {
	suite.Suite
	ctrl              *ControllerMock
	sampleUserID      uuid.UUID
	samplePermissions []store.Permission
}

func (suite *ControllerPermissionsByUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUserID = uuid.New()
	suite.samplePermissions = []store.Permission{
		{Name: "meow"},
		{Name: "woof"},
	}
}

func (suite *ControllerPermissionsByUserSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.PermissionsByUser(timeout, uuid.New())
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerPermissionsByUserSuite) TestAssureUserExistsFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AssureUserExists", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.PermissionsByUser(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerPermissionsByUserSuite) TestStoreRetrievalFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AssureUserExists", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).Return(nil)
	suite.ctrl.Store.On("PermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.PermissionsByUser(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerPermissionsByUserSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AssureUserExists", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).Return(nil)
	suite.ctrl.Store.On("PermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(suite.samplePermissions, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.PermissionsByUser(timeout, suite.sampleUserID)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.samplePermissions, got, "should return correct permissions")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestController_PermissionsByUser(t *testing.T) {
	suite.Run(t, new(ControllerPermissionsByUserSuite))
}

// ControllerUpdatePermissionsByUserSuite tests
// Controller.UpdatePermissionsByUser.
type ControllerUpdatePermissionsByUserSuite struct {
	suite.Suite
	ctrl                     *ControllerMock
	sampleUserID             uuid.UUID
	sampleUpdatedPermissions []store.Permission
}

func (suite *ControllerUpdatePermissionsByUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUserID = uuid.New()
	suite.sampleUpdatedPermissions = []store.Permission{
		{Name: "meow"},
		{Name: "woof"},
	}
}

func (suite *ControllerUpdatePermissionsByUserSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdatePermissionsByUser(timeout, suite.sampleUserID, suite.sampleUpdatedPermissions)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdatePermissionsByUserSuite) TestAssureUserExistsFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AssureUserExists", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.PermissionsByUser(timeout, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdatePermissionsByUserSuite) TestStoreUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AssureUserExists", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).Return(nil)
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID, suite.sampleUpdatedPermissions).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdatePermissionsByUser(timeout, suite.sampleUserID, suite.sampleUpdatedPermissions)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdatePermissionsByUserSuite) TestNotifyFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AssureUserExists", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).Return(nil)
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID, suite.sampleUpdatedPermissions).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyPermissionsUpdated", suite.sampleUserID, suite.sampleUpdatedPermissions).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdatePermissionsByUser(timeout, suite.sampleUserID, suite.sampleUpdatedPermissions)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdatePermissionsByUserSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AssureUserExists", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).Return(nil)
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID, suite.sampleUpdatedPermissions).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyPermissionsUpdated", suite.sampleUserID, suite.sampleUpdatedPermissions).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdatePermissionsByUser(timeout, suite.sampleUserID, suite.sampleUpdatedPermissions)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestController_UpdatePermissionsByUser(t *testing.T) {
	suite.Run(t, new(ControllerUpdatePermissionsByUserSuite))
}
