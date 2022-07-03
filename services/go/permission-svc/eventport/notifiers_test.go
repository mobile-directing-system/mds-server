package eventport

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
	"testing"
)

// PortNotifyPermissionsUpdatedSuite tests Port.NotifyPermissionsUpdated.
type PortNotifyPermissionsUpdatedSuite struct {
	suite.Suite
	port              *PortMock
	sampleUserID      uuid.UUID
	samplePermissions []store.Permission
	expectedMessage   kafka.Message
}

func (suite *PortNotifyPermissionsUpdatedSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleUserID = testutil.NewUUIDV4()
	suite.samplePermissions = []store.Permission{
		{
			Name: "Hello",
		},
		{
			Name:    "World",
			Options: nulls.NewJSONRawMessage(json.RawMessage(`{"hello":"world"}`)),
		},
	}
	permissions := make([]permission.Permission, 0, len(suite.samplePermissions))
	for _, p := range suite.samplePermissions {
		permissions = append(permissions, permission.Permission(p))
	}
	var err error
	suite.expectedMessage, err = kafkautil.KafkaMessageFromMessage(kafkautil.Message{
		Topic:     event.PermissionsTopic,
		Key:       suite.sampleUserID.String(),
		EventType: event.TypePermissionsUpdated,
		Value: event.PermissionsUpdated{
			User:        suite.sampleUserID,
			Permissions: permissions,
		},
	})
	if err != nil {
		panic(err)
	}
}

func (suite *PortNotifyPermissionsUpdatedSuite) TestWriteFail() {
	suite.port.recorder.WriteFail = true
	err := suite.port.Port.NotifyPermissionsUpdated(suite.sampleUserID, suite.samplePermissions)
	suite.Error(err, "should fail")
}

func (suite *PortNotifyPermissionsUpdatedSuite) TestOK() {
	err := suite.port.Port.NotifyPermissionsUpdated(suite.sampleUserID, suite.samplePermissions)
	suite.Require().NoError(err, "should not fail")
	suite.Equal([]kafka.Message{suite.expectedMessage}, suite.port.recorder.Recorded, "should have written correct message")
}

func TestPort_NotifyPermissionsUpdated(t *testing.T) {
	suite.Run(t, new(PortNotifyPermissionsUpdatedSuite))
}
