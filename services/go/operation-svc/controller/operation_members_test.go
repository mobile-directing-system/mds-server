package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerUpdateOperationMembersByOperationSuite tests
// Controller.UpdateOperationMembersByOperation.
type ControllerUpdateOperationMembersByOperationSuite struct {
	suite.Suite
	ctrl              *ControllerMock
	sampleOperationID uuid.UUID
	sampleMembers     []uuid.UUID
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleOperationID = testutil.NewUUIDV4()
	suite.sampleMembers = make([]uuid.UUID, 16)
	for i := range suite.sampleMembers {
		suite.sampleMembers[i] = testutil.NewUUIDV4()
	}
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, suite.sampleOperationID, suite.sampleMembers)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestAssureExistsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID).
		Return(store.Operation{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, suite.sampleOperationID, suite.sampleMembers)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestStoreUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID).
		Return(store.Operation{}, nil)
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0],
		suite.sampleOperationID, suite.sampleMembers).Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, suite.sampleOperationID, suite.sampleMembers)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID).
		Return(store.Operation{}, nil)
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0],
		suite.sampleOperationID, suite.sampleMembers).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyOperationMembersUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID, suite.sampleMembers).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, suite.sampleOperationID, suite.sampleMembers)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID).
		Return(store.Operation{}, nil)
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0],
		suite.sampleOperationID, suite.sampleMembers).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyOperationMembersUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID, suite.sampleMembers).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, suite.sampleOperationID, suite.sampleMembers)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit")
	}()

	wait()
}

func TestController_UpdateOperationMembersByOperation(t *testing.T) {
	suite.Run(t, new(ControllerUpdateOperationMembersByOperationSuite))
}

// ControllerOperationMembersByOperationSuite tests
// Controller.OperationMembersByOperation.
type ControllerOperationMembersByOperationSuite struct {
	suite.Suite
	ctrl              *ControllerMock
	sampleOperationID uuid.UUID
	sampleParams      pagination.Params
	sampleMembers     pagination.Paginated[store.User]
}

func (suite *ControllerOperationMembersByOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleOperationID = testutil.NewUUIDV4()
	suite.sampleParams = pagination.Params{
		Limit:          47,
		Offset:         93,
		OrderBy:        nulls.NewString("intend"),
		OrderDirection: pagination.OrderDirDesc,
	}
	suite.sampleMembers = pagination.NewPaginated(suite.sampleParams, []store.User{
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "power",
			FirstName: "justice",
			LastName:  "demand",
		},
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "during",
			FirstName: "occasion",
			LastName:  "cliff",
		},
	}, 645)
}

func (suite *ControllerOperationMembersByOperationSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.OperationMembersByOperation(timeout, suite.sampleOperationID, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerOperationMembersByOperationSuite) TestAssureOperationFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID).
		Return(store.Operation{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.OperationMembersByOperation(timeout, suite.sampleOperationID, suite.sampleParams)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerOperationMembersByOperationSuite) TestMemberRetrievalFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID).
		Return(store.Operation{}, nil)
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID, suite.sampleParams).
		Return(pagination.Paginated[store.User]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.OperationMembersByOperation(timeout, suite.sampleOperationID, suite.sampleParams)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerOperationMembersByOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID).
		Return(store.Operation{}, nil)
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperationID, suite.sampleParams).
		Return(suite.sampleMembers, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.OperationMembersByOperation(timeout, suite.sampleOperationID, suite.sampleParams)
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit")
		suite.Equal(suite.sampleMembers, got, "should return correct value")
	}()

	wait()
}

func TestController_OperationMembersByOperation(t *testing.T) {
	suite.Run(t, new(ControllerOperationMembersByOperationSuite))
}
