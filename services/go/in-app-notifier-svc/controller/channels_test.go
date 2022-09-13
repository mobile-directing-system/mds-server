package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerUpdateNotificationChannelsByEntrySuite tests
// Controller.UpdateNotificationChannelsByEntry.
type ControllerUpdateNotificationChannelsByEntrySuite struct {
	suite.Suite
	ctrl           *ControllerMock
	tx             *testutil.DBTx
	sampleEntry    uuid.UUID
	sampleChannels []store.NotificationChannel
}

func (suite *ControllerUpdateNotificationChannelsByEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleEntry = testutil.NewUUIDV4()
	genChannel := func() store.NotificationChannel {
		return store.NotificationChannel{
			ID:      testutil.NewUUIDV4(),
			Entry:   suite.sampleEntry,
			Label:   "death",
			Timeout: 776,
		}
	}
	suite.sampleChannels = []store.NotificationChannel{
		genChannel(),
		genChannel(),
		genChannel(),
	}
}

func (suite *ControllerUpdateNotificationChannelsByEntrySuite) TestDeleteInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("DeleteNotificationChannelsByEntry", timeout, suite.tx, suite.sampleEntry).
		Return(errors.New("sad life")).Once()
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateNotificationChannelsByEntry(timeout, suite.tx, suite.sampleEntry, suite.sampleChannels)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateNotificationChannelsByEntrySuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("DeleteNotificationChannelsByEntry", timeout, suite.tx, suite.sampleEntry).
		Return(nil).Once()
	suite.ctrl.Store.On("CreateNotificationChannel", timeout, suite.tx, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateNotificationChannelsByEntry(timeout, suite.tx, suite.sampleEntry, suite.sampleChannels)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateNotificationChannelsByEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("DeleteNotificationChannelsByEntry", timeout, suite.tx, suite.sampleEntry).
		Return(nil).Once()
	for _, channel := range suite.sampleChannels {
		suite.ctrl.Store.On("CreateNotificationChannel", timeout, suite.tx, channel).
			Return(nil).Once()
	}
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateNotificationChannelsByEntry(timeout, suite.tx, suite.sampleEntry, suite.sampleChannels)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateNotificationChannelsByEntry(t *testing.T) {
	suite.Run(t, new(ControllerUpdateNotificationChannelsByEntrySuite))
}

// ControllerDeleteNotificationChannelsByEntrySuite tests
// Controller.DeleteNotificationChannelsByEntry.
type ControllerDeleteNotificationChannelsByEntrySuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleEntry uuid.UUID
}

func (suite *ControllerDeleteNotificationChannelsByEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntry = testutil.NewUUIDV4()
}

func (suite *ControllerDeleteNotificationChannelsByEntrySuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteNotificationChannelsByEntry", timeout, tx, suite.sampleEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteNotificationChannelsByEntry(timeout, tx, suite.sampleEntry)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteNotificationChannelsByEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("DeleteNotificationChannelsByEntry", timeout, tx, suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteNotificationChannelsByEntry(timeout, tx, suite.sampleEntry)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_DeleteNotificationChannelsByEntry(t *testing.T) {
	suite.Run(t, new(ControllerDeleteNotificationChannelsByEntrySuite))
}
