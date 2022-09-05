package endpoints

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/stretchr/testify/mock"
	"time"
)

const timeout = 5 * time.Second

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) AddressBookEntryByID(ctx context.Context, entryID uuid.UUID, visibleBy uuid.NullUUID) (store.AddressBookEntryDetailed, error) {
	args := m.Called(ctx, entryID, visibleBy)
	return args.Get(0).(store.AddressBookEntryDetailed), args.Error(1)
}

func (m *StoreMock) UpdateAddressBookEntry(ctx context.Context, update store.AddressBookEntry, limitToUser uuid.NullUUID) error {
	return m.Called(ctx, update, limitToUser).Error(0)
}

func (m *StoreMock) CreateAddressBookEntry(ctx context.Context, entry store.AddressBookEntry) (store.AddressBookEntryDetailed, error) {
	args := m.Called(ctx, entry)
	return args.Get(0).(store.AddressBookEntryDetailed), args.Error(1)
}

func (m *StoreMock) UpdateChannelsByAddressBookEntry(ctx context.Context, entryID uuid.UUID, newChannels []store.Channel,
	limitToUser uuid.NullUUID) error {
	return m.Called(ctx, entryID, newChannels, limitToUser).Error(0)
}

func (m *StoreMock) AddressBookEntries(ctx context.Context, filters store.AddressBookEntryFilters,
	paginationParams pagination.Params) (pagination.Paginated[store.AddressBookEntryDetailed], error) {
	args := m.Called(ctx, filters, paginationParams)
	return args.Get(0).(pagination.Paginated[store.AddressBookEntryDetailed]), args.Error(1)
}

func (m *StoreMock) DeleteAddressBookEntryByID(ctx context.Context, entryID uuid.UUID, limitToUser uuid.NullUUID) error {
	return m.Called(ctx, entryID, limitToUser).Error(0)
}

func (m *StoreMock) ChannelsByAddressBookEntry(ctx context.Context, entryID uuid.UUID, limitToUser uuid.NullUUID) ([]store.Channel, error) {
	args := m.Called(ctx, entryID, limitToUser)
	var channels []store.Channel
	channels, _ = args.Get(0).([]store.Channel)
	return channels, args.Error(1)
}

func (m *StoreMock) SearchAddressBookEntries(ctx context.Context, filters store.AddressBookEntryFilters,
	searchParams search.Params) (search.Result[store.AddressBookEntryDetailed], error) {
	args := m.Called(ctx, filters, searchParams)
	return args.Get(0).(search.Result[store.AddressBookEntryDetailed]), args.Error(1)
}

func (m *StoreMock) RebuildAddressBookEntrySearch(ctx context.Context) {
	m.Called(ctx)
}
