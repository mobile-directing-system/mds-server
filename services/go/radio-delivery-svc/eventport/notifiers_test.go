package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// PortNotifyRadioDeliveryReadyForPickupSuite tests
// Port.NotifyRadioDeliveryReadyForPickup.
type PortNotifyRadioDeliveryReadyForPickupSuite struct {
	suite.Suite
	port                       *PortMock
	sampleIntelDeliveryAttempt store.AcceptedIntelDeliveryAttempt
	sampleRadioDeliveryNote    string
	expectedMessage            kafkautil.OutboundMessage
}

func (suite *PortNotifyRadioDeliveryReadyForPickupSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleIntelDeliveryAttempt = store.AcceptedIntelDeliveryAttempt{
		ID:              testutil.NewUUIDV4(),
		Intel:           testutil.NewUUIDV4(),
		IntelOperation:  testutil.NewUUIDV4(),
		IntelImportance: 403,
		AssignedTo:      testutil.NewUUIDV4(),
		AssignedToLabel: "explore",
		AssignedToUser:  nulls.NewUUID(testutil.NewUUIDV4()),
		Delivery:        testutil.NewUUIDV4(),
		Channel:         testutil.NewUUIDV4(),
		CreatedAt:       testutil.NewRandomTime(),
		IsActive:        true,
		StatusTS:        testutil.NewRandomTime(),
		Note:            nulls.NewString("gaiety"),
		AcceptedAt:      testutil.NewRandomTime(),
	}
	suite.sampleRadioDeliveryNote = "overflow"
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		Key:       suite.sampleIntelDeliveryAttempt.ID.String(),
		EventType: event.TypeRadioDeliveryReadyForPickup,
		Value: event.RadioDeliveryReadyForPickup{
			Attempt:                suite.sampleIntelDeliveryAttempt.ID,
			Intel:                  suite.sampleIntelDeliveryAttempt.Intel,
			IntelOperation:         suite.sampleIntelDeliveryAttempt.IntelOperation,
			IntelImportance:        suite.sampleIntelDeliveryAttempt.IntelImportance,
			AttemptAssignedTo:      suite.sampleIntelDeliveryAttempt.AssignedTo,
			AttemptAssignedToLabel: suite.sampleIntelDeliveryAttempt.AssignedToLabel,
			Delivery:               suite.sampleIntelDeliveryAttempt.Delivery,
			Channel:                suite.sampleIntelDeliveryAttempt.Channel,
			Note:                   suite.sampleRadioDeliveryNote,
			AttemptAcceptedAt:      suite.sampleIntelDeliveryAttempt.AcceptedAt,
		},
	}
}

func (suite *PortNotifyRadioDeliveryReadyForPickupSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyRadioDeliveryReadyForPickup(timeout, &testutil.DBTx{},
			suite.sampleIntelDeliveryAttempt, suite.sampleRadioDeliveryNote)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyRadioDeliveryReadyForPickupSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyRadioDeliveryReadyForPickup(timeout, &testutil.DBTx{},
			suite.sampleIntelDeliveryAttempt, suite.sampleRadioDeliveryNote)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
	}()

	wait()
}

func TestPort_NotifyRadioDeliveryReadyForPickup(t *testing.T) {
	suite.Run(t, new(PortNotifyRadioDeliveryReadyForPickupSuite))
}

// PortNotifyRadioDeliveryPickedUpSuite tests Port.NotifyRadioDeliveryPickedUp.
type PortNotifyRadioDeliveryPickedUpSuite struct {
	suite.Suite
	port             *PortMock
	sampleAttemptID  uuid.UUID
	samplePickedUpBy uuid.UUID
	samplePickedUpAt time.Time
	expectedMessage  kafkautil.OutboundMessage
}

func (suite *PortNotifyRadioDeliveryPickedUpSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.samplePickedUpBy = testutil.NewUUIDV4()
	suite.samplePickedUpAt = testutil.NewRandomTime()
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		Key:       suite.sampleAttemptID.String(),
		EventType: event.TypeRadioDeliveryPickedUp,
		Value: event.RadioDeliveryPickedUp{
			Attempt:    suite.sampleAttemptID,
			PickedUpBy: suite.samplePickedUpBy,
			PickedUpAt: suite.samplePickedUpAt,
		},
	}
}

