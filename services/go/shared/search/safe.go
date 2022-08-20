package search

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/lefinal/nulls"
	"golang.org/x/sync/errgroup"
	"time"
)

func (c *client) Run(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return meh.NilOrWrap(c.actionProcessor.run(egCtx), "run action processor", nil)
	})
	return eg.Wait()
}

func (c *client) SafeAddOrUpdateDocument(ctx context.Context, tx pgx.Tx, index Index, document Document) error {
	// Try to extract the document id.
	ixCfg, err := c.IndexConfig(index)
	if err != nil {
		return meh.Wrap(err, "get index config", meh.Details{"index": index})
	}
	var documentID nulls.String
	if idFromDoc, ok := document[ixCfg.PrimaryKey]; ok {
		documentID = nulls.NewString(fmt.Sprintf("%v", idFromDoc))
	}
	// Schedule.
	err = scheduleSafeAction(ctx, tx, index, documentID, actionTypeAddOrUpdateDocument,
		actionAddOrUpdateDocumentOptions{Document: document})
	if err != nil {
		return meh.Wrap(err, "schedule safe action", meh.Details{"document": document})
	}
	return nil
}

func (c *client) SafeDeleteDocumentByUUID(ctx context.Context, tx pgx.Tx, index Index, id uuid.UUID) error {
	return c.SafeDeleteDocumentByID(ctx, tx, index, id.String())
}

func (c *client) SafeDeleteDocumentByID(ctx context.Context, tx pgx.Tx, index Index, id string) error {
	err := scheduleSafeAction(ctx, tx, index, nulls.NewString(id), actionTypeDeleteDocument,
		actionDeleteDocumentOptions{DocumentID: id})
	if err != nil {
		return meh.Wrap(err, "schedule safe action", meh.Details{"document_id": id})
	}
	return nil
}

func (c *client) SafeRebuildIndex(ctx context.Context, tx pgx.Tx, index Index) error {
	err := scheduleSafeAction(ctx, tx, index, nulls.String{}, actionTypeRebuildIndex, nil)
	if err != nil {
		return meh.Wrap(err, "schedule safe action", nil)
	}
	return nil
}

const (
	actionDefaultRemainingRetries = 16
	actionDefaultRetryCooldown    = 16 * time.Second
)

// scheduleSafeActions schedules the given action list.
func scheduleSafeAction(ctx context.Context, tx pgx.Tx, index Index, documentID nulls.String, actionType actionType, options any) error {
	optionsRaw, err := json.Marshal(options)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "marshal options", nil)
	}
	q, _, err := goqu.Insert(goqu.T("__search_actions")).Rows(goqu.Record{
		"created":         time.Now().UTC(),
		"index":           index,
		"document_id":     documentID,
		"action_type":     actionType,
		"options":         optionsRaw,
		"remaining_tries": actionDefaultRemainingRetries,
		"retry_cooldown":  actionDefaultRetryCooldown,
		"processing_ts":   time.Now().UTC(),
	}).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}
