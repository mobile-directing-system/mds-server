package app

import (
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/basicconfig"
)

type config struct {
	basicconfig.Config
}

func parseConfigFromEnv() (config, error) {
	baseConfig, err := basicconfig.ParseFromEnv()
	if err != nil {
		return config{}, meh.Wrap(err, "parse basic config from env", nil)
	}
	return config{
		Config: baseConfig,
	}, nil
}
