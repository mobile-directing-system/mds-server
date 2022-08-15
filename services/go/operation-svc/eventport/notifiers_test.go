package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// PortNotifyOperationCreatedSuite tests Port.NotifyOperationCreated.
type PortNotifyOperationCreatedSuite struct {
	suite.Suite
	port            *PortMock
	sampleOperation store.Operation
	expectedMessage kafkautil.OutboundMessage
}

func (suite *PortNotifyOperationCreatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleOperation = store.Operation{
		ID:          testutil.NewUUIDV4(),
		Title:       "win",
		Description: "compose",
		Start:       time.UnixMilli(5),
		End:         nulls.NewTime(time.UnixMilli(2202)),
		IsArchived:  true,
	}
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.OperationsTopic,
		Key:       suite.sampleOperation.ID.String(),
		EventType: event.TypeOperationCreated,
		Value: event.OperationCreated{
			ID:          suite.sampleOperation.ID,
			Title:       suite.sampleOperation.Title,
			Description: suite.sampleOperation.Description,
			Start:       suite.sampleOperation.Start,
			End:         suite.sampleOperation.End,
			IsArchived:  suite.sampleOperation.IsArchived,
		},
	}
}

func (suite *PortNotifyOperationCreatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyOperationCreated(timeout, tx, suite.sampleOperation)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyOperationCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyOperationCreated(timeout, tx, suite.sampleOperation)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
	}()

	wait()
}

func TestPort_NotifyOperationCreated(t *testing.T) {
	suite.Run(t, new(PortNotifyOperationCreatedSuite))
}

// PortNotifyOperationUpdatedSuite tests Port.NotifyOperationUpdated.
type PortNotifyOperationUpdatedSuite struct {
	suite.Suite
	port            *PortMock
	sampleOperation store.Operation
	expectedMessage kafkautil.OutboundMessage
}

func (suite *PortNotifyOperationUpdatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleOperation = store.Operation{
		ID:          testutil.NewUUIDV4(),
		Title:       "win",
		Description: "compose",
		Start:       time.UnixMilli(5),
		End:         nulls.NewTime(time.UnixMilli(2202)),
		IsArchived:  true,
	}
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.OperationsTopic,
		Key:       suite.sampleOperation.ID.String(),
		EventType: event.TypeOperationUpdated,
		Value: event.OperationUpdated{
			ID:          suite.sampleOperation.ID,
			Title:       suite.sampleOperation.Title,
			Description: suite.sampleOperation.Description,
			Start:       suite.sampleOperation.Start,
			End:         suite.sampleOperation.End,
			IsArchived:  suite.sampleOperation.IsArchived,
		},
	}
}

func (suite *PortNotifyOperationUpdatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyOperationUpdated(timeout, tx, suite.sampleOperation)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyOperationUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyOperationUpdated(timeout, tx, suite.sampleOperation)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
	}()

	wait()
}

func TestPort_NotifyOperationUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyOperationUpdatedSuite))
}

// PortNotifyOperationMembersUpdatedSuite tests
// Port.NotifyOperationMembersUpdated.
type PortNotifyOperationMembersUpdatedSuite struct {
	suite.Suite
	port              *PortMock
	sampleOperationID uuid.UUID
	sampleMembers     []uuid.UUID
	expectedMessage   kafkautil.OutboundMessage
}

func (suite *PortNotifyOperationMembersUpdatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleOperationID = testutil.NewUUIDV4()
	suite.sampleMembers = make([]uuid.UUID, 16)
	for i := range suite.sampleMembers {
		suite.sampleMembers[i] = testutil.NewUUIDV4()
	}
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.OperationsTopic,
		Key:       suite.sampleOperationID.String(),
		EventType: event.TypeOperationMembersUpdated,
		Value: event.OperationMembersUpdated{
			Operation: suite.sampleOperationID,
			Members:   suite.sampleMembers,
		},
	}
}

func (suite *PortNotifyOperationMembersUpdatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyOperationMembersUpdated(timeout, tx, suite.sampleOperationID, suite.sampleMembers)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyOperationMembersUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyOperationMembersUpdated(timeout, tx, suite.sampleOperationID, suite.sampleMembers)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
	}()

	wait()
}

func TestPort_NotifyOperationMembersUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyOperationMembersUpdatedSuite))
}
