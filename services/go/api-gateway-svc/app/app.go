package app

import (
	"context"
	"embed"
	"github.com/go-redis/redis/v8"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/endpoints"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/eventport"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/connectutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgconnect"
	"github.com/mobile-directing-system/mds-server/services/go/shared/ready"
	"golang.org/x/sync/errgroup"
	"io/fs"
)

//go:embed db-migrations/*.sql
var dbMigrationsEmbedded embed.FS

var dbMigrations fs.FS

func init() {
	var err error
	dbMigrations, err = fs.Sub(dbMigrationsEmbedded, "db-migrations")
	if err != nil {
		panic(meh.NewInternalErrFromErr(err, "create sub fs for database migrations", nil).Error())
	}
}

const kafkaGroupID = "mds-api-gateway-svc"

// Run the gateway.
func Run(ctx context.Context) error {
	c, err := parseConfigFromEnv()
	if err != nil {
		return meh.Wrap(err, "parse config", nil)
	}
	logger, err := logging.NewLogger("api-gateway-svc", c.LogLevel)
	if err != nil {
		return meh.Wrap(err, "new logger", nil)
	}
	logging.SetDebugLogger(logger.Named("debug"))
	defer func() { _ = logger.Sync() }()
	eg, egCtx := errgroup.WithContext(ctx)
	readyProbeServer, startUpCompleted := ready.NewServer(logger.Named("ready-probe"))
	eg.Go(func() error {
		err := readyProbeServer.Serve(egCtx, c.ReadyProbeServeAddr)
		return meh.NilOrWrap(err, "serve ready-probe-server", meh.Details{"addr": c.ReadyProbeServeAddr})
	})
	// Connect to database.
	sqlDB, err := pgconnect.ConnectAndRunMigrations(ctx, logger, c.DBConnString, dbMigrations)
	if err != nil {
		return meh.Wrap(err, "connect db and run migrations", meh.Details{"db_conn_string": c.DBConnString})
	}
	// Setup Redis.
	redisClient := redis.NewClient(&redis.Options{Addr: c.RedisAddr})
	// Await ready.
	readyCheck := func(ctx context.Context) error {
		eg, egCtx := errgroup.WithContext(ctx)
		// Check hosts.
		eg.Go(func() error {
			err := connectutil.AwaitHostsReachable(egCtx, c.KafkaAddr, c.RedisAddr)
			return meh.NilOrWrap(err, "await hosts reachable", nil)
		})
		// Check Kafka topics.
		eg.Go(func() error {
			err := kafkautil.AwaitTopics(egCtx, c.KafkaAddr, event.PermissionsTopic, event.UsersTopic, event.AuthTopic)
			return meh.NilOrWrap(err, "await topics", meh.Details{"kafka_addr": c.KafkaAddr})
		})
		// Check Redis.
		eg.Go(func() error {
			err := redisClient.Ping(egCtx).Err()
			return meh.NilOrWrap(err, "ping redis", meh.Details{"redis_addr": c.RedisAddr})
		})
		// Check database.
		eg.Go(func() error {
			err := sqlDB.Ping(egCtx)
			return meh.NilOrWrap(err, "ping database", nil)
		})
		return eg.Wait()
	}
	err = ready.Await(ctx, readyCheck)
	if err != nil {
		return meh.Wrap(err, "await ready", nil)
	}
	// Setup Kafka.
	kafkaWriter := kafkautil.NewWriter(logger.Named("kafka-writer"), c.KafkaAddr)
	eventPort := eventport.NewPort(kafkaWriter)
	// Setup controller.
	ctrl := &controller.Controller{
		Logger:                logger.Named("controller"),
		PublicAuthTokenSecret: c.PublicAuthTokenSecret,
		AuthTokenSecret:       c.AuthTokenSecret,
		Store:                 store.NewMall(redisClient),
		DB:                    sqlDB,
		Notifier:              eventPort,
	}
	// Read messages.
	eg.Go(func() error {
		logger := logger.Named("kafka-reader")
		kafkaReader := kafkautil.NewReader(logger, c.KafkaAddr, kafkaGroupID,
			[]event.Topic{event.PermissionsTopic, event.UsersTopic})
		err := kafkautil.Read(egCtx, logger, kafkaReader, eventPort.HandlerFn(ctrl))
		if err != nil {
			return meh.Wrap(err, "read kafka messages", nil)
		}
		return nil
	})
	// Serve endpoints.
	eg.Go(func() error {
		err := endpoints.Serve(egCtx, logger.Named("endpoints"), c.ServeAddr, c.ForwardAddr, ctrl)
		if err != nil {
			return meh.Wrap(err, "serve endpoints", nil)
		}
		return nil
	})
	startUpCompleted(readyCheck)
	return eg.Wait()
}
