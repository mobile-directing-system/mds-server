package testutil

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
)

// MessageRecorder mocks kafkautil.Writer for testing purposes.
type MessageRecorder struct {
	WriteFail bool
	Recorded  []kafka.Message
}

// NewMessageRecorder creates a new MessageRecorder.
func NewMessageRecorder() *MessageRecorder {
	return &MessageRecorder{Recorded: make([]kafka.Message, 0)}
}

// WriteMessages fails if WriteFail is set and records the messages to Recorded
// otherwise.
func (recorder *MessageRecorder) WriteMessages(_ context.Context, messages ...kafka.Message) error {
	if recorder.WriteFail {
		return errors.New("write fail")
	}
	recorder.Recorded = append(recorder.Recorded, messages...)
	return nil
}