func (suite *PortNotifyRadioDeliveryPickedUpSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyRadioDeliveryPickedUp(timeout, &testutil.DBTx{},
			suite.sampleAttemptID, suite.samplePickedUpBy, suite.samplePickedUpAt)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyRadioDeliveryPickedUpSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyRadioDeliveryPickedUp(timeout, &testutil.DBTx{},
			suite.sampleAttemptID, suite.samplePickedUpBy, suite.samplePickedUpAt)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
	}()

	wait()
}

func TestPort_NotifyRadioDeliveryPickedUp(t *testing.T) {
	suite.Run(t, new(PortNotifyRadioDeliveryPickedUpSuite))
}

// PortNotifyRadioDeliveryReleasedSuite tests Port.NotifyRadioDeliveryReleased.
type PortNotifyRadioDeliveryReleasedSuite struct {
	suite.Suite
	port             *PortMock
	sampleAttemptID  uuid.UUID
	sampleReleasedAt time.Time
	expectedMessage  kafkautil.OutboundMessage
}

func (suite *PortNotifyRadioDeliveryReleasedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.sampleReleasedAt = testutil.NewRandomTime()
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		Key:       suite.sampleAttemptID.String(),
		EventType: event.TypeRadioDeliveryReleased,
		Value: event.RadioDeliveryReleased{
			Attempt:    suite.sampleAttemptID,
			ReleasedAt: suite.sampleReleasedAt,
		},
	}
}

func (suite *PortNotifyRadioDeliveryReleasedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyRadioDeliveryReleased(timeout, &testutil.DBTx{}, suite.sampleAttemptID, suite.sampleReleasedAt)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyRadioDeliveryReleasedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyRadioDeliveryReleased(timeout, &testutil.DBTx{}, suite.sampleAttemptID, suite.sampleReleasedAt)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
	}()

	wait()
}

func TestPort_NotifyRadioDeliveryReleased(t *testing.T) {
	suite.Run(t, new(PortNotifyRadioDeliveryReleasedSuite))
}

// PortNotifyRadioDeliveryFinishedSuite tests Port.NotifyRadioDeliveryFinished.
type PortNotifyRadioDeliveryFinishedSuite struct {
	suite.Suite
	port                *PortMock
	sampleRadioDelivery store.RadioDelivery
	expectedMessage     kafkautil.OutboundMessage
}

func (suite *PortNotifyRadioDeliveryFinishedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleRadioDelivery = store.RadioDelivery{
		Attempt:    testutil.NewUUIDV4(),
		PickedUpBy: nulls.NewUUID(testutil.NewUUIDV4()),
		PickedUpAt: nulls.NewTime(testutil.NewRandomTime()),
		Success:    nulls.NewBool(true),
		SuccessTS:  testutil.NewRandomTime(),
		Note:       "cape",
	}
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.RadioDeliveriesTopic,
		Key:       suite.sampleRadioDelivery.Attempt.String(),
		EventType: event.TypeRadioDeliveryFinished,
		Value: event.RadioDeliveryFinished{
			Attempt:    suite.sampleRadioDelivery.Attempt,
			PickedUpBy: suite.sampleRadioDelivery.PickedUpBy,
			PickedUpAt: suite.sampleRadioDelivery.PickedUpAt,
			Success:    suite.sampleRadioDelivery.Success.Bool,
			FinishedAt: suite.sampleRadioDelivery.SuccessTS,
			Note:       suite.sampleRadioDelivery.Note,
		},
	}
}

func (suite *PortNotifyRadioDeliveryFinishedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyRadioDeliveryFinished(timeout, &testutil.DBTx{}, suite.sampleRadioDelivery)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyRadioDeliveryFinishedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyRadioDeliveryFinished(timeout, &testutil.DBTx{}, suite.sampleRadioDelivery)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
	}()

	wait()
}

func TestPort_NotifyRadioDeliveryFinished(t *testing.T) {
	suite.Run(t, new(PortNotifyRadioDeliveryFinishedSuite))
}
