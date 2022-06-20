package eventport

import (
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"time"
)

const timeout = 5 * time.Second

type PortMock struct {
	recorder *testutil.MessageRecorder
	Port     *Port
}

func newMockPort() *PortMock {
	p := &PortMock{
		recorder: testutil.NewMessageRecorder(),
	}
	p.Port = NewPort(p.recorder)
	return p
}
