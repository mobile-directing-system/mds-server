package controller

import (
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
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
	"time"
)

// ControllerCreateIntelSuite tests Controller.CreateIntel.
type ControllerCreateIntelSuite struct {
	suite.Suite
	ctrl          *ControllerMock
	tx            *testutil.DBTx
	sampleCreate  store.CreateIntel
	sampleCreated store.Intel
}

func (suite *ControllerCreateIntelSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleCreate = store.CreateIntel{
		CreatedBy:        testutil.NewUUIDV4(),
		Operation:        testutil.NewUUIDV4(),
		Type:             store.IntelTypePlaintextMessage,
		Content:          json.RawMessage(`{"text":"world"}`),
		SearchText:       nulls.NewString("world"),
		Importance:       234,
		InitialDeliverTo: []uuid.UUID{}, // Empty for simpler tests.
	}
	suite.sampleCreated = store.Intel{
		ID:         testutil.NewUUIDV4(),
		CreatedAt:  time.Date(2022, 9, 1, 11, 12, 57, 0, time.UTC),
		CreatedBy:  suite.sampleCreated.CreatedBy,
		Operation:  suite.sampleCreated.Operation,
		Type:       suite.sampleCreated.Type,
		Content:    suite.sampleCreated.Content,
		SearchText: suite.sampleCreated.SearchText,
		Importance: suite.sampleCreated.Importance,
		IsValid:    suite.sampleCreated.IsValid,
	}
}

func (suite *ControllerCreateIntelSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestOperationMemberCheckFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(false, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestNoOperationMember() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(false, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return correct error code")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("CreateIntel", timeout, suite.tx, suite.sampleCreate).
		Return(store.Intel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestScheduleDeliveriesFails() {
	suite.sampleCreate.InitialDeliverTo = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("CreateIntel", timeout, suite.tx, suite.sampleCreate).
		Return(suite.sampleCreated, nil)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleCreated.ID).
		Return(store.Intel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("CreateIntel", timeout, suite.tx, suite.sampleCreate).
		Return(suite.sampleCreated, nil)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleCreated.ID).
		Return(store.Intel{}, nil)
	suite.ctrl.Notifier.On("NotifyIntelCreated", timeout, suite.tx, suite.sampleCreated).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("CreateIntel", timeout, suite.tx, suite.sampleCreate).
		Return(suite.sampleCreated, nil)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleCreated.ID).
		Return(store.Intel{}, nil)
	suite.ctrl.Notifier.On("NotifyIntelCreated", timeout, suite.tx, suite.sampleCreated).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleCreated, got)
	}()

	wait()
}

func TestController_CreateIntel(t *testing.T) {
	suite.Run(t, new(ControllerCreateIntelSuite))
}

// ControllerInvalidateIntelSuite tests Controller.InvalidateIntel.
type ControllerInvalidateIntelSuite struct {
	suite.Suite
	ctrl        *ControllerMock
	tx          *testutil.DBTx
	sampleIntel store.Intel
	sampleBy    uuid.UUID
}

func (suite *ControllerInvalidateIntelSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleIntel = store.Intel{
		ID:         testutil.NewUUIDV4(),
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       "everyone",
		Content:    json.RawMessage(`null`),
		SearchText: nulls.NewString("gold"),
	}
	suite.sampleBy = testutil.NewUUIDV4()
}

func (suite *ControllerInvalidateIntelSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.InvalidateIntelByID(timeout, suite.sampleIntel.ID, suite.sampleBy)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerInvalidateIntelSuite) TestIntelRetrievalFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntel.ID).
		Return(store.Intel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.InvalidateIntelByID(timeout, suite.sampleIntel.ID, suite.sampleBy)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerInvalidateIntelSuite) TestOperationMemberCheckFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntel.ID).
		Return(suite.sampleIntel, nil)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleBy, suite.sampleIntel.Operation).
		Return(false, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.InvalidateIntelByID(timeout, suite.sampleIntel.ID, suite.sampleBy)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerInvalidateIntelSuite) TestNoOperationMember() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntel.ID).
		Return(suite.sampleIntel, nil)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleBy, suite.sampleIntel.Operation).
		Return(false, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.InvalidateIntelByID(timeout, suite.sampleIntel.ID, suite.sampleBy)
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return correct error code")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerInvalidateIntelSuite) TestInvalidateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntel.ID).
		Return(suite.sampleIntel, nil)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleBy, suite.sampleIntel.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("InvalidateIntelByID", timeout, suite.tx, suite.sampleIntel.ID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.InvalidateIntelByID(timeout, suite.sampleIntel.ID, suite.sampleBy)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerInvalidateIntelSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntel.ID).
		Return(suite.sampleIntel, nil)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleBy, suite.sampleIntel.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("InvalidateIntelByID", timeout, suite.tx, suite.sampleIntel.ID).
		Return(nil)
	suite.ctrl.Notifier.On("NotifyIntelInvalidated", timeout, suite.tx, suite.sampleIntel.ID, suite.sampleBy).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.InvalidateIntelByID(timeout, suite.sampleIntel.ID, suite.sampleBy)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerInvalidateIntelSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleIntel.ID).
		Return(suite.sampleIntel, nil)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleBy, suite.sampleIntel.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("InvalidateIntelByID", timeout, suite.tx, suite.sampleIntel.ID).
		Return(nil)
	suite.ctrl.Notifier.On("NotifyIntelInvalidated", timeout, suite.tx, suite.sampleIntel.ID, suite.sampleBy).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.InvalidateIntelByID(timeout, suite.sampleIntel.ID, suite.sampleBy)
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_InvalidateIntel(t *testing.T) {
	suite.Run(t, new(ControllerInvalidateIntelSuite))
}

// ControllerRebuildIntelSearchSuite tests Controller.RebuildIntelSearch.
type ControllerRebuildIntelSearchSuite struct {
	suite.Suite
	ctrl     *ControllerMock
	tx       *testutil.DBTx
	recorder *zaprec.RecordStore
}

func (suite *ControllerRebuildIntelSearchSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.ctrl.Logger, suite.recorder = zaprec.NewRecorder(zapcore.ErrorLevel)
	suite.ctrl.Ctrl.Logger = suite.ctrl.Logger
}

func (suite *ControllerRebuildIntelSearchSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildIntelSearch(timeout)
		suite.Len(suite.recorder.Records(), 1, "should have logged error")
	}()

	wait()
}

func (suite *ControllerRebuildIntelSearchSuite) TestStoreRebuildFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RebuildIntelSearch", timeout, suite.tx).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildIntelSearch(timeout)
		suite.Len(suite.recorder.Records(), 1, "should have logged error")
		suite.False(suite.tx.IsCommitted, "should not have committed tx")
	}()

	wait()
}

func (suite *ControllerRebuildIntelSearchSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("RebuildIntelSearch", timeout, suite.tx).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildIntelSearch(timeout)
		suite.Len(suite.recorder.Records(), 0, "should not have logged error")
		suite.True(suite.tx.IsCommitted, "should have committed tx")
	}()

	wait()
}

func TestController_RebuildIntelSearch(t *testing.T) {
	suite.Run(t, new(ControllerRebuildIntelSearchSuite))
}

// ControllerIntelByIDSuite tests Controller.IntelByID.
type ControllerIntelByIDSuite struct {
	suite.Suite
	ctrl                      *ControllerMock
	tx                        *testutil.DBTx
	sampleID                  uuid.UUID
	sampleUsersWithDeliveries []uuid.UUID
	sampleIntel               store.Intel
}

func (suite *ControllerIntelByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleID = testutil.NewUUIDV4()
	suite.sampleIntel = store.Intel{
		ID:         suite.sampleID,
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       "everyone",
		Content:    json.RawMessage(`null`),
		SearchText: nulls.NewString("gold"),
	}
	suite.sampleUsersWithDeliveries = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
}

func (suite *ControllerIntelByIDSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.IntelByID(timeout, suite.sampleID, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerIntelByIDSuite) TestRetrieveAssociatedUsersFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UsersWithDeliveriesByIntel", timeout, suite.tx, suite.sampleID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.IntelByID(timeout, suite.sampleID, nulls.NewUUID(testutil.NewUUIDV4()))
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerIntelByIDSuite) TestNonAssociatedUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UsersWithDeliveriesByIntel", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleUsersWithDeliveries, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.IntelByID(timeout, suite.sampleID, nulls.NewUUID(testutil.NewUUIDV4()))
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerIntelByIDSuite) TestRetrieveIntelFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleID).
		Return(store.Intel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.IntelByID(timeout, suite.sampleID, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerIntelByIDSuite) TestOKWithoutUserLimit() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleIntel, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.IntelByID(timeout, suite.sampleID, uuid.NullUUID{})
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.tx.IsCommitted, "should commit tx")
		suite.Equal(suite.sampleIntel, got, "should return correct value")
	}()

	wait()
}

