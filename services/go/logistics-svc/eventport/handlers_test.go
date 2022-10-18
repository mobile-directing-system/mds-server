package eventport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
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

func (m *HandlerMock) CreateGroup(ctx context.Context, tx pgx.Tx, create store.Group) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *HandlerMock) UpdateGroup(ctx context.Context, tx pgx.Tx, update store.Group) error {
	return m.Called(ctx, tx, update).Error(0)
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

func (m *HandlerMock) CreateIntel(ctx context.Context, tx pgx.Tx, create store.Intel, initialDeliverTo []uuid.UUID) error {
	return m.Called(ctx, tx, create, initialDeliverTo).Error(0)
}

func (m *HandlerMock) InvalidateIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error {
	return m.Called(ctx, tx, intelID).Error(0)
}

func (m *HandlerMock) UpdateIntelDeliveryAttemptStatusForActive(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID,
	newStatus store.IntelDeliveryStatus, newNote nulls.String) error {
	return m.Called(ctx, tx, attemptID, newStatus, newNote).Error(0)
}

func (m *HandlerMock) MarkIntelDeliveryAttemptAsDeliveredTx(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, by uuid.NullUUID) error {
	return m.Called(ctx, tx, attemptID, by).Error(0)
}

func (m *HandlerMock) MarkIntelDeliveryAttemptAsFailed(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, note nulls.String) error {
	return m.Called(ctx, tx, attemptID, note).Error(0)
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

// portHandleGroupCreatedSuite tests Port.handleGroupCreated.
type portHandleGroupCreatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.GroupCreated
	sampleCreate store.Group
}

func (suite *portHandleGroupCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	members := make([]uuid.UUID, 16)
	for i := range members {
		members[i] = testutil.NewUUIDV4()
	}
	suite.sampleEvent = event.GroupCreated{
		ID:          testutil.NewUUIDV4(),
		Title:       "part",
		Description: "game",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members:     members,
	}
	suite.sampleCreate = store.Group{
		ID:          suite.sampleEvent.ID,
		Title:       suite.sampleEvent.Title,
		Description: suite.sampleEvent.Description,
		Operation:   suite.sampleEvent.Operation,
		Members:     suite.sampleEvent.Members,
	}
}

func (suite *portHandleGroupCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.GroupsTopic,
		EventType: event.TypeGroupCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleGroupCreatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleGroupCreatedSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateGroup", timeout, tx, suite.sampleCreate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleGroupCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateGroup", timeout, tx, suite.sampleCreate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleGroupCreated(t *testing.T) {
	suite.Run(t, new(portHandleGroupCreatedSuite))
}

// portHandleGroupUpdatedSuite tests Port.handleGroupUpdated.
type portHandleGroupUpdatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.GroupUpdated
	sampleUpdate store.Group
}

func (suite *portHandleGroupUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	members := make([]uuid.UUID, 16)
	for i := range members {
		members[i] = testutil.NewUUIDV4()
	}
	suite.sampleEvent = event.GroupUpdated{
		ID:          testutil.NewUUIDV4(),
		Title:       "part",
		Description: "game",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members:     members,
	}
	suite.sampleUpdate = store.Group{
		ID:          suite.sampleEvent.ID,
		Title:       suite.sampleEvent.Title,
		Description: suite.sampleEvent.Description,
		Operation:   suite.sampleEvent.Operation,
		Members:     suite.sampleEvent.Members,
	}
}

func (suite *portHandleGroupUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.GroupsTopic,
		EventType: event.TypeGroupUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleGroupUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleGroupUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateGroup", timeout, tx, suite.sampleUpdate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleGroupUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateGroup", timeout, tx, suite.sampleUpdate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleGroupUpdated(t *testing.T) {
	suite.Run(t, new(portHandleGroupUpdatedSuite))
}

// portHandleGroupDeletedSuite tests Port.handleGroupDeleted.
type portHandleGroupDeletedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.GroupDeleted
}

func (suite *portHandleGroupDeletedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.GroupDeleted{
		ID: testutil.NewUUIDV4(),
	}
}

func (suite *portHandleGroupDeletedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.GroupsTopic,
		EventType: event.TypeGroupDeleted,
		RawValue:  rawValue,
	})
}

