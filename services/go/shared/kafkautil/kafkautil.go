package kafkautil

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"time"
)

// writeMessagesTimeout is the timeout to use when writing messages to kafka via
// WriteMessages.
const writeMessagesTimeout = 10 * time.Second

// kafkaMessageEventTypeHeader is the header name where event-type metadata is
// stored in.
const kafkaMessageEventTypeHeader = "event-type"

// Message acts as a replacement of kafka.Message for easier usage.
type Message struct {
	// Topic indicates the topic the message was consumed from or should be written
	// to if not specified otherwise.
	Topic string
	// Key is message key (translated to and from byte slice).
	Key string
	// EventType is the type of event, taken from/put into Headers.
	EventType event.Type
	// Value will be marshalled as JSON when used with WriteMessages. When
	// processing a received message, this will be the same as RawValue.
	Value any
	// RawValue is the marshalled Value.
	RawValue json.RawMessage
	// Headers for the message (translated to and from kafka.Header).
	Headers []MessageHeader
}

// MessageHeader equals kafka.Header except that Value is a string and not byte
// slice.
type MessageHeader struct {
	Key   string
	Value string
}

// messageFromKafkaMessage converts a kafka.Message to Message.
func messageFromKafkaMessage(kafkaMessage kafka.Message) Message {
	headers, eventType := headersFromKafkaHeaders(kafkaMessage.Headers)
	return Message{
		Topic:     kafkaMessage.Topic,
		Key:       string(kafkaMessage.Key),
		EventType: eventType,
		Value:     kafkaMessage.Value,
		RawValue:  kafkaMessage.Value,
		Headers:   headers,
	}
}

// headersFromKafkaHeaders converts a kafka.Header list to MessageHeader list.
// Additionally, the event.Type is extracted from kafkaMessageEventTypeHeader.
func headersFromKafkaHeaders(kafkaHeaders []kafka.Header) ([]MessageHeader, event.Type) {
	headers := make([]MessageHeader, 0, len(kafkaHeaders))
	var eventType event.Type
	for _, kafkaHeader := range kafkaHeaders {
		if kafkaHeader.Key == kafkaMessageEventTypeHeader {
			eventType = event.Type(kafkaHeader.Value)
		}
		headers = append(headers, MessageHeader{
			Key:   kafkaHeader.Key,
			Value: string(kafkaHeader.Value),
		})
	}
	return headers, eventType
}

// kafkaHeadersFromHeaders converts a MessageHeader list to kafka.Header list.
// Additionally, the given event.Type is added via kafkaMessageEventTypeHeader.
func kafkaHeadersFromHeaders(headers []MessageHeader, eventType event.Type) []kafka.Header {
	kafkaHeaders := make([]kafka.Header, 0, len(headers)+1)
	for _, header := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{
			Key:   header.Key,
			Value: []byte(header.Value),
		})
	}
	kafkaHeaders = append(kafkaHeaders, kafka.Header{
		Key:   kafkaMessageEventTypeHeader,
		Value: []byte(eventType),
	})
	return kafkaHeaders
}

// kafkaMessageFromMessage converts a Message to kafka.Message and marshals the
// Message.Value as JSON if not nil.
func kafkaMessageFromMessage(message Message) (kafka.Message, error) {
	if message.Value != nil {
		var err error
		message.RawValue, err = json.Marshal(message.Value)
		if err != nil {
			return kafka.Message{}, meh.NewInternalErrFromErr(err, "marshal message value", nil)
		}
	}
	return kafka.Message{
		Topic:   message.Topic,
		Key:     []byte(message.Key),
		Value:   message.RawValue,
		Headers: kafkaHeadersFromHeaders(message.Headers, message.EventType),
	}, nil
}

// WriteMessages writes the given kafka.Message list to the kafka.Writer with a
// predefined timeout.
func WriteMessages(w *kafka.Writer, messages ...Message) error {
	// Convert all messages.
	kafkaMessages := make([]kafka.Message, 0, len(messages))
	for _, message := range messages {
		kafkaMessage, err := kafkaMessageFromMessage(message)
		if err != nil {
			return meh.Wrap(err, "convert message to kafka message", meh.Details{"message": message})
		}
		kafkaMessages = append(kafkaMessages, kafkaMessage)
	}
	// Write messages.
	timeout, cancel := context.WithTimeout(context.Background(), writeMessagesTimeout)
	defer cancel()
	err := w.WriteMessages(timeout, kafkaMessages...)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "write messages", meh.Details{"timeout": writeMessagesTimeout})
	}
	return nil
}

// NewWriter creates a new kafka.Writer with logging middleware.
func NewWriter(logger *zap.Logger, addr string) *kafka.Writer {
	return &kafka.Writer{
		Addr:        kafka.TCP(addr),
		Logger:      kafkaLogger(logger),
		ErrorLogger: kafkaErrorLogger(logger),
		MaxAttempts: 16,
	}
}

func kafkaLogger(logger *zap.Logger) kafka.LoggerFunc {
	return func(message string, args ...interface{}) {
		logger.Debug(fmt.Sprintf(message, args...))
	}
}

func kafkaErrorLogger(logger *zap.Logger) kafka.LoggerFunc {
	return func(message string, args ...interface{}) {
		logger.Error(fmt.Sprintf(message, args...))
	}
}
