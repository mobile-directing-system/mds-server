package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/lefinal/zaprec"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
	"testing"
	"time"
)

// ControllerOperationByIDSuite tests Controller.OperationByID.
type ControllerOperationByIDSuite struct {
	suite.Suite
	ctrl            *ControllerMock
	sampleOperation store.Operation
}

func (suite *ControllerOperationByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleOperation = store.Operation{
		ID:          testutil.NewUUIDV4(),
		Title:       "marry",
		Description: "stand",
		Start:       time.UnixMilli(716),
		End:         nulls.NewTime(time.UnixMilli(12440)),
		IsArchived:  true,
	}
}

func (suite *ControllerOperationByIDSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.OperationByID(timeout, suite.sampleOperation.ID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerOperationByIDSuite) TestRetrieveFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperation.ID).
		Return(store.Operation{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.OperationByID(timeout, suite.sampleOperation.ID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerOperationByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleOperation.ID).
		Return(suite.sampleOperation, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.OperationByID(timeout, suite.sampleOperation.ID)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleOperation, got, "should return correct value")
	}()

	wait()
}

func TestController_OperationByID(t *testing.T) {
	suite.Run(t, new(ControllerOperationByIDSuite))
}

// ControllerOperationsSuite tests Controller.Operations.
type ControllerOperationsSuite struct {
	suite.Suite
	ctrl             *ControllerMock
	sampleFilters    store.OperationRetrievalFilters
	sampleParams     pagination.Params
	sampleOperations pagination.Paginated[store.Operation]
}

func (suite *ControllerOperationsSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleFilters = store.OperationRetrievalFilters{
		OnlyOngoing:     true,
		IncludeArchived: true,
		ForUser:         nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.sampleParams = pagination.Params{
		Limit:          923,
		Offset:         209,
		OrderBy:        nulls.NewString("brick"),
		OrderDirection: "desc",
	}
	suite.sampleOperations = pagination.NewPaginated(suite.sampleParams, []store.Operation{
		{
			ID:          testutil.NewUUIDV4(),
			Title:       "sympathy",
			Description: "civilize",
			Start:       time.UnixMilli(535),
			End:         nulls.NewTime(time.UnixMilli(15234235)),
			IsArchived:  true,
		},
		{
			ID:          testutil.NewUUIDV4(),
			Title:       "regard",
			Description: "throw",
			Start:       time.UnixMilli(123),
			End:         nulls.Time{},
			IsArchived:  false,
		},
	}, 9313)
}

func (suite *ControllerOperationsSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Operations(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerOperationsSuite) TestRetrieveFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Operations", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(pagination.Paginated[store.Operation]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Operations(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerOperationsSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Operations", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(suite.sampleOperations, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.Operations(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleOperations, got, "should return correct value")
	}()

	wait()
}

func TestController_Operations(t *testing.T) {
	suite.Run(t, new(ControllerOperationsSuite))
}

// ControllerCreateOperationSuite tests Controller.CreateOperation.
type ControllerCreateOperationSuite struct {
	suite.Suite
	ctrl   *ControllerMock
	create store.Operation
}

func (suite *ControllerCreateOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.create = store.Operation{
		ID:          testutil.NewUUIDV4(),
		Title:       "moderate",
		Description: "intend",
		Start:       time.UnixMilli(824),
		End:         nulls.NewTime(time.UnixMilli(12563)),
		IsArchived:  true,
	}
}

func (suite *ControllerCreateOperationSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateOperation(timeout, suite.create)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateOperationSuite) TestStoreCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateOperation", timeout, suite.ctrl.DB.Tx[0], suite.create).
		Return(store.Operation{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateOperation(timeout, suite.create)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerCreateOperationSuite) TestNotifyOperationUpdatedFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateOperation", timeout, suite.ctrl.DB.Tx[0], suite.create).
		Return(suite.create, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyOperationCreated", timeout, suite.ctrl.DB.Tx[0], suite.create).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateOperation(timeout, suite.create)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerCreateOperationSuite) TestNotifyOperationMembersUpdatedFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateOperation", timeout, suite.ctrl.DB.Tx[0], suite.create).
		Return(suite.create, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyOperationCreated", timeout, suite.ctrl.DB.Tx[0], suite.create).
		Return(nil)
	suite.ctrl.Notifier.On("NotifyOperationMembersUpdated", timeout, suite.ctrl.DB.Tx[0], suite.create.ID, []uuid.UUID{}).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateOperation(timeout, suite.create)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerCreateOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateOperation", timeout, suite.ctrl.DB.Tx[0], suite.create).
		Return(suite.create, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyOperationCreated", timeout, suite.ctrl.DB.Tx[0], suite.create).
		Return(nil)
	suite.ctrl.Notifier.On("NotifyOperationMembersUpdated", timeout, suite.ctrl.DB.Tx[0], suite.create.ID, []uuid.UUID{}).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.CreateOperation(timeout, suite.create)
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit")
		suite.Equal(suite.create, got, "should return correct value")
	}()

	wait()
}

func TestController_CreateOperation(t *testing.T) {
	suite.Run(t, new(ControllerCreateOperationSuite))
}

// ControllerUpdateOperationSuite tests Controller.UpdateOperation.
type ControllerUpdateOperationSuite struct {
	suite.Suite
	ctrl   *ControllerMock
	update store.Operation
}

func (suite *ControllerUpdateOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.update = store.Operation{
		ID:          testutil.NewUUIDV4(),
		Title:       "moderate",
		Description: "intend",
		Start:       time.UnixMilli(824),
		End:         nulls.NewTime(time.UnixMilli(12563)),
		IsArchived:  true,
	}
}

func (suite *ControllerUpdateOperationSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperation(timeout, suite.update)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationSuite) TestStoreUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateOperation", timeout, suite.ctrl.DB.Tx[0], suite.update).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperation(timeout, suite.update)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerUpdateOperationSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateOperation", timeout, suite.ctrl.DB.Tx[0], suite.update).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyOperationUpdated", timeout, suite.ctrl.DB.Tx[0], suite.update).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperation(timeout, suite.update)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit")
	}()

	wait()
}

