package ws

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wstest"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestConnection_UserID(t *testing.T) {
	authToken := auth.Token{UserID: testutil.NewUUIDV4()}
	conn := newConnection(wstest.NewConnectionMock(context.Background(), authToken))
	assert.Equal(t, authToken.UserID, conn.UserID(), "should return the correct user id")
}

// connectionNotifyNewAvailableSuite tests connection.NotifyNewAvailable.
type connectionNotifyNewAvailableSuite struct {
	suite.Suite
	wsConn                   *wstest.Connection
	conn                     *connection
	sampleOperation          uuid.UUID
	samplePublicNotification publicNewRadioDeliveriesAvailable
}

func (suite *connectionNotifyNewAvailableSuite) SetupTest() {
	suite.wsConn = wstest.NewConnectionMock(context.Background(), auth.Token{
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: permission.DeliverAnyRadioDeliveryPermissionName}},
	})
	suite.conn = newConnection(suite.wsConn)
	suite.sampleOperation = testutil.NewUUIDV4()
	suite.samplePublicNotification = publicNewRadioDeliveriesAvailable{
		Operation: suite.sampleOperation,
	}
}

func (suite *connectionNotifyNewAvailableSuite) TestSendFail() {
	suite.wsConn.SendFail = true

	err := suite.conn.NotifyNewAvailable(context.Background(), suite.sampleOperation)
	suite.Error(err, "should fail")
}

func (suite *connectionNotifyNewAvailableSuite) TestMissingPermission() {
	suite.wsConn = wstest.NewConnectionMock(context.Background(), auth.Token{
		Permissions: []permission.Permission{},
	})
	suite.conn = newConnection(suite.wsConn)
	err := suite.conn.NotifyNewAvailable(context.Background(), suite.sampleOperation)
	suite.Require().NoError(err, "should not fail")
	outbox := suite.wsConn.Outbox()
	suite.Empty(outbox, "should not have sent message")
}

func (suite *connectionNotifyNewAvailableSuite) TestOK() {
	err := suite.conn.NotifyNewAvailable(context.Background(), suite.sampleOperation)
	suite.Require().NoError(err, "should not fail")
	outbox := suite.wsConn.Outbox()
	suite.NotEmpty(outbox, "should have sent message")
	suite.Equal(wsutil.Message{
		Type:    messageTypeNewRadioDeliveriesAvailable,
		Payload: testutil.MarshalJSONMust(suite.samplePublicNotification),
	}, outbox[0], "should have sent correct message")
}

func TestConnection_Notify(t *testing.T) {
	suite.Run(t, new(connectionNotifyNewAvailableSuite))
}
