package eventport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/nulls"
	store "github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
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

func (m *HandlerMock) DeleteGroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	return m.Called(ctx, tx, groupID).Error(0)
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

func (m *HandlerMock) CreateIntel(ctx context.Context, create store.Intel) error {
	return m.Called(ctx, create).Error(0)
}

func (m *HandlerMock) InvalidateIntelByID(ctx context.Context, intelID uuid.UUID) error {
	return m.Called(ctx, intelID).Error(0)
}

func (m *HandlerMock) CreateActiveIntelDelivery(ctx context.Context, create store.ActiveIntelDelivery) error {
	return m.Called(ctx, create).Error(0)
}

func (m *HandlerMock) DeleteActiveIntelDeliveryByID(ctx context.Context, deliveryID uuid.UUID) error {
	return m.Called(ctx, deliveryID).Error(0)
}

func (m *HandlerMock) CreateActiveIntelDeliveryAttempt(ctx context.Context, create store.ActiveIntelDeliveryAttempt) error {
	return m.Called(ctx, create).Error(0)
}

func (m *HandlerMock) DeleteActiveIntelDeliveryAttemptByID(ctx context.Context, attemptID uuid.UUID) error {
	return m.Called(ctx, attemptID).Error(0)
}

func (m *HandlerMock) SetAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID, enabled bool) error {
	return m.Called(ctx, entryID, enabled).Error(0)
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

// portHandleIntelCreatedSuite tests Port.handleIntelCreated.
type portHandleIntelCreatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.IntelCreated
	sampleCreate store.Intel
}

func (suite *portHandleIntelCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.IntelCreated{
		ID:         testutil.NewUUIDV4(),
		CreatedAt:  testutil.NewRandomTime(),
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       event.IntelTypeAnalogRadioMessage,
		Content:    nil,
		SearchText: nulls.NewString("bicycle"),
		Importance: 127,
		IsValid:    true,
	}
	suite.sampleCreate = store.Intel{
		ID:         suite.sampleEvent.ID,
		CreatedAt:  suite.sampleEvent.CreatedAt,
		CreatedBy:  suite.sampleEvent.CreatedBy,
		Operation:  suite.sampleEvent.Operation,
		Importance: suite.sampleEvent.Importance,
		IsValid:    suite.sampleEvent.IsValid,
	}
}

