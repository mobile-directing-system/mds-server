package app

import (
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/basicconfig"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"go.uber.org/zap"
	"os"
)

const (
	// envRedisAddr for config.RedisAddr.
	envRedisAddr = "MDS_REDIS_ADDR"
	// envForwardAddr for config.ForwardAddr.
	envForwardAddr = "MDS_FORWARD_ADDR"
	// envPublicAuthTokenSecret for config.PublicAuthTokenSecret.
	envPublicAuthTokenSecret = "MDS_PUBLIC_AUTH_TOKEN_SECRET"
)

type config struct {
	basicconfig.Config
	// RedisAddr is the address under which Redis is reachable.
	RedisAddr string `json:"redis_addr"`
	// ForwardAddr is the address to forward handled requests to.
	ForwardAddr string `json:"forward_addr"`
	// PublicAuthTokenSecret is the secret to use for signing public JWT tokens.
	PublicAuthTokenSecret string `json:"public_auth_token_secret"`
}

func parseConfigFromEnv() (config, error) {
	var c config
	var err error
	c.Config, err = basicconfig.ParseFromEnv()
	if err != nil {
		return config{}, meh.Wrap(err, "parse basic config from env", nil)
	}
	// Redis address.
	c.RedisAddr = os.Getenv(envRedisAddr)
	if c.RedisAddr == "" {
		return config{}, meh.NewBadInputErr("missing redis address", meh.Details{"env": envRedisAddr})
	}
	// Forward address.
	c.ForwardAddr = os.Getenv(envForwardAddr)
	if c.ForwardAddr == "" {
		return config{}, meh.NewBadInputErr("missing forward address", meh.Details{"env": envForwardAddr})
	}
	// Public auth token secret.
	c.PublicAuthTokenSecret = os.Getenv(envPublicAuthTokenSecret)
	if c.PublicAuthTokenSecret == "" {
		// For development purposes, we only log a warning.
		logging.DebugLogger().Warn("no public auth token secret provided", zap.String("env", envPublicAuthTokenSecret))
	}
	return c, nil
}
