package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// PortNotifyGroupCreatedSuite tests Port.NotifyGroupCreated.
type PortNotifyGroupCreatedSuite struct {
	suite.Suite
	port            *PortMock
	sampleGroup     store.Group
	expectedMessage kafkautil.OutboundMessage
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
	suite.expectedMessage = kafkautil.OutboundMessage{
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
	}
}

func (suite *PortNotifyGroupCreatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyGroupCreated(timeout, &testutil.DBTx{}, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyGroupCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyGroupCreated(timeout, &testutil.DBTx{}, suite.sampleGroup)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyGroupCreated(t *testing.T) {
	suite.Run(t, new(PortNotifyGroupCreatedSuite))
}

// PortNotifyGroupUpdatedSuite tests Port.NotifyGroupUpdated.
type PortNotifyGroupUpdatedSuite struct {
	suite.Suite
	port            *PortMock
	sampleGroup     store.Group
	expectedMessage kafkautil.OutboundMessage
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
	suite.expectedMessage = kafkautil.OutboundMessage{
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
	}
}

func (suite *PortNotifyGroupUpdatedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyGroupUpdated(timeout, &testutil.DBTx{}, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyGroupUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyGroupUpdated(timeout, &testutil.DBTx{}, suite.sampleGroup)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyGroupUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyGroupUpdatedSuite))
}

// PortNotifyGroupDeletedSuite tests Port.NotifyGroupDeleted.
type PortNotifyGroupDeletedSuite struct {
	suite.Suite
	port            *PortMock
	sampleGroupID   uuid.UUID
	expectedMessage kafkautil.OutboundMessage
}

func (suite *PortNotifyGroupDeletedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleGroupID = testutil.NewUUIDV4()
	suite.expectedMessage = kafkautil.OutboundMessage{
		Topic:     event.GroupsTopic,
		Key:       suite.sampleGroupID.String(),
		EventType: event.TypeGroupDeleted,
		Value: event.GroupDeleted{
			ID: suite.sampleGroupID,
		},
	}
}

func (suite *PortNotifyGroupDeletedSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		suite.port.recorder.WriteFail = true
		err := suite.port.Port.NotifyGroupDeleted(timeout, &testutil.DBTx{}, suite.sampleGroupID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyGroupDeletedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyGroupDeleted(timeout, &testutil.DBTx{}, suite.sampleGroupID)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should write correct messages")
	}()

	wait()
}

func TestPort_NotifyGroupDeleted(t *testing.T) {
	suite.Run(t, new(PortNotifyGroupDeletedSuite))
}
