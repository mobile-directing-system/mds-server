package app

import (
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/basicconfig"
	"os"
)

const (
	// envRedisAddr for config.RedisAddr.
	envRedisAddr = "MDS_REDIS_ADDR"
	// envForwardAddr for config.ForwardAddr.
	envForwardAddr = "MDS_FORWARD_ADDR"
)

type config struct {
	basicconfig.Config
	// RedisAddr is the address under which Redis is reachable.
	RedisAddr string `json:"redis_addr"`
	// ForwardAddr is the address to forward handled requests to.
	ForwardAddr string `json:"forward_addr"`
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
	return c, nil
}
