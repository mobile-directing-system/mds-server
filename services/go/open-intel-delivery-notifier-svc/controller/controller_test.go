package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"time"
)

const timeout = 5 * time.Second

type ControllerMock struct {
	Logger *zap.Logger
	DB     *testutil.DBTxSupplier
	Store  *StoreMock
	Ctrl   *Controller
}

func NewMockController() *ControllerMock {
	ctrl := &ControllerMock{
		Logger: zap.NewNop(),
		DB:     &testutil.DBTxSupplier{},
		Store:  &StoreMock{},
	}
	ctrl.Ctrl = NewController(ctrl.Logger, ctrl.DB, ctrl.Store)
	return ctrl
}

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) CreateUser(ctx context.Context, tx pgx.Tx, create store.User) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) UpdateUser(ctx context.Context, tx pgx.Tx, update store.User) error {
	return m.Called(ctx, tx, update).Error(0)
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

func (m *StoreMock) CreateActiveIntelDelivery(ctx context.Context, tx pgx.Tx, create store.ActiveIntelDelivery) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) DeleteActiveIntelDeliveryByID(ctx context.Context, tx pgx.Tx, deliveryID uuid.UUID) error {
	return m.Called(ctx, tx, deliveryID).Error(0)
}

func (m *StoreMock) CreateActiveIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, create store.ActiveIntelDeliveryAttempt) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) DeleteActiveIntelDeliveryAttemptByID(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) error {
	return m.Called(ctx, tx, attemptID).Error(0)
}

func (m *StoreMock) IsAutoIntelDeliveryEnabledForEntry(ctx context.Context, tx pgx.Tx, addressBookEntryID uuid.UUID) (bool, error) {
	args := m.Called(ctx, tx, addressBookEntryID)
	return args.Bool(0), args.Error(1)
}

func (m *StoreMock) SetAutoIntelDeliveryEnabledForEntry(ctx context.Context, tx pgx.Tx, addressBookEntryID uuid.UUID, enabled bool) error {
	return m.Called(ctx, tx, addressBookEntryID, enabled).Error(0)
}

func (m *StoreMock) IntelOperationByDelivery(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, tx, attemptID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *StoreMock) OpenIntelDeliveriesByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) ([]store.OpenIntelDeliverySummary, error) {
	args := m.Called(ctx, tx, operationID)
	var deliveries []store.OpenIntelDeliverySummary
	if a := args.Get(0); a != nil {
		deliveries = a.([]store.OpenIntelDeliverySummary)
	}
	return deliveries, args.Error(1)
}

func (m *StoreMock) IntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) (store.Intel, error) {
	args := m.Called(ctx, tx, intelID)
	return args.Get(0).(store.Intel), args.Error(1)
}

func (m *StoreMock) CreateIntel(ctx context.Context, tx pgx.Tx, create store.Intel) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) InvalidateIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error {
	return m.Called(ctx, tx, intelID).Error(0)
}

func (m *StoreMock) IntelOperationByDeliveryAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, tx, attemptID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *StoreMock) IntelOperationsByActiveIntelDeliveryRecipient(ctx context.Context, tx pgx.Tx, addressBookEntryID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, tx, addressBookEntryID)
	var operations []uuid.UUID
	if a := args.Get(0); a != nil {
		operations = a.([]uuid.UUID)
	}
	return operations, args.Error(1)
}
