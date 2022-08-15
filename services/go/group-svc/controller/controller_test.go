package controller

import (
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/net/context"
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

func (m *StoreMock) CreateUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *StoreMock) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *StoreMock) CreateOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) error {
	return m.Called(ctx, tx, operationID).Error(0)
}

func (m *StoreMock) CreateGroup(ctx context.Context, tx pgx.Tx, create store.Group) (store.Group, error) {
	args := m.Called(ctx, tx, create)
	return args.Get(0).(store.Group), args.Error(1)
}

func (m *StoreMock) UpdateGroup(ctx context.Context, tx pgx.Tx, update store.Group) error {
	return m.Called(ctx, tx, update).Error(0)
}

func (m *StoreMock) GroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) (store.Group, error) {
	args := m.Called(ctx, tx, groupID)
	return args.Get(0).(store.Group), args.Error(1)
}

func (m *StoreMock) Groups(ctx context.Context, tx pgx.Tx, filters store.GroupFilters,
	params pagination.Params) (pagination.Paginated[store.Group], error) {
	args := m.Called(ctx, tx, filters, params)
	return args.Get(0).(pagination.Paginated[store.Group]), args.Error(1)
}

func (m *StoreMock) AssureUserExists(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *StoreMock) DeleteGroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	return m.Called(ctx, tx, groupID).Error(0)
}

func (m *StoreMock) OperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, tx, operationID)
	var members []uuid.UUID
	members, _ = args.Get(0).([]uuid.UUID)
	return members, args.Error(1)
}

func (m *StoreMock) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error {
	return m.Called(ctx, tx, operationID, newMembers).Error(0)
}

// NotifierMock mocks Notifier.
type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) NotifyGroupCreated(ctx context.Context, tx pgx.Tx, group store.Group) error {
	return m.Called(ctx, tx, group).Error(0)
}

func (m *NotifierMock) NotifyGroupUpdated(ctx context.Context, tx pgx.Tx, group store.Group) error {
	return m.Called(ctx, tx, group).Error(0)
}

func (m *NotifierMock) NotifyGroupDeleted(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	return m.Called(ctx, tx, groupID).Error(0)
}
