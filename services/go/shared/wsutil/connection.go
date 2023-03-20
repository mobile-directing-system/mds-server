package wsutil

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"go.uber.org/zap"
)

// BaseConnection allows sending and receiving messages via WebSocket.
type BaseConnection interface {
	ID() uuid.UUID
	Logger() *zap.Logger
	// AuthToken returns the auth.Token for the connection request.
	AuthToken() auth.Token
	// Send marshals the given payload and combines it with the MessageType into a
	// Message. If marshalling fails at any step, a meh.ErrInternal will be
	// returned.
	//
	// Warning: If the connection is closed while sending, NO error will be
	// returned!
	Send(ctx context.Context, messageType MessageType, payload any) error
	// SendErr notifies the client about the given error.
	SendErr(ctx context.Context, err error)
	// SendDirect a Message. If marshalling the Message fails, a meh.ErrInternal
	// will be returned.
	//
	// Warning: If the connection is closed while sending, NO error will be
	// returned!
	SendDirect(ctx context.Context, message Message) error
	// SendRaw returns the raw channel for sending messages.
	SendRaw() chan<- json.RawMessage
	// Done receives, when the connection is closed. You can also read from Receive
	// as the channel is closed as well, when the connection is so.
	Done() <-chan struct{}
}

// RawConnection is the regular connection for receiving raw messages. If you
// want messages to be parsed into wsutil.Message automatically, you can use
// NewAutoParserConnection.
type RawConnection interface {
	BaseConnection
	// ReceiveRaw returns the raw channel for receiving messages.
	ReceiveRaw() <-chan json.RawMessage
}

// AutoParserConnection is created through NewAutoParserConnection from a
// RawConnection and provides received messages, being automatically parsed into
// wsutil.Message.
type AutoParserConnection interface {
	BaseConnection
	// Receive messages from the connection. Closes, when the connection is closed.
	// It is safe to make multiple calls to Receive because of the same channel being
	// returned.
	Receive() <-chan Message
}

// rawConnection is the raw implementation of RawConnection.
type rawConnection struct {
	id        uuid.UUID
	logger    *zap.Logger
	lifetime  context.Context
	cancel    context.CancelFunc
	authToken auth.Token
	receive   chan json.RawMessage
	send      chan json.RawMessage
}

func (conn *rawConnection) ID() uuid.UUID {
	return conn.id
}

func (conn *rawConnection) Logger() *zap.Logger {
	return conn.logger
}

// AuthToken returns the auth.Token that was used for initiating the connection.
func (conn *rawConnection) AuthToken() auth.Token {
	return conn.authToken
}

// Send marshals the given payload and combines it with the MessageType into a
// Message. If marshalling fails at any step, a meh.ErrInternal will be
// returned.
func (conn *rawConnection) Send(ctx context.Context, messageType MessageType, payload any) error {
	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "marshal payload", meh.Details{"payload": fmt.Sprintf("%+v", payload)})
	}
	return conn.SendDirect(ctx, Message{
		Type:    messageType,
		Payload: payloadRaw,
	})
}

// SendErr sends a TypeError message with the given error.
func (conn *rawConnection) SendErr(ctx context.Context, sendErr error) {
	message := ErrorMessageFromErr(sendErr)
	err := conn.Send(ctx, TypeError, message)
	if err != nil {
		mehlog.Log(conn.logger, meh.Wrap(err, "send error message", meh.Details{"message": message}))
	}
}

// SendDirect the given Message. If marshalling fails, a meh.ErrInternal will be
// returned. If the connection is closed during sending, NO error will be
// returned!
func (conn *rawConnection) SendDirect(ctx context.Context, message Message) error {
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

func (conn *rawConnection) Done() <-chan struct{} {
	return conn.lifetime.Done()
}

func (conn *rawConnection) ReceiveRaw() <-chan json.RawMessage {
	return conn.receive
}

func (conn *rawConnection) SendRaw() chan<- json.RawMessage {
	return conn.send
}

// autoParserConnection is the implementation of AutoParserConnection.
type autoParserConnection struct {
	RawConnection
	receiveParsed chan Message
}

// NewAutoParserConnection starts a new AutoParserConnection that runs until the given
// RawConnection is done.
//
// Warning: Do NOT use the given connection afterwards!
func NewAutoParserConnection(conn RawConnection) AutoParserConnection {
	apConn := &autoParserConnection{
		RawConnection: conn,
		receiveParsed: make(chan Message),
	}
	go apConn.receiveParseAndForward()
	return apConn
}

func (apConn *autoParserConnection) receiveParseAndForward() {
	defer close(apConn.receiveParsed)
	for {
		// Receive next message.
		var messageRaw json.RawMessage
		var more bool
		select {
		case <-apConn.Done():
			return
		case messageRaw, more = <-apConn.ReceiveRaw():
		}
		if !more {
			return
		}
		// Unmarshal as message.
		var message Message
		err := json.Unmarshal(messageRaw, &message)
		if err != nil {
			mehlog.Log(apConn.Logger(), meh.NewBadInputErrFromErr(err, "unmarshal message", meh.Details{"raw": messageRaw}))
			continue
		}
		// Forward.
		select {
		case <-apConn.Done():
			apConn.Logger().Debug("dropping message to forward because of connection being closed",
				zap.Any("message", message))
			return
		case apConn.receiveParsed <- message:
		}
	}
}

func (apConn *autoParserConnection) Receive() <-chan Message {
	return apConn.receiveParsed
}