func (suite *ControllerIntelByIDSuite) TestOKWithUserLimit() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UsersWithDeliveriesByIntel", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleUsersWithDeliveries, nil)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleIntel, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.IntelByID(timeout, suite.sampleID, nulls.NewUUID(suite.sampleUsersWithDeliveries[2]))
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.tx.IsCommitted, "should commit tx")
		suite.Equal(suite.sampleIntel, got, "should return correct value")
	}()

	wait()
}

func TestController_IntelByID(t *testing.T) {
	suite.Run(t, new(ControllerIntelByIDSuite))
}

// ControllerIntelSuite tests Controller.Intel.
type ControllerIntelSuite struct {
	suite.Suite
	ctrl          *ControllerMock
	tx            *testutil.DBTx
	sampleFilters store.IntelFilters
	sampleParams  pagination.Params
	sampleIntel   pagination.Paginated[store.Intel]
}

func (suite *ControllerIntelSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleFilters = store.IntelFilters{
		CreatedBy:      nulls.NewUUID(testutil.NewUUIDV4()),
		Operation:      nulls.NewUUID(testutil.NewUUIDV4()),
		IntelType:      nulls.NewJSONNullable(store.IntelTypeAnalogRadioMessage),
		MinImportance:  nulls.NewInt(58),
		IncludeInvalid: nulls.NewBool(true),
		OneOfDeliveryForEntries: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
		OneOfDeliveredToEntries: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	suite.sampleParams = pagination.Params{
		Limit:          724,
		Offset:         849,
		OrderBy:        nulls.String{},
		OrderDirection: "",
	}
	suite.sampleIntel = pagination.NewPaginated(suite.sampleParams, []store.Intel{
		{ID: testutil.NewUUIDV4()},
		{ID: testutil.NewUUIDV4()},
		{ID: testutil.NewUUIDV4()},
	}, 943)
}

func (suite *ControllerIntelSuite) TestBeginTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Intel(timeout, suite.sampleFilters, suite.sampleParams, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerIntelSuite) TestRetrieveIntelFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("Intel", timeout, suite.tx, suite.sampleFilters, suite.sampleParams).
		Return(pagination.Paginated[store.Intel]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Intel(timeout, suite.sampleFilters, suite.sampleParams, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerIntelSuite) TestOKWithoutLimit() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("Intel", timeout, suite.tx, suite.sampleFilters, suite.sampleParams).
		Return(suite.sampleIntel, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.Intel(timeout, suite.sampleFilters, suite.sampleParams, uuid.NullUUID{})
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleIntel, got, "should return correct value")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func (suite *ControllerIntelSuite) TestLimitIntelFiltersToUserFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx, mock.Anything, mock.Anything).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Intel(timeout, suite.sampleFilters, suite.sampleParams, nulls.NewUUID(testutil.NewUUIDV4()))
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerIntelSuite) TestOKWithLimitToUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	entryOfUser1 := testutil.NewUUIDV4()
	entryOfUser2 := testutil.NewUUIDV4()
	entryOfUser3 := testutil.NewUUIDV4()
	suite.sampleFilters.OneOfDeliveryForEntries = []uuid.UUID{
		testutil.NewUUIDV4(), // Other.
		entryOfUser1,
		testutil.NewUUIDV4(), // Other.
		entryOfUser2,
	}
	userID := testutil.NewUUIDV4()
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx,
		store.AddressBookEntryFilters{ByUser: nulls.NewUUID(userID)}, mock.Anything).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{Entries: []store.AddressBookEntryDetailed{
			{AddressBookEntry: store.AddressBookEntry{ID: entryOfUser1}},
			{AddressBookEntry: store.AddressBookEntry{ID: testutil.NewUUIDV4()}},
			{AddressBookEntry: store.AddressBookEntry{ID: entryOfUser2}},
			{AddressBookEntry: store.AddressBookEntry{ID: entryOfUser3}},
		}}, nil)
	suite.ctrl.Store.On("Intel", timeout, suite.tx, store.IntelFilters{
		CreatedBy:      suite.sampleFilters.CreatedBy,
		Operation:      suite.sampleFilters.Operation,
		IntelType:      suite.sampleFilters.IntelType,
		MinImportance:  suite.sampleFilters.MinImportance,
		IncludeInvalid: suite.sampleFilters.IncludeInvalid,
		OneOfDeliveryForEntries: []uuid.UUID{
			entryOfUser1,
			entryOfUser2,
		},
		OneOfDeliveredToEntries: suite.sampleFilters.OneOfDeliveredToEntries,
	}, suite.sampleParams).Return(suite.sampleIntel, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.Intel(timeout, suite.sampleFilters, suite.sampleParams, nulls.NewUUID(userID))
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleIntel, got, "should return correct value")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func (suite *ControllerIntelSuite) TestOKWithOrderBy() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleParams.OrderBy = nulls.NewString("soon")
	suite.ctrl.Store.On("Intel", timeout, suite.tx, suite.sampleFilters, pagination.Params{
		Limit:          suite.sampleParams.Limit,
		Offset:         suite.sampleParams.Offset,
		OrderBy:        nulls.String{},
		OrderDirection: suite.sampleParams.OrderDirection,
	}).Return(suite.sampleIntel, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.Intel(timeout, suite.sampleFilters, suite.sampleParams, uuid.NullUUID{})
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleIntel, got, "should return correct value")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func (suite *ControllerIntelSuite) TestNoEntriesForLimitUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	userID := testutil.NewUUIDV4()
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx,
		store.AddressBookEntryFilters{ByUser: nulls.NewUUID(userID)}, mock.Anything).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{Entries: []store.AddressBookEntryDetailed{}}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.Intel(timeout, suite.sampleFilters, suite.sampleParams, nulls.NewUUID(userID))
		suite.Require().NoError(err, "should not fail")
		suite.Empty(got.Entries, "should return correct value")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_Intel(t *testing.T) {
	suite.Run(t, new(ControllerIntelSuite))
}

