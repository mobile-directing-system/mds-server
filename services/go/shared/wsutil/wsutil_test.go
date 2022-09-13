package wsutil

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	defer cancel()
	ginR := testutil.NewGinEngine()
	httpServer := httptest.NewServer(ginR)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	var testWG sync.WaitGroup
	// Setup client mock.
	clientToSend := make([]Message, 128)
	for i := range clientToSend {
		clientToSend[i] = genRandomMessage()
	}
	serverToSend := make([]Message, 128)
	for i := range serverToSend {
		serverToSend[i] = genRandomMessage()
	}
	clientCount := 128
	// Setup listener.
	connListener := func(conn Connection) {
		testWG.Add(1)
		defer testWG.Done()
		received := make([]Message, 0)
		defer func() {
			assert.Equal(t, clientToSend, received, "should have received all messages from channel")
		}()
		expect := len(clientToSend)
		// Read messages.
		var connWG sync.WaitGroup
		// Read.
		connWG.Add(1)
		go func() {
			defer connWG.Done()
			for message := range conn.Receive() {
				received = append(received, message)
				expect--
				if expect == 0 {
					return
				}
			}
		}()
		// Send.
		connWG.Add(1)
		go func() {
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
			defer func() { _ = serverConn.Close() }()
			// Send messages.
			connWG.Add(1)
			go func() {
				defer connWG.Done()
				for _, message := range clientToSend {
					err := serverConn.WriteJSON(message)
					assert.NoError(t, err, "writing message on server connection should not fail")
				}
			}()
			// Read messages.
			connWG.Add(1)
			go func() {
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
