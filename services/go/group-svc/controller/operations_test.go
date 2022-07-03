package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateOperationSuite tests Controller.CreateOperation.
type ControllerCreateOperationSuite struct {
	suite.Suite
	ctrl            *ControllerMock
	sampleOperation uuid.UUID
}

func (suite *ControllerCreateOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleOperation = testutil.NewUUIDV4()
}

func (suite *ControllerCreateOperationSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateOperation(timeout, suite.sampleOperation)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateOperationSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperation).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateOperation(timeout, suite.sampleOperation)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperation).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateOperation(timeout, suite.sampleOperation)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_CreateOperation(t *testing.T) {
	suite.Run(t, new(ControllerCreateOperationSuite))
}
