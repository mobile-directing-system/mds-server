package app

import (
	"context"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/connectutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"github.com/mobile-directing-system/mds-server/services/go/shared/ready"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"github.com/mobile-directing-system/mds-server/services/go/ws-hub-svc/endpoints"
	"github.com/mobile-directing-system/mds-server/services/go/ws-hub-svc/ws"
	"golang.org/x/sync/errgroup"
)

// Run the hub.
func Run(ctx context.Context) error {
	c, err := parseConfig()
	if err != nil {
		return meh.Wrap(err, "parse config", nil)
	}
	logger, err := logging.NewLogger("ws-hub-svc", c.LogLevel)
	if err != nil {
		return meh.Wrap(err, "new logger", nil)
	}
	defer func() { _ = logger.Sync() }()
	logging.SetDebugLogger(logger.Named("debug"))
	eg, egCtx := errgroup.WithContext(ctx)
	probeServer, startUpCompleted := ready.NewServer(logger.Named("probe-server"))
	eg.Go(func() error {
		err := probeServer.Serve(egCtx, c.ReadyProbeServeAddr)
		return meh.NilOrWrap(err, "serve ready-probe-server", meh.Details{"addr": c.ReadyProbeServeAddr})
	})
	forwardURLs := extractForwardChannelURLs(c.Router.Gates)
	// Await ready.
	readyCheck := func(ctx context.Context) error {
		eg, egCtx := errgroup.WithContext(ctx)
		// Check hosts.
		eg.Go(func() error {
			err := connectutil.AwaitHostsReachable(egCtx, forwardURLs...)
			return meh.NilOrWrap(err, "await forward urls reachable", meh.Details{"forward_urls": forwardURLs})
		})
		return eg.Wait()
	}
	err = ready.Await(ctx, readyCheck)
	if err != nil {
		return meh.Wrap(err, "await ready", nil)
	}
	// Setup hub.
	gateConfigs := gateConfigsFromConfig(c.Router.Gates)
	wsHub := ws.NewNetHub(egCtx, logger.Named("ws"), gateConfigs)
	// Serve endpoints.
	eg.Go(func() error {
		err := endpoints.Serve(egCtx, logger.Named("endpoints"), c.ServeAddr, wsHub)
		return meh.NilOrWrap(err, "serve endpoints", meh.Details{"serve_addr": c.ServeAddr})
	})
	startUpCompleted(readyCheck)
	return eg.Wait()
}

func extractForwardChannelURLs(gateConfigs []gateConfig) []string {
	forwardURLsMap := make(map[string]struct{}, len(gateConfigs))
	for _, gateConfig := range gateConfigs {
		for _, agent := range gateConfig.Channels {
			forwardURLsMap[agent.URL] = struct{}{}
		}
	}
	// Convert to slice.
	forwardURLs := make([]string, 0, len(forwardURLsMap))
	for url := range forwardURLsMap {
		forwardURLs = append(forwardURLs, url)
	}
	return forwardURLs
}

func gateConfigsFromConfig(gateConfigs []gateConfig) map[string]ws.Gate {
	out := make(map[string]ws.Gate, len(gateConfigs))
	for _, gc := range gateConfigs {
		channels := make(map[wsutil.Channel]ws.Channel, len(gc.Channels))
		for _, agent := range gc.Channels {
			channels[wsutil.Channel(agent.Name)] = ws.Channel{
				URL: agent.URL,
			}
		}
		out[gc.Name] = ws.Gate{
			Name:     gc.Name,
			Channels: channels,
		}
	}
	return out
}
