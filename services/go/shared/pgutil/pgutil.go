package pgutil

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"strings"
)

// Rollback the given pgx.Tx. If rollback fails, the error will be logged to
// logging.DebugLogger.
func Rollback(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if err != nil {
		mehlog.Log(logging.DebugLogger(), meh.NewInternalErrFromErr(err, "rollback tx", nil))
		return
	}
}

// SanitizeString ensures that no malicious quotes are present.
func SanitizeString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// RunInTx is a transaction wrapper for the given function, that needs
// isolation. If function execution fails, the transaction is rolled back.
func RunInTx(ctx context.Context, sqlDB *pgxpool.Pool, fn func(ctx context.Context, tx pgx.Tx) error) error {
	// Begin tx.
	tx, err := sqlDB.Begin(ctx)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "begin tx", nil)
	}
	// Run stuff.
	err = fn(ctx, tx)
	if err != nil {
		// Rollback.
		var details meh.Details
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			mehlog.Log(logging.DebugLogger(), meh.NewInternalErrFromErr(err, "rollback tx because of failed fn execution", nil))
			details = meh.Details{"rollback_err": rollbackErr.Error()}
		}
		return meh.Wrap(err, "run fn", details)
	}
	// Commit tx.
	err = tx.Commit(ctx)
	if err != nil {
		// Rollback.
		var details meh.Details
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			mehlog.Log(logging.DebugLogger(), meh.NewInternalErrFromErr(err, "rollback tx because of failed commit", nil))
			details = meh.Details{"rollback_err": rollbackErr.Error()}
		}
		return meh.Wrap(err, "commit tx", details)
	}
	return nil
}
