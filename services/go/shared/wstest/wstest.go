package wstest

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"go.uber.org/zap"
	"sync"
)

// RawConnection mocks wsutil.Connection for testing purposes.
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
		id:         id,
		lifetime:   lifetime,
		disconnect: disconnect,
		authToken:  authToken,
		SendFail:   false,
		receive:    make(chan json.RawMessage),
		outbox:     make([]wsutil.Message, 0),
	}
}

// SetPermissions overwrites the permissions in the held auth.Token.
func (m *RawConnection) SetPermissions(newPermissions []permission.Permission) {
	m.authToken.Permissions = newPermissions
}

// NextReceive passes the given wsutil.Message to the channel from Receive.
func (m *RawConnection) NextReceive(ctx context.Context, message wsutil.Message) {
	raw, err := json.Marshal(message)
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
func (m *RawConnection) SendDirect(_ context.Context, message wsutil.Message) error {
	if m.SendFail {
		return errors.New("send fail")
	}
	m.outboxMutex.Lock()
	defer m.outboxMutex.Unlock()
	m.outbox = append(m.outbox, message)
	return nil
}

// Disconnect makes Done receive.
func (m *RawConnection) Disconnect() {
	m.disconnect()
}

// Done receives when the connection is done.
func (m *RawConnection) Done() <-chan struct{} {
	return m.lifetime.Done()
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
