package app

import (
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/basicconfig"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
)

type config struct {
	basicconfig.Config
	searchConfig search.ServiceConfig
}

func parseConfigFromEnv() (config, error) {
	baseConfig, err := basicconfig.ParseFromEnv()
	if err != nil {
		return config{}, meh.Wrap(err, "parse basic config from env", nil)
	}
	searchConfig := search.ServiceConfigFromEnv()
	return config{
		Config:       baseConfig,
		searchConfig: searchConfig,
	}, nil
}
