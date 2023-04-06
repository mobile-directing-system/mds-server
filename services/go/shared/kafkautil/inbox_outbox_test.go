package kafkautil

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/zaprec"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	"time"
)

// ReaderMock mocks Reader.
type ReaderMock struct {
	mock.Mock
}

func (m *ReaderMock) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *ReaderMock) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.Called(ctx, msgs).Error(0)
}

// WriterMock mocks Writer.
type WriterMock struct {
	mock.Mock
}

func (m *WriterMock) WriteMessages(ctx context.Context, messages ...kafka.Message) error {
	return m.Called(ctx, messages).Error(0)
}

// ConnectorMock mocks Connector.
type ConnectorMock struct {
	mock.Mock
}

func (m *ConnectorMock) AddOutboxMessages(ctx context.Context, tx pgx.Tx, messages ...OutboundMessage) error {
	return m.Called(ctx, tx, messages).Error(0)
}

func (m *ConnectorMock) PumpOutgoing(ctx context.Context, txSupplier pgutil.DBTxSupplier, writer Writer) error {
	return m.Called(ctx, txSupplier, writer).Error(0)
}

func (m *ConnectorMock) Read(ctx context.Context, txSupplier pgutil.DBTxSupplier, reader Reader) error {
	return m.Called(ctx, txSupplier, reader).Error(0)
}

func (m *ConnectorMock) ProcessIncoming(ctx context.Context, txSupplier pgutil.DBTxSupplier, handlerFn HandlerFunc) error {
	return m.Called(ctx, txSupplier, handlerFn).Error(0)
}

// storeMock mocks store.
type storeMock struct {
	mock.Mock
}

func (m *storeMock) addInboxMessages(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messages ...InboundMessage) error {
	return m.Called(ctx, tx, instanceID, messages).Error(0)
}

func (m *storeMock) nextInboxMessage(ctx context.Context, tx pgx.Tx, selectRandomSegment bool) (InboundMessage, bool, error) {
	args := m.Called(ctx, tx, selectRandomSegment)
	return args.Get(0).(InboundMessage), args.Bool(1), args.Error(2)
}

func (m *storeMock) setInboxMessageStatus(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messageID int, status inboxMessageStatus) error {
	return m.Called(ctx, tx, instanceID, messageID, status).Error(0)
}

func (m *storeMock) addOutboxMessages(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messages ...OutboundMessage) error {
	return m.Called(ctx, tx, instanceID, messages).Error(0)
}

func (m *storeMock) nextOutboxMessage(ctx context.Context, tx pgx.Tx) (OutboundMessage, bool, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(OutboundMessage), args.Bool(1), args.Error(2)
}

func (m *storeMock) setOutboxMessageStatus(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messageID int, status outboxMessageStatus) error {
	return m.Called(ctx, tx, instanceID, messageID, status).Error(0)
}

