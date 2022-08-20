package search

import (
	"context"
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
	"time"
)

type actionType string

const (
	actionTypeAddOrUpdateDocument actionType = "add-or-update-document"
	actionTypeDeleteDocument      actionType = "delete-document"
	actionTypeRebuildIndex        actionType = "rebuild-index"
)

// actionAddOrUpdateDocumentOptions are the options for actions with
// actionTypeAddOrUpdateDocument.
type actionAddOrUpdateDocumentOptions struct {
	// Document is the Document to add or update.
	Document Document `json:"document"`
}

// actionDeleteDocumentOptions are the options for actions with
// actionTypeDeleteDocument.
type actionDeleteDocumentOptions struct {
	// DocumentID is the id of the document to delete.
	DocumentID string `json:"document_id"`
}

// RebuildIndexFn is used in the action processor for handling calls for
// index-rebuild.
type RebuildIndexFn func(ctx context.Context, logger *zap.Logger, tx pgx.Tx, searchClient Client, index Index) error

type actionProcessor struct {
	logger       *zap.Logger
	store        actionStore
	searchClient Client
	rebuildIndex RebuildIndexFn
}

type action struct {
	// ID uniquely identifies the action. It also serves ordering purposes.
	ID int
	// Created is the timestamp of action creation.
	Created time.Time
	// Index is the index-name.
	Index Index
	// DocumentID for allowing additional concurrency and serving identifying
	// purposes.
	DocumentID nulls.String
	// ActionType determines the action to do.
	ActionType actionType
	// Options for the action.
	Options json.RawMessage
	// RemainingTries to perform. Also used for indicating open actions.
	RemainingTries int
	// RetryCooldown in case of processing-failure.
	RetryCooldown time.Duration
	// ProcessingTS is the timestamp of the last operation, performed on this
	// action. On Creation, it must be the same as Created.
	ProcessingTS time.Time
	// ErrMessage is valid, when action-processing failed and contains additional
	// error information.
	ErrMesage nulls.String
}

// Cooldowns for actionProcessor.run.
const (
	actionProcessorEmptyCooldown = 2 * time.Second
	actionProcessorErrorCooldown = 16 * time.Second
)

// run blocks until the given context is done. It checks and handles new actions
// and waits when no more available or in case of an error.
func (ap *actionProcessor) run(ctx context.Context) error {
	for {
		cooldown := time.Duration(0)
		more, err := ap.store.next(ctx, ap.logger, ap.handlerFn)
		if err != nil {
			mehlog.Log(ap.logger, meh.Wrap(err, "next in supplier", nil))
			// Probably hard error like database failure -> wait longer.
			cooldown = actionProcessorErrorCooldown
		} else if !more {
			// Polling delay.
			cooldown = actionProcessorEmptyCooldown
		}
		// Wait.
		if cooldown == 0 {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(cooldown):
		}
	}
}

// handlerFn handles events in run.
func (ap *actionProcessor) handlerFn(ctx context.Context, tx pgx.Tx, action action) error {
	switch action.ActionType {
	case actionTypeAddOrUpdateDocument:
		return meh.NilOrWrap(ap.handleActionTypeAddOrUpdateDocument(ctx, tx, action), "handle action-type add-or-update-documents", nil)
	case actionTypeDeleteDocument:
		return meh.NilOrWrap(ap.handleActionTypeDeleteDocument(ctx, tx, action), "handle action-type delete-document", nil)
	case actionTypeRebuildIndex:
		return meh.NilOrWrap(ap.handleActionTypeRebuildIndex(ctx, tx, action), "handle action-type rebuild-index", nil)
	default:
		return meh.NewInternalErr("unsupported action-type", meh.Details{"action_type": action.ActionType})
	}
}

// handleActionTypeAddOrUpdateDocument handles an action in handlerFn with
// actionTypeAddOrUpdateDocument.
func (ap *actionProcessor) handleActionTypeAddOrUpdateDocument(_ context.Context, _ pgx.Tx, action action) error {
	var options actionAddOrUpdateDocumentOptions
	err := json.Unmarshal(action.Options, &options)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal options", meh.Details{"was": options})
	}
	err = AddOrUpdateDocuments(ap.searchClient, action.Index, options.Document)
	if err != nil {
		return meh.Wrap(err, "add or update documents", nil)
	}
	return nil
}

// handleActionTypeDeleteDocument handles an action in handlerFn with
// actionTypeDeleteDocument.
func (ap *actionProcessor) handleActionTypeDeleteDocument(_ context.Context, _ pgx.Tx, action action) error {
	var options actionDeleteDocumentOptions
	err := json.Unmarshal(action.Options, &options)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "unmarshal options", meh.Details{"was": options})
	}
	err = DeleteDocumentsByID(ap.searchClient, action.Index, options.DocumentID)
	if err != nil {
		return meh.Wrap(err, "delete documents by id", nil)
	}
	return nil
}

// handleActionTypeRebuildIndex handles an action in handlerFn with
// actionTypeRebuildIndex.
func (ap *actionProcessor) handleActionTypeRebuildIndex(ctx context.Context, tx pgx.Tx, action action) error {
	if ap.rebuildIndex == nil {
		return meh.NewInternalErr("missing handler fn for index-rebuild", nil)
	}
	err := ap.rebuildIndex(ctx, ap.logger, tx, ap.searchClient, action.Index)
	if err != nil {
		return meh.Wrap(err, "rebuild index", nil)
	}
	return nil
}

type actionProcessorHandlerFn func(ctx context.Context, tx pgx.Tx, action action) error

