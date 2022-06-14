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
	logging.SetDebugLogger(logger.Named("debug"))
	// Connect to database.
	sqlDB, err := pgconnect.ConnectAndRunMigrations(ctx, logger, c.DBConnString, dbMigrations)
	if err != nil {
		return meh.Wrap(err, "connect db and run migrations", meh.Details{"db_conn_string": c.DBConnString})
	}
	// Wait for Kafka.
	err = connectutil.AwaitHostsReachable(ctx, c.KafkaAddr)
	if err != nil {
		return meh.Wrap(err, "await hosts reachable", nil)
	}
	// Setup.
	err = kafkautil.AwaitTopics(ctx, logger, c.KafkaAddr, event.UsersTopic)
	if err != nil {
		return meh.Wrap(err, "await topics", meh.Details{"kafka_addr": c.KafkaAddr})
	}
	eventPort := eventport.NewPort(kafkautil.NewWriter(logger.Named("kafka"), c.KafkaAddr))
	ctrl := &controller.Controller{
		Logger:   logger.Named("controller"),
		DB:       sqlDB,
		Store:    store.NewMall(),
		Notifier: eventPort,
	}
	eg, egCtx := errgroup.WithContext(ctx)
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
	return eg.Wait()
}
