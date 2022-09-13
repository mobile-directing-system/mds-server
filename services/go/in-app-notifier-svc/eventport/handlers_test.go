package eventport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/in-app-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/mobile-directing-system/mds-server/services/go/shared/kafkautil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
	"time"
)

// HandlerMock mocks Handler.
type HandlerMock struct {
	mock.Mock
}

func (m *HandlerMock) CreateUser(ctx context.Context, tx pgx.Tx, userID store.User) error {
	return m.Called(ctx, tx, userID).Error(0)
}

func (m *HandlerMock) UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error {
	return m.Called(ctx, tx, user).Error(0)
}

func (m *HandlerMock) DeleteNotificationChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error {
	return m.Called(ctx, tx, entryID).Error(0)
}

func (m *HandlerMock) UpdateNotificationChannelsByEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID,
	create []store.NotificationChannel) error {
	return m.Called(ctx, tx, entryID, create).Error(0)
}

func (m *HandlerMock) CreateIntelDeliveryAttempt(ctx context.Context, tx pgx.Tx, attempt store.AcceptedIntelDeliveryAttempt,
	intelToDeliver store.IntelToDeliver) error {
	return m.Called(ctx, tx, attempt, intelToDeliver).Error(0)
}

func (m *HandlerMock) UpdateIntelDeliveryAttemptStatus(ctx context.Context, tx pgx.Tx, newStatus store.AcceptedIntelDeliveryAttemptStatus) error {
	return m.Called(ctx, tx, newStatus).Error(0)
}

// portHandleUserCreatedSuite tests Port.handleUserCreated.
type portHandleUserCreatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.UserCreated
	sampleCreate store.User
}

func (suite *portHandleUserCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.UserCreated{
		ID:        testutil.NewUUIDV4(),
		Username:  "small",
		FirstName: "pleasant",
		LastName:  "glad",
		IsAdmin:   true,
		Pass:      nil,
		IsActive:  true,
	}
	suite.sampleCreate = store.User{
		ID:        suite.sampleEvent.ID,
		Username:  suite.sampleEvent.Username,
		FirstName: suite.sampleEvent.FirstName,
		LastName:  suite.sampleEvent.LastName,
		IsActive:  true,
	}
}

func (suite *portHandleUserCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleUserCreatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserCreatedSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateUser", timeout, tx, suite.sampleCreate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateUser", timeout, tx, suite.sampleCreate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleUserCreated(t *testing.T) {
	suite.Run(t, new(portHandleUserCreatedSuite))
}

// portHandleUserUpdatedSuite tests Port.handleUserUpdated.
type portHandleUserUpdatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.UserUpdated
	sampleUpdate store.User
}

func (suite *portHandleUserUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.UserUpdated{
		ID:        testutil.NewUUIDV4(),
		Username:  "small",
		FirstName: "pleasant",
		LastName:  "glad",
		IsAdmin:   true,
	}
	suite.sampleUpdate = store.User{
		ID:        suite.sampleEvent.ID,
		Username:  suite.sampleEvent.Username,
		FirstName: suite.sampleEvent.FirstName,
		LastName:  suite.sampleEvent.LastName,
	}
}

func (suite *portHandleUserUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.UsersTopic,
		EventType: event.TypeUserUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleUserUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateUser", timeout, tx, suite.sampleUpdate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleUserUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateUser", timeout, tx, suite.sampleUpdate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleUserUpdated(t *testing.T) {
	suite.Run(t, new(portHandleUserUpdatedSuite))
}

// portHandleAddressBookEntryDeletedSuite tests
// Port.handleAddressBookEntryDeleted.
type portHandleAddressBookEntryDeletedSuite struct {
	suite.Suite
	handler     *HandlerMock
	port        *PortMock
	sampleEvent event.AddressBookEntryDeleted
}

func (suite *portHandleAddressBookEntryDeletedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.AddressBookEntryDeleted{
		ID: testutil.NewUUIDV4(),
	}
}

func (suite *portHandleAddressBookEntryDeletedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.AddressBookTopic,
		EventType: event.TypeAddressBookEntryDeleted,
		RawValue:  rawValue,
	})
}

func (suite *portHandleAddressBookEntryDeletedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryDeletedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteNotificationChannelsByEntry", timeout, tx, suite.sampleEvent.ID).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryDeletedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("DeleteNotificationChannelsByEntry", timeout, tx, suite.sampleEvent.ID).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleAddressBookEntryDeleted(t *testing.T) {
	suite.Run(t, new(portHandleAddressBookEntryDeletedSuite))
}

// portHandleAddressBookEntryChannelsUpdatedSuite tests
// Port.handleAddressBookEntryChannelsUpdated.
type portHandleAddressBookEntryChannelsUpdatedSuite struct {
	suite.Suite
	handler                *HandlerMock
	port                   *PortMock
	sampleEvent            event.AddressBookEntryChannelsUpdated
	sampleNewNotifChannels []store.NotificationChannel
}

