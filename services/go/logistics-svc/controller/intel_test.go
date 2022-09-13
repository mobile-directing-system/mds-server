package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// ControllerCreateIntelSuite tests Controller.CreateIntel.
type ControllerCreateIntelSuite struct {
	suite.Suite
	ctrl         *ControllerMock
	tx           *testutil.DBTx
	sampleCreate store.Intel
}

func (suite *ControllerCreateIntelSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleCreate = store.Intel{
		ID:         testutil.NewUUIDV4(),
		CreatedAt:  time.Date(2022, 9, 1, 11, 12, 57, 0, time.UTC),
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       "meow",
		Content:    nil,
		SearchText: nulls.NewString("dive"),
		Importance: 234,
		IsValid:    true,
		Assignments: []store.IntelAssignment{
			{To: testutil.NewUUIDV4()},
			{To: testutil.NewUUIDV4()},
		},
	}
}

func (suite *ControllerCreateIntelSuite) TestCreateIntelFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("CreateIntel", timeout, suite.tx, suite.sampleCreate).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.tx, suite.sampleCreate)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestScheduleDeliveriesFails() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("CreateIntel", timeout, suite.tx, suite.sampleCreate).
		Return(nil)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleCreate.ID).
		Return(store.Intel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.tx, suite.sampleCreate)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("CreateIntel", timeout, suite.tx, suite.sampleCreate).
		Return(nil)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleCreate.ID).
		Return(store.Intel{}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.tx, suite.sampleCreate)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_CreateIntel(t *testing.T) {
	suite.Run(t, new(ControllerCreateIntelSuite))
}

// ControllerInvalidateIntelSuite tests Controller.InvalidateIntel.
type ControllerInvalidateIntelSuite struct {
	suite.Suite
	ctrl     *ControllerMock
	tx       *testutil.DBTx
	sampleID uuid.UUID
}

func (suite *ControllerInvalidateIntelSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleID = testutil.NewUUIDV4()
}

func (suite *ControllerInvalidateIntelSuite) TestInvalidateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("InvalidateIntelByID", timeout, suite.tx, suite.sampleID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.InvalidateIntelByID(timeout, suite.tx, suite.sampleID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerInvalidateIntelSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("InvalidateIntelByID", timeout, suite.tx, suite.sampleID).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.InvalidateIntelByID(timeout, suite.tx, suite.sampleID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_InvalidateIntel(t *testing.T) {
	suite.Run(t, new(ControllerInvalidateIntelSuite))
}
