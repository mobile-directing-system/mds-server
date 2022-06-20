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

func (m *HandlerMock) CreateUser(ctx context.Context, user store.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *HandlerMock) UpdatePermissionsByUser(ctx context.Context, userID uuid.UUID, updatedPermissions []permission.Permission) error {
	return m.Called(ctx, userID, updatedPermissions).Error(0)
}

// PortHandlePermissionsUpdatedSuite tests Port.handlePermissionsUpdated.
type PortHandlePermissionsUpdatedSuite struct {
	suite.Suite
	handler                  *HandlerMock
	port                     *PortMock
	sampleUserID             uuid.UUID
	sampleUpdatedPermissions []permission.Permission
}

func (suite *PortHandlePermissionsUpdatedSuite) SetupTest() {
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

func (suite *PortHandlePermissionsUpdatedSuite) handle(ctx context.Context, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, kafkautil.Message{
		Topic:     event.PermissionsTopic,
		EventType: event.TypePermissionsUpdated,
		RawValue:  rawValue,
	})
}

func (suite *PortHandlePermissionsUpdatedSuite) TestBadEventValue() {
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

func (suite *PortHandlePermissionsUpdatedSuite) TestUpdateFail() {
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

func (suite *PortHandlePermissionsUpdatedSuite) TestOK() {
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
	suite.Run(t, new(PortHandlePermissionsUpdatedSuite))
}
