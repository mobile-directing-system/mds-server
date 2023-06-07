package endpoints

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
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

func (m *StoreMock) DeleteAddressBookEntryWithChannelsByID(ctx context.Context, entryID uuid.UUID, limitToUser uuid.NullUUID) error {
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

func (m *StoreMock) MarkIntelDeliveryAsDelivered(ctx context.Context, deliveryID uuid.UUID, by uuid.NullUUID) error {
	return m.Called(ctx, deliveryID, by).Error(0)
}

func (m *StoreMock) MarkIntelDeliveryAttemptAsDelivered(ctx context.Context, attemptID uuid.UUID, by uuid.NullUUID) error {
	return m.Called(ctx, attemptID, by).Error(0)
}

func (m *StoreMock) SearchIntel(ctx context.Context, intelFilters store.IntelFilters, searchParams search.Params,
	limitToUser uuid.NullUUID) (search.Result[store.Intel], error) {
	args := m.Called(ctx, intelFilters, searchParams, limitToUser)
	return args.Get(0).(search.Result[store.Intel]), args.Error(1)
}

func (m *StoreMock) CreateIntel(ctx context.Context, create store.CreateIntel) (store.Intel, error) {
	args := m.Called(ctx, create)
	return args.Get(0).(store.Intel), args.Error(1)
}

func (m *StoreMock) InvalidateIntelByID(ctx context.Context, intelID uuid.UUID, by uuid.UUID) error {
	return m.Called(ctx, intelID, by).Error(0)
}

func (m *StoreMock) RebuildIntelSearch(ctx context.Context) {
	m.Called(ctx)
}

func (m *StoreMock) IntelByID(ctx context.Context, intelID uuid.UUID, limitToUser uuid.NullUUID) (store.Intel, error) {
	args := m.Called(ctx, intelID, limitToUser)
	return args.Get(0).(store.Intel), args.Error(1)
}

func (m *StoreMock) Intel(ctx context.Context, filters store.IntelFilters, paginationParams pagination.Params,
	limitToUser uuid.NullUUID) (pagination.Paginated[store.Intel], error) {
	args := m.Called(ctx, filters, paginationParams, limitToUser)
	return args.Get(0).(pagination.Paginated[store.Intel]), args.Error(1)
}

func (m *StoreMock) CreateIntelDeliveryAttempt(ctx context.Context, deliveryID uuid.UUID, channelID uuid.UUID) (store.IntelDeliveryAttempt, error) {
	args := m.Called(ctx, deliveryID, channelID)
	return args.Get(0).(store.IntelDeliveryAttempt), args.Error(1)
}

func (m *StoreMock) IntelDeliveryAttemptsByDelivery(ctx context.Context, deliveryID uuid.UUID) ([]store.IntelDeliveryAttempt, error) {
	args := m.Called(ctx, deliveryID)
	var attempts []store.IntelDeliveryAttempt
	if a := args.Get(0); a != nil {
		attempts = a.([]store.IntelDeliveryAttempt)
	}
	return attempts, args.Error(1)
}

func (m *StoreMock) SetAddressBookEntriesWithAutoDeliveryEnabled(ctx context.Context, entryIDs []uuid.UUID) error {
	return m.Called(ctx, entryIDs).Error(0)
}

func (m *StoreMock) CancelIntelDeliveryByID(ctx context.Context, deliveryID uuid.UUID, success bool, note nulls.String) error {
	return m.Called(ctx, deliveryID, success, note).Error(0)
}

func (m *StoreMock) IntelDeliveryAttempts(ctx context.Context, filters store.IntelDeliveryAttemptFilters,
	page pagination.Params) (pagination.Paginated[store.IntelDeliveryAttempt], error) {
	args := m.Called(ctx, filters, page)
	return args.Get(0).(pagination.Paginated[store.IntelDeliveryAttempt]), args.Error(1)
}

func (m *StoreMock) SetAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID, enabled bool) error {
	return m.Called(ctx, entryID, enabled).Error(0)
}

func (m *StoreMock) IsAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID) (bool, error) {
	args := m.Called(ctx, entryID)
	return args.Bool(0), args.Error(1)
}
