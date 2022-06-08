package connectutil

import (
	"context"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net"
	"time"
)

// AwaitHostReachable waits until the given host is reachable using
// net.DialTimeout.
func AwaitHostReachable(ctx context.Context, host string) error {
	for {
		select {
		case <-ctx.Done():
			return meh.NewInternalErrFromErr(ctx.Err(), "wait for host reachable", meh.Details{"host": host})
		default:
		}
		conn, err := net.DialTimeout("tcp", host, 3*time.Second)
		if err != nil {
			logging.DebugLogger().Debug("awaiting host reachable", zap.String("host", host), zap.Error(err))
			continue
		}
		err = conn.Close()
		if err != nil {
			return meh.NewInternalErrFromErr(err, "close opened connection", nil)
		}
		return nil
	}
}

// AwaitHostsReachable waits until the given hosts are reachable. This is the
// same as running AwaitHostReachable with goroutines.
func AwaitHostsReachable(ctx context.Context, hosts ...string) error {
	eg, egCtx := errgroup.WithContext(ctx)
	for _, host := range hosts {
		eg.Go(func() error {
			err := AwaitHostReachable(egCtx, host)
			if err != nil {
				return meh.Wrap(err, "await host reachable", meh.Details{"host": host})
			}
			return nil
		})
	}
	return eg.Wait()
}
