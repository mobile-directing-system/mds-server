package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
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

func (m *StoreMock) OperationByID(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) (store.Operation, error) {
	args := m.Called(ctx, tx, operationID)
	return args.Get(0).(store.Operation), args.Error(1)
}

func (m *StoreMock) Operations(ctx context.Context, tx pgx.Tx, params pagination.Params) (pagination.Paginated[store.Operation], error) {
	args := m.Called(ctx, tx, params)
	return args.Get(0).(pagination.Paginated[store.Operation]), args.Error(1)
}

func (m *StoreMock) CreateOperation(ctx context.Context, tx pgx.Tx, operation store.Operation) (store.Operation, error) {
	args := m.Called(ctx, tx, operation)
	return args.Get(0).(store.Operation), args.Error(1)
}

func (m *StoreMock) UpdateOperation(ctx context.Context, tx pgx.Tx, operation store.Operation) error {
	return m.Called(ctx, tx, operation).Error(0)
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

func (m *StoreMock) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, members []uuid.UUID) error {
	return m.Called(ctx, tx, operationID, members).Error(0)
}

func (m *StoreMock) OperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID,
	params pagination.Params) (pagination.Paginated[store.User], error) {
	args := m.Called(ctx, tx, operationID, params)
	return args.Get(0).(pagination.Paginated[store.User]), args.Error(1)
}

func (m *StoreMock) OperationsByMember(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]store.Operation, error) {
	args := m.Called(ctx, tx, userID)
	var operations []store.Operation
	operations, _ = args.Get(0).([]store.Operation)
	return operations, args.Error(1)
}

// NotifierMock mocks Notifier.
type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) NotifyOperationCreated(operation store.Operation) error {
	return m.Called(operation).Error(0)
}

func (m *NotifierMock) NotifyOperationUpdated(operation store.Operation) error {
	return m.Called(operation).Error(0)
}

func (m *NotifierMock) NotifyOperationMembersUpdated(operationID uuid.UUID, members []uuid.UUID) error {
	return m.Called(operationID, members).Error(0)
}