// ControllerSearchIntelSuite tests Controller.Intel.
type ControllerSearchIntelSuite struct {
	suite.Suite
	ctrl          *ControllerMock
	tx            *testutil.DBTx
	sampleFilters store.IntelFilters
	sampleParams  search.Params
	sampleIntel   search.Result[store.Intel]
}

func (suite *ControllerSearchIntelSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{suite.tx}
	suite.sampleFilters = store.IntelFilters{
		CreatedBy:      nulls.NewUUID(testutil.NewUUIDV4()),
		Operation:      nulls.NewUUID(testutil.NewUUIDV4()),
		IntelType:      nulls.NewJSONNullable(store.IntelTypeAnalogRadioMessage),
		MinImportance:  nulls.NewInt(58),
		IncludeInvalid: nulls.NewBool(true),
		OneOfDeliveryForEntries: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
		OneOfDeliveredToEntries: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	suite.sampleParams = search.Params{
		Query:  "sentence",
		Offset: 602,
		Limit:  397,
	}
	suite.sampleIntel = search.Result[store.Intel]{
		Hits: []store.Intel{
			{ID: testutil.NewUUIDV4()},
			{ID: testutil.NewUUIDV4()},
			{ID: testutil.NewUUIDV4()},
		},
	}
}

func (suite *ControllerSearchIntelSuite) TestBeginTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.SearchIntel(timeout, suite.sampleFilters, suite.sampleParams, uuid.NullUUID{})
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerSearchIntelSuite) TestSearchFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("SearchIntel", timeout, suite.tx, suite.sampleFilters, suite.sampleParams).
		Return(search.Result[store.Intel]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.SearchIntel(timeout, suite.sampleFilters, suite.sampleParams, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerSearchIntelSuite) TestOKWithoutLimit() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("SearchIntel", timeout, suite.tx, suite.sampleFilters, suite.sampleParams).
		Return(suite.sampleIntel, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.SearchIntel(timeout, suite.sampleFilters, suite.sampleParams, uuid.NullUUID{})
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleIntel, got, "should return correct value")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func (suite *ControllerSearchIntelSuite) TestLimitIntelFiltersToUserFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx, mock.Anything, mock.Anything).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.SearchIntel(timeout, suite.sampleFilters, suite.sampleParams, nulls.NewUUID(testutil.NewUUIDV4()))
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerSearchIntelSuite) TestOKWithLimitToUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	entryOfUser1 := testutil.NewUUIDV4()
	entryOfUser2 := testutil.NewUUIDV4()
	entryOfUser3 := testutil.NewUUIDV4()
	suite.sampleFilters.OneOfDeliveryForEntries = []uuid.UUID{
		testutil.NewUUIDV4(), // Other.
		entryOfUser1,
		testutil.NewUUIDV4(), // Other.
		entryOfUser2,
	}
	userID := testutil.NewUUIDV4()
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx,
		store.AddressBookEntryFilters{ByUser: nulls.NewUUID(userID)}, mock.Anything).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{Entries: []store.AddressBookEntryDetailed{
			{AddressBookEntry: store.AddressBookEntry{ID: entryOfUser1}},
			{AddressBookEntry: store.AddressBookEntry{ID: testutil.NewUUIDV4()}},
			{AddressBookEntry: store.AddressBookEntry{ID: entryOfUser2}},
			{AddressBookEntry: store.AddressBookEntry{ID: entryOfUser3}},
		}}, nil)
	suite.ctrl.Store.On("SearchIntel", timeout, suite.tx, store.IntelFilters{
		CreatedBy:      suite.sampleFilters.CreatedBy,
		Operation:      suite.sampleFilters.Operation,
		IntelType:      suite.sampleFilters.IntelType,
		MinImportance:  suite.sampleFilters.MinImportance,
		IncludeInvalid: suite.sampleFilters.IncludeInvalid,
		OneOfDeliveryForEntries: []uuid.UUID{
			entryOfUser1,
			entryOfUser2,
		},
		OneOfDeliveredToEntries: suite.sampleFilters.OneOfDeliveredToEntries,
	}, suite.sampleParams).Return(suite.sampleIntel, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.SearchIntel(timeout, suite.sampleFilters, suite.sampleParams, nulls.NewUUID(userID))
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleIntel, got, "should return correct value")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func (suite *ControllerSearchIntelSuite) TestNoEntriesForLimitUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	userID := testutil.NewUUIDV4()
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx,
		store.AddressBookEntryFilters{ByUser: nulls.NewUUID(userID)}, mock.Anything).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{Entries: []store.AddressBookEntryDetailed{}}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.SearchIntel(timeout, suite.sampleFilters, suite.sampleParams, nulls.NewUUID(userID))
		suite.Require().NoError(err, "should not fail")
		suite.Empty(got.Hits, "should return correct value")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_SearchIntel(t *testing.T) {
	suite.Run(t, new(ControllerSearchIntelSuite))
}

