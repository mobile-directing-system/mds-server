// Package basicconfig is used for basic configurations with database and Kafka
// connections.
package basicconfig

import (
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"github.com/mobile-directing-system/mds-server/services/go/shared/ready"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	// EnvDBConnString for Config.DBConnString.
	EnvDBConnString = "MDS_DB_CONN_STRING"
	// EnvLogLevel for Config.LogLevel.
	EnvLogLevel = "MDS_LOG_LEVEL"
	// EnvServeAddr for Config.ServeAddr.
	EnvServeAddr = "MDS_SERVE_ADDR"
	// EnvKafkaAddr for Config.KafkaAddr.
	EnvKafkaAddr = "MDS_KAFKA_ADDR"
	// EnvAuthTokenSecret for Config.AuthTokenSecret.
	EnvAuthTokenSecret = "MDS_AUTH_TOKEN_SECRET"
	// EnvReadyProbeServeAddr for Config.ReadyProbeServeAddr.
	EnvReadyProbeServeAddr = "MDS_READY_PROBE_SERVE_ADDR"
)

// Config is a basic configuration with support for database and Kafka
// connections, log level and endpoint serving.
type Config struct {
	// DBConnString is the database connection string.
	DBConnString string `json:"db_conn_string"`
	// LogLevel for logging.
	LogLevel zapcore.Level `json:"log_level"`
	// ServeAddr is the address under which to serve endpoints.
	ServeAddr string `json:"serve_addr"`
	// KafkaAddr is the address under which Kafka is reachable.
	KafkaAddr string `json:"kafka_addr"`
	// AuthTokenSecret is the secret to use for signing and validating internal
	// authentication tokens.
	AuthTokenSecret string `json:"auth_token_secret"`
	// ReadyProbeServeAddr is the address undew chich to serve the
	// ready-probe-endpoints. ParseFromEnv will set this to ready.DefaultServeAddr
	// if not provided otherwise.
	ReadyProbeServeAddr string `json:"ready_probe_serve_addr"`
}

// ParseFromEnv parses a Config from the related environment variables like
// EnvDBConnString.
func ParseFromEnv() (Config, error) {
	var c Config
	// Database connection string.
	c.DBConnString = os.Getenv(EnvDBConnString)
	if c.DBConnString == "" {
		return Config{}, meh.NewBadInputErr("missing database connection string", meh.Details{"env": EnvDBConnString})
	}
	// Log level.
	logLevelStr := os.Getenv(EnvLogLevel)
	if logLevelStr != "" {
		logLevel, err := zapcore.ParseLevel(logLevelStr)
		if err != nil {
			return Config{}, meh.NewInternalErrFromErr(err, "parse log level", meh.Details{
				"env": EnvLogLevel,
				"was": logLevelStr,
			})
		}
		c.LogLevel = logLevel
	}
	// Serve address.
	c.ServeAddr = os.Getenv(EnvServeAddr)
	if c.ServeAddr == "" {
		return Config{}, meh.NewBadInputErr("missing serve address", meh.Details{"env": EnvServeAddr})
	}
	// Kafka address.
	c.KafkaAddr = os.Getenv(EnvKafkaAddr)
	if c.KafkaAddr == "" {
		return Config{}, meh.NewBadInputErr("missing kafka address", meh.Details{"env": EnvKafkaAddr})
	}
	// Auth token secret.
	c.AuthTokenSecret = os.Getenv(EnvAuthTokenSecret)
	if c.AuthTokenSecret == "" {
		// For development purposes, we only print a warning.
		logging.DebugLogger().Warn("no authentication token secret provided", zap.String("env", EnvAuthTokenSecret))
	}
	// Ready probe serve address.
	c.ReadyProbeServeAddr = os.Getenv(EnvReadyProbeServeAddr)
	if c.ReadyProbeServeAddr == "" {
		c.ReadyProbeServeAddr = ready.DefaultServeAddr
	}
	return c, nil
}
