package controller

import (
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       store.IntelTypePlaintextMessage,
		Content:    json.RawMessage(`{"text":"world"}`),
		SearchText: nulls.NewString("world"),
		Assignments: []store.IntelAssignment{
			{
				ID:    uuid.UUID{},
				Intel: uuid.UUID{},
				To:    testutil.NewUUIDV4(),
			},
			{
				ID:    uuid.UUID{},
				Intel: uuid.UUID{},
				To:    testutil.NewUUIDV4(),
			},
		},
	}
	suite.sampleCreated = store.Intel{
		ID:          testutil.NewUUIDV4(),
		CreatedAt:   time.Now().UTC(),
		CreatedBy:   suite.sampleCreate.CreatedBy,
		Operation:   suite.sampleCreate.Operation,
		Type:        suite.sampleCreate.Type,
		Content:     suite.sampleCreate.Content,
		SearchText:  suite.sampleCreate.SearchText,
		IsValid:     true,
		Assignments: suite.sampleCreate.Assignments,
	}
	suite.sampleCreated.Assignments[0].ID = testutil.NewUUIDV4()
	suite.sampleCreated.Assignments[0].Intel = suite.sampleCreated.ID
	suite.sampleCreated.Assignments[1].ID = testutil.NewUUIDV4()
	suite.sampleCreated.Assignments[1].Intel = suite.sampleCreated.ID
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

func (suite *ControllerCreateIntelSuite) TestCheckAssignmentToExistsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, mock.Anything).
		Return(store.AddressBookEntry{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestAssignmentToDoesNotExist() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, mock.Anything).
		Return(store.AddressBookEntry{}, meh.NewNotFoundErr("", nil))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrBadInput, meh.ErrorCode(err), "should return correct error code")
		suite.False(suite.tx.IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateIntelSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, mock.Anything).
		Return(store.AddressBookEntry{}, nil)
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

func (suite *ControllerCreateIntelSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IsUserOperationMember", timeout, suite.tx, suite.sampleCreate.CreatedBy, suite.sampleCreate.Operation).
		Return(true, nil)
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, mock.Anything).
		Return(store.AddressBookEntry{}, nil)
	suite.ctrl.Store.On("CreateIntel", timeout, suite.tx, suite.sampleCreate).
		Return(suite.sampleCreated, nil)
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
	suite.ctrl.Store.On("AddressBookEntryByID", timeout, suite.tx, mock.Anything).
		Return(store.AddressBookEntry{}, nil)
	suite.ctrl.Store.On("CreateIntel", timeout, suite.tx, suite.sampleCreate).
		Return(suite.sampleCreated, nil)
	suite.ctrl.Notifier.On("NotifyIntelCreated", timeout, suite.tx, suite.sampleCreated).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.CreateIntel(timeout, suite.sampleCreate)
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.tx.IsCommitted, "should not tx")
		suite.Equal(suite.sampleCreated, got, "should return correct value")
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
	suite.sampleIntel.Assignments = []store.IntelAssignment{
		{
			ID:    testutil.NewUUIDV4(),
			Intel: suite.sampleIntel.ID,
			To:    testutil.NewUUIDV4(),
		},
		{
			ID:    testutil.NewUUIDV4(),
			Intel: suite.sampleIntel.ID,
			To:    testutil.NewUUIDV4(),
		},
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

// ControllerIntelByIDSuite tests Controller.IntelByID.
type ControllerIntelByIDSuite struct {
	suite.Suite
	ctrl        *ControllerMock
	tx          *testutil.DBTx
	sampleID    uuid.UUID
	sampleIntel store.Intel
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
	suite.sampleIntel.Assignments = []store.IntelAssignment{
		{
			ID:    testutil.NewUUIDV4(),
			Intel: suite.sampleIntel.ID,
			To:    testutil.NewUUIDV4(),
		},
		{
			ID:    testutil.NewUUIDV4(),
			Intel: suite.sampleIntel.ID,
			To:    testutil.NewUUIDV4(),
		},
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

func (suite *ControllerIntelByIDSuite) TestRetrieveFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleID).
		Return(store.Intel{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.IntelByID(timeout, suite.sampleID, uuid.NullUUID{})
		suite.Error(err, "should fail")
		suite.False(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func (suite *ControllerIntelByIDSuite) TestNotAssignedToGivenUserLimit() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	limitTo := testutil.NewUUIDV4()
	suite.sampleIntel.Assignments = []store.IntelAssignment{
		{To: testutil.NewUUIDV4()},
		{To: testutil.NewUUIDV4()},
		{To: testutil.NewUUIDV4()},
	}
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleIntel, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.IntelByID(timeout, suite.sampleID, nulls.NewUUID(limitTo))
		suite.Require().Error(err, "should fail")
		suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return correct error code")
		suite.False(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func (suite *ControllerIntelByIDSuite) TestOK() {
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
	limitTo := testutil.NewUUIDV4()
	suite.sampleIntel.Assignments = []store.IntelAssignment{
		{To: testutil.NewUUIDV4()},
		{To: limitTo},
		{To: testutil.NewUUIDV4()},
	}
	suite.ctrl.Store.On("IntelByID", timeout, suite.tx, suite.sampleID).
		Return(suite.sampleIntel, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.IntelByID(timeout, suite.sampleID, nulls.NewUUID(limitTo))
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.tx.IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_IntelByID(t *testing.T) {
	suite.Run(t, new(ControllerIntelByIDSuite))
}