func (suite *portHandleAddressBookEntryChannelsUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	entryID := testutil.NewUUIDV4()
	genChannel := func(channelType event.AddressBookEntryChannelType) event.AddressBookEntryChannelsUpdatedChannel {
		return event.AddressBookEntryChannelsUpdatedChannel{
			ID:            testutil.NewUUIDV4(),
			Entry:         entryID,
			Label:         "tail",
			Type:          channelType,
			Priority:      rand.Int31(),
			MinImportance: rand.Float64(),
			Details:       nil,
			Timeout:       time.Duration(rand.Int31()),
		}
	}
	suite.sampleEvent = event.AddressBookEntryChannelsUpdated{
		Entry: entryID,
		Channels: []event.AddressBookEntryChannelsUpdatedChannel{
			genChannel(event.AddressBookEntryChannelTypeDirect),
			genChannel(event.AddressBookEntryChannelTypeInAppNotification),
			genChannel(event.AddressBookEntryChannelTypeEmail),
			genChannel(event.AddressBookEntryChannelTypeInAppNotification),
			genChannel(event.AddressBookEntryChannelTypePhoneCall),
			genChannel(event.AddressBookEntryChannelTypeInAppNotification),
		},
	}
	notifChannelFromEvent := func(eventChannel event.AddressBookEntryChannelsUpdatedChannel) store.NotificationChannel {
		if eventChannel.Type != event.AddressBookEntryChannelTypeInAppNotification {
			panic("bad channel-type")
		}
		return store.NotificationChannel{
			ID:      eventChannel.ID,
			Entry:   eventChannel.Entry,
			Label:   eventChannel.Label,
			Timeout: eventChannel.Timeout,
		}
	}
	suite.sampleNewNotifChannels = []store.NotificationChannel{
		notifChannelFromEvent(suite.sampleEvent.Channels[1]),
		notifChannelFromEvent(suite.sampleEvent.Channels[3]),
		notifChannelFromEvent(suite.sampleEvent.Channels[5]),
	}
}

func (suite *portHandleAddressBookEntryChannelsUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.AddressBookTopic,
		EventType: event.TypeAddressBookEntryChannelsUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleAddressBookEntryChannelsUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryChannelsUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateNotificationChannelsByEntry", timeout, tx, suite.sampleEvent.Entry, suite.sampleNewNotifChannels).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleAddressBookEntryChannelsUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateNotificationChannelsByEntry", timeout, tx, suite.sampleEvent.Entry, suite.sampleNewNotifChannels).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleAddressBookEntryChannelsUpdated(t *testing.T) {
	suite.Run(t, new(portHandleAddressBookEntryChannelsUpdatedSuite))
}

// portHandleIntelDeliveryAttemptStatusUpdatedSuite tests
// Port.handleIntelDeliveryAttemptStatusUpdated.
type portHandleIntelDeliveryAttemptStatusUpdatedSuite struct {
	suite.Suite
	handler      *HandlerMock
	port         *PortMock
	sampleEvent  event.IntelDeliveryAttemptStatusUpdated
	sampleUpdate store.AcceptedIntelDeliveryAttemptStatus
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	suite.sampleEvent = event.IntelDeliveryAttemptStatusUpdated{
		ID:       testutil.NewUUIDV4(),
		IsActive: true,
		Status:   "lesson",
		StatusTS: time.Date(2022, 9, 8, 0, 23, 25, 0, time.UTC),
		Note:     nulls.NewString("educate"),
	}
	suite.sampleUpdate = store.AcceptedIntelDeliveryAttemptStatus{
		ID:       suite.sampleEvent.ID,
		IsActive: suite.sampleEvent.IsActive,
		StatusTS: suite.sampleEvent.StatusTS,
		Note:     suite.sampleEvent.Note,
	}
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		EventType: event.TypeIntelDeliveryAttemptStatusUpdated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatus", timeout, tx, suite.sampleUpdate).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleIntelDeliveryAttemptStatusUpdatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("UpdateIntelDeliveryAttemptStatus", timeout, tx, suite.sampleUpdate).Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleIntelDeliveryAttemptStatusUpdated(t *testing.T) {
	suite.Run(t, new(portHandleIntelDeliveryAttemptStatusUpdatedSuite))
}

