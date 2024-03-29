package connectutil

import (
	"context"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net"
	"net/url"
	"strings"
	"time"
)

// AwaitHostReachableCooldown is the cooldown to use in AwaitHostReachable when
// the host was not reachable.
var AwaitHostReachableCooldown = 3 * time.Second

// AwaitHostReachable waits until the given host is reachable using
// net.DialTimeout.
func AwaitHostReachable(ctx context.Context, host string) error {
	for {
		err := AssureHostReachable(host)
		if err != nil {
			mehlog.LogToLevel(logging.DebugLogger(), zap.DebugLevel, meh.Wrap(err, "await host reachable",
				meh.Details{"host": host}))
			// Wait.
			select {
			case <-ctx.Done():
				return meh.NewInternalErrFromErr(ctx.Err(), "wait for host reachable", meh.Details{"host": host})
			case <-time.After(AwaitHostReachableCooldown):
			}
			continue
		}
		return nil
	}
}

//goland:noinspection HttpUrlsUsage
var urlPrefixes = []string{"https://", "http://", "wss://", "ws://"}

// AssureHostReachable checks if the given host is reachable with a timeout of 3
// seconds.
func AssureHostReachable(host string) error {
	for _, prefix := range urlPrefixes {
		if !strings.HasPrefix(host, prefix) {
			continue
		}
		u, err := url.Parse(host)
		if err != nil {
			return meh.NewInternalErrFromErr(err, "parse host", meh.Details{"was": host})
		}
		host = u.Host
	}
	// Dial.
	conn, err := net.DialTimeout("tcp", host, 3*time.Second)
	if err != nil {
		return meh.Wrap(err, "dial tcp", meh.Details{"host": host})
	}
	err = conn.Close()
	if err != nil {
		return meh.Wrap(err, "close connection in host-reachable-check", meh.Details{"host": host})
	}
	return nil
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
