package ws

import (
	"context"
	"github.com/lefinal/zaprec"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wstest"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	"time"
)

const timeout = 5 * time.Second

// ForwardListenerMock mocks ForwardListener.
type ForwardListenerMock struct {
	mock.Mock
}

func (m *ForwardListenerMock) AcceptNewConnection(connection controller.Connection) {
	m.Called(connection)
}

// ConnListenerSuite tests ConnListener.
type ConnListenerSuite struct {
	suite.Suite
	forwardListener *ForwardListenerMock
	sampleConn      wsutil.RawConnection
	listener        wsutil.ConnListener
}

func (suite *ConnListenerSuite) SetupTest() {
	suite.forwardListener = &ForwardListenerMock{}
	suite.sampleConn = wstest.NewConnectionMock(context.Background(), auth.Token{IsAuthenticated: true})
	suite.listener = ConnListener(zap.NewNop(), suite.forwardListener)
}

func (suite *ConnListenerSuite) TestNotAuthenticated() {
	conn := wstest.NewConnectionMock(context.Background(), auth.Token{IsAuthenticated: false})
	logger, recorder := zaprec.NewRecorder(zap.ErrorLevel)
	listener := ConnListener(logger, suite.forwardListener)

	listener(conn)

	suite.NotEmpty(recorder.Records(), "should have logged error")
}

func (suite *ConnListenerSuite) TestOK() {
	suite.forwardListener.On("AcceptNewConnection", mock.Anything).Once()
	defer suite.forwardListener.AssertExpectations(suite.T())

	suite.listener(suite.sampleConn)
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