// controllerLimitIntelFiltersToUserSuite tests
// Controller.limitIntelFiltersToUser.
type controllerLimitIntelFiltersToUserSuite struct {
	suite.Suite
	ctrl          *ControllerMock
	tx            *testutil.DBTx
	sampleFilters store.IntelFilters
	sampleUserID  uuid.UUID
	userEntry1    uuid.UUID
	userEntry2    uuid.UUID
	userEntry3    uuid.UUID
}

func (suite *controllerLimitIntelFiltersToUserSuite) copyIntelFilters(old store.IntelFilters, newFor []uuid.UUID) store.IntelFilters {
	return store.IntelFilters{
		CreatedBy:               old.CreatedBy,
		Operation:               old.Operation,
		IntelType:               old.IntelType,
		MinImportance:           old.MinImportance,
		IncludeInvalid:          old.IncludeInvalid,
		OneOfDeliveryForEntries: newFor,
		OneOfDeliveredToEntries: old.OneOfDeliveredToEntries,
	}
}

func (suite *controllerLimitIntelFiltersToUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.tx = &testutil.DBTx{}
	suite.sampleUserID = testutil.NewUUIDV4()
	suite.userEntry1 = testutil.NewUUIDV4()
	suite.userEntry2 = testutil.NewUUIDV4()
	suite.userEntry3 = testutil.NewUUIDV4()
	suite.sampleFilters = store.IntelFilters{
		CreatedBy:      nulls.NewUUID(testutil.NewUUIDV4()),
		Operation:      nulls.NewUUID(testutil.NewUUIDV4()),
		IntelType:      nulls.NewJSONNullable(store.IntelTypePlaintextMessage),
		MinImportance:  nulls.NewInt(129),
		IncludeInvalid: nulls.NewBool(true),
		OneOfDeliveryForEntries: []uuid.UUID{
			testutil.NewUUIDV4(),
			suite.userEntry1,
			testutil.NewUUIDV4(),
			suite.userEntry2,
			testutil.NewUUIDV4(),
		},
		OneOfDeliveredToEntries: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
}

func (suite *controllerLimitIntelFiltersToUserSuite) TestRetrieveEntriesFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx,
		store.AddressBookEntryFilters{ByUser: nulls.NewUUID(suite.sampleUserID)}, pagination.Params{Limit: 0}).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, err := suite.ctrl.Ctrl.limitIntelFiltersToUser(timeout, suite.tx, suite.sampleFilters, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *controllerLimitIntelFiltersToUserSuite) TestNoEntriesForUser() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx,
		store.AddressBookEntryFilters{ByUser: nulls.NewUUID(suite.sampleUserID)}, pagination.Params{Limit: 0}).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, ok, err := suite.ctrl.Ctrl.limitIntelFiltersToUser(timeout, suite.tx, suite.sampleFilters, suite.sampleUserID)
		suite.Require().NoError(err, "should not fail")
		suite.False(ok, "should return correct value")
	}()

	wait()
}

