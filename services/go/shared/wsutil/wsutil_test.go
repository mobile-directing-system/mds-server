package wsutil

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"math/rand"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const timeout = 5 * time.Second

func genRandomMessage() Message {
	return Message{
		Type:    MessageType(testutil.NewUUIDV4().String()),
		Payload: json.RawMessage(fmt.Sprintf(`{"hello":%d}`, rand.Int())),
	}
}

// TestWebSocket tests hub and client functionality. It sets up a test server
// and then creates WebSocket connections. Expected messages are sent and
// received.
func TestWebSocket(t *testing.T) {
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout*3)
	defer cancel()
	ginR := testutil.NewGinEngine()
	httpServer := httptest.NewServer(ginR)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	var testWG sync.WaitGroup
	// Setup client mock.
	clientToSend := make([]Message, 16)
	for i := range clientToSend {
		clientToSend[i] = genRandomMessage()
	}
	serverToSend := make([]Message, 32)
	for i := range serverToSend {
		serverToSend[i] = genRandomMessage()
	}
	clientCount := 16
	// We need to keep track of requests to keep all connections open. This is
	// because when a connection is prematurely closed while incoming/outgoing
	// messages are still being passed around because of parsing, etc., channel sends
	// will abort if the connection is already closed. We therefore create a wait
	// group with 4 expected operations per client (2 times client send/receive and 2
	// times server send/receive).
	var keepConnectionOpen sync.WaitGroup
	keepConnectionOpen.Add(clientCount * 4)
	// Setup listener.
	connListener := func(rawConn RawConnection) {
		conn := NewAutoParserConnection(rawConn)
		testWG.Add(1)
		defer testWG.Done()
		received := make([]Message, 0)
		expect := len(clientToSend)
		defer func() {
			assert.Equal(t, clientToSend, received, "should have received all messages from channel")
		}()
		// Read messages.
		var connWG sync.WaitGroup
		// Read.
		connWG.Add(1)
		go func() {
			defer keepConnectionOpen.Done()
			defer connWG.Done()
			for message := range conn.Receive() {
				received = append(received, message)
				expect--
				if expect == 0 {
					return
				}
			}
			assert.Fail(t, "early close", "receive closed before all messages where received")
		}()
		// Send.
		connWG.Add(1)
		go func() {
			defer keepConnectionOpen.Done()
			defer connWG.Done()
			for _, message := range serverToSend {
				err := conn.SendDirect(timeout, message)
				assert.NoError(t, err, "sending message should not fail")
			}
		}()
		connWG.Wait()
	}
	// Setup hub.
	hub := NewHub(timeout, zap.NewNop(), nil, connListener).(*hub)
	ginR.GET("/ws", func(c *gin.Context) {
		err := hub.UpgradeHandler()(c, auth.Token{
			IsAuthenticated: true,
		})
		assert.NoError(t, err, "upgrade-handler should not fail")
	})
	// Run each client (connect, send and receive messages).
	for i := 0; i < clientCount; i++ {
		testWG.Add(1)
		go func() {
			defer testWG.Done()
			var connWG sync.WaitGroup
			// Connect to the server
			serverConn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/ws", serverURL), nil)
			require.NoError(t, err, "should not fail")
			defer func() {
				keepConnectionOpen.Wait()
				_ = serverConn.Close()
			}()
			// Send messages.
			connWG.Add(1)
			go func() {
				defer keepConnectionOpen.Done()
				defer connWG.Done()
				for _, message := range clientToSend {
					err := serverConn.WriteJSON(message)
					assert.NoError(t, err, "writing message on server connection should not fail")
				}
			}()
			// Read messages.
			connWG.Add(1)
			go func() {
				defer keepConnectionOpen.Done()
				defer connWG.Done()
				for _, messageToSend := range serverToSend {
					_, messageRaw, err := serverConn.ReadMessage()
					require.NoError(t, err, "reading message should not fail")
					var message Message
					err = json.Unmarshal(messageRaw, &message)
					require.NoError(t, err, "server should send valid message")
					assert.Equal(t, messageToSend, message, "server should send expected message")
				}
			}()
			connWG.Wait()
		}()
	}
	// Wait until all connections are done.
	testWG.Wait()
	cancel()
	wait()
}

func TestConnection_ParallelReceive(t *testing.T) {
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	l, _ := zap.NewDevelopment()
	conn := NewClient(timeout, l, auth.Token{}, nil).RawConnection().(*rawConnection)
	var wg sync.WaitGroup
	workersProcessing := 0
	received := 0
	block := true
	cond := sync.Cond{L: &sync.Mutex{}}

	const workerCount = 16
	const messagesToSend = workerCount * 1024
	for workerNum := 0; workerNum < workerCount; workerNum++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range conn.ReceiveRaw() {
				// Wait until all are processing, then wait until all are done.
				cond.L.Lock()
				workersProcessing++
				cond.Broadcast()
				for block && workersProcessing < workerCount {
					select {
					case <-timeout.Done():
						assert.Fail(t, "timeout while waiting for all workers processing")
						cond.L.Unlock()
						return
					default:
					}
					cond.Wait()
				}
				block = false
				received++
				workersProcessing--
				cond.Broadcast()
				for !block && workersProcessing > 0 {
					select {
					case <-timeout.Done():
						assert.Fail(t, "timeout while waiting for all workers idle")
						cond.L.Unlock()
						return
					default:
					}
					cond.Wait()
				}
				block = true
				cond.Broadcast()
				cond.L.Unlock()
			}
		}()
	}

	go func() {
		defer cancel()
		defer cond.Broadcast()
		for i := 0; i < messagesToSend; i++ {
			select {
			case <-timeout.Done():
				return
			case conn.receive <- json.RawMessage(`{}`):
			}
		}
		close(conn.receive)
		wg.Wait()
		assert.Equal(t, messagesToSend, received, "should have received all messages")
	}()

	wait()
	wg.Wait()
}

type myStruct struct {
	An int `json:"an"`
}

// ParseAndHandleSuite tests ParseAndHandle.
type ParseAndHandleSuite struct {
	suite.Suite
}

func (suite *ParseAndHandleSuite) TestParseFail() {
	err := ParseAndHandle(Message{
		Payload: json.RawMessage(`{invalid`),
	}, func(_ myStruct) error {
		suite.Fail("should not have called handler")
		return nil
	})
	suite.Require().Error(err, "should fail")
	suite.Equal(meh.ErrBadInput, meh.ErrorCode(err), "should return correct error code")
}

func (suite *ParseAndHandleSuite) TestHandleFail() {
	ec := meh.Code(testutil.NewUUIDV4().String())
	err := ParseAndHandle(Message{
		Payload: json.RawMessage(`{"an":2}`),
	}, func(_ myStruct) error {
		return meh.NewErr(ec, "", nil)
	})
	suite.Require().Error(err, "should fail")
	suite.Equal(ec, meh.ErrorCode(err), "should return correct error code")
}

func (suite *ParseAndHandleSuite) TestOK() {
	called := false
	err := ParseAndHandle(Message{
		Payload: json.RawMessage(`{"an":2}`),
	}, func(s myStruct) error {
		suite.Equal(myStruct{An: 2}, s, "should call handler with correct payload")
		called = true
		return nil
	})
	suite.NoError(err, "should not fail")
	suite.True(called, "should have called handler")
}

func TestParseAndHandle(t *testing.T) {
	suite.Run(t, new(ParseAndHandleSuite))
}
