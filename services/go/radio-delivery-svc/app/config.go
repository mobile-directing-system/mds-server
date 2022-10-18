package app

import (
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/basicconfig"
	"os"
	"time"
)

// envPickedUpTimeout for config.pickedUpTimeout.
const envPickedUpTimeout = "MDS_PICKED_UP_TIMEOUT"

type config struct {
	basicconfig.Config
	// pickedUpTimeout is the timeout for picked up radio deliveries to be
	// automatically released.
	pickedUpTimeout time.Duration
}

func parseConfigFromEnv() (config, error) {
	baseConfig, err := basicconfig.ParseFromEnv()
	if err != nil {
		return config{}, meh.Wrap(err, "parse basic config from env", nil)
	}
	c := config{Config: baseConfig}
	// Parse picked-up-timeout.
	pickedUpTimeoutStr := os.Getenv(envPickedUpTimeout)
	if pickedUpTimeoutStr == "" {
		return config{}, meh.NewBadInputErr("missing picked-up-timeout string", meh.Details{"env": envPickedUpTimeout})
	}
	c.pickedUpTimeout, err = time.ParseDuration(pickedUpTimeoutStr)
	if err != nil {
		return config{}, meh.NewBadInputErrFromErr(err, "parse picked-up-timeout", meh.Details{"was": pickedUpTimeoutStr})
	}
	return c, nil
}
