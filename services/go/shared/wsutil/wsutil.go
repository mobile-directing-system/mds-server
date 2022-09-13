package wsutil

import "time"

const (
	// WriteWait is the time allowed to write a message to the peer.
	WriteWait = 16 * time.Second
	// MaxMessageSize is the maximum message size allowed from the peer.
	MaxMessageSize = 128 * 1000 * 1000 // 128MB
	// PongWait is the time allowed to read the next pong message from the peer.
	PongWait = 10 * time.Second
	// PingPeriod is the interval duration in which to send pings to the peer.
	PingPeriod = (PongWait * 9) / 10
	// ReadBufferSize for the websocket.Upgrader.
	ReadBufferSize = 2048
	// WriteBufferSize for the websocket.Upgrader.
	WriteBufferSize = 2048
)

var (
	// NewLine for reading.
	NewLine = []byte{'\n'}
	// Space for reading.
	Space = []byte{' '}
)
