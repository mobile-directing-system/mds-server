package pgutil

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"strings"
	"time"
)

// Rollback the given pgx.Tx. If rollback fails, the error will be logged to
// logging.DebugLogger.
func Rollback(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if err != nil && err != pgx.ErrTxClosed {
		mehlog.Log(logging.DebugLogger(), meh.NewInternalErrFromErr(err, "rollback tx", nil))
		return
	}
}

// SanitizeString ensures that no malicious quotes are present.
func SanitizeString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// commitTimeout is the timeout to use in RunInTx for committing transactions as
// well as for rolling back in case of failed commit.
const commitTimeout = 15 * time.Second

// DBTxSupplier servers as an abstraction for pgxpool.Pool in RunInTx for better
// testability.
type DBTxSupplier interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// RunInTx is a transaction wrapper for the given function, that needs
// isolation. If function execution fails, the transaction is rolled back.
func RunInTx(ctx context.Context, txSupplier DBTxSupplier, fn func(ctx context.Context, tx pgx.Tx) error) error {
	// Begin tx.
	tx, err := txSupplier.Begin(ctx)
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
	// Commit tx with timeout. The reason for this is that we may run other
	// important actions in the function which might take long. An example for this
	// could be creating Kafka events. If the transaction is then rolled back
	// because of the context being done while trying to commit, this could lead to
	// unwanted behavior.
	commitTimeoutCtx, cancelCommitTimeout := context.WithTimeout(context.Background(), commitTimeout)
	defer cancelCommitTimeout()
	err = tx.Commit(commitTimeoutCtx)
	if err != nil {
		// Rollback.
		var details meh.Details
		rollbackTimeoutCtx, cancelRollbackTimeout := context.WithTimeout(context.Background(), commitTimeout)
		defer cancelRollbackTimeout()
		rollbackErr := tx.Rollback(rollbackTimeoutCtx)
		if rollbackErr != nil {
			mehlog.Log(logging.DebugLogger(), meh.NewInternalErrFromErr(err, "rollback tx because of failed commit", nil))
			details = meh.Details{"rollback_err": rollbackErr.Error()}
		}
		return meh.Wrap(err, "commit tx", details)
	}
	return nil
}