func (suite *portHandleGroupDeletedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleGroupDeletedSuite) TestDeleteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteGroupByID", timeout, tx, suite.sampleEvent.ID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleGroupDeletedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteGroupByID", timeout, tx, suite.sampleEvent.ID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleGroupDeleted(t *testing.T) {
	suite.Run(t, new(portHandleGroupDeletedSuite))
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

// portHandleInAppNotificationForIntelPendingSuite tests
// Port.handleInAppNotificationForIntelPending.
type portHandleInAppNotificationForIntelPendingSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.InAppNotificationForIntelPending
}

func (suite *portHandleInAppNotificationForIntelPendingSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.InAppNotificationForIntelPending{
		Attempt: testutil.NewUUIDV4(),
		Since:   time.Date(2022, 9, 8, 13, 58, 39, 0, time.UTC),
	}
}

func (suite *portHandleInAppNotificationForIntelPendingSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.InAppNotificationsTopic,
		EventType: event.TypeInAppNotificationForIntelPending,
		RawValue:  rawValue,
	})
}

func (suite *portHandleInAppNotificationForIntelPendingSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleInAppNotificationForIntelPendingSuite) TestUpdateStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusAwaitingDelivery, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleInAppNotificationForIntelPendingSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusAwaitingDelivery, mock.Anything).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleInAppNotificationForIntelPendingSuite(t *testing.T) {
	suite.Run(t, new(portHandleInAppNotificationForIntelPendingSuite))
}

// portHandleInAppNotificationForIntelSentSuite tests
// Port.handleInAppNotificationForIntelSent.
type portHandleInAppNotificationForIntelSentSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.InAppNotificationForIntelSent
}

func (suite *portHandleInAppNotificationForIntelSentSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.InAppNotificationForIntelSent{
		Attempt: testutil.NewUUIDV4(),
		SentAt:  time.Date(2022, 9, 8, 13, 59, 39, 0, time.UTC),
	}
}

func (suite *portHandleInAppNotificationForIntelSentSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.InAppNotificationsTopic,
		EventType: event.TypeInAppNotificationForIntelSent,
		RawValue:  rawValue,
	})
}

func (suite *portHandleInAppNotificationForIntelSentSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleInAppNotificationForIntelSentSuite) TestUpdateStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusAwaitingAck, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleInAppNotificationForIntelSentSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusAwaitingAck, mock.Anything).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleInAppNotificationForIntelSentSuite(t *testing.T) {
	suite.Run(t, new(portHandleInAppNotificationForIntelSentSuite))
}

// portHandleRadioDeliveryReadyForPickupSuite tests
// Port.handleRadioDeliveryReadyForPickup.
type portHandleRadioDeliveryReadyForPickupSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.RadioDeliveryReadyForPickup
}

func (suite *portHandleRadioDeliveryReadyForPickupSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.RadioDeliveryReadyForPickup{
		Attempt:                testutil.NewUUIDV4(),
		Intel:                  testutil.NewUUIDV4(),
		IntelOperation:         testutil.NewUUIDV4(),
		IntelImportance:        979,
		AttemptAssignedTo:      testutil.NewUUIDV4(),
		AttemptAssignedToLabel: "appear",
		Delivery:               testutil.NewUUIDV4(),
		Channel:                testutil.NewUUIDV4(),
		Note:                   "jaw",
		AttemptAcceptedAt:      time.Date(2022, 10, 13, 14, 11, 40, 0, time.UTC),
	}
}

func (suite *portHandleRadioDeliveryReadyForPickupSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		EventType: event.TypeRadioDeliveryReadyForPickup,
		RawValue:  rawValue,
	})
}

func (suite *portHandleRadioDeliveryReadyForPickupSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryReadyForPickupSuite) TestUpdateStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusAwaitingDelivery, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryReadyForPickupSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusAwaitingDelivery, mock.Anything).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleRadioDeliveryReadyForPickupSuite(t *testing.T) {
	suite.Run(t, new(portHandleRadioDeliveryReadyForPickupSuite))
}

