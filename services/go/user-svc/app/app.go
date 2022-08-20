package app

import (
	"context"
	"embed"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/connectutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgconnect"
	"github.com/mobile-directing-system/mds-server/services/go/shared/ready"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/endpoints"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/eventport"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
	"golang.org/x/sync/errgroup"
	"io/fs"
)

//go:embed db-migrations/*.sql
var dbMigrationsEmbedded embed.FS

var dbMigrations fs.FS

const dbScope = "app"

func init() {
	var err error
	dbMigrations, err = fs.Sub(dbMigrationsEmbedded, "db-migrations")
	if err != nil {
		panic(meh.NewInternalErrFromErr(err, "create sub fs for database migrations", nil).Error())
	}
}

// Run the application.
func Run(ctx context.Context) error {
	c, err := parseConfigFromEnv()
	if err != nil {
		return meh.Wrap(err, "parse config", nil)
	}
	logger, err := logging.NewLogger("user-svc", c.LogLevel)
	if err != nil {
		return meh.Wrap(err, "new logger", nil)
	}
	defer func() { _ = logger.Sync() }()
	logging.SetDebugLogger(logger.Named("debug"))
	eg, egCtx := errgroup.WithContext(ctx)
	probeServer, startUpCompleted := ready.NewServer(logger.Named("probe-server"))
	eg.Go(func() error {
		err := probeServer.Serve(egCtx, c.ReadyProbeServeAddr)
		return meh.NilOrWrap(err, "serve ready-probe-server", meh.Details{"addr": c.ReadyProbeServeAddr})
	})
	// Connect to database.
	sqlDB, err := pgconnect.ConnectAndRunMigrations(ctx, logger, c.DBConnString, dbScope, dbMigrations)
	if err != nil {
		return meh.Wrap(err, "connect db and run migrations", meh.Details{"db_conn_string": c.DBConnString})
	}
	// Await ready.
	readyCheck := func(ctx context.Context) error {
		eg, egCtx := errgroup.WithContext(ctx)
		// Check hosts.
		eg.Go(func() error {
			err := connectutil.AwaitHostsReachable(egCtx, c.KafkaAddr, c.searchConfig.Host)
			return meh.NilOrWrap(err, "await hosts reachable", nil)
		})
		// Check Kafka topics.
		eg.Go(func() error {
			err := kafkautil.AwaitTopics(egCtx, c.KafkaAddr, event.UsersTopic)
			return meh.NilOrWrap(err, "await topics", meh.Details{"kafka_addr": c.KafkaAddr})
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
	// Setup.
	kafkaConnector, err := kafkautil.InitNewConnector(ctx, logger.Named("kafka-connector"), sqlDB)
	if err != nil {
		return meh.Wrap(err, "init new kafka connector", nil)
	}
	eventPort := eventport.NewPort(kafkaConnector)
	mall, err := store.InitNewMall(ctx, logger.Named("mall"), sqlDB, c.searchConfig.Host, c.searchConfig.MasterKey)
	if err != nil {
		return meh.Wrap(err, "init new mall", nil)
	}
	ctrl := &controller.Controller{
		Logger:   logger.Named("controller"),
		DB:       sqlDB,
		Store:    mall,
		Notifier: eventPort,
	}
	// Open mall.
	eg.Go(func() error {
		err := mall.Open(egCtx)
		if err != nil {
			return meh.Wrap(err, "open mall", nil)
		}
		return nil
	})
	// Run controller.
	eg.Go(func() error {
		err := ctrl.Run(egCtx)
		if err != nil {
			return meh.Wrap(err, "run controller", nil)
		}
		return nil
	})
	// Serve endpoints.
	eg.Go(func() error {
		err = endpoints.Serve(egCtx, logger.Named("endpoints"), c.ServeAddr, c.AuthTokenSecret, ctrl)
		if err != nil {
			return meh.Wrap(err, "serve endpoints", meh.Details{"serve_addr": c.ServeAddr})
		}
		return nil
	})
	// Run Kafka connector.
	eg.Go(func() error {
		logger := logger.Named("kafka-reader")
		kafkaWriter := kafkautil.NewWriter(logger.Named("kafka"), c.KafkaAddr)
		err := kafkaConnector.PumpOutgoing(egCtx, sqlDB, kafkaWriter)
		if err != nil {
			return meh.Wrap(err, "pump outgoing", nil)
		}
		return nil
	})
	startUpCompleted(readyCheck)
	return eg.Wait()
}
