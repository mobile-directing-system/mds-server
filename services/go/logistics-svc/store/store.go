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
	"log"
)

// Mall provides all store access methods.
type Mall struct {
	dialect          goqu.DialectWrapper
	channelOperators map[ChannelType]channelOperator
	searchClient     search.SafeClient
	searchConfig     search.ClientConfig
}

// ChannelTypeSupplier is the global channelOperatorSupplier that is
// used when creating Mall with InitNewMall and also for validation of supported
// channel types. Will be set in init.
var ChannelTypeSupplier channelTypesSupplier

func init() {
	ChannelTypeSupplier = channelTypesSupplier{
		ChannelTypes: map[ChannelType]struct{}{
			ChannelTypeDirect:            {},
			ChannelTypeEmail:             {},
			ChannelTypeForwardToGroup:    {},
			ChannelTypeForwardToUser:     {},
			ChannelTypeInAppNotification: {},
			ChannelTypePhoneCall:         {},
			ChannelTypeRadio:             {},
		},
	}
	_ = ChannelTypeSupplier.operators(nil)
}

// channelTypesSupplier is a central supplier for channel operators. As is
// holds a list of supported channel types and is provided via a global instance
// (ChannelTypeSupplier), it is also used in Channel.Validate.
type channelTypesSupplier struct {
	// ChannelTypes holds all supported channel types. They are held in a map in
	// order to provide fast access for entity validation.
	ChannelTypes map[ChannelType]struct{}
}

func (supplier channelTypesSupplier) operators(m *Mall) map[ChannelType]channelOperator {
	operators := make(map[ChannelType]channelOperator, len(supplier.ChannelTypes))
	for channelType := range supplier.ChannelTypes {
		var operator channelOperator
		switch channelType {
		case ChannelTypeDirect:
			operator = &directChannelOperator{m: m}
		case ChannelTypeEmail:
			operator = &emailChannelOperator{m: m}
		case ChannelTypeForwardToGroup:
			operator = &forwardToGroupChannelOperator{m: m}
		case ChannelTypeForwardToUser:
			operator = &forwardToUserChannelOperator{m: m}
		case ChannelTypeInAppNotification:
			operator = &inAppNotificationChannelOperator{m: m}
		case ChannelTypePhoneCall:
			operator = &phoneCallChannelOperator{m: m}
		case ChannelTypeRadio:
			operator = &radioChannelOperator{m: m}
		default:
			log.Fatalf("missing channel operator for channel type %v", channelType)
		}
		operators[channelType] = operator
	}
	return operators
}

// InitNewMall creates a new Mall with postgres dialect.
func InitNewMall(ctx context.Context, logger *zap.Logger, db pgutil.DBTxSupplier, searchHost string, searchMasterKey string) (*Mall, error) {
	m := &Mall{
		dialect: goqu.Dialect("postgres"),
	}
	m.channelOperators = ChannelTypeSupplier.operators(m)
	// Add search.
	m.searchConfig = search.ClientConfig{
		Host:      searchHost,
		MasterKey: searchMasterKey,
		IndexConfigs: map[search.Index]search.IndexConfig{
			abEntrySearchIndex: abEntrySearchIndexConfig,
		},
		Logger:  logger.Named("search"),
		Timeout: search.DefaultClientTimeout,
		RebuildIndexFn: func(ctx context.Context, logger *zap.Logger, tx pgx.Tx, searchClient search.Client, index search.Index) error {
			switch index {
			case abEntrySearchIndex:
				return meh.NilOrWrap(m.rebuildAddressBookEntrySearchIndex(ctx, tx, searchClient),
					"rebuild address-book-entry-search-index", nil)
			default:
				return meh.NewInternalErr("unsupported index", meh.Details{"index": index})
			}
		},
	}
	searchClient, err := search.LaunchSafe(ctx, logger.Named("search"), db, m.searchConfig)
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

// rebuildAddressBookEntrySearchIndex in search.ClientConfig in InitNewMall.
func (m *Mall) rebuildAddressBookEntrySearchIndex(ctx context.Context, tx pgx.Tx, searchClient search.Client) error {
	err := search.Rebuild(ctx, searchClient, abEntrySearchIndex, search.DefaultBatchSize,
		func(ctx context.Context, next chan<- search.Document) error {
			defer close(next)
			// Retrieve ids for all entries.
			q, _, err := m.dialect.From(goqu.T("address_book_entries")).
				Select(goqu.C("id")).ToSQL()
			if err != nil {
				return meh.NewInternalErrFromErr(err, "query to sql", nil)
			}
			rows, err := tx.Query(ctx, q)
			if err != nil {
				return mehpg.NewQueryDBErr(err, "query db", q)
			}
			defer rows.Close()
			entries := make([]uuid.UUID, 0)
			for rows.Next() {
				var entryID uuid.UUID
				err = rows.Scan(&entryID)
				if err != nil {
					return mehpg.NewScanRowsErr(err, "scan row", q)
				}
				entries = append(entries, entryID)
			}
			rows.Close()
			// Retrieve document for each (database operation).
			for _, entryID := range entries {
				d, err := m.documentFromAddressBookEntryByID(ctx, tx, entryID)
				if err != nil {
					return meh.Wrap(err, "document from address book entry by id", meh.Details{"entry_id": entryID})
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case next <- d:
				}
			}
			return nil
		})
	if err != nil {
		return meh.Wrap(err, "rebuild search", nil)
	}
	return nil
}
