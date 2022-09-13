package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateAddressBookEntrySuite tests
// Controller.CreateAddressBookEntry.
type ControllerCreateAddressBookEntrySuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleEntry store.AddressBookEntry
}

func (suite *ControllerCreateAddressBookEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntry = store.AddressBookEntry{
		Label:       "hurry",
		Description: "relief",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
}

func (suite *ControllerCreateAddressBookEntrySuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateAddressBookEntry", timeout, tx, suite.sampleEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateAddressBookEntry(timeout, tx, suite.sampleEntry)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateAddressBookEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateAddressBookEntry", timeout, tx, suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateAddressBookEntry(timeout, tx, suite.sampleEntry)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_CreateAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerCreateAddressBookEntrySuite))
}

// ControllerUpdateAddressBookEntrySuite tests
// Controller.UpdateAddressBookEntry.
type ControllerUpdateAddressBookEntrySuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleEntry store.AddressBookEntry
}

func (suite *ControllerUpdateAddressBookEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntry = store.AddressBookEntry{
		ID:          testutil.NewUUIDV4(),
		Label:       "hurry",
		Description: "relief",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
}

func (suite *ControllerUpdateAddressBookEntrySuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateAddressBookEntry", timeout, tx, suite.sampleEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateAddressBookEntry(timeout, tx, suite.sampleEntry)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateAddressBookEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("UpdateAddressBookEntry", timeout, tx, suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateAddressBookEntry(timeout, tx, suite.sampleEntry)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerUpdateAddressBookEntrySuite))
}

// ControllerDeleteAddressBookEntryByIDSuite tests
// Controller.DeleteAddressBookEntryByID.
type ControllerDeleteAddressBookEntryByIDSuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleEntry uuid.UUID
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntry = testutil.NewUUIDV4()
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteAddressBookEntryByID", timeout, tx, suite.sampleEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteAddressBookEntryByID(timeout, tx, suite.sampleEntry)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteAddressBookEntryByID", timeout, tx, suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteAddressBookEntryByID(timeout, tx, suite.sampleEntry)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_DeleteAddressBookEntryByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteAddressBookEntryByIDSuite))
}
