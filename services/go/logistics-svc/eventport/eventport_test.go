package eventport

import (
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"time"
)

const timeout = 5 * time.Second

type PortMock struct {
	recorder *kafkautil.MessageRecorder
	Port     *Port
}

func newMockPort() *PortMock {
	p := &PortMock{
		recorder: kafkautil.NewMessageRecorder(),
	}
	p.Port = NewPort(p.recorder)
	return p
}
