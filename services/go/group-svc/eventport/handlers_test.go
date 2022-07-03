package eventport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
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

func (m *HandlerMock) CreateOperation(ctx context.Context, operationID uuid.UUID) error {
	return m.Called(ctx, operationID).Error(0)
}

func (m *HandlerMock) CreateUser(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *HandlerMock) DeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

// portHandleOperationCreatedSuite tests Port.handleOperationCreated.
type portHandleOperationCreatedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.OperationCreated
}

func (suite *portHandleOperationCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.OperationCreated{
		ID:          testutil.NewUUIDV4(),
		Title:       "drag",
		Description: "cushion",
		Start:       time.Date(2022, 06, 02, 18, 55, 35, 0, time.UTC),
		End:         nulls.NewTime(time.Date(2022, 06, 03, 12, 0, 0, 0, time.UTC)),
		IsArchived:  false,
	}
}

func (suite *portHandleOperationCreatedSuite) handle(ctx context.Context, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, kafkautil.Message{
		Topic:     event.OperationsTopic,
		EventType: event.TypeOperationCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleOperationCreatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()

	go func() {
		defer cancel()
		err := suite.handle(timeout, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleOperationCreatedSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.handler.On("CreateOperation", timeout, suite.sampleEvent.ID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleOperationCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.handler.On("CreateOperation", timeout, suite.sampleEvent.ID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleOperationCreatedSuite(t *testing.T) {
	suite.Run(t, new(portHandleOperationCreatedSuite))
}

// portHandleUserCreatedSuite tests Port.handleUserCreated.
type portHandleUserCreatedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.UserCreated
}

func (suite *portHandleUserCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.UserCreated{
		ID: testutil.NewUUIDV4(),
	}
}

func (suite *portHandleUserCreatedSuite) handle(ctx context.Context, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, kafkautil.Message{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleUserCreatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()

	go func() {
		defer cancel()
		err := suite.handle(timeout, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserCreatedSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.handler.On("CreateUser", timeout, suite.sampleEvent.ID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.handler.On("CreateUser", timeout, suite.sampleEvent.ID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handlerUserCreated(t *testing.T) {
	suite.Run(t, new(portHandleUserCreatedSuite))
}

// portHandleUserDeletedSuite tests Port.handleUserDeleted.
type portHandleUserDeletedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.UserDeleted
}

func (suite *portHandleUserDeletedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.UserDeleted{
		ID: testutil.NewUUIDV4(),
	}
}

func (suite *portHandleUserDeletedSuite) handle(ctx context.Context, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, kafkautil.Message{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserDeleted,
		RawValue:  rawValue,
	})
}

func (suite *portHandleUserDeletedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()

	go func() {
		defer cancel()
		err := suite.handle(timeout, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserDeletedSuite) TestDeleteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.handler.On("DeleteUserByID", timeout, suite.sampleEvent.ID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserDeletedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.handler.On("DeleteUserByID", timeout, suite.sampleEvent.ID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handlerUserDeleted(t *testing.T) {
	suite.Run(t, new(portHandleUserDeletedSuite))
}
