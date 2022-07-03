package controller

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerUpdatePermissionsByUserSuite tests
// Controller.UpdatePermissionsByUser.
type ControllerUpdatePermissionsByUserSuite struct {
	suite.Suite
	ctrl                     *ControllerMock
	sampleUserID             uuid.UUID
	sampleUpdatedPermissions []permission.Permission
}

func (suite *ControllerUpdatePermissionsByUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUserID = testutil.NewUUIDV4()
	suite.sampleUpdatedPermissions = []permission.Permission{
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

func (suite *ControllerUpdatePermissionsByUserSuite) TestStoreUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
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

func (suite *ControllerUpdatePermissionsByUserSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID, suite.sampleUpdatedPermissions).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

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
