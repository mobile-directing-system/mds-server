package wsutil

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"time"
)

// ErrWSCommunication is used for all errors regarding raw WebSocket
// communication.
const ErrWSCommunication meh.Code = "__mds-server_shared_ws_ws-communication"

func init() {
	logging.AddToDefaultLevelTranslator(ErrWSCommunication, zap.DebugLevel)
}

// Client is an abstraction for a raw WebSocket connection including
// lifesupport.
type Client interface {
	// RunAndClose runs lifesupport for the connection and closes it after the
	// connection lifetime is done.
	RunAndClose() error
	Run() error
	RawConnection() RawConnection
	Close()
}

// client is a container for an actual WebSocket connection and the internal
// processing connection.
type client struct {
	wsConn *websocket.Conn
	conn   *rawConnection
}

// NewClient creates a new Client. Do not forget to call Client.RunAndClose!
func NewClient(lifetime context.Context, logger *zap.Logger, authToken auth.Token, wsConn *websocket.Conn) Client {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	if logger == nil {
		logger = logging.DebugLogger().Named("tmp").Named("ws-conn").Named(id.String())
	}
	connLifetime, cancel := context.WithCancel(lifetime)
	return &client{
		wsConn: wsConn,
		conn: &rawConnection{
			id:        id,
			lifetime:  connLifetime,
			cancel:    cancel,
			authToken: authToken,
			logger:    logger,
			receive:   make(chan json.RawMessage, 0),
			send:      make(chan json.RawMessage, 0),
		},
	}
}

func (c *client) Close() {
	c.conn.cancel()
	err := c.wsConn.Close()
	if err != nil {
		mehlog.Log(c.conn.logger, meh.NewErrFromErr(err, ErrWSCommunication, "close connection", nil))
	}
}

// Run runs WebSocket lifesupport and returns when no lifesupport is needed
// anymore, e.g., the set lifetime context is done.
func (c *client) Run() error {
	c.wsConn.SetReadLimit(MaxMessageSize)
	err := pongHandler(c.wsConn)("")
	if err != nil {
		return meh.Wrap(err, "initial pong-handler", nil)
	}
	c.wsConn.SetPongHandler(pongHandler(c.wsConn))
	eg, egCtx := errgroup.WithContext(c.conn.lifetime)
	eg.Go(func() error {
		defer close(c.conn.receive)
		err := c.readPump(egCtx)
		if err != nil {
			return meh.Wrap(err, "read pump", nil)
		}
		return nil
	})
	eg.Go(func() error {
		err := c.writePump(egCtx)
		if err != nil {
			return meh.Wrap(err, "write pump", nil)
		}
		return nil
	})
	return eg.Wait()
}

// RunAndClose runs WebSocket lifesupport and closes after the set lifetime
// context is done.
func (c *client) RunAndClose() error {
	defer c.Close()
	err := c.Run()
	if err != nil {
		return meh.Wrap(err, "run", nil)
	}
	return nil
}

func (c *client) Connection() BaseConnection {
	return c.conn
}

func (c *client) RawConnection() RawConnection {
	return c.conn
}

// readPump reads from the wsConn and forwards to the receive-channel.
func (c *client) readPump(ctx context.Context) error {
	for {
		_, message, err := c.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return meh.NewErrFromErr(err, ErrWSCommunication, "unexpected close", nil)
			}
			return meh.NewErrFromErr(err, ErrWSCommunication, "regular close", nil)
		}
		message = bytes.TrimSpace(bytes.Replace(message, NewLine, Space, -1))
		select {
		case <-ctx.Done():
			return nil
		case c.conn.receive <- message:
		}
	}
}

// writePump sends messages from send to the wsConn until the given context is
// done.
func (c *client) writePump(ctx context.Context) error {
	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()
	resetWriteDeadline := func() error {
		err := c.wsConn.SetWriteDeadline(time.Now().Add(WriteWait))
		if err != nil {
			return meh.NewErrFromErr(err, ErrWSCommunication, "set write-deadline",
				meh.Details{"deadline": WriteWait})
		}
		return nil
	}
	for {
		err := resetWriteDeadline()
		if err != nil {
			return meh.Wrap(err, "reset write-deadline", nil)
		}
		if err := c.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return meh.NewErrFromErr(err, ErrWSCommunication, "write ping-message", nil)
		}
		select {
		case <-ctx.Done():
			return nil
		case message, more := <-c.conn.send:
			if !more {
				return nil
			}
			err := resetWriteDeadline()
			if err != nil {
				return meh.Wrap(err, "reset write-deadline", nil)
			}
			// Send message.
			err = c.wsConn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return meh.NewErrFromErr(err, ErrWSCommunication, "write message", nil)
			}
		case <-ticker.C:
			// Loop over.
		}
	}
}

func pongHandler(conn *websocket.Conn) func(string) error {
	return func(_ string) error {
		err := conn.SetReadDeadline(time.Now().Add(PongWait))
		if err != nil {
			return meh.NewErrFromErr(err, ErrWSCommunication, "set read-deadline", meh.Details{"deadline": PongWait})
		}
		return nil
	}
}
