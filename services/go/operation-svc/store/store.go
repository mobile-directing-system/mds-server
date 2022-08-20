package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehpg"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Mall provides all store access methods.
type Mall struct {
	dialect      goqu.DialectWrapper
	searchClient search.SafeClient
	searchConfig search.ClientConfig
}

// InitNewMall creates and sets up a new Mall with postgres dialect. Do not
// forget to call Mall.Open!
func InitNewMall(ctx context.Context, logger *zap.Logger, txSupplier pgutil.DBTxSupplier, searchHost string, searchMasterKey string) (*Mall, error) {
	m := &Mall{
		dialect: goqu.Dialect("postgres"),
	}
	m.searchConfig = search.ClientConfig{
		Host:      searchHost,
		MasterKey: searchMasterKey,
		IndexConfigs: map[search.Index]search.IndexConfig{
			operationSearchIndex: operationSearchIndexConfig,
		},
		Logger:  logger.Named("search"),
		Timeout: search.DefaultClientTimeout,
		RebuildIndexFn: func(ctx context.Context, logger *zap.Logger, tx pgx.Tx, searchClient search.Client, index search.Index) error {
			switch index {
			case operationSearchIndex:
				return meh.NilOrWrap(rebuildOperationSearchIndex(ctx, m.dialect, tx, searchClient), "rebuild operation-search-index", nil)
			default:
				return meh.NewInternalErr("unsupported index", meh.Details{"index": index})
			}
		},
	}
	// Launch search.
	searchClient, err := search.LaunchSafe(ctx, logger.Named("search"), txSupplier, m.searchConfig)
	if err != nil {
		return nil, meh.Wrap(err, "launch search client", meh.Details{"config": m.searchConfig})
	}
	m.searchClient = searchClient
	return m, nil
}

// Open the Mall. This blocks until the given context is done.
func (m *Mall) Open(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return meh.NilOrWrap(m.searchClient.Run(egCtx), "run search-client", nil)
	})
	return eg.Wait()
}

// rebuildOperationSearchIndex in search.ClientConfig in InitNewMall.
func rebuildOperationSearchIndex(ctx context.Context, dialect goqu.DialectWrapper, tx pgx.Tx, searchClient search.Client) error {
	err := search.Rebuild(ctx, searchClient, operationSearchIndex, search.DefaultBatchSize,
		func(ctx context.Context, next chan<- search.Document) error {
			defer close(next)
			// Build query.
			q, _, err := dialect.From(goqu.T("operations")).
				Select(goqu.I("operations.id"),
					goqu.I("operations.title"),
					goqu.I("operations.description"),
					goqu.I("operations.start_ts"),
					goqu.I("operations.end_ts"),
					goqu.I("operations.is_archived"),
					dialect.From(goqu.T("operation_members")).
						Select(goqu.I("operation_members.user")).
						Where(goqu.I("operation_members.operation").Eq(goqu.I("operations.id")))).ToSQL()
			if err != nil {
				return meh.Wrap(err, "query to sql", nil)
			}
			// Query.
			rows, err := tx.Query(ctx, q)
			if err != nil {
				return mehpg.NewQueryDBErr(err, "query db", q)
			}
			defer rows.Close()
			// Scan.
			for rows.Next() {
				var operation Operation
				var operationMembers []string
				err = rows.Scan(&operation.ID,
					&operation.Title,
					&operation.Description,
					&operation.Start,
					&operation.End,
					&operation.IsArchived,
					&operationMembers)
				if err != nil {
					return mehpg.NewScanRowsErr(err, "scan row", q)
				}
				operationMembersUUID := make([]uuid.UUID, 0, len(operationMembers))
				for _, member := range operationMembers {
					id, err := uuid.FromString(member)
					if err != nil {
						return meh.NewInternalErrFromErr(err, "parse member uuid", meh.Details{"was": member})
					}
					operationMembersUUID = append(operationMembersUUID, id)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case next <- documentFromOperation(operation, operationMembersUUID):
				}
			}
			return nil
		})
	if err != nil {
		return meh.Wrap(err, "rebuild search", nil)
	}
	return nil
}