func (suite *portHandleIntelCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.IntelTopic,
		EventType: event.TypeIntelCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleIntelCreatedSuite) TestBadEventValue() {
	tx := &testutil.DBTx{}
	err := suite.handle(context.Background(), tx, json.RawMessage(`{invalid`))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelCreatedSuite) TestCreateFail() {
	tx := &testutil.DBTx{}
	suite.handler.On("CreateIntel", mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelCreatedSuite) TestOK() {
	tx := &testutil.DBTx{}
	suite.handler.On("CreateIntel", mock.Anything, suite.sampleCreate).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func TestPort_handleIntelCreated(t *testing.T) {
	suite.Run(t, new(portHandleIntelCreatedSuite))
}

// portHandleIntelInvalidatedSuite tests Port.handleIntelInvalidated.
type portHandleIntelInvalidatedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.IntelInvalidated
	intelID     uuid.UUID
}

func (suite *portHandleIntelInvalidatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.IntelInvalidated{
		ID: testutil.NewUUIDV4(),
		By: testutil.NewUUIDV4(),
	}
	suite.intelID = suite.sampleEvent.ID
}

func (suite *portHandleIntelInvalidatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.IntelTopic,
		EventType: event.TypeIntelInvalidated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleIntelInvalidatedSuite) TestBadEventValue() {
	tx := &testutil.DBTx{}
	err := suite.handle(context.Background(), tx, json.RawMessage(`{invalid`))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelInvalidatedSuite) TestCreateFail() {
	tx := &testutil.DBTx{}
	suite.handler.On("InvalidateIntelByID", mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelInvalidatedSuite) TestOK() {
	tx := &testutil.DBTx{}
	suite.handler.On("InvalidateIntelByID", mock.Anything, suite.intelID).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func TestPort_handleIntelInvalidated(t *testing.T) {
	suite.Run(t, new(portHandleIntelInvalidatedSuite))
}

// portHandleIntelDeliveryCreatedSuite tests Port.handleIntelDeliveryCreated.
type portHandleIntelDeliveryCreatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.IntelDeliveryCreated
	sampleCreate store.ActiveIntelDelivery
}

func (suite *portHandleIntelDeliveryCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.IntelDeliveryCreated{
		ID:       testutil.NewUUIDV4(),
		Intel:    testutil.NewUUIDV4(),
		To:       testutil.NewUUIDV4(),
		IsActive: true,
		Success:  false,
		Note:     nulls.String{},
	}
	suite.sampleCreate = store.ActiveIntelDelivery{
		ID:    suite.sampleEvent.ID,
		Intel: suite.sampleEvent.Intel,
		To:    suite.sampleEvent.To,
		Note:  suite.sampleEvent.Note,
	}
}

func (suite *portHandleIntelDeliveryCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		EventType: event.TypeIntelDeliveryCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleIntelDeliveryCreatedSuite) TestBadEventValue() {
	tx := &testutil.DBTx{}
	err := suite.handle(context.Background(), tx, json.RawMessage(`{invalid`))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelDeliveryCreatedSuite) TestCreateFail() {
	tx := &testutil.DBTx{}
	suite.handler.On("CreateActiveIntelDelivery", mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelDeliveryCreatedSuite) TestInactiveDelivery() {
	suite.sampleEvent.IsActive = false
	tx := &testutil.DBTx{}
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func (suite *portHandleIntelDeliveryCreatedSuite) TestOK() {
	tx := &testutil.DBTx{}
	suite.handler.On("CreateActiveIntelDelivery", mock.Anything, suite.sampleCreate).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func TestPort_handleIntelDeliveryCreated(t *testing.T) {
	suite.Run(t, new(portHandleIntelDeliveryCreatedSuite))
}

// portHandleIntelDeliveryAttemptCreatedSuite tests
// Port.handleIntelDeliveryAttemptCreated.
type portHandleIntelDeliveryAttemptCreatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.IntelDeliveryAttemptCreated
	sampleCreate store.ActiveIntelDeliveryAttempt
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	intelID := testutil.NewUUIDV4()
	entryID := testutil.NewUUIDV4()
	suite.sampleEvent = event.IntelDeliveryAttemptCreated{
		ID: testutil.NewUUIDV4(),
		Delivery: event.IntelDeliveryAttemptCreatedDelivery{
			ID:       testutil.NewUUIDV4(),
			Intel:    intelID,
			To:       entryID,
			IsActive: true,
			Success:  false,
			Note:     nulls.String{},
		},
		AssignedEntry: event.IntelDeliveryAttemptCreatedAssignedEntry{
			ID:          entryID,
			Label:       "hot",
			Description: "medical",
			Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
			User:        nulls.NewUUID(testutil.NewUUIDV4()),
			UserDetails: nulls.JSONNullable[event.IntelDeliveryAttemptCreatedAssignedEntryUserDetails]{},
		},
		Intel: event.IntelDeliveryAttemptCreatedIntel{
			ID:         intelID,
			CreatedAt:  testutil.NewRandomTime(),
			CreatedBy:  testutil.NewUUIDV4(),
			Operation:  testutil.NewUUIDV4(),
			Type:       event.IntelTypeAnalogRadioMessage,
			Content:    nil,
			SearchText: nulls.NewString("south"),
			Importance: 314,
			IsValid:    true,
		},
		Channel:   testutil.NewUUIDV4(),
		CreatedAt: testutil.NewRandomTime(),
		IsActive:  true,
		Status:    event.IntelDeliveryStatusOpen,
		StatusTS:  testutil.NewRandomTime(),
		Note:      nulls.String{},
	}
	suite.sampleCreate = store.ActiveIntelDeliveryAttempt{
		ID:       suite.sampleEvent.ID,
		Delivery: suite.sampleEvent.Delivery.ID,
	}
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		EventType: event.TypeIntelDeliveryAttemptCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) TestBadEventValue() {
	tx := &testutil.DBTx{}
	err := suite.handle(context.Background(), tx, json.RawMessage(`{invalid`))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) TestCreateFail() {
	tx := &testutil.DBTx{}
	suite.handler.On("CreateActiveIntelDeliveryAttempt", mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) TestInactiveDeliveryAttempt() {
	suite.sampleEvent.IsActive = false
	tx := &testutil.DBTx{}
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) TestOK() {
	tx := &testutil.DBTx{}
	suite.handler.On("CreateActiveIntelDeliveryAttempt", mock.Anything, suite.sampleCreate).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func TestPort_handleIntelDeliveryAttemptCreated(t *testing.T) {
	suite.Run(t, new(portHandleIntelDeliveryAttemptCreatedSuite))
}

// portHandleIntelDeliveryAttemptStatusUpdatedSuite tests
// Port.handleIntelDeliveryAttemptStatusUpdated.
type portHandleIntelDeliveryAttemptStatusUpdatedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.IntelDeliveryAttemptStatusUpdated
	attemptID   uuid.UUID
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.IntelDeliveryAttemptStatusUpdated{
		ID:       testutil.NewUUIDV4(),
		IsActive: false,
		Status:   event.IntelDeliveryStatusDelivered,
		StatusTS: testutil.NewRandomTime(),
		Note:     nulls.String{},
	}
	suite.attemptID = suite.sampleEvent.ID
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		EventType: event.TypeIntelDeliveryAttemptStatusUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) TestBadEventValue() {
	tx := &testutil.DBTx{}
	err := suite.handle(context.Background(), tx, json.RawMessage(`{invalid`))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) TestOKStillActive() {
	suite.sampleEvent.IsActive = true
	tx := &testutil.DBTx{}
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) TestDeleteFail() {
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteActiveIntelDeliveryAttemptByID", mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) TestOKInactive() {
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteActiveIntelDeliveryAttemptByID", mock.Anything, suite.attemptID).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func TestPort_handleIntelDeliveryAttemptStatusUpdated(t *testing.T) {
	suite.Run(t, new(portHandleIntelDeliveryAttemptStatusUpdatedSuite))
}

// portHandleIntelDeliveryStatusUpdatedSuite tests
// Port.handleIntelDeliveryStatusUpdated.
type portHandleIntelDeliveryStatusUpdatedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.IntelDeliveryStatusUpdated
	deliveryID  uuid.UUID
}

func (suite *portHandleIntelDeliveryStatusUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.IntelDeliveryStatusUpdated{
		ID:       testutil.NewUUIDV4(),
		IsActive: false,
		Success:  true,
		Note:     nulls.String{},
	}
	suite.deliveryID = suite.sampleEvent.ID
}

func (suite *portHandleIntelDeliveryStatusUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		EventType: event.TypeIntelDeliveryStatusUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleIntelDeliveryStatusUpdatedSuite) TestBadEventValue() {
	tx := &testutil.DBTx{}
	err := suite.handle(context.Background(), tx, json.RawMessage(`{invalid`))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelDeliveryStatusUpdatedSuite) TestOKStillActive() {
	suite.sampleEvent.IsActive = true
	tx := &testutil.DBTx{}
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func (suite *portHandleIntelDeliveryStatusUpdatedSuite) TestDeleteFail() {
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteActiveIntelDeliveryByID", mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.Error(err, "should fail")
}

func (suite *portHandleIntelDeliveryStatusUpdatedSuite) TestOKInactive() {
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteActiveIntelDeliveryByID", mock.Anything, suite.deliveryID).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func TestPort_handleIntelDeliveryStatusUpdated(t *testing.T) {
	suite.Run(t, new(portHandleIntelDeliveryStatusUpdatedSuite))
}

// portHandleAddressBookEntryAutoDeliveryUpdatedSuite tests
// Port.handleAddressBookEntryAutoDeliveryUpdated.
type portHandleAddressBookEntryAutoDeliveryUpdatedSuite struct {
	suite.Suite
	handler               *HandlerMock
	port                  *PortMock
	sampleEvent           event.AddressBookEntryAutoDeliveryUpdated
	entryID               uuid.UUID
	isAutoDeliveryEnabled bool
}

func (suite *portHandleAddressBookEntryAutoDeliveryUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.AddressBookEntryAutoDeliveryUpdated{
		ID:                    testutil.NewUUIDV4(),
		IsAutoDeliveryEnabled: true,
	}
	suite.entryID = suite.sampleEvent.ID
	suite.isAutoDeliveryEnabled = suite.sampleEvent.IsAutoDeliveryEnabled
}

func (suite *portHandleAddressBookEntryAutoDeliveryUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		EventType: event.TypeAddressBookEntryAutoDeliveryUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleAddressBookEntryAutoDeliveryUpdatedSuite) TestBadEventValue() {
	tx := &testutil.DBTx{}
	err := suite.handle(context.Background(), tx, json.RawMessage(`{invalid`))
	suite.Error(err, "should fail")
}

func (suite *portHandleAddressBookEntryAutoDeliveryUpdatedSuite) TestSetFail() {
	tx := &testutil.DBTx{}
	suite.handler.On("SetAutoIntelDeliveryEnabledForAddressBookEntry", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.Error(err, "should fail")
}

func (suite *portHandleAddressBookEntryAutoDeliveryUpdatedSuite) TestOK() {
	tx := &testutil.DBTx{}
	suite.handler.On("SetAutoIntelDeliveryEnabledForAddressBookEntry", mock.Anything, suite.entryID, suite.isAutoDeliveryEnabled).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	err := suite.handle(context.Background(), tx, testutil.MarshalJSONMust(suite.sampleEvent))
	suite.NoError(err, "should not fail")
}

func TestPort_handleAddressBookEntryAutoDeliveryUpdated(t *testing.T) {
	suite.Run(t, new(portHandleAddressBookEntryAutoDeliveryUpdatedSuite))
}
