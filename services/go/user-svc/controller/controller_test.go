package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
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

func (m *StoreMock) UserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (store.User, error) {
	args := m.Called(ctx, tx, userID)
	return args.Get(0).(store.User), args.Error(1)
}

func (m *StoreMock) UserByUsername(ctx context.Context, tx pgx.Tx, username string) (store.User, error) {
	args := m.Called(ctx, tx, username)
	return args.Get(0).(store.User), args.Error(1)
}

func (m *StoreMock) Users(ctx context.Context, tx pgx.Tx, params pagination.Params) (pagination.Paginated[store.User], error) {
	args := m.Called(ctx, tx, params)
	return args.Get(0).(pagination.Paginated[store.User]), args.Error(1)
}

func (m *StoreMock) CreateUser(ctx context.Context, tx pgx.Tx, user store.UserWithPass) (store.User, error) {
	args := m.Called(ctx, tx, user)
	return args.Get(0).(store.User), args.Error(1)
}

func (m *StoreMock) UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *StoreMock) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *StoreMock) UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, pass []byte) error {
	return m.Called(ctx, tx, userID, pass).Error(0)
}

func (m *StoreMock) SearchUsers(ctx context.Context, tx pgx.Tx, searchParams search.Params) (search.Result[store.User], error) {
	args := m.Called(ctx, tx, searchParams)
	return args.Get(0).(search.Result[store.User]), args.Error(1)
}

func (m *StoreMock) RebuildUserSearch(ctx context.Context, tx pgx.Tx) error {
	return m.Called(ctx, tx).Error(0)
}

// NotifierMock mocks Notifier.
type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) NotifyUserCreated(ctx context.Context, tx pgx.Tx, user store.UserWithPass) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *NotifierMock) NotifyUserUpdated(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *NotifierMock) NotifyUserPassUpdated(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newPass []byte) error {
	return m.Called(ctx, tx, userID, newPass).Error(0)
}

func (m *NotifierMock) NotifyUserDeleted(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	return m.Called(ctx, tx, userID).Error(0)
}
