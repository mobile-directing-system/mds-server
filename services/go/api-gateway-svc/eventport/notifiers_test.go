package eventport

import (
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
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
	expectedMessage       kafkautil.OutboundMessage
}

func (suite *PortNotifyUserLoggedInSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleUserID = testutil.NewUUIDV4()
	suite.sampleUsername = "grow"
	suite.sampleRequestMetadata = controller.AuthRequestMetadata{
		Host:       "wheel",
		UserAgent:  "puzzle",
		RemoteAddr: "desert",
	}
	suite.expectedMessage = kafkautil.OutboundMessage{
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
	}
}

func (suite *PortNotifyUserLoggedInSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyUserLoggedIn(timeout, &testutil.DBTx{}, suite.sampleUserID, suite.sampleUsername, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyUserLoggedInSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyUserLoggedIn(timeout, &testutil.DBTx{}, suite.sampleUserID, suite.sampleUsername, suite.sampleRequestMetadata)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should have written correct message")
	}()

	wait()
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
	expectedMessage       kafkautil.OutboundMessage
}

func (suite *PortNotifyUserLoggedOutSuite) SetupTest() {
	suite.port = newMockPort()
	suite.sampleUserID = testutil.NewUUIDV4()
	suite.sampleUsername = "kick"
	suite.sampleRequestMetadata = controller.AuthRequestMetadata{
		Host:       "pause",
		UserAgent:  "parent",
		RemoteAddr: "against",
	}
	suite.expectedMessage = kafkautil.OutboundMessage{
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
	}
}

func (suite *PortNotifyUserLoggedOutSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.port.recorder.WriteFail = true

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyUserLoggedOut(timeout, &testutil.DBTx{}, suite.sampleUserID, suite.sampleUsername, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *PortNotifyUserLoggedOutSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)

	go func() {
		defer cancel()
		err := suite.port.Port.NotifyUserLoggedOut(timeout, &testutil.DBTx{}, suite.sampleUserID, suite.sampleUsername, suite.sampleRequestMetadata)
		suite.Require().NoError(err, "should not fail")
		suite.Equal([]kafkautil.OutboundMessage{suite.expectedMessage}, suite.port.recorder.Recorded, "should have written correct message")
	}()

	wait()
}

func TestPort_NotifyUserLoggedOut(t *testing.T) {
	suite.Run(t, new(PortNotifyUserLoggedOutSuite))
}
