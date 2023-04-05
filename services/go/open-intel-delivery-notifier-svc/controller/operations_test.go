package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// ControllerCreateOperationSuite tests Controller.CreateOperation.
type ControllerCreateOperationSuite struct {
	suite.Suite
	ctrl            *ControllerMock
	sampleOperation store.Operation
}

func (suite *ControllerCreateOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleOperation = store.Operation{
		ID:          testutil.NewUUIDV4(),
		Title:       "strict",
		Description: "applaud",
		Start:       time.Date(2022, 8, 5, 0, 33, 30, 1, time.UTC),
		End:         nulls.NewTime(time.Date(2022, 8, 6, 0, 33, 30, 1, time.UTC)),
		IsArchived:  true,
	}
}

func (suite *ControllerCreateOperationSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateOperation", timeout, tx, suite.sampleOperation).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateOperation(timeout, tx, suite.sampleOperation)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateOperation", timeout, tx, suite.sampleOperation).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateOperation(timeout, tx, suite.sampleOperation)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_CreateOperation(t *testing.T) {
	suite.Run(t, new(ControllerCreateOperationSuite))
}

// ControllerUpdateOperationSuite tests Controller.UpdateOperation.
type ControllerUpdateOperationSuite struct {
	suite.Suite
	ctrl            *ControllerMock
	sampleOperation store.Operation
}

func (suite *ControllerUpdateOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleOperation = store.Operation{
		ID:          testutil.NewUUIDV4(),
		Title:       "pain",
		Description: "spade",
		Start:       time.Date(2022, 8, 5, 0, 39, 34, 0, time.UTC),
		End:         nulls.NewTime(time.Date(2022, 8, 5, 0, 39, 34, 1, time.UTC)),
		IsArchived:  true,
	}
}

func (suite *ControllerUpdateOperationSuite) TestUpdateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateOperation", timeout, tx, suite.sampleOperation).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperation(timeout, tx, suite.sampleOperation)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateOperation", timeout, tx, suite.sampleOperation).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperation(timeout, tx, suite.sampleOperation)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateOperation(t *testing.T) {
	suite.Run(t, new(ControllerUpdateOperationSuite))
}

// ControllerUpdateOperationMembersByOperationSuite tests Controller.UpdateOperationMembersByOperation.
type ControllerUpdateOperationMembersByOperationSuite struct {
	suite.Suite
	ctrl                     *ControllerMock
	sampleOperation          uuid.UUID
	sampleMembers            []uuid.UUID
	sampleOldMembersWithMore []uuid.UUID
	sampleRemovedMembers     []uuid.UUID
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleOperation = testutil.NewUUIDV4()
	suite.sampleMembers = make([]uuid.UUID, 16)
	for i := range suite.sampleMembers {
		suite.sampleMembers[i] = testutil.NewUUIDV4()
	}
	suite.sampleOldMembersWithMore = make([]uuid.UUID, 0, len(suite.sampleMembers))
	for _, member := range suite.sampleMembers {
		suite.sampleOldMembersWithMore = append(suite.sampleOldMembersWithMore, member)
	}
	suite.sampleRemovedMembers = make([]uuid.UUID, 4)
	for i := range suite.sampleRemovedMembers {
		suite.sampleRemovedMembers[i] = testutil.NewUUIDV4()
	}
	suite.sampleOldMembersWithMore = append(suite.sampleOldMembersWithMore, suite.sampleRemovedMembers...)
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, tx, suite.sampleOperation, suite.sampleMembers).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, tx, suite.sampleOperation, suite.sampleMembers)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, tx, suite.sampleOperation, suite.sampleMembers).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, tx, suite.sampleOperation, suite.sampleMembers)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateOperationMembersByOperation(t *testing.T) {
	suite.Run(t, new(ControllerUpdateOperationMembersByOperationSuite))
}
