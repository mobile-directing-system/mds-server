package wsutil

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"go.uber.org/zap"
)

// Connection allows sending and receiving messages via WebSocket.
type Connection interface {
	// AuthToken returns the auth.Token for the connection request.
	AuthToken() auth.Token
	// Receive messages from the connection. Closes, when the connection is closed.
	Receive() <-chan Message
	// Send marshals the given payload and combines it with the MessageType into a
	// Message. If marshalling fails at any step, a meh.ErrInternal will be
	// returned.
	//
	// Warning: If the connection is closed while sending, NO error will be
	// returned!
	Send(ctx context.Context, messageType MessageType, payload any) error
	// SendDirect a Message. If marshalling the Message fails, a meh.ErrInternal
	// will be returned.
	//
	// Warning: If the connection is closed while sending, NO error will be
	// returned!
	SendDirect(ctx context.Context, message Message) error
	// Done receives, when the connection is closed. You can also read from Receive
	// as the channel is closed as well, when the connection is so.
	Done() <-chan struct{}
}

// RawConnection is used for raw access instead of Connection.
type RawConnection interface {
	// ReceiveRaw returns the raw channel for receiving messages.
	ReceiveRaw() <-chan json.RawMessage
	// SendRaw returns the raw channel for sending messages.
	SendRaw() chan<- json.RawMessage
}

// connection is the raw implementation of Connection.
type connection struct {
	logger    *zap.Logger
	lifetime  context.Context
	cancel    context.CancelFunc
	authToken auth.Token
	receive   chan json.RawMessage
	send      chan json.RawMessage
}

// AuthToken returns the auth.Token that was used for initiating the connection.
func (conn *connection) AuthToken() auth.Token {
	return conn.authToken
}

// Receive reads raw messages, unmarshals as Message and forwards them to the
// returned channel. Messages, that cannot be unmarshalled, will be logged as
// meh.ErrBadInput.
func (conn *connection) Receive() <-chan Message {
	c := make(chan Message)
	go func() {
		defer close(c)
		for {
			// Receive next message.
			var messageRaw json.RawMessage
			var more bool
			select {
			case <-conn.lifetime.Done():
				return
			case messageRaw, more = <-conn.receive:
			}
			if !more {
				return
			}
			// Unmarshal as message.
			var message Message
			err := json.Unmarshal(messageRaw, &message)
			if err != nil {
				mehlog.Log(conn.logger, meh.NewBadInputErrFromErr(err, "unmarshal message", meh.Details{"raw": messageRaw}))
				continue
			}
			// Forward.
			select {
			case <-conn.lifetime.Done():
				conn.logger.Debug("dropping message to forward because of connection being closed",
					zap.Any("message", message))
				return
			case c <- message:
			}
		}
	}()
	return c
}

// Send marshals the given payload and combines it with the MessageType into a
// Message. If marshalling fails at any step, a meh.ErrInternal will be
// returned.
func (conn *connection) Send(ctx context.Context, messageType MessageType, payload any) error {
	paylodRaw, err := json.Marshal(payload)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "marshal payload", meh.Details{"payload": fmt.Sprintf("%+v", payload)})
	}
	return conn.SendDirect(ctx, Message{
		Type:    messageType,
		Payload: paylodRaw,
	})
}

// SendDirect the given Message. If marshalling fails, a meh.ErrInternal will be
// returned. If the connection is closed during sending, NO error will be
// returned!
func (conn *connection) SendDirect(ctx context.Context, message Message) error {
	messageRaw, err := json.Marshal(message)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "marshal message", meh.Details{"message": message})
	}
	select {
	case <-ctx.Done():
		conn.logger.Debug("connection done while sending message via websocket",
			zap.Any("message", messageRaw))
	case conn.send <- messageRaw:
	}
	return nil
}

func (conn *connection) Done() <-chan struct{} {
	return conn.lifetime.Done()
}

func (conn *connection) ReceiveRaw() <-chan json.RawMessage {
	return conn.receive
}

func (conn *connection) SendRaw() chan<- json.RawMessage {
	return conn.send
}
