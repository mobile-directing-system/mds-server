package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// PortNotifyIntelDeliveryNotificationSentSuite tests
// Port.NotifyIntelDeliveryNotificationSent.
type PortNotifyIntelDeliveryNotificationSentSuite struct {
	suite.Suite
	port             *PortMock
	tx               *testutil.DBTx
	sampleAttempt    uuid.UUID
	sampleSentTS     time.Time
	expectedMessages []kafkautil.OutboundMessage
}

func (suite *PortNotifyIntelDeliveryNotificationSentSuite) SetupTest() {
	suite.port = newMockPort()
	suite.tx = &testutil.DBTx{}
	suite.sampleAttempt = testutil.NewUUIDV4()
	suite.sampleSentTS = time.Date(2022, 9, 8, 0, 4, 19, 0, time.UTC)
	suite.expectedMessages = []kafkautil.OutboundMessage{
		{
			Topic:     event.InAppNotificationsTopic,
			Key:       suite.sampleAttempt.String(),
			EventType: event.TypeInAppNotificationForIntelSent,
			Value: event.InAppNotificationForIntelSent{
				Attempt: suite.sampleAttempt,
				SentAt:  suite.sampleSentTS,
			},
			Headers: nil,
		},
	}
}

func (suite *PortNotifyIntelDeliveryNotificationSentSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryNotificationSent(timeout, suite.tx, suite.sampleAttempt, suite.sampleSentTS)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelDeliveryNotificationSentSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryNotificationSent(timeout, suite.tx, suite.sampleAttempt, suite.sampleSentTS)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.expectedMessages, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyIntelDeliveryNotificationSent(t *testing.T) {
	suite.Run(t, new(PortNotifyIntelDeliveryNotificationSentSuite))
}

// PortNotifyIntelDeliveryNotificationPendingSuite tests
// Port.NotifyIntelDeliveryNotificationPending.
type PortNotifyIntelDeliveryNotificationPendingSuite struct {
	suite.Suite
	port             *PortMock
	tx               *testutil.DBTx
	sampleAttempt    uuid.UUID
	sampleSince      time.Time
	expectedMessages []kafkautil.OutboundMessage
}

func (suite *PortNotifyIntelDeliveryNotificationPendingSuite) SetupTest() {
	suite.port = newMockPort()
	suite.tx = &testutil.DBTx{}
	suite.sampleAttempt = testutil.NewUUIDV4()
	suite.sampleSince = time.Date(2022, 9, 8, 0, 4, 19, 0, time.UTC)
	suite.expectedMessages = []kafkautil.OutboundMessage{
		{
			Topic:     event.InAppNotificationsTopic,
			Key:       suite.sampleAttempt.String(),
			EventType: event.TypeInAppNotificationForIntelPending,
			Value: event.InAppNotificationForIntelPending{
				Attempt: suite.sampleAttempt,
				Since:   suite.sampleSince,
			},
			Headers: nil,
		},
	}
}

func (suite *PortNotifyIntelDeliveryNotificationPendingSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryNotificationPending(timeout, suite.tx, suite.sampleAttempt, suite.sampleSince)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyIntelDeliveryNotificationPendingSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyIntelDeliveryNotificationPending(timeout, suite.tx, suite.sampleAttempt, suite.sampleSince)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.expectedMessages, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyIntelDeliveryNotificationPending(t *testing.T) {
	suite.Run(t, new(PortNotifyIntelDeliveryNotificationPendingSuite))
}
