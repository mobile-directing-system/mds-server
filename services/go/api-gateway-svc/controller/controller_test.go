package controller

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
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

func (m *StoreMock) StoreSessionTokenForUser(ctx context.Context, tx pgx.Tx, token string, userID uuid.UUID) error {
	return m.Called(ctx, tx, token, userID).Error(0)
}

func (m *StoreMock) GetAndDeleteUserIDBySessionToken(ctx context.Context, tx pgx.Tx, token string) (uuid.UUID, error) {
	args := m.Called(ctx, tx, token)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *StoreMock) DeleteSessionTokensByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *StoreMock) UserWithPassByUsername(ctx context.Context, tx pgx.Tx, username string) (store.UserWithPass, error) {
	args := m.Called(ctx, tx, username)
	return args.Get(0).(store.UserWithPass), args.Error(1)
}

func (m *StoreMock) UserWithPassByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (store.UserWithPass, error) {
	args := m.Called(ctx, tx, userID)
	return args.Get(0).(store.UserWithPass), args.Error(1)
}

func (m *StoreMock) CreateUser(ctx context.Context, tx pgx.Tx, user store.UserWithPass) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *StoreMock) UpdateUserPassByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newPass []byte) error {
	return m.Called(ctx, tx, userID, newPass).Error(0)
}

func (m *StoreMock) DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *StoreMock) PermissionsByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]permission.Permission, error) {
	args := m.Called(ctx, tx, userID)
	var p []permission.Permission
	if argsPermissions := args.Get(0); argsPermissions != nil {
		p = argsPermissions.([]permission.Permission)
	}
	return p, args.Error(1)
}

func (m *StoreMock) UserIDBySessionToken(ctx context.Context, txSupplier pgutil.DBTxSupplier, token string) (uuid.UUID, error) {
	args := m.Called(ctx, txSupplier, token)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *StoreMock) PassByUsername(ctx context.Context, tx pgx.Tx, username string) ([]byte, error) {
	args := m.Called(ctx, tx, username)
	var b []byte
	if argsByteSlice := args.Get(0); argsByteSlice != nil {
		b = argsByteSlice.([]byte)
	}
	return b, args.Error(1)
}

func (m *StoreMock) UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *StoreMock) UpdatePermissionsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID, permissions []permission.Permission) error {
	return m.Called(ctx, tx, userID, permissions).Error(0)
}

// NotifierMock mocks Notifier.
type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) NotifyUserLoggedIn(userID uuid.UUID, username string, requestMetadata AuthRequestMetadata) error {
	return m.Called(userID, username, requestMetadata).Error(0)
}

func (m *NotifierMock) NotifyUserLoggedOut(userID uuid.UUID, username string, requestMetadata AuthRequestMetadata) error {
	return m.Called(userID, username, requestMetadata).Error(0)
}
