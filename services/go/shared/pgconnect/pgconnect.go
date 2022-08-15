package pgconnect

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgmigrate"
	"go.uber.org/zap"
	"io/fs"
	"time"
)

const (
	// RetryTimeout is the timeout to wait while retrying the connection to the
	// database in Connect.
	RetryTimeout = 3 * time.Second
	// DefaultMigrationLogTable is the default table to use for keeping the
	// migration log.
	DefaultMigrationLogTable = "__db_migration_log"
)

// Connect tries to connect to the given postgres database. If connection fails,
// we wait for RetryTimeout and then try again. Connection errors will be logged
// to the given zap.Logger.
func Connect(ctx context.Context, logger *zap.Logger, connString string) (*pgxpool.Pool, error) {
	// Create config.
	connConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "parse connection string config", meh.Details{"conn_string": connString})
	}
	connConfig.MaxConns = 64
	connConfig.MinConns = 0
	connConfig.MaxConnIdleTime = 10 * time.Second
	connConfig.HealthCheckPeriod = 5 * time.Second
	// Connect.
	conn, ok := tryConnect(ctx, logger, connConfig)
	if ok {
		return conn, nil
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(RetryTimeout):
			conn, ok = tryConnect(ctx, logger, connConfig)
			if ok {
				return conn, nil
			}
		}
	}
}

// tryConnect tries to connect to the given postgres database. If connection
// fails, the error will be logged to the given zap.Logger and false returned as
// second return value.
func tryConnect(ctx context.Context, logger *zap.Logger, config *pgxpool.Config) (*pgxpool.Pool, bool) {
	conn, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		mehlog.Log(logger, meh.NewInternalErrFromErr(err, "connect to database", nil))
		return nil, false
	}
	return conn, true
}

// ConnectAndRunMigrations runs Connect for the given connection string. It then
// extracts migrations from the given fs.FS using pgmigrate.MigrationsFromFS and
// runs them using the pgmigrate.Migrator and DefaultMigrationLogTable.
func ConnectAndRunMigrations(ctx context.Context, logger *zap.Logger, connString string, scope string, migrationsFS fs.FS) (*pgxpool.Pool, error) {
	// Connect.
	connPool, err := Connect(ctx, logger, connString)
	if err != nil {
		return nil, meh.Wrap(err, "connect", meh.Details{"conn_string": connPool})
	}
	// Extract migrations.
	migrations, err := pgmigrate.MigrationsFromFS(migrationsFS)
	if err != nil {
		return nil, meh.Wrap(err, "migrations from fs", nil)
	}
	// Run migrations.
	migrator, err := pgmigrate.NewMigrator(migrations, DefaultMigrationLogTable, scope)
	if err != nil {
		return nil, meh.Wrap(err, "new migrator", meh.Details{"migration_log_table": DefaultMigrationLogTable})
	}
	singleConn, err := connPool.Acquire(ctx)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "acquire pgx conn", nil)
	}
	defer singleConn.Release()
	err = migrator.Up(ctx, logger, singleConn.Conn())
	if err != nil {
		return nil, meh.Wrap(err, "migrator up", nil)
	}
	return connPool, nil
}
