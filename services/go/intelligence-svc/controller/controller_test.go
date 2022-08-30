package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"time"
)

const timeout = 5 * time.Second

// ControllerMock mocks Controller.
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

func (m *StoreMock) CreateUser(ctx context.Context, tx pgx.Tx, userID store.User) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *StoreMock) UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
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

func (m *StoreMock) CreateAddressBookEntry(ctx context.Context, tx pgx.Tx, create store.AddressBookEntry) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) UpdateAddressBookEntry(ctx context.Context, tx pgx.Tx, update store.AddressBookEntry) error {
	return m.Called(ctx, tx, update).Error(0)
}

func (m *StoreMock) DeleteAddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	return m.Called(ctx, tx, entryID).Error(0)
}

func (m *StoreMock) CreateIntel(ctx context.Context, tx pgx.Tx, create store.CreateIntel) (store.Intel, error) {
	args := m.Called(ctx, tx, create)
	return args.Get(0).(store.Intel), args.Error(1)
}

func (m *StoreMock) InvalidateIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error {
	return m.Called(ctx, tx, intelID).Error(0)
}

func (m *StoreMock) IntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) (store.Intel, error) {
	args := m.Called(ctx, tx, intelID)
	return args.Get(0).(store.Intel), args.Error(1)
}

func (m *StoreMock) IsUserOperationMember(ctx context.Context, tx pgx.Tx, userID uuid.UUID, operationID uuid.UUID) (bool, error) {
	args := m.Called(ctx, tx, userID, operationID)
	return args.Bool(0), args.Error(1)
}

func (m *StoreMock) AddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) (store.AddressBookEntry, error) {
	args := m.Called(ctx, tx, entryID)
	return args.Get(0).(store.AddressBookEntry), args.Error(1)
}

// NotifierMock mocks Notifier.
type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) NotifyIntelCreated(ctx context.Context, tx pgx.Tx, created store.Intel) error {
	return m.Called(ctx, tx, created).Error(0)
}

func (m *NotifierMock) NotifyIntelInvalidated(ctx context.Context, tx pgx.Tx, intelID uuid.UUID, by uuid.UUID) error {
	return m.Called(ctx, tx, intelID, by).Error(0)
}