func TestRunConnector(t *testing.T) {
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	c := &ConnectorMock{}
	txSupplier := &testutil.DBTxSupplier{}
	writer := &WriterMock{}
	reader := &ReaderMock{}
	handlerFn := func(_ context.Context, _ pgx.Tx, _ InboundMessage) error {
		return nil
	}
	c.On("PumpOutgoing", mock.Anything, txSupplier, writer).
		Return(errors.New("sad life")).Once()
	c.On("Read", mock.Anything, txSupplier, reader).
		Return(errors.New("sad life")).Once()
	c.On("ProcessIncoming", mock.Anything, txSupplier, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer c.AssertExpectations(t)

	go func() {
		defer cancel()
		err := RunConnector(timeout, c, txSupplier, writer, reader, handlerFn)
		assert.Error(t, err, "should fail")
	}()

	wait()
}

// newConnectorSuite tests newConnector.
type newConnectorSuite struct {
	suite.Suite
}

func (suite *newConnectorSuite) TestNilLogger() {
	c, err := newConnector(nil, nil)
	suite.Require().NoError(err, "should not fail")
	suite.NotNil(c.logger, "should have set shared debug logger")
}

func (suite *newConnectorSuite) TestOK() {
	c, err := newConnector(zap.NewNop(), &storeMock{})
	suite.Require().NoError(err, "should not fail")
	suite.NotNil(c.logger, "should have set logger")
	suite.NotEmpty(c.id, "should have set id")
	suite.NotNil(c.logger, "should have set logger")
	suite.NotNil(c.store, "should have set store")
}

func Test_newConnector(t *testing.T) {
	suite.Run(t, new(newConnectorSuite))
}

func Test_connectorAddOutboxMessages(t *testing.T) {
	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	tx := &testutil.DBTx{}
	messages := []OutboundMessage{{}, {}}
	s := &storeMock{}
	c, err := newConnector(zap.NewNop(), s)
	require.NoError(t, err, "connector creation should not fail")
	s.On("addOutboxMessages", timeout, tx, c.id, messages).Return(errors.New("sad life"))
	defer s.AssertExpectations(t)

	go func() {
		defer cancel()
		err := c.AddOutboxMessages(timeout, tx, messages...)
		assert.Error(t, err, "should fail")
	}()

	wait()
}

// pumpOutgoingSuite tests pumpOutgoing.
type pumpOutgoingSuite struct {
	suite.Suite
	instanceID    uuid.UUID
	store         *storeMock
	txSupplier    *testutil.DBTxSupplier
	writer        *WriterMock
	logger        *zap.Logger
	recorder      *zaprec.RecordStore
	sampleMessage OutboundMessage
}

func (suite *pumpOutgoingSuite) SetupTest() {
	suite.instanceID = testutil.NewUUIDV4()
	suite.store = &storeMock{}
	suite.txSupplier = &testutil.DBTxSupplier{}
	suite.writer = &WriterMock{}
	suite.logger, suite.recorder = zaprec.NewRecorder(zap.ErrorLevel)
	suite.sampleMessage = OutboundMessage{
		id:        14,
		Topic:     "read",
		Key:       "heaven",
		EventType: "because",
		Value:     OutboundMessage{},
		Headers:   nil,
	}
}

func (suite *pumpOutgoingSuite) test(ctx context.Context) error {
	// Exit immediately because we do not test delays or cooldowns.
	ctx, done := context.WithCancel(ctx)
	done()
	return pumpOutgoing(ctx, suite.logger, suite.instanceID, suite.store, suite.txSupplier, suite.writer)
}

func (suite *pumpOutgoingSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.BeginFail = true

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
	}()

	wait()
}

func (suite *pumpOutgoingSuite) TestNextFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.store.On("nextOutboxMessage", mock.Anything, suite.txSupplier.Tx[0]).
		Return(OutboundMessage{}, false, errors.New("sad life"))
	defer suite.store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
	}()

	wait()
}

func (suite *pumpOutgoingSuite) TestNoNext() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.store.On("nextOutboxMessage", mock.Anything, suite.txSupplier.Tx[0]).
		Return(OutboundMessage{}, false, nil)
	defer suite.store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Empty(suite.recorder.Records(), "should not have logged errors")
	}()

	wait()
}

func (suite *pumpOutgoingSuite) TestWriteFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.store.On("nextOutboxMessage", mock.Anything, suite.txSupplier.Tx[0]).
		Return(suite.sampleMessage, true, nil)
	suite.writer.On("WriteMessages", mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.store.AssertExpectations(suite.T())
	defer suite.writer.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
	}()

	wait()
}

func (suite *pumpOutgoingSuite) TestSentStatusUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.store.On("nextOutboxMessage", mock.Anything, suite.txSupplier.Tx[0]).
		Return(suite.sampleMessage, true, nil)
	suite.writer.On("WriteMessages", mock.Anything, mock.Anything).
		Return(nil)
	suite.store.On("setOutboxMessageStatus", mock.Anything, suite.txSupplier.Tx[0], suite.instanceID,
		suite.sampleMessage.id, outboxMessageStatusSent).
		Return(errors.New("sad life"))
	defer suite.store.AssertExpectations(suite.T())
	defer suite.writer.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged both errors")
	}()

	wait()
}

func (suite *pumpOutgoingSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}, {}} // 2 because of immediate proceed.
	suite.store.On("nextOutboxMessage", mock.Anything, suite.txSupplier.Tx[0]).
		Return(suite.sampleMessage, true, nil).Once()
	suite.writer.On("WriteMessages", mock.Anything, mock.Anything).
		Return(nil)
	suite.store.On("setOutboxMessageStatus", mock.Anything, suite.txSupplier.Tx[0], suite.instanceID,
		suite.sampleMessage.id, outboxMessageStatusSent).
		Return(nil)
	suite.store.On("nextOutboxMessage", mock.Anything, suite.txSupplier.Tx[1]).
		Return(OutboundMessage{}, false, nil).Once()
	defer suite.store.AssertExpectations(suite.T())
	defer suite.writer.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Empty(suite.recorder.Records(), "should not have logged errors")
	}()

	wait()
}

