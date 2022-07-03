package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
	"testing"
)

// PortNotifyGroupCreatedSuite tests Port.NotifyGroupCreated.
type PortNotifyGroupCreatedSuite struct {
	suite.Suite
	port            *PortMock
	sampleGroup     store.Group
	expectedMessage kafka.Message
}

func (suite *PortNotifyGroupCreatedSuite) SetupTest() {
	suite.port = newMockPort()
	members := make([]uuid.UUID, 16)
	for i := range members {
		members[i] = testutil.NewUUIDV4()
	}
	suite.sampleGroup = store.Group{
		ID:          testutil.NewUUIDV4(),
		Title:       "each",
		Description: "sharp",
		Operation: uuid.NullUUID{
			UUID:  testutil.NewUUIDV4(),
			Valid: true,
		},
		Members: nil,
	}
	var err error
	suite.expectedMessage, err = kafkautil.KafkaMessageFromMessage(kafkautil.Message{
		Topic:     event.GroupsTopic,
		Key:       suite.sampleGroup.ID.String(),
		EventType: event.TypeGroupCreated,
		Value: event.GroupCreated{
			ID:          suite.sampleGroup.ID,
			Title:       suite.sampleGroup.Title,
			Description: suite.sampleGroup.Description,
			Operation:   suite.sampleGroup.Operation,
			Members:     suite.sampleGroup.Members,
		},
	})
	if err != nil {
		panic(err)
	}
}

func (suite *PortNotifyGroupCreatedSuite) TestWriteFail() {
	suite.port.recorder.WriteFail = true
	err := suite.port.Port.NotifyGroupCreated(suite.sampleGroup)
	suite.Error(err, "should fail")
}

func (suite *PortNotifyGroupCreatedSuite) TestOK() {
	err := suite.port.Port.NotifyGroupCreated(suite.sampleGroup)
	suite.Require().NoError(err, "should not fail")
	suite.Equal([]kafka.Message{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct messages")
}

func TestPort_NotifyGroupCreated(t *testing.T) {
	suite.Run(t, new(PortNotifyGroupCreatedSuite))
}

// PortNotifyGroupUpdatedSuite tests Port.NotifyGroupUpdated.
type PortNotifyGroupUpdatedSuite struct {
	suite.Suite
	port            *PortMock
	sampleGroup     store.Group
	expectedMessage kafka.Message
}

func (suite *PortNotifyGroupUpdatedSuite) SetupTest() {
	suite.port = newMockPort()
	members := make([]uuid.UUID, 16)
	for i := range members {
		members[i] = testutil.NewUUIDV4()
	}
	suite.sampleGroup = store.Group{
		ID:          testutil.NewUUIDV4(),
		Title:       "each",
		Description: "sharp",
		Operation: uuid.NullUUID{
			UUID:  testutil.NewUUIDV4(),
			Valid: true,
		},
		Members: nil,
	}
	var err error
	suite.expectedMessage, err = kafkautil.KafkaMessageFromMessage(kafkautil.Message{
		Topic:     event.GroupsTopic,
		Key:       suite.sampleGroup.ID.String(),
		EventType: event.TypeGroupUpdated,
		Value: event.GroupUpdated{
			ID:          suite.sampleGroup.ID,
			Title:       suite.sampleGroup.Title,
			Description: suite.sampleGroup.Description,
			Operation:   suite.sampleGroup.Operation,
			Members:     suite.sampleGroup.Members,
		},
	})
	if err != nil {
		panic(err)
	}
}

func (suite *PortNotifyGroupUpdatedSuite) TestWriteFail() {
	suite.port.recorder.WriteFail = true
	err := suite.port.Port.NotifyGroupUpdated(suite.sampleGroup)
	suite.Error(err, "should fail")
}

func (suite *PortNotifyGroupUpdatedSuite) TestOK() {
	err := suite.port.Port.NotifyGroupUpdated(suite.sampleGroup)
	suite.Require().NoError(err, "should not fail")
	suite.Equal([]kafka.Message{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct messages")
}

func TestPort_NotifyGroupUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyGroupUpdatedSuite))
}

// PortNotifyGroupDeletedSuite tests Port.NotifyGroupDeleted.
type PortNotifyGroupDeletedSuite struct {
	suite.Suite
	port            *PortMock
	sampleGroupID   uuid.UUID
	expectedMessage kafka.Message
}

func (suite *PortNotifyGroupDeletedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleGroupID = testutil.NewUUIDV4()
	var err error
	suite.expectedMessage, err = kafkautil.KafkaMessageFromMessage(kafkautil.Message{
		Topic:     event.GroupsTopic,
		Key:       suite.sampleGroupID.String(),
		EventType: event.TypeGroupDeleted,
		Value: event.GroupDeleted{
			ID: suite.sampleGroupID,
		},
	})
	if err != nil {
		panic(err)
	}
}

func (suite *PortNotifyGroupDeletedSuite) TestWriteFail() {
	suite.port.recorder.WriteFail = true
	err := suite.port.Port.NotifyGroupDeleted(suite.sampleGroupID)
	suite.Error(err, "should fail")
}

func (suite *PortNotifyGroupDeletedSuite) TestOK() {
	err := suite.port.Port.NotifyGroupDeleted(suite.sampleGroupID)
	suite.Require().NoError(err, "should not fail")
	suite.Equal([]kafka.Message{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct messages")
}

func TestPort_NotifyGroupDeleted(t *testing.T) {
	suite.Run(t, new(PortNotifyGroupDeletedSuite))
}
