package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
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