// portHandleRadioDeliveryPickedUpSuite tests Port.handleRadioDeliveryPickedUp.
type portHandleRadioDeliveryPickedUpSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.RadioDeliveryPickedUp
}

func (suite *portHandleRadioDeliveryPickedUpSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.RadioDeliveryPickedUp{
		Attempt:    testutil.NewUUIDV4(),
		PickedUpBy: testutil.NewUUIDV4(),
		PickedUpAt: time.Date(2022, 10, 13, 14, 12, 55, 0, time.UTC),
	}
}

func (suite *portHandleRadioDeliveryPickedUpSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		EventType: event.TypeRadioDeliveryPickedUp,
		RawValue:  rawValue,
	})
}

func (suite *portHandleRadioDeliveryPickedUpSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryPickedUpSuite) TestUpdateStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusDelivering, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryPickedUpSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusDelivering, mock.Anything).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleRadioDeliveryPickedUpSuite(t *testing.T) {
	suite.Run(t, new(portHandleRadioDeliveryPickedUpSuite))
}

// portHandleRadioDeliveryReleasedSuite tests Port.handleRadioDeliveryReleased.
type portHandleRadioDeliveryReleasedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.RadioDeliveryReleased
}

func (suite *portHandleRadioDeliveryReleasedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.RadioDeliveryReleased{
		Attempt:    testutil.NewUUIDV4(),
		ReleasedAt: time.Date(2022, 10, 13, 14, 12, 55, 0, time.UTC),
	}
}

func (suite *portHandleRadioDeliveryReleasedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		EventType: event.TypeRadioDeliveryReleased,
		RawValue:  rawValue,
	})
}

func (suite *portHandleRadioDeliveryReleasedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryReleasedSuite) TestUpdateStatusFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusAwaitingDelivery, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryReleasedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatusForActive", timeout, tx, suite.sampleEvent.Attempt,
		store.IntelDeliveryStatusAwaitingDelivery, mock.Anything).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleRadioDeliveryReleasedSuite(t *testing.T) {
	suite.Run(t, new(portHandleRadioDeliveryReleasedSuite))
}

// portHandleRadioDeliveryFinishedSuite tests Port.handleRadioDeliveryFinished.
type portHandleRadioDeliveryFinishedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.RadioDeliveryFinished
}

func (suite *portHandleRadioDeliveryFinishedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.RadioDeliveryFinished{
		Attempt:    testutil.NewUUIDV4(),
		PickedUpBy: nulls.NewUUID(testutil.NewUUIDV4()),
		PickedUpAt: nulls.NewTime(time.Date(2022, 10, 13, 14, 15, 23, 0, time.UTC)),
		Success:    true,
		FinishedAt: time.Date(2022, 10, 13, 14, 15, 55, 0, time.UTC),
		Note:       "instead",
	}
}

func (suite *portHandleRadioDeliveryFinishedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		EventType: event.TypeRadioDeliveryFinished,
		RawValue:  rawValue,
	})
}

func (suite *portHandleRadioDeliveryFinishedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryFinishedSuite) TestMarkAsDeliveredFail() {
	suite.sampleEvent.Success = true
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("MarkIntelDeliveryAttemptAsDeliveredTx", timeout, tx, suite.sampleEvent.Attempt, uuid.NullUUID{}).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryFinishedSuite) TestOKDelivered() {
	suite.sampleEvent.Success = true
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("MarkIntelDeliveryAttemptAsDeliveredTx", timeout, tx, suite.sampleEvent.Attempt, uuid.NullUUID{}).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryFinishedSuite) TestMarkAsFailedFail() {
	suite.sampleEvent.Success = false
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("MarkIntelDeliveryAttemptAsFailed", timeout, tx, suite.sampleEvent.Attempt, nulls.NewString(suite.sampleEvent.Note)).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleRadioDeliveryFinishedSuite) TestOKFailed() {
	suite.sampleEvent.Success = false
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("MarkIntelDeliveryAttemptAsFailed", timeout, tx, suite.sampleEvent.Attempt, nulls.NewString(suite.sampleEvent.Note)).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleRadioDeliveryFinishedSuite(t *testing.T) {
	suite.Run(t, new(portHandleRadioDeliveryFinishedSuite))
}
