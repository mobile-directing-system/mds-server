package eventport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
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

func (m *HandlerMock) CreateUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *HandlerMock) UpdatePermissionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, permissions []store.Permission) error {
	return m.Called(ctx, tx, userID, permissions).Error(0)
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
	suite.sampleUserID = testutil.NewUUIDV4()
}

func (suite *PortHandleUserCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserCreated,
		RawValue:  rawValue,
	})
}

func (suite *PortHandleUserCreatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortHandleUserCreatedSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateUser", timeout, tx, suite.sampleUserID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(event.UserCreated{
			ID: suite.sampleUserID,
		}))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortHandleUserCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateUser", timeout, tx, suite.sampleUserID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(event.UserCreated{
			ID: suite.sampleUserID,
		}))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handlerUserCreated(t *testing.T) {
	suite.Run(t, new(PortHandleUserCreatedSuite))
}
