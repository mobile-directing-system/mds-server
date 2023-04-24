package controller

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/lefinal/zaprec"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
	"testing"
)

// ControllerCreateAddressBookEntriesuite tests Controller.CreateAddressBookEntry.
type ControllerCreateAddressBookEntriesuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleEntry store.AddressBookEntry
}

func (suite *ControllerCreateAddressBookEntriesuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntry = store.AddressBookEntry{
		Label:       "idle",
		Description: "root",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
}

func (suite *ControllerCreateAddressBookEntriesuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateAddressBookEntry(timeout, suite.sampleEntry)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateAddressBookEntriesuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(store.AddressBookEntryDetailed{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateAddressBookEntry(timeout, suite.sampleEntry)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateAddressBookEntriesuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	created := store.AddressBookEntryDetailed{
		AddressBookEntry: suite.sampleEntry,
		UserDetails: nulls.NewJSONNullable(store.User{
			ID:        suite.sampleEntry.ID,
			Username:  "field",
			FirstName: "saddle",
			LastName:  "eye",
		}),
	}
	created.ID = testutil.NewUUIDV4()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(created, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyAddressBookEntryCreated", timeout, suite.ctrl.DB.Tx[0], created.AddressBookEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateAddressBookEntry(timeout, suite.sampleEntry)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateAddressBookEntriesuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	created := store.AddressBookEntryDetailed{
		AddressBookEntry: suite.sampleEntry,
		UserDetails: nulls.NewJSONNullable(store.User{
			ID:        suite.sampleEntry.ID,
			Username:  "field",
			FirstName: "saddle",
			LastName:  "eye",
		}),
	}
	created.ID = testutil.NewUUIDV4()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(created, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyAddressBookEntryCreated", timeout, suite.ctrl.DB.Tx[0], created.AddressBookEntry).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.CreateAddressBookEntry(timeout, suite.sampleEntry)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(created, got, "should return correct value")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_CreateAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerCreateAddressBookEntriesuite))
}

// ControllerUpdateAddressBookEntrySuite tests Controller.UpdateAddressBookEntry.
type ControllerUpdateAddressBookEntrySuite struct {
	suite.Suite
	ctrl                *ControllerMock
	sampleEntry         store.AddressBookEntry
	sampleEntryDetailed store.AddressBookEntryDetailed
}

func (suite *ControllerUpdateAddressBookEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntry = store.AddressBookEntry{
		ID:          testutil.NewUUIDV4(),
		Label:       "tomorrow",
		Description: "fit",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.sampleEntryDetailed = store.AddressBookEntryDetailed{
		AddressBookEntry: store.AddressBookEntry{
			ID:          suite.sampleEntry.ID,
			Label:       "noise",
			Description: "low",
			Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
			User:        nulls.NewUUID(testutil.NewUUIDV4()),
		},
	}
}

func (suite *ControllerUpdateAddressBookEntrySuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateAddressBookEntry(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateAddressBookEntrySuite) TestRetrieveEntryFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry.ID, uuid.NullUUID{}).
		Return(store.AddressBookEntryDetailed{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateAddressBookEntry(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateAddressBookEntrySuite) TestLimitViolationForGlobal() {
	limitToUser := testutil.NewUUIDV4()
	globalEntry := suite.sampleEntryDetailed
	globalEntry.User = uuid.NullUUID{}
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry.ID, uuid.NullUUID{}).
		Return(globalEntry, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateAddressBookEntry(timeout, suite.sampleEntry, nulls.NewUUID(limitToUser))
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateAddressBookEntrySuite) TestLimitViolationForForeignUsersEntry() {
	limitToUser := testutil.NewUUIDV4()
	globalEntry := suite.sampleEntryDetailed
	globalEntry.User = nulls.NewUUID(testutil.NewUUIDV4())
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry.ID, uuid.NullUUID{}).
		Return(globalEntry, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateAddressBookEntry(timeout, suite.sampleEntry, nulls.NewUUID(limitToUser))
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateAddressBookEntrySuite) TestUpdateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry.ID, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	suite.ctrl.Store.On("UpdateAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateAddressBookEntry(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateAddressBookEntrySuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry.ID, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	suite.ctrl.Store.On("UpdateAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyAddressBookEntryUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateAddressBookEntry(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateAddressBookEntrySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry.ID, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	suite.ctrl.Store.On("UpdateAddressBookEntry", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyAddressBookEntryUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateAddressBookEntry(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_UpdateAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerUpdateAddressBookEntrySuite))
}

// ControllerDeleteAddressBookEntryByIDSuite tests Controller.DeleteAddressBookEntryByID.
type ControllerDeleteAddressBookEntryByIDSuite struct {
	suite.Suite
	ctrl                *ControllerMock
	sampleEntry         uuid.UUID
	sampleEntryDetailed store.AddressBookEntryDetailed
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntry = testutil.NewUUIDV4()
	suite.sampleEntryDetailed = store.AddressBookEntryDetailed{
		AddressBookEntry: store.AddressBookEntry{
			ID:          suite.sampleEntry,
			Label:       "shape",
			Description: "doctor",
			Operation:   uuid.NullUUID{},
			User:        uuid.NullUUID{},
		},
		UserDetails: nulls.JSONNullable[store.User]{},
	}
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteAddressBookEntryByID(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) TestRetrieveFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry, uuid.NullUUID{}).
		Return(store.AddressBookEntryDetailed{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteAddressBookEntryByID(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) TestLimitViolationForGlobal() {
	limitToUser := testutil.NewUUIDV4()
	suite.sampleEntryDetailed.User = uuid.NullUUID{}
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteAddressBookEntryByID(timeout, suite.sampleEntry, nulls.NewUUID(limitToUser))
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) TestLimitViolationForForeignUsersEntry() {
	limitToUser := testutil.NewUUIDV4()
	suite.sampleEntryDetailed.User = nulls.NewUUID(testutil.NewUUIDV4())
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteAddressBookEntryByID(timeout, suite.sampleEntry, nulls.NewUUID(limitToUser))
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) TestDeleteInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	suite.ctrl.Store.On("DeleteAddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteAddressBookEntryByID(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	suite.ctrl.Store.On("DeleteAddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyAddressBookEntryDeleted", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteAddressBookEntryByID(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteAddressBookEntryByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry, uuid.NullUUID{}).
		Return(suite.sampleEntryDetailed, nil)
	suite.ctrl.Store.On("DeleteAddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyAddressBookEntryDeleted", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteAddressBookEntryByID(timeout, suite.sampleEntry, uuid.NullUUID{})
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_DeleteAddressBookEntryByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteAddressBookEntryByIDSuite))
}

// ControllerAddressBookEntryByIDSuite tests Controller.AddressBookEntryByID.
type ControllerAddressBookEntryByIDSuite struct {
	suite.Suite
	ctrl            *ControllerMock
	sampleEntry     store.AddressBookEntryDetailed
	sampleVisibleBy uuid.UUID
}

func (suite *ControllerAddressBookEntryByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleEntry = store.AddressBookEntryDetailed{
		AddressBookEntry: store.AddressBookEntry{
			ID:          testutil.NewUUIDV4(),
			Label:       "lung",
			Description: "stand",
			Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
			User:        nulls.NewUUID(testutil.NewUUIDV4()),
		},
		UserDetails: nulls.NewJSONNullable(store.User{
			ID:        uuid.UUID{},
			Username:  "lipstick",
			FirstName: "month",
			LastName:  "buy",
		}),
	}
	suite.sampleVisibleBy = testutil.NewUUIDV4()
}

func (suite *ControllerAddressBookEntryByIDSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.AddressBookEntryByID(timeout, suite.sampleEntry.ID, nulls.NewUUID(suite.sampleVisibleBy))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerAddressBookEntryByIDSuite) TestRetrieveFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry.ID, nulls.NewUUID(suite.sampleVisibleBy)).
		Return(store.AddressBookEntryDetailed{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.AddressBookEntryByID(timeout, suite.sampleEntry.ID, nulls.NewUUID(suite.sampleVisibleBy))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerAddressBookEntryByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleEntry.ID, nulls.NewUUID(suite.sampleVisibleBy)).
		Return(suite.sampleEntry, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.AddressBookEntryByID(timeout, suite.sampleEntry.ID, nulls.NewUUID(suite.sampleVisibleBy))
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleEntry, got, "should return correct value")
	}()

	wait()
}

func TestController_AddressBookEntryByID(t *testing.T) {
	suite.Run(t, new(ControllerAddressBookEntryByIDSuite))
}

// ControllerAddressBookEntriesSuite tests Controller.AddressBookEntries.
type ControllerAddressBookEntriesSuite struct {
	suite.Suite
	ctrl          *ControllerMock
	sampleFilters store.AddressBookEntryFilters
	sampleParams  pagination.Params
	sampleEntries pagination.Paginated[store.AddressBookEntryDetailed]
}

func (suite *ControllerAddressBookEntriesSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleFilters = store.AddressBookEntryFilters{
		ByUser:        nulls.NewUUID(testutil.NewUUIDV4()),
		ForOperation:  nulls.NewUUID(testutil.NewUUIDV4()),
		ExcludeGlobal: true,
	}
	suite.sampleParams = pagination.Params{
		Limit:          923,
		Offset:         209,
		OrderBy:        nulls.NewString("brick"),
		OrderDirection: "desc",
	}
	suite.sampleEntries = pagination.NewPaginated(suite.sampleParams, []store.AddressBookEntryDetailed{
		{
			AddressBookEntry: store.AddressBookEntry{
				ID:          testutil.NewUUIDV4(),
				Label:       "sympathy",
				Description: "civilize",
				Operation:   uuid.NullUUID{},
				User:        nulls.NewUUID(testutil.NewUUIDV4()),
			},
		},
		{
			AddressBookEntry: store.AddressBookEntry{
				ID:          testutil.NewUUIDV4(),
				Label:       "regard",
				Description: "throw",
				Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
				User:        uuid.NullUUID{},
			},
			UserDetails: nulls.NewJSONNullable(store.User{
				ID:        uuid.UUID{},
				Username:  "around",
				FirstName: "other",
				LastName:  "purple",
				IsActive:  true,
			}),
		},
	}, 9313)
}

func (suite *ControllerAddressBookEntriesSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.AddressBookEntries(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerAddressBookEntriesSuite) TestRetrieveFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.AddressBookEntries(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerAddressBookEntriesSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(suite.sampleEntries, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.AddressBookEntries(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleEntries, got, "should return correct value")
	}()

	wait()
}

func TestController_AddressBookEntries(t *testing.T) {
	suite.Run(t, new(ControllerAddressBookEntriesSuite))
}

// ControllerSearchAddressBookEntriesSuite tests
// Controller.SearchAddressBookEntries
type ControllerSearchAddressBookEntriesSuite struct {
	suite.Suite
	ctrl          *ControllerMock
	tx            *testutil.DBTx
	sampleFilters store.AddressBookEntryFilters
	sampleParams  search.Params
	sampleEntries []store.AddressBookEntryDetailed
}

func (suite *ControllerSearchAddressBookEntriesSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleFilters = store.AddressBookEntryFilters{
		ByUser:                  nulls.NewUUID(testutil.NewUUIDV4()),
		ForOperation:            nulls.NewUUID(testutil.NewUUIDV4()),
		ExcludeGlobal:           true,
		VisibleBy:               nulls.NewUUID(testutil.NewUUIDV4()),
		IncludeForInactiveUsers: true,
	}
	suite.sampleParams = search.Params{
		Query:  "among",
		Offset: 93,
		Limit:  286,
	}
	suite.sampleEntries = []store.AddressBookEntryDetailed{
		{
			AddressBookEntry: store.AddressBookEntry{
				ID: testutil.NewUUIDV4(),
			},
		},
		{
			AddressBookEntry: store.AddressBookEntry{
				ID: testutil.NewUUIDV4(),
			},
		},
	}
}

func (suite *ControllerSearchAddressBookEntriesSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.SearchAddressBookEntries(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerSearchAddressBookEntriesSuite) TestSearchFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("SearchAddressBookEntries", timeout, suite.tx, suite.sampleFilters, suite.sampleParams).
		Return(search.Result[store.AddressBookEntryDetailed]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.SearchAddressBookEntries(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerSearchAddressBookEntriesSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("SearchAddressBookEntries", timeout, suite.tx, suite.sampleFilters, suite.sampleParams).
		Return(search.Result[store.AddressBookEntryDetailed]{Hits: suite.sampleEntries}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.SearchAddressBookEntries(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleEntries, got.Hits, "shoudl return correct result")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_SearchAddressBookEntries(t *testing.T) {
	suite.Run(t, new(ControllerSearchAddressBookEntriesSuite))
}

// ControllerRebuildAddressBookEntrySearchSuite tests
// Controller.RebuildAddressBookEntrySearch.
type ControllerRebuildAddressBookEntrySearchSuite struct {
	suite.Suite
	ctrl     *ControllerMock
	tx       *testutil.DBTx
	recorder *zaprec.RecordStore
}

func (suite *ControllerRebuildAddressBookEntrySearchSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.ctrl.Logger, suite.recorder = zaprec.NewRecorder(zapcore.ErrorLevel)
	suite.ctrl.Ctrl.Logger = suite.ctrl.Logger
}

func (suite *ControllerRebuildAddressBookEntrySearchSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildAddressBookEntrySearch(timeout)
		suite.Len(suite.recorder.Records(), 1, "should have logged error")
	}()

	wait()
}

func (suite *ControllerRebuildAddressBookEntrySearchSuite) TestStoreRebuildFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RebuildAddressBookEntrySearch", timeout, suite.tx).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildAddressBookEntrySearch(timeout)
		suite.Len(suite.recorder.Records(), 1, "should have logged error")
		suite.False(suite.tx.IsCommitted, "should not have committed tx")
	}()

	wait()
}

func (suite *ControllerRebuildAddressBookEntrySearchSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RebuildAddressBookEntrySearch", timeout, suite.tx).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildAddressBookEntrySearch(timeout)
		suite.Len(suite.recorder.Records(), 0, "should not have logged error")
		suite.True(suite.tx.IsCommitted, "should have committed tx")
	}()

	wait()
}

func TestController_RebuildAddressBookEntrySearch(t *testing.T) {
	suite.Run(t, new(ControllerRebuildAddressBookEntrySearchSuite))
}

// ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite tests
// Controller.SetAddressBookEntriesWithAutoDeliveryEnabled.
type ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite struct {
	suite.Suite
	ctrl     *ControllerMock
	tx       *testutil.DBTx
	disabled []uuid.UUID
	entryIDs []uuid.UUID
}

func (suite *ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.entryIDs = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
	suite.disabled = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}

	suite.ctrl.Store.On("SetAddressBookEntriesWithAutoDeliveryEnabled", mock.Anything, suite.tx, suite.entryIDs).
		Return(suite.disabled, nil).Maybe()
	for _, entryID := range suite.disabled {
		suite.ctrl.Notifier.On("NotifyAddressBookEntryAutoDeliveryUpdated", mock.Anything, suite.tx, entryID, false).
			Return(nil).Maybe()
	}
	for _, entryID := range suite.entryIDs {
		suite.ctrl.Notifier.On("NotifyAddressBookEntryAutoDeliveryUpdated", mock.Anything, suite.tx, entryID, true).
			Return(nil).Maybe()
	}
}

func (suite *ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TeardownTest() {
	suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.AssertExpectations(suite.T())
}

func (suite *ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestBeginTxFail() {
	suite.ctrl.DB.BeginFail = true

	err := suite.ctrl.Ctrl.SetAddressBookEntriesWithAutoDeliveryEnabled(context.Background(), suite.entryIDs)
	suite.Error(err, "should fail")
}

func (suite *ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestSetEntriesEnabledFail() {
	testutil.UnsetAndOn(&suite.ctrl.Store.Mock, "SetAddressBookEntriesWithAutoDeliveryEnabled", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("sad life"))

	err := suite.ctrl.Ctrl.SetAddressBookEntriesWithAutoDeliveryEnabled(context.Background(), suite.entryIDs)
	suite.Error(err, "should fail")
}

func (suite *ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestNotifyFail1() {
	testutil.UnsetAndOn(&suite.ctrl.Notifier.Mock, "NotifyAddressBookEntryAutoDeliveryUpdated", mock.Anything, suite.tx, suite.disabled[1]).
		Return(errors.New("sad life"))

	err := suite.ctrl.Ctrl.SetAddressBookEntriesWithAutoDeliveryEnabled(context.Background(), suite.entryIDs)
	suite.NoError(err, "should not fail")
}

func (suite *ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestNotifyFail2() {
	testutil.UnsetAndOn(&suite.ctrl.Notifier.Mock, "NotifyAddressBookEntryAutoDeliveryUpdated", mock.Anything, suite.tx, suite.entryIDs[1]).
		Return(errors.New("sad life"))

	err := suite.ctrl.Ctrl.SetAddressBookEntriesWithAutoDeliveryEnabled(context.Background(), suite.entryIDs)
	suite.NoError(err, "should not fail")
}

func (suite *ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite) TestOK() {
	err := suite.ctrl.Ctrl.SetAddressBookEntriesWithAutoDeliveryEnabled(context.Background(), suite.entryIDs)
	suite.NoError(err, "should not fail")
}

func TestController_SetAddressBookEntriesWithAutoDeliveryEnabled(t *testing.T) {
	suite.Run(t, new(ControllerSetAddressBookEntriesWithAutoDeliveryEnabledSuite))
}

// ControllerIsAutoIntelDeliveryEnabledForAddressBookEntrySuite tests
// Controller.IsAutoIntelDeliveryEnabledForAddressBookEntry.
type ControllerIsAutoIntelDeliveryEnabledForAddressBookEntrySuite struct {
	suite.Suite
	ctrl    *ControllerMock
	tx      *testutil.DBTx
	entryID uuid.UUID
}

func (suite *ControllerIsAutoIntelDeliveryEnabledForAddressBookEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.entryID = testutil.NewUUIDV4()
}

func (suite *ControllerIsAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestBeginTxFail() {
	suite.ctrl.DB.BeginFail = true

	_, err := suite.ctrl.Ctrl.IsAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID)
	suite.Error(err, "should fail")
}

func (suite *ControllerIsAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestRetrieveFail() {
	suite.ctrl.Store.On("IsAutoDeliveryEnabledForAddressBookEntry", mock.Anything, mock.Anything, mock.Anything).
		Return(false, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	_, err := suite.ctrl.Ctrl.IsAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID)
	suite.Error(err, "should fail")
}

func (suite *ControllerIsAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestOKEnabled() {
	suite.ctrl.Store.On("IsAutoDeliveryEnabledForAddressBookEntry", mock.Anything, suite.tx, suite.entryID).
		Return(true, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	got, err := suite.ctrl.Ctrl.IsAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID)
	suite.Require().NoError(err, "should not fail")
	suite.True(got, "should return correct value")
}

func (suite *ControllerIsAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestOKDisabled() {
	suite.ctrl.Store.On("IsAutoDeliveryEnabledForAddressBookEntry", mock.Anything, suite.tx, suite.entryID).
		Return(false, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	got, err := suite.ctrl.Ctrl.IsAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID)
	suite.Require().NoError(err, "should not fail")
	suite.False(got, "should return correct value")
}

func TestController_IsAutoIntelDeliveryEnabledForAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerIsAutoIntelDeliveryEnabledForAddressBookEntrySuite))
}

// ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite tests
// Controller.SetAutoIntelDeliveryEnabledForAddressBookEntry.
type ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite struct {
	suite.Suite
	ctrl    *ControllerMock
	tx      *testutil.DBTx
	entryID uuid.UUID
	enabled bool
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.entryID = testutil.NewUUIDV4()
	suite.enabled = true
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestBeginTxFail() {
	suite.ctrl.DB.BeginFail = true

	err := suite.ctrl.Ctrl.SetAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID, suite.enabled)
	suite.Error(err, "should fail")
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestRetrieveFail() {
	suite.ctrl.Store.On("SetAutoDeliveryEnabledForAddressBookEntry",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	err := suite.ctrl.Ctrl.SetAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID, suite.enabled)
	suite.Error(err, "should fail")
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestOKEnabled() {
	suite.enabled = true
	suite.ctrl.Store.On("SetAutoDeliveryEnabledForAddressBookEntry",
		mock.Anything, suite.tx, suite.entryID, suite.enabled).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	err := suite.ctrl.Ctrl.SetAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID, suite.enabled)
	suite.Require().NoError(err, "should not fail")
}

func (suite *ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite) TestOKDisabled() {
	suite.enabled = false
	suite.ctrl.Store.On("SetAutoDeliveryEnabledForAddressBookEntry",
		mock.Anything, suite.tx, suite.entryID, suite.enabled).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	err := suite.ctrl.Ctrl.SetAutoIntelDeliveryEnabledForAddressBookEntry(context.Background(), suite.entryID, suite.enabled)
	suite.Require().NoError(err, "should not fail")
}

func TestController_SetAutoIntelDeliveryEnabledForAddressBookEntry(t *testing.T) {
	suite.Run(t, new(ControllerSetAutoIntelDeliveryEnabledForAddressBookEntrySuite))
}
