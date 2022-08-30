package eventport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// HandlerMock mocks Handler.
type HandlerMock struct {
	mock.Mock
}

func (m *HandlerMock) CreateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *HandlerMock) UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *HandlerMock) CreateOperation(ctx context.Context, tx pgx.Tx, create store.Operation) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *HandlerMock) UpdateOperation(ctx context.Context, tx pgx.Tx, update store.Operation) error {
	return m.Called(ctx, tx, update).Error(0)
}

func (m *HandlerMock) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error {
	return m.Called(ctx, tx, operationID, newMembers).Error(0)
}

func (m *HandlerMock) CreateAddressBookEntry(ctx context.Context, tx pgx.Tx, create store.AddressBookEntry) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *HandlerMock) UpdateAddressBookEntry(ctx context.Context, tx pgx.Tx, update store.AddressBookEntry) error {
	return m.Called(ctx, tx, update).Error(0)
}

func (m *HandlerMock) DeleteAddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	return m.Called(ctx, tx, entryID).Error(0)
}

// portHandleUserCreatedSuite tests Port.handleUserCreated.
type portHandleUserCreatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.UserCreated
	sampleCreate store.User
}

func (suite *portHandleUserCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.UserCreated{
		ID:        testutil.NewUUIDV4(),
		Username:  "small",
		FirstName: "pleasant",
		LastName:  "glad",
		IsAdmin:   true,
		Pass:      nil,
		IsActive:  true,
	}
	suite.sampleCreate = store.User{
		ID:        suite.sampleEvent.ID,
		Username:  suite.sampleEvent.Username,
		FirstName: suite.sampleEvent.FirstName,
		LastName:  suite.sampleEvent.LastName,
		IsActive:  true,
	}
}

func (suite *portHandleUserCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleUserCreatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserCreatedSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateUser", timeout, tx, suite.sampleCreate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateUser", timeout, tx, suite.sampleCreate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleUserCreated(t *testing.T) {
	suite.Run(t, new(portHandleUserCreatedSuite))
}

// portHandleUserUpdatedSuite tests Port.handleUserUpdated.
type portHandleUserUpdatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.UserUpdated
	sampleUpdate store.User
}

func (suite *portHandleUserUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.UserUpdated{
		ID:        testutil.NewUUIDV4(),
		Username:  "small",
		FirstName: "pleasant",
		LastName:  "glad",
		IsAdmin:   true,
	}
	suite.sampleUpdate = store.User{
		ID:        suite.sampleEvent.ID,
		Username:  suite.sampleEvent.Username,
		FirstName: suite.sampleEvent.FirstName,
		LastName:  suite.sampleEvent.LastName,
	}
}

func (suite *portHandleUserUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleUserUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateUser", timeout, tx, suite.sampleUpdate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateUser", timeout, tx, suite.sampleUpdate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleUserUpdated(t *testing.T) {
	suite.Run(t, new(portHandleUserUpdatedSuite))
}

// portHandleAddressBookEntryCreatedSuite tests Port.handleAddressBookEntryCreated.
type portHandleAddressBookEntryCreatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.AddressBookEntryCreated
	sampleCreate store.AddressBookEntry
}

func (suite *portHandleAddressBookEntryCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	members := make([]uuid.UUID, 16)
	for i := range members {
		members[i] = testutil.NewUUIDV4()
	}
	suite.sampleEvent = event.AddressBookEntryCreated{
		ID:          testutil.NewUUIDV4(),
		Label:       "meow",
		Description: "woof",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.sampleCreate = store.AddressBookEntry{
		ID:          suite.sampleEvent.ID,
		Label:       suite.sampleEvent.Label,
		Description: suite.sampleEvent.Description,
		Operation:   suite.sampleEvent.Operation,
		User:        suite.sampleEvent.User,
	}
}

func (suite *portHandleAddressBookEntryCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.AddressBookTopic,
		EventType: event.TypeAddressBookEntryCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleAddressBookEntryCreatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryCreatedSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateAddressBookEntry", timeout, tx, suite.sampleCreate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateAddressBookEntry", timeout, tx, suite.sampleCreate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleAddressBookEntryCreated(t *testing.T) {
	suite.Run(t, new(portHandleAddressBookEntryCreatedSuite))
}

// portHandleAddressBookEntryUpdatedSuite tests Port.handleAddressBookEntryUpdated.
type portHandleAddressBookEntryUpdatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.AddressBookEntryUpdated
	sampleUpdate store.AddressBookEntry
}

func (suite *portHandleAddressBookEntryUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	members := make([]uuid.UUID, 16)
	for i := range members {
		members[i] = testutil.NewUUIDV4()
	}
	suite.sampleEvent = event.AddressBookEntryUpdated{
		ID:          testutil.NewUUIDV4(),
		Label:       "match",
		Description: "game",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		User:        nulls.NewUUID(testutil.NewUUIDV4()),
	}
	suite.sampleUpdate = store.AddressBookEntry{
		ID:          suite.sampleEvent.ID,
		Label:       suite.sampleEvent.Label,
		Description: suite.sampleEvent.Description,
		Operation:   suite.sampleEvent.Operation,
		User:        suite.sampleEvent.User,
	}
}

func (suite *portHandleAddressBookEntryUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.AddressBookTopic,
		EventType: event.TypeAddressBookEntryUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleAddressBookEntryUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateAddressBookEntry", timeout, tx, suite.sampleUpdate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateAddressBookEntry", timeout, tx, suite.sampleUpdate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleAddressBookEntryUpdated(t *testing.T) {
	suite.Run(t, new(portHandleAddressBookEntryUpdatedSuite))
}

// portHandleAddressBookEntryDeletedSuite tests Port.handleAddressBookEntryDeleted.
type portHandleAddressBookEntryDeletedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.AddressBookEntryDeleted
}

func (suite *portHandleAddressBookEntryDeletedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.AddressBookEntryDeleted{
		ID: testutil.NewUUIDV4(),
	}
}

func (suite *portHandleAddressBookEntryDeletedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.AddressBookTopic,
		EventType: event.TypeAddressBookEntryDeleted,
		RawValue:  rawValue,
	})
}

func (suite *portHandleAddressBookEntryDeletedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryDeletedSuite) TestDeleteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteAddressBookEntryByID", timeout, tx, suite.sampleEvent.ID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryDeletedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteAddressBookEntryByID", timeout, tx, suite.sampleEvent.ID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleAddressBookEntryDeleted(t *testing.T) {
	suite.Run(t, new(portHandleAddressBookEntryDeletedSuite))
}

// portHandleOperationCreatedSuite tests Port.handleOperationCreated.
type portHandleOperationCreatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.OperationCreated
	sampleCreate store.Operation
}

func (suite *portHandleOperationCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.OperationCreated{
		ID:          testutil.NewUUIDV4(),
		Title:       "disagree",
		Description: "set",
		Start:       time.Date(2022, 8, 04, 0, 12, 0, 0, time.UTC),
		End:         nulls.NewTime(time.Date(2022, 9, 10, 3, 1, 0, 0, time.UTC)),
		IsArchived:  true,
	}
	suite.sampleCreate = store.Operation{
		ID:          suite.sampleEvent.ID,
		Title:       suite.sampleEvent.Title,
		Description: suite.sampleEvent.Description,
		Start:       suite.sampleEvent.Start,
		End:         suite.sampleEvent.End,
		IsArchived:  suite.sampleEvent.IsArchived,
	}
}

func (suite *portHandleOperationCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.OperationsTopic,
		EventType: event.TypeOperationCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleOperationCreatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleOperationCreatedSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateOperation", timeout, tx, suite.sampleCreate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleOperationCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateOperation", timeout, tx, suite.sampleCreate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleOperationCreated(t *testing.T) {
	suite.Run(t, new(portHandleOperationCreatedSuite))
}

// portHandleOperationUpdatedSuite tests Port.handleOperationUpdated.
type portHandleOperationUpdatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.OperationUpdated
	sampleUpdate store.Operation
}

func (suite *portHandleOperationUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.OperationUpdated{
		ID:          testutil.NewUUIDV4(),
		Title:       "disagree",
		Description: "set",
		Start:       time.Date(2022, 8, 04, 0, 12, 0, 0, time.UTC),
		End:         nulls.NewTime(time.Date(2022, 9, 10, 3, 1, 0, 0, time.UTC)),
		IsArchived:  true,
	}
	suite.sampleUpdate = store.Operation{
		ID:          suite.sampleEvent.ID,
		Title:       suite.sampleEvent.Title,
		Description: suite.sampleEvent.Description,
		Start:       suite.sampleEvent.Start,
		End:         suite.sampleEvent.End,
		IsArchived:  suite.sampleEvent.IsArchived,
	}
}

func (suite *portHandleOperationUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.OperationsTopic,
		EventType: event.TypeOperationUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleOperationUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleOperationUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateOperation", timeout, tx, suite.sampleUpdate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleOperationUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateOperation", timeout, tx, suite.sampleUpdate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleOperationUpdated(t *testing.T) {
	suite.Run(t, new(portHandleOperationUpdatedSuite))
}

// portHandleOperationMembersUpdatedSuite tests Port.handleOperationMembersUpdated.
type portHandleOperationMembersUpdatedSuite struct {
	suite.Suite
	handler          *HandlerMock
	port             *PortMock
	sampleEvent      event.OperationMembersUpdated
	sampleOperation  uuid.UUID
	sampleNewMembers []uuid.UUID
}

func (suite *portHandleOperationMembersUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.OperationMembersUpdated{
		Operation: testutil.NewUUIDV4(),
		Members: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	suite.sampleOperation = suite.sampleEvent.Operation
	suite.sampleNewMembers = suite.sampleEvent.Members
}

func (suite *portHandleOperationMembersUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.OperationsTopic,
		EventType: event.TypeOperationMembersUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleOperationMembersUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleOperationMembersUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateOperationMembersByOperation", timeout, tx, suite.sampleOperation, suite.sampleNewMembers).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleOperationMembersUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateOperationMembersByOperation", timeout, tx, suite.sampleOperation, suite.sampleNewMembers).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleOperationMembersUpdated(t *testing.T) {
	suite.Run(t, new(portHandleOperationMembersUpdatedSuite))
}
