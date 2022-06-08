package pgmigrate

import (
	"context"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
	"sort"
	"strings"
	"time"
)

// ZeroVersion is the initial version number to use when no migrations have been
// performed.
const ZeroVersion = -1

// Migrator allows running all needed database migrations. Create a Migrator
// with NewMigrator.
type Migrator interface {
	// Up creates the migration log table if not existing, performs any migrations
	// to-do and logs information and results to the given zap.Logger.
	Up(ctx context.Context, logger *zap.Logger, db *pgx.Conn) error
}

// Migration holds all metadata for an available Migration.
type Migration struct {
	// Name is a human-readable name of the migration.
	Name string
	// TargetVersion is the version SQL migrates to.
	TargetVersion int
	// SQL is the query to run.
	SQL string
}

// migrator is the actual implementation of Migrator, created via NewMigrator.
type migrator struct {
	// migrations that are set up for the migrator. When created using NewMigrator,
	// migrations holds no one with the same target version.
	migrations []Migration
	// migrationLogTable is the name of the table the migration log is kept in.
	migrationLogTable string
}

// NewMigrator creates a new Migrator using the given Migration list with all
// available migrations as well as the name of the table, the migration log is
// kept in.
func NewMigrator(migrations []Migration, migrationLogTable string) (Migrator, error) {
	// Assure no duplicate target versions.
	migrationVersions := make(map[int]struct{})
	for _, migration := range migrations {
		if _, ok := migrationVersions[migration.TargetVersion]; ok {
			return nil, meh.NewInternalErr("migration with duplicate target version", meh.Details{
				"migration_name":           migration.Name,
				"migration_target_version": migration.TargetVersion,
			})
		}
		migrationVersions[migration.TargetVersion] = struct{}{}
	}
	return &migrator{
		migrations:        migrations,
		migrationLogTable: migrationLogTable,
	}, nil
}

// Up creates the migration log table if not existing, performs any migrations
// to-do and logs information and results to the given zap.Logger.
func (m *migrator) Up(ctx context.Context, logger *zap.Logger, db *pgx.Conn) error {
	// Assure log table exists.
	err := assureMigrationsLogTable(ctx, db, m.migrationLogTable)
	if err != nil {
		return meh.Wrap(err, "assure migrations log table", meh.Details{"migration_log_table": m.migrationLogTable})
	}
	// Load migration log.
	logEntries, err := getMigrationLog(ctx, db, m.migrationLogTable)
	if err != nil {
		return meh.Wrap(err, "get migration log", meh.Details{"migration_log_table": m.migrationLogTable})
	}
	currentVersion := currentVersion(logEntries)
	migrationsToDo := migrationsToDoOrdered(currentVersion, m.migrations)
	if len(migrationsToDo) == 0 {
		logger.Info("database schema up-to-date", zap.Int("current_version", currentVersion))
		return nil
	}
	// Run migrations.
	for i, migration := range migrationsToDo {
		logger.Info(fmt.Sprintf("running database migration %d/%d", i, len(migrationsToDo)),
			zap.String("name", migration.Name),
			zap.Int("target_version", migration.TargetVersion))
		err = m.runMigration(ctx, db, migration)
		if err != nil {
			return meh.Wrap(err, "run migration", meh.Details{
				"migration_name":           migration.Name,
				"migration_target_version": migration.TargetVersion,
			})
		}
	}
	return nil
}

// runMigration runs the given Migration in a transaction and logs the result.
func (m *migrator) runMigration(ctx context.Context, db *pgx.Conn, migration Migration) error {
	// Begin tx.
	tx, err := db.Begin(ctx)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "begin tx", nil)
	}
	defer pgutil.Rollback(ctx, tx)
	// Run migration.
	_, err = tx.Exec(ctx, migration.SQL)
	if err != nil {
		err = meh.Wrap(err, "exec migration", meh.Details{"query": migration.SQL})
	}
	logErr := m.logMigrationResult(ctx, db, migration, err)
	if err == nil && logErr != nil {
		err = meh.Wrap(err, "log migration result", nil)
	}
	if err != nil {
		return err
	}
	// Commit tx.
	err = tx.Commit(ctx)
	if err != nil {
		return meh.Wrap(err, "commit tx", nil)
	}
	return nil
}

