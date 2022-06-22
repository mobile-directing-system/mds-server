package eventport

import (
	"github.com/google/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
	"testing"
)

// PortNotifyUserLoggedInSuite tests Port.NotifyUserLoggedIn.
type PortNotifyUserLoggedInSuite struct {
	suite.Suite
	port                  *PortMock
	sampleUserID          uuid.UUID
	sampleUsername        string
	sampleRequestMetadata controller.AuthRequestMetadata
	expectedMessage       kafka.Message
}

func (suite *PortNotifyUserLoggedInSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleUserID = uuid.New()
	suite.sampleUsername = "grow"
	suite.sampleRequestMetadata = controller.AuthRequestMetadata{
		Host:       "wheel",
		UserAgent:  "puzzle",
		RemoteAddr: "desert",
	}
	var err error
	suite.expectedMessage, err = kafkautil.KafkaMessageFromMessage(kafkautil.Message{
		Topic:     event.AuthTopic,
		Key:       suite.sampleUserID.String(),
		EventType: event.TypeUserLoggedIn,
		Value: event.UserLoggedIn{
			User:       suite.sampleUserID,
			Username:   suite.sampleUsername,
			Host:       suite.sampleRequestMetadata.Host,
			UserAgent:  suite.sampleRequestMetadata.UserAgent,
			RemoteAddr: suite.sampleRequestMetadata.RemoteAddr,
		},
	})
	if err != nil {
		panic(err)
	}
}

func (suite *PortNotifyUserLoggedInSuite) TestWriteFail() {
	suite.port.recorder.WriteFail = true
	err := suite.port.Port.NotifyUserLoggedIn(suite.sampleUserID, suite.sampleUsername, suite.sampleRequestMetadata)
	suite.Error(err, "should fail")
}

func (suite *PortNotifyUserLoggedInSuite) TestOK() {
	err := suite.port.Port.NotifyUserLoggedIn(suite.sampleUserID, suite.sampleUsername, suite.sampleRequestMetadata)
	suite.Require().NoError(err, "should not fail")
	suite.Equal([]kafka.Message{suite.expectedMessage}, suite.port.recorder.Recorded, "should have written correct message")
}

func TestPort_NotifyUserLoggedIn(t *testing.T) {
	suite.Run(t, new(PortNotifyUserLoggedInSuite))
}

// PortNotifyUserLoggedOutSuite tests Port.NotifyUserLoggedOut.
type PortNotifyUserLoggedOutSuite struct {
	suite.Suite
	port                  *PortMock
	sampleUserID          uuid.UUID
	sampleUsername        string
	sampleRequestMetadata controller.AuthRequestMetadata
	expectedMessage       kafka.Message
}

func (suite *PortNotifyUserLoggedOutSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleUserID = uuid.New()
	suite.sampleUsername = "kick"
	suite.sampleRequestMetadata = controller.AuthRequestMetadata{
		Host:       "pause",
		UserAgent:  "parent",
		RemoteAddr: "against",
	}
	var err error
	suite.expectedMessage, err = kafkautil.KafkaMessageFromMessage(kafkautil.Message{
		Topic:     event.AuthTopic,
		Key:       suite.sampleUserID.String(),
		EventType: event.TypeUserLoggedOut,
		Value: event.UserLoggedOut{
			User:       suite.sampleUserID,
			Username:   suite.sampleUsername,
			Host:       suite.sampleRequestMetadata.Host,
			UserAgent:  suite.sampleRequestMetadata.UserAgent,
			RemoteAddr: suite.sampleRequestMetadata.RemoteAddr,
		},
	})
	if err != nil {
		panic(err)
	}
}

func (suite *PortNotifyUserLoggedOutSuite) TestWriteFail() {
	suite.port.recorder.WriteFail = true
	err := suite.port.Port.NotifyUserLoggedOut(suite.sampleUserID, suite.sampleUsername, suite.sampleRequestMetadata)
	suite.Error(err, "should fail")
}

func (suite *PortNotifyUserLoggedOutSuite) TestOK() {
	err := suite.port.Port.NotifyUserLoggedOut(suite.sampleUserID, suite.sampleUsername, suite.sampleRequestMetadata)
	suite.Require().NoError(err, "should not fail")
	suite.Equal([]kafka.Message{suite.expectedMessage}, suite.port.recorder.Recorded, "should have written correct message")
}

func TestPort_NotifyUserLoggedOut(t *testing.T) {
	suite.Run(t, new(PortNotifyUserLoggedOutSuite))
}
