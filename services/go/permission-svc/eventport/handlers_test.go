package eventport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

// HandlerMock mocks Handler.
type HandlerMock struct {
	mock.Mock
}

func (m *HandlerMock) CreateUser(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *HandlerMock) DeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *HandlerMock) UpdatePermissionsByUser(ctx context.Context, userID uuid.UUID, permissions []store.Permission) error {
	return m.Called(ctx, userID, permissions).Error(0)
}

// PortHandleUserDeletedSuite tests Port.handleUserDeleted.
type PortHandleUserDeletedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleUserID uuid.UUID
}

func (suite *PortHandleUserDeletedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleUserID = uuid.New()
}

func (suite *PortHandleUserDeletedSuite) handle(ctx context.Context, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, kafkautil.Message{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserDeleted,
		RawValue:  rawValue,
	})
}

func (suite *PortHandleUserDeletedSuite) TestBadEventValue() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		defer cancel()
		err := suite.handle(timeout, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *PortHandleUserDeletedSuite) TestUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("DeleteUserByID", timeout, suite.sampleUserID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(event.UserDeleted{
			ID: suite.sampleUserID,
		}))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *PortHandleUserDeletedSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("DeleteUserByID", timeout, suite.sampleUserID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(event.UserDeleted{
			ID: suite.sampleUserID,
		}))
		suite.NoError(err, "should not fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestPort_handlerUserDeleted(t *testing.T) {
	suite.Run(t, new(PortHandleUserDeletedSuite))
}

// PortHandleUserCreatedSuite tests Port.handleUserCreated.
type PortHandleUserCreatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleUserID uuid.UUID
}

func (suite *PortHandleUserCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleUserID = uuid.New()
}

func (suite *PortHandleUserCreatedSuite) handle(ctx context.Context, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, kafkautil.Message{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserCreated,
		RawValue:  rawValue,
	})
}

func (suite *PortHandleUserCreatedSuite) TestBadEventValue() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		defer cancel()
		err := suite.handle(timeout, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *PortHandleUserCreatedSuite) TestCreateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("CreateUser", timeout, suite.sampleUserID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(event.UserCreated{
			ID: suite.sampleUserID,
		}))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *PortHandleUserCreatedSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("CreateUser", timeout, suite.sampleUserID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(event.UserCreated{
			ID: suite.sampleUserID,
		}))
		suite.NoError(err, "should not fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestPort_handlerUserCreated(t *testing.T) {
	suite.Run(t, new(PortHandleUserCreatedSuite))
}
