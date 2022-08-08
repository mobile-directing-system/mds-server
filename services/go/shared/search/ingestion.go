package search

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"golang.org/x/sync/errgroup"
)

// DefaultBatchSize is the default size for batches in Rebuild.
const DefaultBatchSize = 64

// AddOrUpdateDocuments adds/updates the given Document for the Index. It does
// not wait for task completion.
func AddOrUpdateDocuments(c Client, index Index, doc ...Document) error {
	mapped := make([]map[string]any, 0, len(doc))
	for _, document := range doc {
		mapped = append(mapped, msDocumentFromDocument(document))
	}
	_, err := c.Index(index).UpdateDocuments(mapped)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "update documents", nil)
	}
	return nil
}

// DeleteDocumentsByUUID deletes the documents with the given ids.
func DeleteDocumentsByUUID(c Client, index Index, ids ...uuid.UUID) error {
	mapped := make([]string, 0, len(ids))
	for _, id := range ids {
		mapped = append(mapped, id.String())
	}
	_, err := c.Index(index).DeleteDocuments(mapped)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "delete documents", meh.Details{"ids": mapped})
	}
	return nil
}

// Rebuild deletes all documents for the given index, waits for task completion,
// and then adds documents by reading from the passed channel. If all documents
// were passed, the channel must be closed from the document retriever. The
// document retrieves is called as soon as all documents have been deleted.
func Rebuild(ctx context.Context, c Client, index Index, batchSize int,
	documentRetriever func(ctx context.Context, next chan<- Document) error) error {
	indexConfig, err := c.IndexConfig(index)
	if err != nil {
		return meh.Wrap(err, "index config from client", meh.Details{"index": index})
	}
	// Delete all documents.
	task, err := c.Index(index).DeleteAllDocuments()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "delete all documents", nil)
	}
	err = awaitTask(ctx, c, index, task.TaskUID)
	if err != nil {
		return meh.Wrap(err, "await task", meh.Details{"task_uid": task.TaskUID})
	}
	// Setup index again.
	err = launchIndex(ctx, c, index, indexConfig)
	if err != nil {
		return meh.Wrap(err, "launch index", meh.Details{
			"index":        index,
			"index_config": indexConfig,
		})
	}
	// Add all documents.
	nextDocuments := make(chan Document, 256)
	eg, egCtx := errgroup.WithContext(ctx)
	// Run document retriever.
	eg.Go(func() error {
		err := documentRetriever(egCtx, nextDocuments)
		if err != nil {
			return meh.Wrap(err, "run document retriever", nil)
		}
		return nil
	})
	// Run to-search-forwarder.
	eg.Go(func() error {
		// Optimized because of possibly high throughput.
		batch := make([]map[string]any, batchSize)
		done := false
		for !done {
			// Collect batch.
			currentBatchIndex := 0
		readNextDocuments:
			for {
				select {
				case <-egCtx.Done():
					return egCtx.Err()
				case nextDocuments, more := <-nextDocuments:
					if !more {
						done = true
						// Remove tail of batch.
						batch = batch[:currentBatchIndex]
						break readNextDocuments
					}
					batch[currentBatchIndex] = msDocumentFromDocument(nextDocuments)
					currentBatchIndex++
				}
			}
			// Add to index.
			if len(batch) == 0 {
				continue
			}
			_, err = c.Index(index).UpdateDocuments(batch)
			if err != nil {
				return meh.Wrap(err, "update documents", nil)
			}
			currentBatchIndex = 0
		}
		return nil
	})
	return eg.Wait()
}

func msDocumentFromDocument(d Document) map[string]any {
	m := make(map[string]any, len(d))
	for k, v := range d {
		m[string(k)] = v
	}
	return m
}
