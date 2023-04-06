package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/zaprec"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/checkpointrec"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wstest"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	"time"
)

const timeout = 5 * time.Second

type ControllerSinkMock struct {
	mock.Mock
}

func (m *ControllerSinkMock) ServeOpenIntelDeliveriesListener(lifetime context.Context, operationID uuid.UUID, notifier controller.OpenIntelDeliveriesListener) {
	m.Called(lifetime, operationID, notifier)
}

// ConnListenerSuite tests ConnListener.
type ConnListenerSuite struct {
	suite.Suite
	sink       *ControllerSinkMock
	sampleConn *wstest.RawConnection
	listener   wsutil.ConnListener
}

func (suite *ConnListenerSuite) SetupTest() {
	suite.sink = &ControllerSinkMock{}
	suite.sampleConn = wstest.NewConnectionMock(context.Background(), auth.Token{IsAuthenticated: true})
	suite.listener = ConnListener(zap.NewNop(), suite.sink)
}

func (suite *ConnListenerSuite) TestNotAuthenticated() {
	conn := wstest.NewConnectionMock(context.Background(), auth.Token{IsAuthenticated: false})
	logger, recorder := zaprec.NewRecorder(zap.ErrorLevel)
	listener := ConnListener(logger, suite.sink)

	listener(conn)

	suite.NotEmpty(recorder.Records(), "should have logged error")
}

func (suite *ConnListenerSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	keepConnectionAlive, disconnect := context.WithCancel(timeout)
	defer cancel()
	operationID := testutil.NewUUIDV4()
	suite.sampleConn.SetPermissions([]permission.Permission{{Name: permission.ManageIntelDeliveryPermissionName}})
	suite.sink.On("ServeOpenIntelDeliveriesListener", mock.Anything, operationID, mock.Anything).
		Run(func(_ mock.Arguments) {
			disconnect()
		}).Return().Once()
	defer suite.sink.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		suite.sampleConn.NextReceive(timeout, messageTypeSubscribeOpenIntelDeliveries, messageSubscribeOpenIntelDeliveries{
			Operation: operationID,
		})
		<-keepConnectionAlive.Done()
		suite.Require().NotEqual(context.DeadlineExceeded, keepConnectionAlive.Err(), "should not time out while waiting for disconnect-call")
		suite.sampleConn.Disconnect()
	}()

	suite.listener(suite.sampleConn)

	wait()
}

func TestConnListener(t *testing.T) {
	suite.Run(t, new(ConnListenerSuite))
}

// GatekeeperSuite tests Gatekeeper.
type GatekeeperSuite struct {
	suite.Suite
	tokenOK auth.Token
}

func (suite *GatekeeperSuite) SetupTest() {
	suite.tokenOK = auth.Token{IsAuthenticated: true}
}

func (suite *GatekeeperSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	err := Gatekeeper()(token)
	suite.Error(err, "should fail")
}

func (suite *GatekeeperSuite) TestOK() {
	err := Gatekeeper()(suite.tokenOK)
	suite.NoError(err, "should not fail")
}

func TestGatekeeper(t *testing.T) {
	suite.Run(t, new(GatekeeperSuite))
}

// connectionSubscribeOpenIntelDeliveriesSuite tests
// connection.subscribeOpenIntelDeliveries.
type connectionSubscribeOpenIntelDeliveriesSuite struct {
	suite.Suite
	sink           *ControllerSinkMock
	conn           *connection
	rawConn        *wstest.RawConnection
	messageContent messageSubscribeOpenIntelDeliveries
	operationID    uuid.UUID
}

func (suite *connectionSubscribeOpenIntelDeliveriesSuite) SetupTest() {
	suite.sink = &ControllerSinkMock{}
	connLifetime, cancelConn := context.WithCancel(context.Background())
	suite.T().Cleanup(cancelConn)
	suite.rawConn = wstest.NewConnectionMock(connLifetime, auth.Token{
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: permission.ManageIntelDeliveryPermissionName}},
	})
	suite.conn = newConnection(zap.NewNop(), suite.rawConn, suite.sink)
	suite.operationID = testutil.NewUUIDV4()
	suite.messageContent = messageSubscribeOpenIntelDeliveries{
		Operation: suite.operationID,
	}
}

func (suite *connectionSubscribeOpenIntelDeliveriesSuite) handle() error {
	return suite.conn.handleReceivedMessage(wsutil.Message{
		Type:    messageTypeSubscribeOpenIntelDeliveries,
		Payload: testutil.MarshalJSONMust(suite.messageContent),
	})
}

