package search

import "os"

// EnvHost for ServiceConfig.Host.
const EnvHost = "MDS_SEARCH_HOST"

// EnvMasterKey for ServiceConfig.MasterKey.
const EnvMasterKey = "MDS_SEARCH_MASTER_KEY"

// ServiceConfig is a general-purpose config for search to use in services.
type ServiceConfig struct {
	Host      string
	MasterKey string
}

// ServiceConfigFromEnv returns a ServiceConfig based on environent variables.
func ServiceConfigFromEnv() ServiceConfig {
	return ServiceConfig{
		Host:      os.Getenv(EnvHost),
		MasterKey: os.Getenv(EnvMasterKey),
	}
}
