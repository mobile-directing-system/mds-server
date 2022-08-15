package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"time"
)

const timeout = 5 * time.Second

type ControllerMock struct {
	Logger   *zap.Logger
	DB       *testutil.DBTxSupplier
	Store    *StoreMock
	Notifier *NotifierMock
	Ctrl     *Controller
}

func NewMockController() *ControllerMock {
	ctrl := &ControllerMock{
		Logger:   zap.NewNop(),
		DB:       &testutil.DBTxSupplier{},
		Store:    &StoreMock{},
		Notifier: &NotifierMock{},
	}
	ctrl.Ctrl = &Controller{
		Logger:   ctrl.Logger,
		DB:       ctrl.DB,
		Store:    ctrl.Store,
		Notifier: ctrl.Notifier,
	}
	return ctrl
}

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) ChannelsByAddressBookEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) ([]store.Channel, error) {
	args := m.Called(ctx, tx, entryID)
	var channels []store.Channel
	channels, _ = args.Get(0).([]store.Channel)
	return channels, args.Error(1)
}

func (m *StoreMock) AssureAddressBookEntryExists(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	return m.Called(ctx, tx, entryID).Error(0)
}

func (m *StoreMock) DeleteChannelWithDetailsByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID, channelType store.ChannelType) error {
	return m.Called(ctx, tx, channelID, channelType).Error(0)
}

func (m *StoreMock) CreateChannelWithDetails(ctx context.Context, tx pgx.Tx, channel store.Channel) error {
	return m.Called(ctx, tx, channel).Error(0)
}

func (m *StoreMock) UpdateChannelWithDetails(ctx context.Context, tx pgx.Tx, channel store.Channel) error {
	return m.Called(ctx, tx, channel).Error(0)
}

func (m *StoreMock) AddressBookEntries(ctx context.Context, tx pgx.Tx, filters store.AddressBookEntryFilters,
	paginationParams pagination.Params) (pagination.Paginated[store.AddressBookEntryDetailed], error) {
	args := m.Called(ctx, tx, filters, paginationParams)
	return args.Get(0).(pagination.Paginated[store.AddressBookEntryDetailed]), args.Error(1)
}

func (m *StoreMock) CreateAddressBookEntry(ctx context.Context, tx pgx.Tx, entry store.AddressBookEntry) (store.AddressBookEntryDetailed, error) {
	args := m.Called(ctx, tx, entry)
	return args.Get(0).(store.AddressBookEntryDetailed), args.Error(1)
}

func (m *StoreMock) UpdateAddressBookEntry(ctx context.Context, tx pgx.Tx, entry store.AddressBookEntry) error {
	return m.Called(ctx, tx, entry).Error(0)
}

func (m *StoreMock) DeleteAddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	return m.Called(ctx, tx, entryID).Error(0)
}

func (m *StoreMock) CreateGroup(ctx context.Context, tx pgx.Tx, create store.Group) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) UpdateGroup(ctx context.Context, tx pgx.Tx, update store.Group) error {
	return m.Called(ctx, tx, update).Error(0)
}

func (m *StoreMock) DeleteGroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	return m.Called(ctx, tx, groupID).Error(0)
}

func (m *StoreMock) CreateUser(ctx context.Context, tx pgx.Tx, create store.User) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) UpdateUser(ctx context.Context, tx pgx.Tx, update store.User) error {
	return m.Called(ctx, tx, update).Error(0)
}

func (m *StoreMock) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *StoreMock) CreateOperation(ctx context.Context, tx pgx.Tx, create store.Operation) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) UpdateOperation(ctx context.Context, tx pgx.Tx, update store.Operation) error {
	return m.Called(ctx, tx, update).Error(0)
}

func (m *StoreMock) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error {
	return m.Called(ctx, tx, operationID, newMembers).Error(0)
}

func (m *StoreMock) AddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID,
	visibleBy uuid.NullUUID) (store.AddressBookEntryDetailed, error) {
	args := m.Called(ctx, tx, entryID, visibleBy)
	return args.Get(0).(store.AddressBookEntryDetailed), args.Error(1)
}

func (m *StoreMock) DeleteForwardToGroupChannelsByGroup(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, tx, groupID)
	var channels []uuid.UUID
	channels, _ = args.Get(0).([]uuid.UUID)
	return channels, args.Error(1)
}

func (m *StoreMock) DeleteForwardToUserChannelsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, tx, userID)
	var channels []uuid.UUID
	channels, _ = args.Get(0).([]uuid.UUID)
	return channels, args.Error(1)
}

// NotifierMock mocks Notifier.
type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) NotifyAddressBookEntryCreated(ctx context.Context, tx pgx.Tx, entry store.AddressBookEntry) error {
	return m.Called(ctx, tx, entry).Error(0)
}

func (m *NotifierMock) NotifyAddressBookEntryUpdated(ctx context.Context, tx pgx.Tx, entry store.AddressBookEntry) error {
	return m.Called(ctx, tx, entry).Error(0)
}

func (m *NotifierMock) NotifyAddressBookEntryDeleted(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	return m.Called(ctx, tx, entryID).Error(0)
}

func (m *NotifierMock) NotifyAddressBookEntryChannelsUpdated(ctx context.Context, tx pgx.Tx, entryID uuid.UUID, channels []store.Channel) error {
	return m.Called(ctx, tx, entryID, channels).Error(0)
}