// logMigrationResult logs the given Migration with the result based on the
// passed error.
func (m *migrator) logMigrationResult(ctx context.Context, db *pgx.Conn, migration Migration, e error) error {
	var errMessage string
	if e != nil {
		errMessage = e.Error()
	}
	// Build query.
	q, _, err := goqu.Insert(goqu.T(m.migrationLogTable)).Rows(goqu.Record{
		"ts":             time.Now(),
		"name":           migration.Name,
		"target_version": migration.TargetVersion,
		"success":        e == nil,
		"err_message":    errMessage,
	}).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Exec.
	_, err = db.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// migrationsToDoOrdered returns an ordered list of the given Migration list
// that have target versions above the given current version.
func migrationsToDoOrdered(currentVersion int, migrations []Migration) []Migration {
	migrationsToDo := make([]Migration, 0)
	for _, migration := range migrations {
		if migration.TargetVersion > currentVersion {
			migrationsToDo = append(migrationsToDo, migration)
		}
	}
	// Sort.
	sort.Slice(migrationsToDo, func(i, j int) bool {
		return migrationsToDo[i].TargetVersion < migrationsToDo[j].TargetVersion
	})
	return migrationsToDo
}

// assureMigrationsLogTable runs the migrationLogCreateQ with the given table
// name.
func assureMigrationsLogTable(ctx context.Context, db *pgx.Conn, migrationLogTable string) error {
	if strings.Contains(migrationLogTable, "'") || strings.Contains(migrationLogTable, `"`) {
		return meh.NewInternalErr("no quotes allowed in migration log table name", meh.Details{"was": migrationLogTable})
	}
	migrationLogCreateQ := fmt.Sprintf(`create table if not exists %s
(
    id             serial
        constraint table_name_pk
            primary key,
    ts             timestamp not null,
    name           varchar not null,
    target_version int     not null,
    success        boolean not null,
    err_message    text
);`, migrationLogTable)
	_, err := db.Exec(ctx, migrationLogCreateQ)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec migration log creation query", migrationLogCreateQ)
	}
	return nil
}

// MigrationLogEntry holds all metadata for an attempted database migration.
type MigrationLogEntry struct {
	// ID identifies the log entry.
	ID int
	// Timestamp is when the migration had finished.
	Timestamp time.Time
	// Name is a human-readable name of the migration that was performed.
	Name string
	// TargetVersion is the version, the migration attempted to migrate to.
	TargetVersion int
	// Success describes if the migration was successful.
	Success bool
	// ErrMessage holds an optional error message.
	ErrMessage nulls.String
}

// currentVersion checks for the latest successful target version, a migration
// has been performed for.
func currentVersion(logEntries []MigrationLogEntry) int {
	maxVersion := ZeroVersion
	for _, entry := range logEntries {
		if !entry.Success {
			continue
		}
		if entry.TargetVersion > maxVersion {
			maxVersion = entry.TargetVersion
		}
	}
	return maxVersion
}

// getMigrationLog retrieves all available migrations in the migration log table
// with the given name.
func getMigrationLog(ctx context.Context, db *pgx.Conn, migrationLogTable string) ([]MigrationLogEntry, error) {
	// Build query.
	q, _, err := goqu.From(goqu.T(migrationLogTable)).
		Select(goqu.C("id"),
			goqu.C("ts"),
			goqu.C("name"),
			goqu.C("target_version"),
			goqu.C("success"),
			goqu.C("err_message")).ToSQL()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	// Query.
	rows, err := db.Query(ctx, q)
	if err != nil {
		return nil, mehpg.NewQueryDBErr(err, "query db", q)
	}
	defer rows.Close()
	// Scan.
	logEntries := make([]MigrationLogEntry, 0)
	for rows.Next() {
		var logEntry MigrationLogEntry
		err = rows.Scan(&logEntry.ID,
			&logEntry.Timestamp,
			&logEntry.Name,
			&logEntry.TargetVersion,
			&logEntry.Success,
			&logEntry.ErrMessage)
		if err != nil {
			return nil, mehpg.NewScanRowsErr(err, "scan row", q)
		}
		logEntries = append(logEntries, logEntry)
	}
	return logEntries, nil
}
