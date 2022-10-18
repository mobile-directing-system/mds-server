package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"time"
)

const timeout = 5 * time.Second

type ControllerMock struct {
	Logger                *zap.Logger
	DB                    *testutil.DBTxSupplier
	Store                 *StoreMock
	Notifier              *NotifierMock
	Ctrl                  *Controller
	samplePickedUpTimeout time.Duration
}

func NewMockController() *ControllerMock {
	ctrl := &ControllerMock{
		Logger:                zap.NewNop(),
		DB:                    &testutil.DBTxSupplier{},
		Store:                 &StoreMock{},
		Notifier:              &NotifierMock{},
		samplePickedUpTimeout: 10 * time.Second,
	}
	ctrl.Ctrl = NewController(ctrl.Logger, ctrl.DB, ctrl.Store, ctrl.Notifier, ctrl.samplePickedUpTimeout)
	return ctrl
}

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) CreateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *StoreMock) UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *StoreMock) CreateRadioChannel(ctx context.Context, tx pgx.Tx, create store.RadioChannel) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) DeleteRadioChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	return m.Called(ctx, tx, entryID).Error(0)
}

func (m *StoreMock) CreateAcceptedIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, create store.AcceptedIntelDeliveryAttempt) error {
	return m.Called(ctx, tx, create).Error(0)
}

func (m *StoreMock) UpdateAcceptedIntelDeliveryAttemptStatus(ctx context.Context, tx pgx.Tx, update store.AcceptedIntelDeliveryAttemptStatus) error {
	return m.Called(ctx, tx, update).Error(0)
}

func (m *StoreMock) RadioChannelByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID) (store.RadioChannel, error) {
	args := m.Called(ctx, tx, channelID)
	return args.Get(0).(store.RadioChannel), args.Error(1)
}

func (m *StoreMock) UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error {
	return m.Called(ctx, tx, operationID, newMembers).Error(0)
}

func (m *StoreMock) OperationsByMember(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, tx, userID)
	var members []uuid.UUID
	if a := args.Get(0); a != nil {
		members = a.([]uuid.UUID)
	}
	return members, args.Error(1)
}

func (m *StoreMock) CreateRadioDelivery(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) error {
	return m.Called(ctx, tx, attemptID).Error(0)
}

func (m *StoreMock) RadioDeliveryByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID) (store.RadioDelivery, error) {
	args := m.Called(ctx, tx, attemptID)
	return args.Get(0).(store.RadioDelivery), args.Error(1)
}

func (m *StoreMock) MarkRadioDeliveryAsPickedUpByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID,
	by uuid.NullUUID, newNote string) error {
	return m.Called(ctx, tx, attemptID, by, newNote).Error(0)
}

func (m *StoreMock) UpdateRadioDeliveryStatusByAttempt(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID,
	newSuccess nulls.Bool, newNote string) error {
	return m.Called(ctx, tx, attemptID, newSuccess, newNote).Error(0)
}

func (m *StoreMock) ActiveRadioDeliveriesAndLockOrWait(ctx context.Context, tx pgx.Tx,
	byOperation uuid.NullUUID) ([]store.ActiveRadioDelivery, error) {
	args := m.Called(ctx, tx, byOperation)
	var deliveries []store.ActiveRadioDelivery
	if a := args.Get(0); a != nil {
		deliveries = a.([]store.ActiveRadioDelivery)
	}
	return deliveries, args.Error(1)
}

func (m *StoreMock) AcceptedIntelDeliveryAttemptByID(ctx context.Context, tx pgx.Tx,
	attemptID uuid.UUID) (store.AcceptedIntelDeliveryAttempt, error) {
	args := m.Called(ctx, tx, attemptID)
	return args.Get(0).(store.AcceptedIntelDeliveryAttempt), args.Error(1)
}

// NotifierMock mocks Notifier.
type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) NotifyRadioDeliveryReadyForPickup(ctx context.Context, tx pgx.Tx,
	intelDeliveryAttempt store.AcceptedIntelDeliveryAttempt, radioDeliveryNote string) error {
	return m.Called(ctx, tx, intelDeliveryAttempt, radioDeliveryNote).Error(0)
}

func (m *NotifierMock) NotifyRadioDeliveryPickedUp(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID,
	pickedUpBy uuid.UUID, pickedUpAt time.Time) error {
	return m.Called(ctx, tx, attemptID, pickedUpBy, pickedUpAt).Error(0)
}

func (m *NotifierMock) NotifyRadioDeliveryReleased(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, releasedAt time.Time) error {
	return m.Called(ctx, tx, attemptID, releasedAt).Error(0)
}

func (m *NotifierMock) NotifyRadioDeliveryFinished(ctx context.Context, tx pgx.Tx, radioDelivery store.RadioDelivery) error {
	return m.Called(ctx, tx, radioDelivery).Error(0)
}
