package kafkautil

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
)

// MessageRecorder mocks kafkautil.OutboxWriter for testing purposes.
type MessageRecorder struct {
	WriteFail bool
	Recorded  []OutboundMessage
}

// NewMessageRecorder creates a new MessageRecorder.
func NewMessageRecorder() *MessageRecorder {
	return &MessageRecorder{Recorded: make([]OutboundMessage, 0)}
}

// AddOutboxMessages adds the given message to the Recorded ones.
func (recorder *MessageRecorder) AddOutboxMessages(_ context.Context, _ pgx.Tx, messages ...OutboundMessage) error {
	if recorder.WriteFail {
		return errors.New("write fail")
	}
	recorder.Recorded = append(recorder.Recorded, messages...)
	return nil
}
