package wstest

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"sync"
)

// Connection mocks wsutil.Connection for testing purposes.
type Connection struct {
	lifetime   context.Context
	disconnect context.CancelFunc
	authToken  auth.Token
	// SendFail lets Send and SendDirect return an error.
	SendFail bool
	// receive is a message queue for returned messages in Receive.
	receive chan wsutil.Message
	// outbox holds all messages from Send and SendDirect where sending was
	// successful.
	outbox []wsutil.Message
	// outboxMutex locks outbox
	outboxMutex sync.RWMutex
}

// NewConnectionMock creates a new Connection with the given auth.Token. The
// context is used as parent context for the connection lifetime that can be
// checked via Connection.Done. If the connection should be disconnected
// manually, call Connection.Disconnect.
func NewConnectionMock(lifetime context.Context, authToken auth.Token) *Connection {
	lifetime, disconnect := context.WithCancel(lifetime)
	return &Connection{
		lifetime:   lifetime,
		disconnect: disconnect,
		authToken:  authToken,
		SendFail:   false,
		receive:    make(chan wsutil.Message),
		outbox:     make([]wsutil.Message, 0),
	}
}

// NextReceive passes the given wsutil.Message to the channel from Receive.
func (m *Connection) NextReceive(ctx context.Context, message wsutil.Message) {
	select {
	case <-ctx.Done():
		return
	case m.receive <- message:
	}
}

// Outbox returns the wsutil.Message list of messages that were sent.
func (m *Connection) Outbox() []wsutil.Message {
	m.outboxMutex.RLock()
	defer m.outboxMutex.RUnlock()
	outboxCopy := make([]wsutil.Message, 0, len(m.outbox))
	for _, message := range m.outbox {
		outboxCopy = append(outboxCopy, message)
	}
	return outboxCopy
}

// AuthToken returns the set auth.Token.
func (m *Connection) AuthToken() auth.Token {
	return m.authToken
}

// Receive returns the receive-channel for incoming wsutil.Message.
func (m *Connection) Receive() <-chan wsutil.Message {
	return m.receive
}

// Send marshals the given payload and calls SendDirect.
func (m *Connection) Send(ctx context.Context, messageType wsutil.MessageType, payload any) error {
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
func (m *Connection) SendDirect(_ context.Context, message wsutil.Message) error {
	if m.SendFail {
		return errors.New("send fail")
	}
	m.outboxMutex.Lock()
	defer m.outboxMutex.Unlock()
	m.outbox = append(m.outbox, message)
	return nil
}

// Disconnect makes Done receive.
func (m *Connection) Disconnect() {
	m.disconnect()
}

// Done receives when the connection is done.
func (m *Connection) Done() <-chan struct{} {
	return m.lifetime.Done()
}
