package controller

import (
	"context"
	"errors"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateIntelSuite tests Controller.CreateIntel.
type ControllerCreateIntelSuite struct {
	suite.Suite
	ctrl   *ControllerMock
	create store.Intel
}

func (suite *ControllerCreateIntelSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.ctrl.DB.GenTx = true
	suite.create = store.Intel{
		ID:         testutil.NewUUIDV4(),
		CreatedAt:  testutil.NewRandomTime(),
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Importance: 208,
		IsValid:    true,
	}

	suite.ctrl.Store.On("CreateIntel", mock.Anything, mock.Anything, suite.create).
		Return(nil).Maybe()
}

func (suite *ControllerCreateIntelSuite) TeardownTest() {
	suite.ctrl.Store.AssertExpectations(suite.T())
}

func (suite *ControllerCreateIntelSuite) TestBeginTxFail() {
	suite.ctrl.DB.BeginFail = true

	err := suite.ctrl.Ctrl.CreateIntel(context.Background(), suite.create)
	suite.Error(err, "should fail")
}

func (suite *ControllerCreateIntelSuite) TestCreateFail() {
	testutil.UnsetAndOn(&suite.ctrl.Store.Mock, "CreateIntel", mock.Anything, mock.Anything, suite.create).
		Return(errors.New("sad life"))

	err := suite.ctrl.Ctrl.CreateIntel(context.Background(), suite.create)
	suite.Error(err, "should fail")
}

func TestController_CreateIntel(t *testing.T) {
	suite.Run(t, new(ControllerCreateIntelSuite))
}
