package wstest

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

// RawConnection mocks wsutil.Connection for testing purposes. If a call to
// OutboxChan was made, messages to send will also be forwarded to the returned
// channel. Otherwise, they will only be added to the Outbox.
type RawConnection struct {
	id         uuid.UUID
	lifetime   context.Context
	disconnect context.CancelFunc
	authToken  auth.Token
	// SendFail lets Send and SendDirect return an error.
	SendFail bool
	// receive is a message queue for returned messages in Receive.
	receive chan json.RawMessage
	// outbox holds all messages from Send and SendDirect where sending was
	// successful.
	outbox []wsutil.Message
	// outboxChan is the channel to send outgoing messages to, if a call to
	// ReceiveRaw was made.
	outboxChan chan wsutil.Message
	// forwardToOutboxChan describes whether outgoing messages should also be routed
	// to outboxChan.
	forwardToOutboxChan *atomic.Bool
	// outboxMutex locks outbox
	outboxMutex sync.RWMutex
}

// NewConnectionMock creates a new RawConnection with the given auth.Token. The
// context is used as parent context for the connection lifetime that can be
// checked via RawConnection.Done. If the connection should be disconnected
// manually, call RawConnection.Disconnect.
func NewConnectionMock(lifetime context.Context, authToken auth.Token) *RawConnection {
	lifetime, disconnect := context.WithCancel(lifetime)
	id, _ := uuid.NewV4()
	return &RawConnection{
		id:                  id,
		lifetime:            lifetime,
		disconnect:          disconnect,
		authToken:           authToken,
		SendFail:            false,
		receive:             make(chan json.RawMessage),
		outbox:              make([]wsutil.Message, 0),
		outboxChan:          make(chan wsutil.Message),
		forwardToOutboxChan: atomic.NewBool(false),
	}
}

// SetPermissions overwrites the permissions in the held auth.Token.
func (m *RawConnection) SetPermissions(newPermissions []permission.Permission) {
	m.authToken.Permissions = newPermissions
}

// NextReceive marshals and passes the given wsutil.Message to the channel from
// ReceiveRaw.
func (m *RawConnection) NextReceive(ctx context.Context, messageType wsutil.MessageType, payload any) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	raw, err := json.Marshal(wsutil.Message{
		Type:    messageType,
		Payload: rawPayload,
	})
	if err != nil {
		panic(err)
	}
	select {
	case <-ctx.Done():
		return
	case m.receive <- raw:
	}
}

// Outbox returns the wsutil.Message list of messages that were sent.
func (m *RawConnection) Outbox() []wsutil.Message {
	m.outboxMutex.RLock()
	defer m.outboxMutex.RUnlock()
	outboxCopy := make([]wsutil.Message, 0, len(m.outbox))
	for _, message := range m.outbox {
		outboxCopy = append(outboxCopy, message)
	}
	return outboxCopy
}

// OutboxChan enables forwarding of outbox messages and returns the channel that
// outbox messages are being forwarded to.
func (m *RawConnection) OutboxChan() <-chan wsutil.Message {
	m.forwardToOutboxChan.Store(true)
	return m.outboxChan
}

// ID returns the id of the connection.
func (m *RawConnection) ID() uuid.UUID {
	return m.id
}

// AuthToken returns the set auth.Token.
func (m *RawConnection) AuthToken() auth.Token {
	return m.authToken
}

// Send marshals the given payload and calls SendDirect.
func (m *RawConnection) Send(ctx context.Context, messageType wsutil.MessageType, payload any) error {
	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "marshal payload", nil)
	}
	return m.SendDirect(ctx, wsutil.Message{
		Type:    messageType,
		Payload: payloadRaw,
	})
}

// SendDirect adds the given wsutil.Message to the outbox if SendFail is not set.
func (m *RawConnection) SendDirect(ctx context.Context, message wsutil.Message) error {
	if m.SendFail {
		return errors.New("send fail")
	}
	m.outboxMutex.Lock()
	defer m.outboxMutex.Unlock()
	m.outbox = append(m.outbox, message)
	if m.forwardToOutboxChan.Load() {
		select {
		case <-ctx.Done():
			return errors.New("context done")
		case m.outboxChan <- message:
		}
	}
	return nil
}

// Disconnect makes Done receive.
func (m *RawConnection) Disconnect() {
	m.disconnect()
}

// Lifetime of the connection.
func (m *RawConnection) Lifetime() context.Context {
	return m.lifetime
}

// Logger returns a zap.NewNop-logger.
func (m *RawConnection) Logger() *zap.Logger {
	return zap.NewNop()
}

// SendErr sends the given error via Send.
func (m *RawConnection) SendErr(ctx context.Context, err error) {
	_ = m.Send(ctx, wsutil.TypeError, wsutil.ErrorMessageFromErr(err))
}

// SendRaw is not implemented.
func (m *RawConnection) SendRaw() chan<- json.RawMessage {
	panic("implement me")
}

// ReceiveRaw is not implemented.
func (m *RawConnection) ReceiveRaw() <-chan json.RawMessage {
	return m.receive
}

// HubMock mocks wsutil.Hub.
type HubMock struct {
	// UpgradeCalled is the amount of times the upgrade handler was called.
	UpgradeCalled int
	// UpgradeFail describes whether upgrade should succeed (return http.StatusOK) or
	// fail (return http.StatusInternalServerError).
	UpgradeFail bool
}

// UpgradeHandler increments UpgradeCalled and returns an error if UpgradeFail is
// set to true. Otherwise, responds with http.StatusOK..
func (m *HubMock) UpgradeHandler() httpendpoints.HandlerFunc {
	return func(c *gin.Context, _ auth.Token) error {
		m.UpgradeCalled++
		if m.UpgradeFail {
			return errors.New("sad life")
		}
		c.Status(http.StatusOK)
		return nil
	}
}
