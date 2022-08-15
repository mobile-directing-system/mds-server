package controller

import (
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

func (suite *ControllerUpdatePermissionsByUserSuite) TestStoreUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, tx, suite.sampleUserID, suite.sampleUpdatedPermissions).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdatePermissionsByUser(timeout, tx, suite.sampleUserID, suite.sampleUpdatedPermissions)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdatePermissionsByUserSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, tx, suite.sampleUserID, suite.sampleUpdatedPermissions).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdatePermissionsByUser(timeout, tx, suite.sampleUserID, suite.sampleUpdatedPermissions)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdatePermissionsByUser(t *testing.T) {
	suite.Run(t, new(ControllerUpdatePermissionsByUserSuite))
}
