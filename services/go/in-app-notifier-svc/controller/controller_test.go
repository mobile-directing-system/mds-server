package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
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
	ctrl.Ctrl = NewController(ctrl.Logger, ctrl.DB, ctrl.Store, ctrl.Notifier)
	return ctrl
}

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) OldestPendingAttemptToNotifyByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (uuid.UUID, bool, error) {
	args := m.Called(ctx, tx, userID)
	return args.Get(0).(uuid.UUID), args.Bool(1), args.Error(2)
}

func (m *StoreMock) OutgoingNotificationByAttemptWithoutTriesAndLockOrSkip(ctx context.Context, tx pgx.Tx,
	attemptID uuid.UUID) (store.OutgoingIntelDeliveryNotification, error) {
	args := m.Called(ctx, tx, attemptID)
	return args.Get(0).(store.OutgoingIntelDeliveryNotification), args.Error(1)
}

func (m *StoreMock) CreateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *StoreMock) UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *StoreMock) CreateNotificationChannel(ctx context.Context, tx pgx.Tx, create store.NotificationChannel) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) DeleteNotificationChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	return m.Called(ctx, tx, entryID).Error(0)
}

func (m *StoreMock) CreateIntelNotificationHistoryEntry(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, ts time.Time) error {
	return m.Called(ctx, tx, attemptID, ts).Error(0)
}

func (m *StoreMock) CreateAcceptedIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, create store.AcceptedIntelDeliveryAttempt) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) UpdateAcceptedIntelDeliveryAttemptStatus(ctx context.Context, tx pgx.Tx, update store.AcceptedIntelDeliveryAttemptStatus) error {
	return m.Called(ctx, tx, update).Error(0)
}

func (m *StoreMock) NotificationChannelByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (store.NotificationChannel, error) {
	args := m.Called(ctx, tx, channelID)
	return args.Get(0).(store.NotificationChannel), args.Error(1)
}

func (m *StoreMock) CreateIntelToDeliver(ctx context.Context, tx pgx.Tx, create store.IntelToDeliver) error {
	return m.Called(ctx, tx, create).Error(0)
}

// NotifierMock mocks Notifier.
type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) NotifyIntelDeliveryNotificationSent(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, sentTS time.Time) error {
	return m.Called(ctx, tx, attemptID, sentTS).Error(0)
}

func (m *NotifierMock) NotifyIntelDeliveryNotificationPending(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, acceptedTS time.Time) error {
	return m.Called(ctx, tx, attemptID, acceptedTS).Error(0)
}
