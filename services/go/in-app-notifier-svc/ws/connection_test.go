package ws

import (
	"context"
	"encoding/json"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wstest"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

func TestConnection_UserID(t *testing.T) {
	timeout, cancel, _ := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	defer cancel()
	authToken := auth.Token{UserID: testutil.NewUUIDV4()}
	conn := newConnection(wsutil.NewAutoParserConnection(wstest.NewConnectionMock(timeout, authToken)))
	assert.Equal(t, authToken.UserID, conn.UserID(), "should return the correct user id")
}

// connectionNotifySuite tests connection.Notify.
type connectionNotifySuite struct {
	suite.Suite
	wsConn                   *wstest.RawConnection
	conn                     *connection
	sampleNotification       store.OutgoingIntelDeliveryNotification
	samplePublicNotification publicIntelDeliveryNotification
}

func (suite *connectionNotifySuite) SetupTest() {
	connLifetime, shutdownConn := context.WithCancel(context.Background())
	suite.T().Cleanup(func() {
		shutdownConn()
	})
	suite.wsConn = wstest.NewConnectionMock(connLifetime, auth.Token{})
	suite.conn = newConnection(wsutil.NewAutoParserConnection(suite.wsConn))
	attemptID := testutil.NewUUIDV4()
	assignedToUser := testutil.NewUUIDV4()
	suite.sampleNotification = store.OutgoingIntelDeliveryNotification{
		IntelToDeliver: store.IntelToDeliver{
			Attempt:    attemptID,
			ID:         testutil.NewUUIDV4(),
			CreatedAt:  time.Date(2022, 9, 8, 1, 26, 19, 0, time.UTC),
			CreatedBy:  testutil.NewUUIDV4(),
			Operation:  testutil.NewUUIDV4(),
			Type:       "together",
			Content:    json.RawMessage(`{"hello":"world"}`),
			Importance: 189,
		},
		DeliveryAttempt: store.AcceptedIntelDeliveryAttempt{
			ID:              attemptID,
			AssignedTo:      testutil.NewUUIDV4(),
			AssignedToLabel: "same",
			AssignedToUser:  nulls.NewUUID(assignedToUser),
			Delivery:        testutil.NewUUIDV4(),
			Channel:         testutil.NewUUIDV4(),
			CreatedAt:       time.Date(2022, 9, 8, 1, 26, 45, 0, time.UTC),
			IsActive:        true,
			StatusTS:        time.Date(2022, 9, 8, 1, 26, 54, 0, time.UTC),
			Note:            nulls.NewString("most"),
			AcceptedAt:      time.Date(2022, 9, 8, 1, 27, 11, 0, time.UTC),
		},
		Channel: store.NotificationChannel{
			ID:      testutil.NewUUIDV4(),
			Entry:   testutil.NewUUIDV4(),
			Label:   "expense",
			Timeout: 940,
		},
		CreatorDetails: store.User{
			ID:        testutil.NewUUIDV4(),
			Username:  "indeed",
			FirstName: "white",
			LastName:  "that",
			IsActive:  true,
		},
		RecipientDetails: nulls.NewJSONNullable(store.User{
			ID:        assignedToUser,
			Username:  "any",
			FirstName: "various",
			LastName:  "earth",
			IsActive:  true,
		}),
	}
	suite.samplePublicNotification = publicIntelDeliveryNotification{
		IntelToDeliver: publicIntelToDeliver{
			Attempt:    suite.sampleNotification.IntelToDeliver.Attempt,
			ID:         suite.sampleNotification.IntelToDeliver.ID,
			CreatedAt:  suite.sampleNotification.IntelToDeliver.CreatedAt,
			CreatedBy:  suite.sampleNotification.IntelToDeliver.CreatedBy,
			Operation:  suite.sampleNotification.IntelToDeliver.Operation,
			Type:       string(suite.sampleNotification.IntelToDeliver.Type),
			Content:    suite.sampleNotification.IntelToDeliver.Content,
			Importance: suite.sampleNotification.IntelToDeliver.Importance,
		},
		DeliveryAttempt: publicIntelDeliveryAttempt{
			ID:              suite.sampleNotification.DeliveryAttempt.ID,
			AssignedTo:      suite.sampleNotification.DeliveryAttempt.AssignedTo,
			AssignedToLabel: suite.sampleNotification.DeliveryAttempt.AssignedToLabel,
			AssignedToUser:  suite.sampleNotification.DeliveryAttempt.AssignedToUser,
			Delivery:        suite.sampleNotification.DeliveryAttempt.Delivery,
			Channel:         suite.sampleNotification.DeliveryAttempt.Channel,
			CreatedAt:       suite.sampleNotification.DeliveryAttempt.CreatedAt,
			IsActive:        suite.sampleNotification.DeliveryAttempt.IsActive,
			StatusTS:        suite.sampleNotification.DeliveryAttempt.StatusTS,
			Note:            suite.sampleNotification.DeliveryAttempt.Note,
			AcceptedAt:      suite.sampleNotification.DeliveryAttempt.AcceptedAt,
		},
		Channel: publicNotificationChannel{
			ID:      suite.sampleNotification.Channel.ID,
			Entry:   suite.sampleNotification.Channel.Entry,
			Label:   suite.sampleNotification.Channel.Label,
			Timeout: suite.sampleNotification.Channel.Timeout,
		},
		CreatorDetails: publicUser{
			ID:        suite.sampleNotification.CreatorDetails.ID,
			Username:  suite.sampleNotification.CreatorDetails.Username,
			FirstName: suite.sampleNotification.CreatorDetails.FirstName,
			LastName:  suite.sampleNotification.CreatorDetails.LastName,
			IsActive:  suite.sampleNotification.CreatorDetails.IsActive,
		},
		RecipientDetails: nulls.NewJSONNullable(publicUser{
			ID:        suite.sampleNotification.RecipientDetails.V.ID,
			Username:  suite.sampleNotification.RecipientDetails.V.Username,
			FirstName: suite.sampleNotification.RecipientDetails.V.FirstName,
			LastName:  suite.sampleNotification.RecipientDetails.V.LastName,
			IsActive:  suite.sampleNotification.RecipientDetails.V.IsActive,
		}),
	}
}

func (suite *connectionNotifySuite) TestSendFail() {
	suite.wsConn.SendFail = true

	err := suite.conn.Notify(context.Background(), suite.sampleNotification)
	suite.Error(err, "should fail")
}

func (suite *connectionNotifySuite) TestOK() {
	err := suite.conn.Notify(context.Background(), suite.sampleNotification)
	suite.Require().NoError(err, "should not fail")
	outbox := suite.wsConn.Outbox()
	suite.NotEmpty(outbox, "should have send message")
	suite.Equal(wsutil.Message{
		Type:    messageTypeIntelNotification,
		Payload: testutil.MarshalJSONMust(suite.samplePublicNotification),
	}, outbox[0], "should have send correct message")
}

func TestConnection_Notify(t *testing.T) {
	suite.Run(t, new(connectionNotifySuite))
}
