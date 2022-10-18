package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerUpdateRadioChannelsByEntrySuite tests
// Controller.UpdateRadioChannelsByEntry.
type ControllerUpdateRadioChannelsByEntrySuite struct {
	suite.Suite
	ctrl           *ControllerMock
	tx             *testutil.DBTx
	sampleEntry    uuid.UUID
	sampleChannels []store.RadioChannel
}

func (suite *ControllerUpdateRadioChannelsByEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleEntry = testutil.NewUUIDV4()
	genChannel := func() store.RadioChannel {
		return store.RadioChannel{
			ID:      testutil.NewUUIDV4(),
			Entry:   suite.sampleEntry,
			Label:   "death",
			Timeout: 776,
		}
	}
	suite.sampleChannels = []store.RadioChannel{
		genChannel(),
		genChannel(),
		genChannel(),
	}
}

func (suite *ControllerUpdateRadioChannelsByEntrySuite) TestDeleteInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("DeleteRadioChannelsByEntry", timeout, suite.tx, suite.sampleEntry).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateRadioChannelsByEntry(timeout, suite.tx, suite.sampleEntry, suite.sampleChannels)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateRadioChannelsByEntrySuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("DeleteRadioChannelsByEntry", timeout, suite.tx, suite.sampleEntry).
		Return(nil).Once()
	suite.ctrl.Store.On("CreateRadioChannel", timeout, suite.tx, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateRadioChannelsByEntry(timeout, suite.tx, suite.sampleEntry, suite.sampleChannels)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateRadioChannelsByEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("DeleteRadioChannelsByEntry", timeout, suite.tx, suite.sampleEntry).
		Return(nil).Once()
	for _, channel := range suite.sampleChannels {
		suite.ctrl.Store.On("CreateRadioChannel", timeout, suite.tx, channel).
			Return(nil).Once()
	}
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateRadioChannelsByEntry(timeout, suite.tx, suite.sampleEntry, suite.sampleChannels)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateRadioChannelsByEntry(t *testing.T) {
	suite.Run(t, new(ControllerUpdateRadioChannelsByEntrySuite))
}

// ControllerDeleteRadioChannelsByEntrySuite tests
// Controller.DeleteRadioChannelsByEntry.
type ControllerDeleteRadioChannelsByEntrySuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleEntry uuid.UUID
}

func (suite *ControllerDeleteRadioChannelsByEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntry = testutil.NewUUIDV4()
}

func (suite *ControllerDeleteRadioChannelsByEntrySuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteRadioChannelsByEntry", timeout, tx, suite.sampleEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteRadioChannelsByEntry(timeout, tx, suite.sampleEntry)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteRadioChannelsByEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteRadioChannelsByEntry", timeout, tx, suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteRadioChannelsByEntry(timeout, tx, suite.sampleEntry)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_DeleteRadioChannelsByEntry(t *testing.T) {
	suite.Run(t, new(ControllerDeleteRadioChannelsByEntrySuite))
}
