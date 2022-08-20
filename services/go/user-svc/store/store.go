package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
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

// InitNewMall prepares a new Mall with postgres-dialect and sets up search. Do
// not forget to call Mall.Open!
func InitNewMall(ctx context.Context, logger *zap.Logger, txSupplier pgutil.DBTxSupplier, searchHost string, searchMasterKey string) (*Mall, error) {
	m := &Mall{
		dialect: goqu.Dialect("postgres"),
	}
	m.searchConfig = search.ClientConfig{
		Host:      searchHost,
		MasterKey: searchMasterKey,
		IndexConfigs: map[search.Index]search.IndexConfig{
			userSearchIndex: userSearchIndexConfig,
		},
		Logger:  logger.Named("search"),
		Timeout: search.DefaultClientTimeout,
		RebuildIndexFn: func(ctx context.Context, logger *zap.Logger, tx pgx.Tx, searchClient search.Client, index search.Index) error {
			switch index {
			case userSearchIndex:
				return meh.NilOrWrap(rebuildUserSearchIndex(ctx, m.dialect, tx, searchClient), "rebuild user-search-index", nil)
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

// rebuildUserSearchIndex in search.ClientConfig in InitNewMall.
func rebuildUserSearchIndex(ctx context.Context, dialect goqu.DialectWrapper, tx pgx.Tx, searchClient search.Client) error {
	err := search.Rebuild(ctx, searchClient, userSearchIndex, search.DefaultBatchSize,
		func(ctx context.Context, next chan<- search.Document) error {
			defer close(next)
			// Build query.
			q, _, err := dialect.From(goqu.T("users")).
				Select(goqu.C("id"),
					goqu.C("username"),
					goqu.C("first_name"),
					goqu.C("last_name"),
					goqu.C("is_admin")).ToSQL()
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
				var user User
				err = rows.Scan(&user.ID,
					&user.Username,
					&user.FirstName,
					&user.LastName,
					&user.IsAdmin)
				if err != nil {
					return mehpg.NewScanRowsErr(err, "scan row", q)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case next <- documentFromUser(user):
				}
			}
			return nil
		})
	if err != nil {
		return meh.Wrap(err, "rebuild search", nil)
	}
	return nil
}
