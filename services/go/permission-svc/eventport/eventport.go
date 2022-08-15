package eventport

import (
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// Port manages event messaging.
type Port struct {
	writer kafkautil.OutboxWriter
}

// NewPort creates a new Port.
func NewPort(writer kafkautil.OutboxWriter) *Port {
	return &Port{
		writer: writer,
	}
}
