package eventport

import (
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// PortNotifyOperationCreatedSuite tests Port.NotifyOperationCreated.
type PortNotifyOperationCreatedSuite struct {
	suite.Suite
	port            *PortMock
	sampleOperation store.Operation
	expectedMessage kafka.Message
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
	var err error
	suite.expectedMessage, err = kafkautil.KafkaMessageFromMessage(kafkautil.Message{
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
	})
	if err != nil {
		panic(err)
	}
}

func (suite *PortNotifyOperationCreatedSuite) TestWriteFail() {
	suite.port.recorder.WriteFail = true
	err := suite.port.Port.NotifyOperationCreated(suite.sampleOperation)
	suite.Error(err, "should fail")
}

func (suite *PortNotifyOperationCreatedSuite) TestOK() {
	err := suite.port.Port.NotifyOperationCreated(suite.sampleOperation)
	suite.Require().NoError(err, "should not fail")
	suite.Equal([]kafka.Message{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
}

func TestPort_NotifyOperationCreated(t *testing.T) {
	suite.Run(t, new(PortNotifyOperationCreatedSuite))
}

// PortNotifyOperationUpdatedSuite tests Port.NotifyOperationUpdated.
type PortNotifyOperationUpdatedSuite struct {
	suite.Suite
	port            *PortMock
	sampleOperation store.Operation
	expectedMessage kafka.Message
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
	var err error
	suite.expectedMessage, err = kafkautil.KafkaMessageFromMessage(kafkautil.Message{
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
	})
	if err != nil {
		panic(err)
	}
}

func (suite *PortNotifyOperationUpdatedSuite) TestWriteFail() {
	suite.port.recorder.WriteFail = true
	err := suite.port.Port.NotifyOperationUpdated(suite.sampleOperation)
	suite.Error(err, "should fail")
}

func (suite *PortNotifyOperationUpdatedSuite) TestOK() {
	err := suite.port.Port.NotifyOperationUpdated(suite.sampleOperation)
	suite.Require().NoError(err, "should not fail")
	suite.Equal([]kafka.Message{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct message")
}

func TestPort_NotifyOperationUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyOperationUpdatedSuite))
}
