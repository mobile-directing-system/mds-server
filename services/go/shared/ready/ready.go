package ready

import (
	"context"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"go.uber.org/zap"
	"time"
)

// CheckFn is the function that reports the applications ready-state.
type CheckFn func(ctx context.Context) error

// ReadyCooldown is the cooldown to use in Await when the CheckFn fails.
var ReadyCooldown = 3 * time.Second

// Await waits until the given CheckFn signals being ready. If the CheckFn
// fails, we wait for the time specified in ReadyCooldown.
func Await(ctx context.Context, checkFn CheckFn) error {
	for {
		err := checkFn(ctx)
		if err != nil {
			mehlog.LogToLevel(logging.DebugLogger(), zap.DebugLevel, meh.Wrap(err, "check ready fn", nil))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(ReadyCooldown):
				continue
			}
		}
		return nil
	}
}