func (suite *controllerLimitIntelFiltersToUserSuite) TestRemoveForeign() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx,
		store.AddressBookEntryFilters{ByUser: nulls.NewUUID(suite.sampleUserID)}, pagination.Params{Limit: 0}).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{Entries: []store.AddressBookEntryDetailed{
			{AddressBookEntry: store.AddressBookEntry{ID: testutil.NewUUIDV4()}},
			{AddressBookEntry: store.AddressBookEntry{ID: suite.userEntry1}},
			{AddressBookEntry: store.AddressBookEntry{ID: testutil.NewUUIDV4()}},
			{AddressBookEntry: store.AddressBookEntry{ID: suite.userEntry2}},
			{AddressBookEntry: store.AddressBookEntry{ID: suite.userEntry3}},
			{AddressBookEntry: store.AddressBookEntry{ID: testutil.NewUUIDV4()}},
		}}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, ok, err := suite.ctrl.Ctrl.limitIntelFiltersToUser(timeout, suite.tx, suite.sampleFilters, suite.sampleUserID)
		suite.Require().NoError(err, "should not fail")
		suite.True(ok, "should return correct value")
		suite.Equal(suite.copyIntelFilters(suite.sampleFilters, []uuid.UUID{
			suite.userEntry1,
			suite.userEntry2,
		}), got, "should return correct filters")
	}()

	wait()
}

func (suite *controllerLimitIntelFiltersToUserSuite) TestNoneLeftAfterRemoval() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleFilters.OneOfDeliveryForEntries = []uuid.UUID{
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
		testutil.NewUUIDV4(),
	}
	suite.ctrl.Store.On("AddressBookEntries", timeout, suite.tx,
		store.AddressBookEntryFilters{ByUser: nulls.NewUUID(suite.sampleUserID)}, pagination.Params{Limit: 0}).
		Return(pagination.Paginated[store.AddressBookEntryDetailed]{Entries: []store.AddressBookEntryDetailed{
			{AddressBookEntry: store.AddressBookEntry{ID: suite.userEntry1}},
			{AddressBookEntry: store.AddressBookEntry{ID: suite.userEntry2}},
			{AddressBookEntry: store.AddressBookEntry{ID: suite.userEntry3}},
		}}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, ok, err := suite.ctrl.Ctrl.limitIntelFiltersToUser(timeout, suite.tx, suite.sampleFilters, suite.sampleUserID)
		suite.Require().NoError(err, "should not fail")
		suite.True(ok, "should return correct value")
		suite.Equal(suite.copyIntelFilters(suite.sampleFilters, []uuid.UUID{
			suite.userEntry1,
			suite.userEntry2,
			suite.userEntry3,
		}), got, "should return correct filters")
	}()

	wait()
}

func TestController_limitIntelFiltersToUser(t *testing.T) {
	suite.Run(t, new(controllerLimitIntelFiltersToUserSuite))
}