type actionStore interface {
	// next retrieves the next action to process and calls the given
	// actionProcessorHandlerFn for it. It then returns whether new actions might be
	// available for avoiding unnecessary cooldown. If action processing fails, an
	// actionProcessingError will be returned. The given zap.Logger is used for
	// logging processing // errors.
	next(ctx context.Context, logger *zap.Logger, handlerFn actionProcessorHandlerFn) (bool, error)
}

type dbActionStore struct {
	txSupplier pgutil.DBTxSupplier
	dialect    goqu.DialectWrapper
}

// next retrieves the next action to process, calls the actionProcessorHandlerFn
// and writes the result. The given zap.Logger is used for logging processing
// errors.
func (aps dbActionStore) next(ctx context.Context, logger *zap.Logger, handlerFn actionProcessorHandlerFn) (bool, error) {
	var more bool
	err := pgutil.RunInTx(ctx, aps.txSupplier, func(ctx context.Context, tx pgx.Tx) error {
		// Select oldest per group, that is still open.
		oldestPerGroupQuery, _, err := aps.dialect.From(goqu.T("__search_actions")).
			Select(goqu.MIN(goqu.C("id"))).
			GroupBy(goqu.C("index"), goqu.C("document_id")).
			Where(goqu.C("remaining_tries").Gt(0)).ToSQL()
		if err != nil {
			return meh.NewInternalErrFromErr(err, "oldest-per-group-query to sql", nil)
		}
		oldestPerGroupRows, err := tx.Query(ctx, oldestPerGroupQuery)
		if err != nil {
			return mehpg.NewQueryDBErr(err, "exec oldest-per-group-query", oldestPerGroupQuery)
		}
		defer oldestPerGroupRows.Close()
		oldestPerGroup := make([]int, 0)
		var id int
		for oldestPerGroupRows.Next() {
			err = oldestPerGroupRows.Scan(&id)
			if err != nil {
				return mehpg.NewScanRowsErr(err, "scan oldest-per-group-row", oldestPerGroupQuery)
			}
			oldestPerGroup = append(oldestPerGroup, id)
		}
		oldestPerGroupRows.Close()
		if len(oldestPerGroup) == 0 {
			// Nothing to do.
			more = false
			return nil
		}
		// Select next.
		nextQuery, _, err := aps.dialect.From(goqu.T("__search_actions")).
			Select(goqu.C("id"),
				goqu.C("created"),
				goqu.C("index"),
				goqu.C("document_id"),
				goqu.C("action_type"),
				goqu.C("options"),
				goqu.C("remaining_tries"),
				goqu.C("retry_cooldown"),
				goqu.C("processing_ts"),
				goqu.C("err_message")).
			ForUpdate(exp.SkipLocked).
			Where(goqu.C("remaining_tries").Gt(0),
				goqu.C("id").In(oldestPerGroup),
				goqu.Or(goqu.C("err_message").IsNull(),
					goqu.C("processing_ts").Lt(goqu.L("now() - interval '1 ms' * retry_cooldown / 1000000")))).
			Order(goqu.C("id").Asc()).
			Limit(1).ToSQL()
		if err != nil {
			return meh.NewInternalErrFromErr(err, "next-query to sql", nil)
		}
		nextRows, err := tx.Query(ctx, nextQuery)
		if err != nil {
			return mehpg.NewQueryDBErr(err, "exec next-query", nextQuery)
		}
		defer nextRows.Close()
		if !nextRows.Next() {
			// All are currently processed (we skipped locked ones).
			more = false
			return nil
		}
		more = true
		var next action
		err = nextRows.Scan(&next.ID,
			&next.Created,
			&next.Index,
			&next.DocumentID,
			&next.ActionType,
			&next.Options,
			&next.RemainingTries,
			&next.RetryCooldown,
			&next.ProcessingTS,
			&next.ErrMesage)
		if err != nil {
			return mehpg.NewScanRowsErr(err, "scan next-row", nextQuery)
		}
		nextRows.Close()
		// Handle.
		var newErrMessage nulls.String
		newRemainingTries := next.RemainingTries - 1
		handlerErr := pgutil.RunInTx(ctx, aps.txSupplier, func(ctx context.Context, tx pgx.Tx) error {
			return meh.NilOrWrap(handlerFn(ctx, tx, next), "handle action", nil)
		})
		if handlerErr != nil {
			newErrMessage = nulls.NewString(handlerErr.Error())
			// Log error.
			handlerErr = meh.Wrap(handlerErr, "handle action in tx", meh.Details{"action": next})
			handlerErr = meh.ApplyCode(handlerErr, meh.ErrInternal)
			mehlog.Log(logger, handlerErr)
		} else {
			newRemainingTries = 0
		}
		// Write result.
		writeResultQuery, _, err := aps.dialect.Update(goqu.T("__search_actions")).Set(goqu.Record{
			"remaining_tries": newRemainingTries,
			"processing_ts":   time.Now().UTC(),
			"err_message":     newErrMessage,
		}).Where(goqu.C("id").Eq(next.ID)).ToSQL()
		if err != nil {
			return meh.Wrap(err, "write-result-query to sql", nil)
		}
		_, err = tx.Exec(ctx, writeResultQuery)
		if err != nil {
			return mehpg.NewQueryDBErr(err, "exec write-result-query", writeResultQuery)
		}
		return nil
	})
	if err != nil {
		return false, meh.Wrap(err, "run in tx", nil)
	}
	return more, nil
}
