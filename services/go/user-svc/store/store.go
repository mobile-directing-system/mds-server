package store

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"go.uber.org/zap"
)

// Mall provides all store access methods.
type Mall struct {
	dialect      goqu.DialectWrapper
	searchClient search.Client
	searchConfig search.ClientConfig
}

// NewMall creates a new Mall with postgres dialect.
func NewMall(logger *zap.Logger, searchHost string, searchMasterKey string) *Mall {
	return &Mall{
		dialect: goqu.Dialect("postgres"),
		searchConfig: search.ClientConfig{
			Host:      searchHost,
			MasterKey: searchMasterKey,
			IndexConfigs: map[search.Index]search.IndexConfig{
				userSearchIndex: userSearchIndexConfig,
			},
			Logger:  logger.Named("search"),
			Timeout: search.DefaultClientTimeout,
		},
	}
}

// Open sets up and connects the Mall.
func (m *Mall) Open(ctx context.Context) error {
	searchClient, err := search.Launch(ctx, m.searchConfig)
	if err != nil {
		return meh.Wrap(err, "launch search client", meh.Details{"config": m.searchConfig})
	}
	m.searchClient = searchClient
	return nil
}
