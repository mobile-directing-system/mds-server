package eventport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

// HandlerMock mocks Handler.
type HandlerMock struct {
	mock.Mock
}

func (m *HandlerMock) UpdatePermissionsByUser(ctx context.Context, userID uuid.UUID, updatedPermissions []permission.Permission) error {
	return m.Called(ctx, userID, updatedPermissions).Error(0)
}

func (m *HandlerMock) CreateUser(ctx context.Context, user store.UserWithPass) error {
	return m.Called(ctx, user).Error(0)
}

func (m *HandlerMock) UpdateUser(ctx context.Context, user store.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *HandlerMock) UpdateUserPassByUserID(ctx context.Context, userID uuid.UUID, newPass []byte) error {
	return m.Called(ctx, userID, newPass).Error(0)
}

func (m *HandlerMock) DeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

// portHandlePermissionsUpdatedSuite tests Port.handlePermissionsUpdated.
type portHandlePermissionsUpdatedSuite struct {
	suite.Suite
	handler                  *HandlerMock
	port                     *PortMock
	sampleUserID             uuid.UUID
	sampleUpdatedPermissions []permission.Permission
}

func (suite *portHandlePermissionsUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleUserID = uuid.New()
	suite.sampleUpdatedPermissions = []permission.Permission{
		{
			Name: "Hello",
		},
		{
			Name:    "World",
			Options: nulls.NewJSONRawMessage(json.RawMessage(`{"world":"!"}`)),
		},
	}
}

func (suite *portHandlePermissionsUpdatedSuite) handle(ctx context.Context, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, kafkautil.Message{
		Topic:     event.PermissionsTopic,
		EventType: event.TypePermissionsUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandlePermissionsUpdatedSuite) TestBadEventValue() {
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

func (suite *portHandlePermissionsUpdatedSuite) TestUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("UpdatePermissionsByUser", timeout, suite.sampleUserID, suite.sampleUpdatedPermissions).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(event.PermissionsUpdated{
			User:        suite.sampleUserID,
			Permissions: suite.sampleUpdatedPermissions,
		}))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *portHandlePermissionsUpdatedSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("UpdatePermissionsByUser", timeout, suite.sampleUserID, suite.sampleUpdatedPermissions).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(event.PermissionsUpdated{
			User:        suite.sampleUserID,
			Permissions: suite.sampleUpdatedPermissions,
		}))
		suite.NoError(err, "should not fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestPort_handlerUserDeleted(t *testing.T) {
	suite.Run(t, new(portHandlePermissionsUpdatedSuite))
}

// portHandleUserCreatedSuite test Port.handleUserCreated.
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
		ID:        uuid.New(),
		Username:  "scent",
		FirstName: "motor",
		LastName:  "english",
		IsAdmin:   true,
		Pass:      []byte("meow"),
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
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		defer cancel()
		err := suite.handle(timeout, []byte("{invalid"))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *portHandleUserCreatedSuite) TestCreateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("CreateUser", timeout, mock.Anything).Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *portHandleUserCreatedSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("CreateUser", timeout, store.UserWithPass{
		User: store.User{
			ID:       suite.sampleEvent.ID,
			Username: suite.sampleEvent.Username,
			IsAdmin:  suite.sampleEvent.IsAdmin,
		},
		Pass: suite.sampleEvent.Pass,
	}).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestPort_handleUserCreated(t *testing.T) {
	suite.Run(t, new(portHandleUserCreatedSuite))
}

// portHandleUserUpdatedSuite test Port.handleUserUpdated.
type portHandleUserUpdatedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.UserUpdated
}

func (suite *portHandleUserUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.UserUpdated{
		ID:        uuid.New(),
		Username:  "scent",
		FirstName: "motor",
		LastName:  "english",
		IsAdmin:   true,
	}
}

func (suite *portHandleUserUpdatedSuite) handle(ctx context.Context, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, kafkautil.Message{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleUserUpdatedSuite) TestBadEventValue() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		defer cancel()
		err := suite.handle(timeout, []byte("{invalid"))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *portHandleUserUpdatedSuite) TestUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("UpdateUser", timeout, mock.Anything).Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *portHandleUserUpdatedSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("UpdateUser", timeout, store.User{
		ID:       suite.sampleEvent.ID,
		Username: suite.sampleEvent.Username,
		IsAdmin:  suite.sampleEvent.IsAdmin,
	}).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestPort_handleUserUpdated(t *testing.T) {
	suite.Run(t, new(portHandleUserUpdatedSuite))
}

// portHandleUserPassUpdatedSuite test Port.handleUserPassUpdated.
type portHandleUserPassUpdatedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.UserPassUpdated
}

func (suite *portHandleUserPassUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.UserPassUpdated{
		User:    uuid.New(),
		NewPass: []byte("woof"),
	}
}

func (suite *portHandleUserPassUpdatedSuite) handle(ctx context.Context, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, kafkautil.Message{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserPassUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleUserPassUpdatedSuite) TestBadEventValue() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		defer cancel()
		err := suite.handle(timeout, []byte("{invalid"))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *portHandleUserPassUpdatedSuite) TestUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("UpdateUserPassByUserID", timeout, suite.sampleEvent.User, suite.sampleEvent.NewPass).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *portHandleUserPassUpdatedSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("UpdateUserPassByUserID", timeout, suite.sampleEvent.User, suite.sampleEvent.NewPass).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestPort_handleUserPassUpdated(t *testing.T) {
	suite.Run(t, new(portHandleUserPassUpdatedSuite))
}

// portHandleUserDeletedSuite test Port.handleUserDeleted.
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
		ID: uuid.New(),
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
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		defer cancel()
		err := suite.handle(timeout, []byte("{invalid"))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *portHandleUserDeletedSuite) TestUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("DeleteUserByID", timeout, suite.sampleEvent.ID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *portHandleUserDeletedSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.handler.On("DeleteUserByID", timeout, suite.sampleEvent.ID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	<-timeout.Done()
	suite.Require().NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestPort_handleUserDeleted(t *testing.T) {
	suite.Run(t, new(portHandleUserDeletedSuite))
}
