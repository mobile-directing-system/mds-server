package app

import (
	"encoding/json"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/ready"
	"go.uber.org/zap/zapcore"
	"net/url"
	"os"
)

const (
	// envLogLevel for config.LogLevel.
	envLogLevel = "MDS_LOG_LEVEL"
	// envServeAddr for config.ServeAddr.
	envServeAddr = "MDS_SERVE_ADDR"
	// envReadyProbeServeAddr for config.ReadyProbeServeAddr.
	envReadyProbeServeAddr = "MDS_READY_PROBE_SERVE_ADDR"
	// envAuthTokenResolveURL for config.AuthTokenResolveURL.
	envAuthTokenResolveURL = "MDS_AUTH_TOKEN_RESOLVE_URL"
	// envRouterConfigPath holds the filepath of the router config.
	envRouterConfigPath = "MDS_ROUTER_CONFIG_PATH"
)

type config struct {
	// LogLevel for logging.
	LogLevel zapcore.Level `json:"log_level"`
	// ServeAddr is the address under which to serve endpoints.
	ServeAddr string `json:"serve_addr"`
	// ReadyProbeServeAddr is the address undew chich to serve the
	// ready-probe-endpoints. parseFromEnv will set this to ready.DefaultServeAddr
	// if not provided otherwise.
	ReadyProbeServeAddr string `json:"ready_probe_serve_addr"`
	// AuthTokenResolveURL is the URL under which public authentication tokens can be
	// resolved to internal ones.
	AuthTokenResolveURL *url.URL `json:"auth_token_resolve_url"`
	// Router is the config for the router.
	Router routerConfig `json:"router"`
}

type routerConfig struct {
	// Gates for the router.
	Gates []gateConfig `json:"gates"`
}

type gateConfig struct {
	// Name of the gate.
	Name string `json:"name"`
	// Channels for the gate.
	Channels []channelConfig `json:"channels"`
}

type channelConfig struct {
	// Name of the channel.
	Name string `json:"name"`
	// URL to which the WebSocket request is sent to.
	URL string `json:"url"`
}

// parseConfig parses a config from the related environment variables like envLogLevel
// and loads the config.Router from filesystem.
func parseConfig() (config, error) {
	var c config
	// Log level.
	logLevelStr := os.Getenv(envLogLevel)
	if logLevelStr != "" {
		logLevel, err := zapcore.ParseLevel(logLevelStr)
		if err != nil {
			return config{}, meh.NewInternalErrFromErr(err, "parse log level", meh.Details{
				"env": envLogLevel,
				"was": logLevelStr,
			})
		}
		c.LogLevel = logLevel
	}
	// Serve address.
	c.ServeAddr = os.Getenv(envServeAddr)
	if c.ServeAddr == "" {
		return config{}, meh.NewBadInputErr("missing serve address", meh.Details{"env": envServeAddr})
	}
	// Ready probe serve address.
	c.ReadyProbeServeAddr = os.Getenv(envReadyProbeServeAddr)
	if c.ReadyProbeServeAddr == "" {
		c.ReadyProbeServeAddr = ready.DefaultServeAddr
	}
	// Auth token resolve URL.
	authTokenResolveURLStr := os.Getenv(envAuthTokenResolveURL)
	if authTokenResolveURLStr == "" {
		return config{}, meh.NewBadInputErr("missing authentication token resolve url", meh.Details{"env": envAuthTokenResolveURL})
	}
	var err error
	c.AuthTokenResolveURL, err = url.Parse(authTokenResolveURLStr)
	if err != nil {
		return config{}, meh.NewBadInputErrFromErr(err, "parse authentication token resolve url", meh.Details{"was": authTokenResolveURLStr})
	}
	// Read and parse router config.
	routerConfigPath := os.Getenv(envRouterConfigPath)
	if routerConfigPath == "" {
		return config{}, meh.NewBadInputErr("missing path for router-config", meh.Details{"env": envRouterConfigPath})
	}
	rc, err := parseRouterConfig(routerConfigPath)
	if err != nil {
		return config{}, meh.Wrap(err, "parse router config", nil)
	}
	c.Router = rc
	return c, nil
}

// parseRouterConfig reads and parses the routerConfig from the given filepath.
func parseRouterConfig(filepath string) (routerConfig, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return routerConfig{}, meh.NewInternalErrFromErr(err, "open file", meh.Details{"path": filepath})
	}
	defer func() { _ = f.Close() }()
	var rc routerConfig
	err = json.NewDecoder(f).Decode(&rc)
	if err != nil {
		return routerConfig{}, meh.NewInternalErrFromErr(err, "read and decode file", nil)
	}
	return rc, nil
}
