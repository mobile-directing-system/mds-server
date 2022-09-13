package wsutil

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"go.uber.org/zap"
	"time"
)

// Hub provides an UpgradeHandler that accepts websocket connections. Create one
// using NewHub.
type Hub interface {
	// UpgradeHandler is the httpendpoints.HandlerFunc for upgrading websocket
	// connections.
	UpgradeHandler() httpendpoints.HandlerFunc
}

// MessageType of Message.
type MessageType string

// Message is a container for a Type and Payload.
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ConnListener for accepting new connections.
type ConnListener func(conn Connection)

// Gatekeeper is used for checking whether WebSocket upgrade is permitted. If no
// error is returned, the connection upgrade is permitted.
type Gatekeeper func(token auth.Token) error

// hub is the implementation of Hub.
type hub struct {
	// connLifetime is the lifetime context for all accepted connections via
	// acceptNewWSConn.
	connLifetime context.Context
	// logger for all WebSocket related activities.
	logger *zap.Logger
	// upgrader performs the WebSocket upgrade.
	upgrader websocket.Upgrader
	// gatekeeper that will be called on new connection requests.
	gatekeeper Gatekeeper
	// connListener is called in acceptNewWSConn when a new client is connected.
	connListener ConnListener
}

// NewHub creates a new Hub. Use Hub.UpgradeHandler for accepting WebSocket
// connections. The given context is used as a lifetime for all connections. If
// the context is done, all active clients will be disconnected.
func NewHub(connLifetime context.Context, logger *zap.Logger, gatekeeper Gatekeeper, connListener ConnListener) Hub {
	return &hub{
		connLifetime: connLifetime,
		logger:       logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:   ReadBufferSize,
			WriteBufferSize:  WriteBufferSize,
			HandshakeTimeout: 2 * time.Second,
		},
		gatekeeper:   gatekeeper,
		connListener: connListener,
	}
}

// acceptNewWSConn handles the given WebSocket connection. It blocks until the
// connection is no longer needed.
func (h *hub) acceptNewWSConn(authToken auth.Token, wsConn *websocket.Conn) error {
	id, err := uuid.NewV4()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "new uuidv4", nil)
	}
	c := NewClient(h.connLifetime, h.logger.Named("client").Named(id.String()), authToken, wsConn)
	if h.connListener != nil {
		go h.connListener(c.Connection())
	}
	err = c.RunAndClose()
	if err != nil {
		return meh.Wrap(err, "run and close", meh.Details{"client_id": id})
	}
	return nil
}

// UpgradeHandler is the httpendpoints.HandlerFunc for upgrading websocket
// connections.
func (h *hub) UpgradeHandler() httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		if h.gatekeeper != nil {
			err := h.gatekeeper(token)
			if err != nil {
				return meh.Wrap(err, "call gatekeeper", nil)
			}
		}
		// Upgrade the client's connection.
		wsConn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return meh.NewErrFromErr(err, ErrWSCommunication, "upgrade client connection", nil)
		}
		defer func() {
			_ = wsConn.Close()
		}()
		// Handle.
		err = h.acceptNewWSConn(token, wsConn)
		if err != nil {
			mehlog.Log(h.logger, meh.Wrap(err, "accept new websocket connection", nil))
		}
		return nil
	}
}