func Test_pumpOutgoing(t *testing.T) {
	suite.Run(t, new(pumpOutgoingSuite))
}

// connectorReadSuite tests connector.Read.
type connectorReadSuite struct {
	suite.Suite
	c                  *connector
	store              *storeMock
	txSupplier         *testutil.DBTxSupplier
	reader             *ReaderMock
	logger             *zap.Logger
	recorder           *zaprec.RecordStore
	sampleKafkaMessage kafka.Message
	sampleMessage      InboundMessage
}

func (suite *connectorReadSuite) SetupTest() {
	suite.logger, suite.recorder = zaprec.NewRecorder(zap.ErrorLevel)
	suite.store = &storeMock{}
	suite.txSupplier = &testutil.DBTxSupplier{}
	suite.reader = &ReaderMock{}
	var err error
	suite.c, err = newConnector(suite.logger, suite.store)
	suite.Require().NoError(err, "connector creation should not fail")
	suite.sampleKafkaMessage = kafka.Message{
		Topic:         "read",
		Partition:     806,
		Offset:        688,
		HighWaterMark: 344,
		Key:           []byte("heaven"),
		Headers:       nil,
	}
	suite.sampleMessage = inboundMessageFromKafkaMessage(suite.sampleKafkaMessage)
}

func (suite *connectorReadSuite) test(ctx context.Context) error {
	// Exit immediately because we do not test delays or cooldowns.
	ctx, done := context.WithCancel(ctx)
	done()
	return suite.c.Read(ctx, suite.txSupplier, suite.reader)
}

func (suite *connectorReadSuite) TestFetchFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.reader.On("FetchMessage", mock.Anything).
		Return(kafka.Message{}, errors.New("sad life"))
	defer suite.reader.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
	}()

	wait()
}

func (suite *connectorReadSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.BeginFail = true
	suite.reader.On("FetchMessage", mock.Anything).
		Return(suite.sampleKafkaMessage, nil)
	defer suite.reader.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
	}()

	wait()
}

func (suite *connectorReadSuite) TestAddFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.reader.On("FetchMessage", mock.Anything).
		Return(suite.sampleKafkaMessage, nil)
	suite.store.On("addInboxMessages", mock.Anything, suite.txSupplier.Tx[0], suite.c.id,
		[]InboundMessage{suite.sampleMessage}).
		Return(errors.New("sad life"))
	defer suite.reader.AssertExpectations(suite.T())
	defer suite.store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
		suite.False(suite.txSupplier.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *connectorReadSuite) TestCommitFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.reader.On("FetchMessage", mock.Anything).
		Return(suite.sampleKafkaMessage, nil)
	suite.store.On("addInboxMessages", mock.Anything, suite.txSupplier.Tx[0], suite.c.id,
		[]InboundMessage{suite.sampleMessage}).
		Return(nil)
	suite.reader.On("CommitMessages", mock.Anything, []kafka.Message{suite.sampleKafkaMessage}).
		Return(errors.New("sad life"))
	defer suite.reader.AssertExpectations(suite.T())
	defer suite.store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
		suite.False(suite.txSupplier.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *connectorReadSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}, {}}
	suite.reader.On("FetchMessage", mock.Anything).
		Return(suite.sampleKafkaMessage, nil).Once()
	suite.reader.On("FetchMessage", mock.Anything).
		Return(kafka.Message{}, context.Canceled).Once()
	suite.store.On("addInboxMessages", mock.Anything, suite.txSupplier.Tx[0], suite.c.id,
		[]InboundMessage{suite.sampleMessage}).
		Return(nil)
	suite.reader.On("CommitMessages", mock.Anything, []kafka.Message{suite.sampleKafkaMessage}).
		Return(nil)
	defer suite.reader.AssertExpectations(suite.T())
	defer suite.store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.True(suite.txSupplier.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func Test_connectorRead(t *testing.T) {
	suite.Run(t, new(connectorReadSuite))
}

// handlerFnMock is a container for a mocked HandlerFunc.
type handlerFnMock struct {
	mock.Mock
}

func (m *handlerFnMock) fn(ctx context.Context, tx pgx.Tx, message InboundMessage) error {
	return m.Called(ctx, tx, message).Error(0)
}

// connectorProcessIncomingSuite tests connector.ProcessIncoming.
type connectorProcessIncomingSuite struct {
	suite.Suite
	c             *connector
	store         *storeMock
	txSupplier    *testutil.DBTxSupplier
	reader        *ReaderMock
	logger        *zap.Logger
	recorder      *zaprec.RecordStore
	handler       *handlerFnMock
	sampleMessage InboundMessage
}

func (suite *connectorProcessIncomingSuite) SetupTest() {
	suite.logger, suite.recorder = zaprec.NewRecorder(zap.ErrorLevel)
	suite.store = &storeMock{}
	suite.txSupplier = &testutil.DBTxSupplier{}
	suite.reader = &ReaderMock{}
	suite.handler = &handlerFnMock{}
	var err error
	suite.c, err = newConnector(suite.logger, suite.store)
	suite.Require().NoError(err, "connector creation should not fail")
	suite.sampleMessage = InboundMessage{
		id:            665,
		Topic:         "read",
		Partition:     806,
		Offset:        688,
		HighWaterMark: 344,
		TS:            time.Date(2022, 8, 15, 17, 13, 27, 0, time.UTC),
		Key:           "heaven",
		EventType:     "hello-world",
		RawValue:      json.RawMessage(`{"hello":"world"}`),
		Headers:       nil,
	}
}

func (suite *connectorProcessIncomingSuite) test(ctx context.Context) error {
	// Exit immediately because we do not test delays or cooldowns.
	ctx, done := context.WithCancel(ctx)
	done()
	return suite.c.ProcessIncoming(ctx, suite.txSupplier, suite.handler.fn)
}

func (suite *connectorProcessIncomingSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.BeginFail = true

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
	}()

	wait()
}

