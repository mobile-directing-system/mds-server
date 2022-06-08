package app

import (
	"context"
	"embed"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/endpoints"
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
	defer func() { _ = logger.Sync() }()
	// Connect to database.
	_, err = pgconnect.ConnectAndRunMigrations(ctx, logger, c.DBConnString, dbMigrations)
	if err != nil {
		return meh.Wrap(err, "connect db and run migrations", meh.Details{"db_conn_string": c.DBConnString})
	}
	// Setup controller.
	ctrl := &controller.Controller{
		Logger:     logger.Named("controller"),
		HMACSecret: "meow", // TODO: CHANGE
	}
	eg, egCtx := errgroup.WithContext(ctx)
	// Serve endpoints.
	eg.Go(func() error {
		err := endpoints.Serve(egCtx, logger.Named("endpoints"), c.ServeAddr, c.ForwardAddr, ctrl)
		if err != nil {
			return meh.Wrap(err, "serve endpoints", nil)
		}
		return nil
	})
	// TODO
	// redis.NewClient(&redis.Options{Addr: c.RedisAddr})
	return eg.Wait()
}