func (suite *connectionSubscribeOpenIntelDeliveriesSuite) TestInvalidContent() {
	err := suite.conn.handleReceivedMessage(wsutil.Message{
		Type:    messageTypeSubscribeOpenIntelDeliveries,
		Payload: json.RawMessage(`{invalid`),
	})
	suite.Require().Error(err, "should fail")
	suite.Equal(meh.ErrBadInput, meh.ErrorCode(err), "should return correct error code")
}

func (suite *connectionSubscribeOpenIntelDeliveriesSuite) TestMissingPermission() {
	suite.rawConn.SetPermissions(nil)

	err := suite.handle()
	suite.Require().Error(err, "should fail")
	suite.Equal(meh.ErrForbidden, meh.ErrorCode(err), "should return correct error code")
}

func (suite *connectionSubscribeOpenIntelDeliveriesSuite) TestNotAcceptingAnymore() {
	suite.conn.accept = false

	err := suite.handle()
	suite.Error(err, "should fail")
}

func (suite *connectionSubscribeOpenIntelDeliveriesSuite) TestOK() {
	openIntelDeliveries := make([]store.OpenIntelDeliverySummary, 16)
	const messageCount = 32
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.sink.On("ServeOpenIntelDeliveriesListener", mock.Anything, suite.operationID, mock.MatchedBy(func(l openIntelDeliveriesListener) bool {
		suite.Equal(suite.operationID, l.operationID, "should register listener with correct operation id")
		suite.NotNil(l.notify, "should register listener with notify fn")
		return true
	})).Run(func(args mock.Arguments) {
		listener := args.Get(2).(openIntelDeliveriesListener)
		for messageNum := 0; messageNum < messageCount; messageNum++ {
			listener.notify(timeout, suite.operationID, openIntelDeliveries)
		}
	})
	defer suite.sink.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle()
		suite.Require().NoError(err, "should not fail")
		// Read subscribed-message.
		select {
		case <-timeout.Done():
			suite.Fail("timeout", "timeout while waiting for subscribed-message")
			return
		case message := <-suite.rawConn.OutboxChan():
			suite.Require().Equal(messageTypeSubscribedOpenIntelDeliveries, message.Type, "should send subscribed-message with correct message type")
			var subscribedMessageContent messageSubscribedOpenIntelDeliveries
			suite.Require().NoError(json.Unmarshal(message.Payload, &subscribedMessageContent), "should return valid subscribed-message payload")
			suite.Require().Equal([]uuid.UUID{suite.operationID}, subscribedMessageContent.Operations, "should have operation id in subscribed-message")
		}
		// Read notify-messages.
		for messageNum := 0; messageNum < messageCount; messageNum++ {
			select {
			case <-timeout.Done():
				suite.Failf("timeout", "timeout while waiting for message %d of %d", messageNum+1, messageCount)
				return
			case message := <-suite.rawConn.OutboxChan():
				suite.Require().Equal(messageTypeOpenIntelDeliveries, message.Type, "should send notify-message with correct message type")
				var notifyMessageContent messageOpenIntelDeliveries
				suite.Require().NoError(json.Unmarshal(message.Payload, &notifyMessageContent), "should have valid notify-message payload")
				suite.Require().Equal(suite.operationID, notifyMessageContent.Operation, "should have correct operation id in notify-message payload")
				suite.Require().Len(notifyMessageContent.OpenIntelDeliveries, len(openIntelDeliveries), "should have correct open intel deliveries in notify-message payload")
			}
		}
	}()

	wait()
}

func Test_connectionSubscribeOpenIntelDeliveries(t *testing.T) {
	suite.Run(t, new(connectionSubscribeOpenIntelDeliveriesSuite))
}

