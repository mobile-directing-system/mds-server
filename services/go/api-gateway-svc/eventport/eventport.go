package eventport

import (
	"github.com/segmentio/kafka-go"
)

// Port manages event messaging.
type Port struct {
	kafkaWriter *kafka.Writer
}

// NewPort creates a new port.
func NewPort(kafkaWriter *kafka.Writer) *Port {
	return &Port{
		kafkaWriter: kafkaWriter,
	}
}
