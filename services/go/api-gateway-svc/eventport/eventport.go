package eventport

import (
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
)

// Port manages event messaging.
type Port struct {
	writer kafkautil.Writer
}

// NewPort creates a new port.
func NewPort(writer kafkautil.Writer) *Port {
	return &Port{
		writer: writer,
	}
}
