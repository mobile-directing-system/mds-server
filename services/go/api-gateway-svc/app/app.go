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
	err = connectutil.AwaitHostsReachable(ctx, c.KafkaAddr, c.RedisAddr)
	if err != nil {
		return meh.Wrap(err, "await hosts reachable", nil)
	}
	// Connect to database.
	sqlDB, err := pgconnect.ConnectAndRunMigrations(ctx, logger, c.DBConnString, dbMigrations)
	if err != nil {
		return meh.Wrap(err, "connect db and run migrations", meh.Details{"db_conn_string": c.DBConnString})
	}
	// Setup Redis.
	redisClient := redis.NewClient(&redis.Options{Addr: c.RedisAddr})
	// Setup Kafka.
	err = kafkautil.AwaitTopics(ctx, logger, c.KafkaAddr, event.UsersTopic, event.AuthTopic)
	if err != nil {
		return meh.Wrap(err, "await topics", meh.Details{"kafka_addr": c.KafkaAddr})
	}
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
	eg, egCtx := errgroup.WithContext(ctx)
	// Read messages.
	eg.Go(func() error {
		logger := logger.Named("kafka-reader")
		kafkaReader := kafkautil.NewReader(logger, c.KafkaAddr, kafkaGroupID,
			[]event.Topic{event.UsersTopic})
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
	return eg.Wait()
}
