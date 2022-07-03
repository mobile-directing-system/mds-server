package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateGroupSuite tests Controller.CreateGroup.
type ControllerCreateGroupSuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleGroup store.Group
}

func (suite *ControllerCreateGroupSuite) SetupTest() {
	suite.ctrl = NewMockController()
	members := make([]uuid.UUID, 16)
	for i := range members {
		members[i] = testutil.NewUUIDV4()
	}
	suite.sampleGroup = store.Group{
		Title:       "command",
		Description: "ready",
		Operation:   uuid.NullUUID{UUID: testutil.NewUUIDV4(), Valid: true},
		Members:     members,
	}
}

func (suite *ControllerCreateGroupSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateGroupSuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(store.Group{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateGroupSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	created := suite.sampleGroup
	created.ID = testutil.NewUUIDV4()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(created, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupCreated", created).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateGroupSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	created := suite.sampleGroup
	created.ID = testutil.NewUUIDV4()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(created, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupCreated", created).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(created, got, "should return correct value")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_CreateGroup(t *testing.T) {
	suite.Run(t, new(ControllerCreateGroupSuite))
}

// ControllerUpdateGroupSuite tests Controller.UpdateGroup.
type ControllerUpdateGroupSuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleGroup store.Group
}

func (suite *ControllerUpdateGroupSuite) SetupTest() {
	suite.ctrl = NewMockController()
	members := make([]uuid.UUID, 16)
	for i := range members {
		members[i] = testutil.NewUUIDV4()
	}
	suite.sampleGroup = store.Group{
		ID:          testutil.NewUUIDV4(),
		Title:       "command",
		Description: "ready",
		Operation:   uuid.NullUUID{UUID: testutil.NewUUIDV4(), Valid: true},
		Members:     members,
	}
}

func (suite *ControllerUpdateGroupSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateGroupSuite) TestUpdateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateGroupSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupUpdated", suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateGroupSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupUpdated", suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, suite.sampleGroup)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_UpdateGroup(t *testing.T) {
	suite.Run(t, new(ControllerUpdateGroupSuite))
}

// ControllerDeleteGroupByIDSuite tests Controller.DeleteGroupByID.
type ControllerDeleteGroupByIDSuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleGroup uuid.UUID
}

func (suite *ControllerDeleteGroupByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleGroup = testutil.NewUUIDV4()
}

func (suite *ControllerDeleteGroupByIDSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestDeleteInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("DeleteGroupByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("DeleteGroupByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupDeleted", suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("DeleteGroupByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupDeleted", suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, suite.sampleGroup)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_DeleteGroupByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteGroupByIDSuite))
}