func Test_connectionOpenIntelDeliveries(t *testing.T) {
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	defer cancel()
	sink := &ControllerSinkMock{}
	rawConn := wstest.NewConnectionMock(timeout, auth.Token{
		IsAuthenticated: true,
		Permissions: []permission.Permission{
			{Name: permission.ManageIntelDeliveryPermissionName},
		},
	})
	conn := newConnection(zap.NewNop(), rawConn, sink)
	cpRec := checkpointrec.NewRecorder()

	operation1 := testutil.NewUUIDV4()
	operation2 := testutil.NewUUIDV4()

	sink.On("ServeOpenIntelDeliveriesListener", mock.Anything, operation1, mock.Anything).Run(func(args mock.Arguments) {
		cpRec.Checkpoint("op1_serve")
		listener := args.Get(2).(openIntelDeliveriesListener)
		// Send 3 messages.
		listener.notify(timeout, operation1, []store.OpenIntelDeliverySummary{})
		listener.notify(timeout, operation1, []store.OpenIntelDeliverySummary{})
		listener.notify(timeout, operation1, []store.OpenIntelDeliverySummary{})
		cpRec.Checkpoint("op1_close")
	})
	sink.On("ServeOpenIntelDeliveriesListener", mock.Anything, operation2, mock.Anything).Run(func(args mock.Arguments) {
		cpRec.Checkpoint("op2_serve")
		listener := args.Get(2).(openIntelDeliveriesListener)
		// Send message.
		listener.notify(timeout, operation2, []store.OpenIntelDeliverySummary{})
		// Expect unsubscribe.
		select {
		case <-timeout.Done():
			assert.Fail(t, "timeout", "timeout while waiting for operation 2 subscription to be done")
		case <-args.Get(0).(context.Context).Done():
		}
		cpRec.Checkpoint("op2_unsubscribed")
	})
	defer sink.AssertExpectations(t)

	waitForOperation1Close := cpRec.WaitForCheckpoint("op1_close")
	waitForOperation2Unsubscribed := cpRec.WaitForCheckpoint("op2_unsubscribed")

	// Subscribe.
	go func() {
		err := conn.handleReceivedMessage(wsutil.Message{
			Type: messageTypeSubscribeOpenIntelDeliveries,
			Payload: testutil.MarshalJSONMust(messageSubscribeOpenIntelDeliveries{
				Operation: operation1,
			}),
		})
		require.NoError(t, err, "handle first subscribe-message should not fail")
		err = conn.handleReceivedMessage(wsutil.Message{
			Type: messageTypeSubscribeOpenIntelDeliveries,
			Payload: testutil.MarshalJSONMust(messageSubscribeOpenIntelDeliveries{
				Operation: operation2,
			}),
		})
		require.NoError(t, err, "handle second subscribe-message should not fail")
	}()

	go func() {
		operation1Messages := 1
		for {
			select {
			case <-timeout.Done():
				return
			case outMessage := <-rawConn.OutboxChan():
				switch outMessage.Type {
				case messageTypeSubscribedOpenIntelDeliveries:
					var subscribedPayload messageSubscribedOpenIntelDeliveries
					require.NoError(t, json.Unmarshal(outMessage.Payload, &subscribedPayload), "unmarshal subscribed-message paylod should not fail")
					cpRec.Checkpoint(fmt.Sprintf("subscribed_%d_operations", len(subscribedPayload.Operations)))
				case messageTypeOpenIntelDeliveries:
					var deliveriesPayload messageOpenIntelDeliveries
					require.NoError(t, json.Unmarshal(outMessage.Payload, &deliveriesPayload), "unmarshal deliveries-message paylod should not fail")
					require.True(t, deliveriesPayload.Operation == operation1 || deliveriesPayload.Operation == operation2,
						"operation in deliveries-message should either be for operation 1 or 2 but was for none of them")
					if deliveriesPayload.Operation == operation1 {
						cpRec.Checkpoint(fmt.Sprintf("op1_got_deliveries_message_%d", operation1Messages))
						operation1Messages++
					} else {
						cpRec.Checkpoint("op2_got_deliveries_message")
						// Unsubscribe operation 2 after first message.
						go func() {
							err := conn.handleReceivedMessage(wsutil.Message{
								Type:    messageTypeUnsubscribeOpenIntelDeliveries,
								Payload: testutil.MarshalJSONMust(messageUnsubscribeOpenIntelDeliveries{Operation: operation2}),
							})
							require.NoError(t, err, "handle unsubscribe-message should not fail")
							cpRec.Checkpoint("op2_sent_unsubscribe_message")
						}()
					}
				}
			}
		}
	}()

	// Wait until both done.
	select {
	case <-timeout.Done():
		cpRec.Require().Fail(t, "timeout", "timeout while waiting for operation 1 close")
	case <-waitForOperation1Close:
	}
	select {
	case <-timeout.Done():
		cpRec.Require().Fail(t, "timeout", "timeout while waiting for operation 2 unsubscribed")
	case <-waitForOperation2Unsubscribed:
	}
	select {
	case <-timeout.Done():
		cpRec.Require().Fail(t, "timeout", "timeout while waiting for third deliveries-messages for operation 1")
	case <-cpRec.WaitForCheckpoint("op1_got_deliveries_message_3"):
	}

	cpRec.Includes(t, "subscribed_2_operations")
	cpRec.IncludesOrdered(t, []string{
		"op1_serve",
		"op1_close",
	})
	cpRec.IncludesOrdered(t, []string{
		"op1_got_deliveries_message_1",
		"op1_got_deliveries_message_2",
		"op1_got_deliveries_message_3",
	})
	cpRec.IncludesOrdered(t, []string{
		"op2_serve",
		"op2_got_deliveries_message",
		"op2_unsubscribed",
	})

	cancel()
	wait()
}