func (suite *connectorProcessIncomingSuite) TestNextFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.store.On("nextInboxMessage", mock.Anything, suite.txSupplier.Tx[0], false).
		Return(InboundMessage{}, false, errors.New("sad life"))
	defer suite.store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
		suite.False(suite.txSupplier.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *connectorProcessIncomingSuite) TestNoNext() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.store.On("nextInboxMessage", mock.Anything, suite.txSupplier.Tx[0], false).
		Return(InboundMessage{}, false, nil)
	defer suite.store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Empty(suite.recorder.Records(), "should not have logged errors")
		suite.True(suite.txSupplier.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func (suite *connectorProcessIncomingSuite) TestHandleFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.store.On("nextInboxMessage", mock.Anything, suite.txSupplier.Tx[0], false).
		Return(suite.sampleMessage, true, nil).Once()
	suite.handler.On("fn", mock.Anything, suite.txSupplier.Tx[0], suite.sampleMessage).
		Return(errors.New("sad life"))
	defer suite.store.AssertExpectations(suite.T())
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
		suite.False(suite.txSupplier.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *connectorProcessIncomingSuite) TestProcessedStatusUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}}
	suite.store.On("nextInboxMessage", mock.Anything, suite.txSupplier.Tx[0], false).
		Return(suite.sampleMessage, true, nil).Once()
	suite.handler.On("fn", mock.Anything, suite.txSupplier.Tx[0], suite.sampleMessage).
		Return(nil)
	suite.store.On("setInboxMessageStatus", mock.Anything, suite.txSupplier.Tx[0], suite.c.id,
		suite.sampleMessage.id, inboxMessageStatusProcessed).
		Return(errors.New("sad life"))
	defer suite.store.AssertExpectations(suite.T())
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Len(suite.recorder.Records(), 1, "should have logged errors")
		suite.False(suite.txSupplier.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *connectorProcessIncomingSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.txSupplier.Tx = []*testutil.DBTx{{}, {}} // 2 because of no cooldown.
	suite.store.On("nextInboxMessage", mock.Anything, suite.txSupplier.Tx[0], false).
		Return(suite.sampleMessage, true, nil).Once()
	suite.handler.On("fn", mock.Anything, suite.txSupplier.Tx[0], suite.sampleMessage).
		Return(nil)
	suite.store.On("setInboxMessageStatus", mock.Anything, suite.txSupplier.Tx[0], suite.c.id,
		suite.sampleMessage.id, inboxMessageStatusProcessed).
		Return(nil)
	suite.store.On("nextInboxMessage", mock.Anything, suite.txSupplier.Tx[1], false).
		Return(InboundMessage{}, false, nil).Once()
	defer suite.store.AssertExpectations(suite.T())
	defer suite.handler.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.test(timeout)
		suite.NoError(err, "should not fail")
		suite.Empty(suite.recorder.Records(), 0, "should not have logged errors")
		suite.True(suite.txSupplier.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func Test_connectorProcessIncoming(t *testing.T) {
	suite.Run(t, new(connectorProcessIncomingSuite))
}
