package eventport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
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

func (m *HandlerMock) UpdatePermissionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, updatedPermissions []permission.Permission) error {
	return m.Called(ctx, tx, userID, updatedPermissions).Error(0)
}

func (m *HandlerMock) CreateUser(ctx context.Context, tx pgx.Tx, user store.UserWithPass) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *HandlerMock) UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *HandlerMock) UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newPass []byte) error {
	return m.Called(ctx, tx, userID, newPass).Error(0)
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
	suite.sampleUserID = testutil.NewUUIDV4()
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

func (suite *portHandlePermissionsUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.PermissionsTopic,
		EventType: event.TypePermissionsUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandlePermissionsUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandlePermissionsUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdatePermissionsByUser", timeout, tx, suite.sampleUserID, suite.sampleUpdatedPermissions).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(event.PermissionsUpdated{
			User:        suite.sampleUserID,
			Permissions: suite.sampleUpdatedPermissions,
		}))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandlePermissionsUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdatePermissionsByUser", timeout, tx, suite.sampleUserID, suite.sampleUpdatedPermissions).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(event.PermissionsUpdated{
			User:        suite.sampleUserID,
			Permissions: suite.sampleUpdatedPermissions,
		}))
		suite.NoError(err, "should not fail")
	}()

	wait()
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
		ID:        testutil.NewUUIDV4(),
		Username:  "scent",
		FirstName: "motor",
		LastName:  "english",
		IsAdmin:   true,
		Pass:      []byte("meow"),
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
		err := suite.handle(timeout, tx, []byte("{invalid"))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserCreatedSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateUser", timeout, tx, mock.Anything).Return(errors.New("sad life"))
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
	suite.handler.On("CreateUser", timeout, tx, store.UserWithPass{
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
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
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
		ID:        testutil.NewUUIDV4(),
		Username:  "scent",
		FirstName: "motor",
		LastName:  "english",
		IsAdmin:   true,
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
		err := suite.handle(timeout, tx, []byte("{invalid"))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateUser", timeout, tx, mock.Anything).Return(errors.New("sad life"))
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
	suite.handler.On("UpdateUser", timeout, tx, store.User{
		ID:       suite.sampleEvent.ID,
		Username: suite.sampleEvent.Username,
		IsAdmin:  suite.sampleEvent.IsAdmin,
	}).Return(nil)
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
		User:    testutil.NewUUIDV4(),
		NewPass: []byte("woof"),
	}
}

func (suite *portHandleUserPassUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserPassUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleUserPassUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, []byte("{invalid"))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserPassUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateUserPassByUserID", timeout, tx, suite.sampleEvent.User, suite.sampleEvent.NewPass).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserPassUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateUserPassByUserID", timeout, tx, suite.sampleEvent.User, suite.sampleEvent.NewPass).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleUserPassUpdated(t *testing.T) {
	suite.Run(t, new(portHandleUserPassUpdatedSuite))
}