func (suite *ControllerUpdateOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateOperation", timeout, suite.ctrl.DB.Tx[0], suite.update).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyOperationUpdated", timeout, suite.ctrl.DB.Tx[0], suite.update).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperation(timeout, suite.update)
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit")
	}()

	wait()
}

func TestController_UpdateOperation(t *testing.T) {
	suite.Run(t, new(ControllerUpdateOperationSuite))
}

// ControllerSearchOperationsSuite tests Controller.SearchOperations.
type ControllerSearchOperationsSuite struct {
	suite.Suite
	ctrl             *ControllerMock
	sampleFilters    store.OperationRetrievalFilters
	sampleParams     search.Params
	sampleOperations []store.Operation
}

func (suite *ControllerSearchOperationsSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleFilters = store.OperationRetrievalFilters{
		OnlyOngoing:     true,
		IncludeArchived: true,
		ForUser:         nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.sampleParams = search.Params{
		Query:  "freedom",
		Offset: 734,
		Limit:  389,
	}
	suite.sampleOperations = []store.Operation{
		{
			ID:          testutil.NewUUIDV4(),
			Title:       "rapid",
			Description: "throw",
			Start:       time.Date(2022, 8, 20, 19, 10, 25, 0, time.UTC),
			End:         nulls.Time{},
			IsArchived:  false,
		},
		{
			ID:          testutil.NewUUIDV4(),
			Title:       "profit",
			Description: "deceit",
			Start:       time.Date(2022, 8, 20, 19, 10, 43, 0, time.UTC),
			End:         nulls.NewTime(time.Date(2022, 8, 19, 20, 10, 0, 0, time.UTC)),
			IsArchived:  true,
		},
	}
}

func (suite *ControllerSearchOperationsSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.SearchOperations(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerSearchOperationsSuite) TestStoreSearchFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("SearchOperations", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(search.Result[store.Operation]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.SearchOperations(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerSearchOperationsSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("SearchOperations", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(search.Result[store.Operation]{Hits: suite.sampleOperations}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.SearchOperations(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleOperations, got.Hits, "should return correct result")
	}()

	wait()
}

func TestController_SearchOperations(t *testing.T) {
	suite.Run(t, new(ControllerSearchOperationsSuite))
}

// ControllerRebuildOperationSearchSuite tests Controller.RebuildOperationSearch.
type ControllerRebuildOperationSearchSuite struct {
	suite.Suite
	ctrl     *ControllerMock
	recorder *zaprec.RecordStore
}

func (suite *ControllerRebuildOperationSearchSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.ctrl.Logger, suite.recorder = zaprec.NewRecorder(zapcore.ErrorLevel)
	suite.ctrl.Ctrl.Logger = suite.ctrl.Logger
}

func (suite *ControllerRebuildOperationSearchSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildOperationSearch(timeout)
		suite.Len(suite.recorder.Records(), 1, "should have logged error")
	}()

	wait()
}

func (suite *ControllerRebuildOperationSearchSuite) TestStoreRebuildFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("RebuildOperationSearch", timeout, suite.ctrl.DB.Tx[0]).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildOperationSearch(timeout)
		suite.Len(suite.recorder.Records(), 1, "should have logged error")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	wait()
}

func (suite *ControllerRebuildOperationSearchSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("RebuildOperationSearch", timeout, suite.ctrl.DB.Tx[0]).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildOperationSearch(timeout)
		suite.Len(suite.recorder.Records(), 0, "should not have logged error")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	wait()
}

func TestController_RebuildOperationSearch(t *testing.T) {
	suite.Run(t, new(ControllerRebuildOperationSearchSuite))
}