// portHandleIntelDeliveryAttemptCreatedSuite tests
// Port.handleIntelDeliveryAttemptCreated.
type portHandleIntelDeliveryAttemptCreatedSuite struct {
	suite.Suite
	handler              *HandlerMock
	port                 *PortMock
	sampleEvent          event.IntelDeliveryAttemptCreated
	sampleCreate         store.AcceptedIntelDeliveryAttempt
	sampleIntelToDeliver store.IntelToDeliver
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) SetupTest() {
	suite.handler = &HandlerMock{}
	suite.port = newMockPort()
	assignmentID := testutil.NewUUIDV4()
	assignedEntryID := testutil.NewUUIDV4()
	intelID := testutil.NewUUIDV4()
	assignedEntryUserID := testutil.NewUUIDV4()
	operationID := testutil.NewUUIDV4()
	suite.sampleEvent = event.IntelDeliveryAttemptCreated{
		ID: testutil.NewUUIDV4(),
		Delivery: event.IntelDeliveryAttemptCreatedDelivery{
			ID:         testutil.NewUUIDV4(),
			Assignment: assignmentID,
			IsActive:   true,
			Success:    false,
			Note:       nulls.NewString("pass"),
		},
		Assignment: event.IntelDeliveryAttemptCreatedAssignment{
			ID:    assignmentID,
			Intel: intelID,
			To:    assignedEntryID,
		},
		AssignedEntry: event.IntelDeliveryAttemptCreatedAssignedEntry{
			ID:          assignedEntryID,
			Label:       "produce",
			Description: "shoot",
			Operation:   nulls.NewUUID(operationID),
			User:        nulls.NewUUID(assignedEntryUserID),
			UserDetails: nulls.NewJSONNullable(event.IntelDeliveryAttemptCreatedAssignedEntryUserDetails{
				ID:        assignedEntryUserID,
				Username:  "cousin",
				FirstName: "breath",
				LastName:  "stomach",
				IsActive:  true,
			}),
		},
		Intel: event.IntelDeliveryAttemptCreatedIntel{
			ID:         intelID,
			CreatedAt:  time.Date(2022, 9, 13, 10, 30, 53, 0, time.UTC),
			CreatedBy:  testutil.NewUUIDV4(),
			Operation:  operationID,
			Type:       event.IntelTypePlaintextMessage,
			Content:    []byte(`{"hello":"world"}`),
			SearchText: nulls.NewString("wife"),
			Importance: 585,
			IsValid:    true,
		},
		Channel:   testutil.NewUUIDV4(),
		CreatedAt: time.Date(2022, 9, 13, 10, 31, 38, 0, time.UTC),
		IsActive:  true,
		Status:    event.IntelDeliveryStatusOpen,
		StatusTS:  time.Date(2022, 9, 13, 10, 31, 57, 0, time.UTC),
		Note:      nulls.NewString("stupid"),
	}
	suite.sampleCreate = store.AcceptedIntelDeliveryAttempt{
		ID:              suite.sampleEvent.ID,
		AssignedTo:      suite.sampleEvent.Assignment.To,
		AssignedToLabel: suite.sampleEvent.AssignedEntry.Label,
		AssignedToUser:  suite.sampleEvent.AssignedEntry.User,
		Delivery:        suite.sampleEvent.Delivery.ID,
		Channel:         suite.sampleEvent.Channel,
		CreatedAt:       suite.sampleEvent.CreatedAt,
		IsActive:        suite.sampleEvent.IsActive,
		StatusTS:        suite.sampleEvent.StatusTS,
		Note:            suite.sampleEvent.Note,
		AcceptedAt:      time.Time{},
	}
	suite.sampleIntelToDeliver = store.IntelToDeliver{
		Attempt:    suite.sampleEvent.ID,
		ID:         suite.sampleEvent.Intel.ID,
		CreatedAt:  suite.sampleEvent.Intel.CreatedAt,
		CreatedBy:  suite.sampleEvent.Intel.CreatedBy,
		Operation:  suite.sampleEvent.Intel.Operation,
		Type:       store.IntelType(suite.sampleEvent.Intel.Type),
		Content:    suite.sampleEvent.Intel.Content,
		Importance: suite.sampleEvent.Intel.Importance,
	}
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) handle(ctx context.Context, tx pgx.Tx, rawValue json.RawMessage) error {
	return suite.port.Port.HandlerFn(suite.handler)(ctx, tx, kafkautil.InboundMessage{
		Topic:     event.IntelDeliveriesTopic,
		EventType: event.TypeIntelDeliveryAttemptCreated,
		RawValue:  rawValue,
	})
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) TestBadEventValue() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, json.RawMessage(`{invalid`))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateIntelDeliveryAttempt", timeout, tx, suite.sampleCreate, suite.sampleIntelToDeliver).
		Return(errors.New("sad life"))
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *portHandleIntelDeliveryAttemptCreatedSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.handler.On("CreateIntelDeliveryAttempt", timeout, tx, suite.sampleCreate, suite.sampleIntelToDeliver).
		Return(nil)
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.handle(timeout, tx, testutil.MarshalJSONMust(suite.sampleEvent))
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestPort_handleIntelDeliveryAttemptCreated(t *testing.T) {
	suite.Run(t, new(portHandleIntelDeliveryAttemptCreatedSuite))
}
