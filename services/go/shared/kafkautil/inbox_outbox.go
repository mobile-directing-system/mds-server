package kafkautil

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/lefinal/meh/mehpg"
	"github.com/lib/pq"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgconnect"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgmigrate"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"io/fs"
	"math/rand"
	"time"
)

//go:embed db-migrations/*.sql
var dbMigrationsEmbedded embed.FS

var dbMigrations fs.FS

// dbScope is the scope for pgmigrate.Migrator.
const dbScope = "__message_inbox_outbox"

func init() {
	var err error
	dbMigrations, err = fs.Sub(dbMigrationsEmbedded, "db-migrations")
	if err != nil {
		panic("create sub fs for database migration")
	}
}

// inboxMessageStatus for messages in the inbox.
type inboxMessageStatus int

const (
	// inboxMessageStatusPending for messages that still need to be processed. This
	// is similar to inboxMessageStatusError, but these messages were already tried
	// to process.
	inboxMessageStatusPending inboxMessageStatus = 0
	// inboxMessageStatusProcessed for messages that have been processed.
	inboxMessageStatusProcessed inboxMessageStatus = 200
)

// outboxMessageStatus for messages in the outbox.
type outboxMessageStatus int

const (
	// outboxMessageStatusPending for messages that were never tried to send.
	outboxMessageStatusPending outboxMessageStatus = 0
	// outboxMessageStatusSent for messages that have been sent.
	outboxMessageStatusSent outboxMessageStatus = 200
)

// OutboxWriter allows writing messages to the event outbox.
type OutboxWriter interface {
	// AddOutboxMessages adds the given messages to the outbox.
	AddOutboxMessages(ctx context.Context, tx pgx.Tx, messages ...OutboundMessage) error
}

// Reader for Connector.PumpIncoming is an abstraction of kafka.Reader.
type Reader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}

// Connector implements the inbox/outbox-pattern for events/messages. It allows
// adding messages to the outbox by implementing OutboxWriter
type Connector interface {
	OutboxWriter
	// PumpOutgoing sends events from the outbox. It blocks until the given context
	// is done.
	PumpOutgoing(ctx context.Context, txSupplier pgutil.DBTxSupplier, writer Writer) error
	// Read fetches messages via the given Reader, adds them to the inbox and
	// commits them. It blocks until the given context is done.
	Read(ctx context.Context, txSupplier pgutil.DBTxSupplier, reader Reader) error
	// ProcessIncoming processes messages in the inbox with the given HandlerFunc
	// and blocks until the given context is done.
	ProcessIncoming(ctx context.Context, txSupplier pgutil.DBTxSupplier, handlerFn HandlerFunc) error
}

// RunConnector serves as a wrapper for calling Connector.PumpOutgoing,
// Connector.Read and Connector.ProcessIncoming and blocks.
func RunConnector(ctx context.Context, c Connector, txSupplier pgutil.DBTxSupplier, writer Writer, reader Reader, handlerFn HandlerFunc) error {
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return meh.NilOrWrap(c.PumpOutgoing(egCtx, txSupplier, writer), "pump outgoing", nil)
	})
	eg.Go(func() error {
		return meh.NilOrWrap(c.Read(egCtx, txSupplier, reader), "read", nil)
	})
	eg.Go(func() error {
		return meh.NilOrWrap(c.ProcessIncoming(egCtx, txSupplier, handlerFn), "process incoming", nil)
	})
	return eg.Wait()
}

// connector is the implementation of Connector.
type connector struct {
	logger *zap.Logger
	// id is an id for the instance. This is only useful for checking which instance
	// updated status for events.
	id    uuid.UUID
	store store
}

// InitNewConnector creates and initializes a new Connector.
func InitNewConnector(ctx context.Context, logger *zap.Logger, connPool *pgxpool.Pool) (Connector, error) {
	c, err := newConnector(logger, &dbStore{
		dialect: goqu.Dialect("postgres"),
	})
	if err != nil {
		return nil, meh.Wrap(err, "new connector", nil)
	}
	err = runDBMigrations(ctx, c.logger, connPool, dbMigrations)
	if err != nil {
		return nil, meh.Wrap(err, "run migrations", nil)
	}
	return c, nil
}

func newConnector(logger *zap.Logger, store store) (*connector, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "new id", nil)
	}
	c := &connector{
		logger: logger,
		id:     id,
		store:  store,
	}
	if c.logger == nil {
		c.logger = logging.DebugLogger()
	}
	return c, nil
}

func runDBMigrations(ctx context.Context, logger *zap.Logger, connPool *pgxpool.Pool, migrationsFS fs.FS) error {
	// Extract migrations.
	migrations, err := pgmigrate.MigrationsFromFS(migrationsFS)
	if err != nil {
		return meh.Wrap(err, "migrations from fs", nil)
	}
	// Run migrations.
	migrator, err := pgmigrate.NewMigrator(migrations, pgconnect.DefaultMigrationLogTable, dbScope)
	if err != nil {
		return meh.Wrap(err, "new migrator", meh.Details{"migration_log_table": pgconnect.DefaultMigrationLogTable})
	}
	singleConn, err := connPool.Acquire(ctx)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "acquire pgx conn", nil)
	}
	defer singleConn.Release()
	err = migrator.Up(ctx, logger, singleConn.Conn())
	if err != nil {
		return meh.Wrap(err, "migrator up", nil)
	}
	return nil
}

// AddOutboxMessages adds the given messages to the outbox table in the
// database.
func (c *connector) AddOutboxMessages(ctx context.Context, tx pgx.Tx, messages ...OutboundMessage) error {
	return c.store.addOutboxMessages(ctx, tx, c.id, messages...)
}

func (c *connector) PumpOutgoing(ctx context.Context, txSupplier pgutil.DBTxSupplier, writer Writer) error {
	logger := c.logger.Named("outgoing-pump")
	// Spawn workers.
	eg, egCtx := errgroup.WithContext(ctx)
	for i := 0; i < writerBatchSize; i++ {
		eg.Go(func() error {
			return pumpOutgoing(egCtx, logger, c.id, c.store, txSupplier, writer)
		})
	}
	return eg.Wait()
}

const (
	pumpOutgoingPollWait      = 500 * time.Millisecond
	pumpOutgoingErrorCooldown = 3 * time.Second
)

// pumpOutgoing is a worker that sends outbound messages. It only returns an
// error in fatal cases or when the context.Context is done.
func pumpOutgoing(ctx context.Context, logger *zap.Logger, instanceID uuid.UUID, store store,
	txSupplier pgutil.DBTxSupplier, writer Writer) error {
	for {
		wait := time.Duration(0)
		err := pgutil.RunInTx(ctx, txSupplier, func(ctx context.Context, tx pgx.Tx) error {
			// Retrieve next.
			next, ok, err := store.nextOutboxMessage(ctx, tx)
			if err != nil {
				return meh.Wrap(err, "next outbox message from store", nil)
			}
			if !ok {
				wait = pumpOutgoingPollWait
				return nil
			}
			// Send.
			err = WriteMessages(writer, next)
			if err != nil {
				return meh.Wrap(err, "write messages", meh.Details{"message": next})
			}
			// Update status.
			err = store.setOutboxMessageStatus(ctx, tx, instanceID, next.id, outboxMessageStatusSent)
			if err != nil {
				return meh.Wrap(err, "set outbox-message-status to sent", meh.Details{"message_id": next.id})
			}
			return nil
		})
		if err != nil {
			mehlog.Log(logger, meh.Wrap(err, "run in tx", nil))
			wait = pumpOutgoingErrorCooldown
		}
		if wait == 0 {
			continue
		}
		// Wait.
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
	}
}

const readErrorCooldown = 3 * time.Second

func (c *connector) Read(ctx context.Context, txSupplier pgutil.DBTxSupplier, reader Reader) error {
	logger := c.logger.Named("read")
	lastMessageOK := true
	var kafkaMessage kafka.Message
	for {
		wait := time.Duration(0)
		// Fetch and add.
		err := func() error {
			var err error
			if lastMessageOK {
				kafkaMessage, err = reader.FetchMessage(ctx)
				if err != nil {
					return meh.NewInternalErrFromErr(err, "fetch message", nil)
				}
				lastMessageOK = false
			}
			err = pgutil.RunInTx(ctx, txSupplier, func(ctx context.Context, tx pgx.Tx) error {
				m := inboundMessageFromKafkaMessage(kafkaMessage)
				// Add to inbox.
				err := c.store.addInboxMessages(ctx, tx, c.id, m)
				if err != nil {
					return meh.Wrap(err, "add inbox message to store", meh.Details{"message": m})
				}
				// Commit.
				err = reader.CommitMessages(ctx, kafkaMessage)
				if err != nil {
					return meh.NewInternalErrFromErr(err, "commit messages", meh.Details{"kafka_message": kafkaMessage})
				}
				return nil
			})
			if err != nil {
				return meh.Wrap(err, "run in tx", nil)
			}
			lastMessageOK = true
			return nil
		}()
		if err != nil {
			mehlog.Log(logger, err)
			wait = readErrorCooldown
		}
		if wait == 0 {
			continue
		}
		// Wait.
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
	}
}

const (
	processIncomingPollWait      = 500 * time.Millisecond
	processIncomingErrorCooldown = 3 * time.Second
)

func (c *connector) ProcessIncoming(ctx context.Context, txSupplier pgutil.DBTxSupplier, handlerFn HandlerFunc) error {
	logger := c.logger.Named("process-incoming")
	lastProcessFailed := false
	for {
		wait := time.Duration(0)
		err := pgutil.RunInTx(ctx, txSupplier, func(ctx context.Context, tx pgx.Tx) error {
			// Retrieve next.
			next, ok, err := c.store.nextInboxMessage(ctx, tx, lastProcessFailed)
			if err != nil {
				return meh.Wrap(err, "next inbox message from store", nil)
			}
			if !ok {
				wait = processIncomingPollWait
				return nil
			}
			// Process.
			err = handlerFn(ctx, tx, next)
			if err != nil {
				return meh.Wrap(err, "run handler", meh.Details{"message": next})
			}
			// Update status.
			err = c.store.setInboxMessageStatus(ctx, tx, c.id, next.id, inboxMessageStatusProcessed)
			if err != nil {
				return meh.Wrap(err, "set inbox-message-status to processed", meh.Details{
					"topic":     next.Topic,
					"partition": next.Partition,
					"offset":    next.Offset,
				})
			}
			return nil
		})
		lastProcessFailed = err != nil
		if err != nil {
			mehlog.Log(logger, meh.Wrap(err, "run in tx", nil))
			wait = processIncomingErrorCooldown
		}
		if wait == 0 {
			continue
		}
		// Wait.
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
	}
}

type store interface {
	// addInboxMessages adds the given messages to the inbox.
	addInboxMessages(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messages ...InboundMessage) error
	// nextInboxMessage retrieves the next inbox message to process and locks it. If
	// the random-flag is set, a random one of possible messages will be chosen. This
	// may lead to falsy return values but is an accepted tradeoff when handling
	// processing errors. The target use-case is a message that cannot be processed
	// because of requiring one from another topic and therefore creating a deadlock.
	nextInboxMessage(ctx context.Context, tx pgx.Tx, selectRandomSegment bool) (InboundMessage, bool, error)
	// setInboxMessageStatus updates the status for the given message.
	setInboxMessageStatus(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messageID int, status inboxMessageStatus) error
	// addOutboxMessages adds the given messages to the message outbox.
	addOutboxMessages(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messages ...OutboundMessage) error
	// nextOutboxMessage retrieves the next message to send allows concurrency.
	nextOutboxMessage(ctx context.Context, tx pgx.Tx) (OutboundMessage, bool, error)
	// setOutboxMessageStatus sets the status for the message with the given id.
	setOutboxMessageStatus(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messageID int, status outboxMessageStatus) error
}

type dbStore struct {
	dialect goqu.DialectWrapper
}

// addInboxMessages adds the given messages to the inbox table in the database.
func (s *dbStore) addInboxMessages(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messages ...InboundMessage) error {
	if len(messages) == 0 {
		return nil
	}
	rows := make([]any, 0, len(messages))
	for _, message := range messages {
		headerKeys := make([]string, 0, len(message.Headers))
		headerValues := make([]string, 0, len(message.Headers))
		for _, header := range message.Headers {
			headerKeys = append(headerKeys, header.Key)
			headerValues = append(headerValues, header.Value)
		}
		rows = append(rows, goqu.Record{
			"topic":           message.Topic,
			"partition":       message.Partition,
			"offset":          message.Offset,
			"ts":              message.TS.UTC(),
			"high_water_mark": message.HighWaterMark,
			"key":             message.Key,
			"value":           string(message.RawValue),
			"event_type":      message.EventType,
			"header_keys":     pq.Array(headerKeys),
			"header_values":   pq.Array(headerValues),
			"status":          inboxMessageStatusPending,
			"status_ts":       time.Now().UTC(),
			"status_by":       instanceID,
		})
	}
	q, _, err := s.dialect.Insert(goqu.T("__message_inbox")).
		Rows(rows...).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// nextInboxMessage retrieves the next inbox message to process and locks it. If
// the random-flag is set, a random one of possible messages will be chosen. This
// may lead to falsy return values but is an accepted tradeoff when handling
// processing errors. The target use-case is a message that cannot be processed
// because of requiring one from another topic and therefore creating a deadlock.
func (s *dbStore) nextInboxMessage(ctx context.Context, tx pgx.Tx, selectRandomSegment bool) (InboundMessage, bool, error) {
	oldestPendingPerSegment := s.dialect.From(goqu.T("__message_inbox")).As("oldest_pending").
		Select(goqu.I("__message_inbox.topic"),
			goqu.I("__message_inbox.partition"),
			goqu.I("__message_inbox.key"),
			goqu.MIN(goqu.I("__message_inbox.offset")).As("offset")).
		Where(goqu.I("__message_inbox.status").Neq(inboxMessageStatusProcessed)).
		GroupBy(goqu.I("__message_inbox.topic"),
			goqu.I("__message_inbox.partition"),
			goqu.I("__message_inbox.key"))
	possibleNextQuery, _, err := s.dialect.From(goqu.T("__message_inbox")).
		InnerJoin(oldestPendingPerSegment, goqu.On(
			goqu.I("__message_inbox.status").Neq(inboxMessageStatusProcessed), // Because of partial index in database.
			goqu.I("__message_inbox.topic").Eq(goqu.I("oldest_pending.topic")),
			goqu.I("__message_inbox.partition").Eq(goqu.I("oldest_pending.partition")),
			goqu.I("__message_inbox.offset").Eq(goqu.I("oldest_pending.offset")),
		)).
		Select(goqu.I("__message_inbox.id")).ToSQL()
	if err != nil {
		return InboundMessage{}, false, meh.NewInternalErrFromErr(err, "possible-next-query to sql", nil)
	}
	// Retrieve.
	possibleNextQueryRows, err := tx.Query(ctx, possibleNextQuery)
	if err != nil {
		return InboundMessage{}, false, mehpg.NewQueryDBErr(err, "exec possible-next-query", possibleNextQuery)
	}
	defer possibleNextQueryRows.Close()
	possibleNext := make([]int, 0)
	var possibleNextID int
	for possibleNextQueryRows.Next() {
		err = possibleNextQueryRows.Scan(&possibleNextID)
		if err != nil {
			return InboundMessage{}, false, mehpg.NewScanRowsErr(err, "scan possible-next-row", possibleNextQuery)
		}
		possibleNext = append(possibleNext, possibleNextID)
	}
	possibleNextQueryRows.Close()
	if len(possibleNext) == 0 {
		return InboundMessage{}, false, nil
	}
	if selectRandomSegment {
		i := rand.Intn(len(possibleNext))
		possibleNext = possibleNext[i : i+1]
	}
	// Choose the oldest one that matches conditions.
	nextQuery, _, err := s.dialect.From(goqu.T("__message_inbox")).
		Select(goqu.I("__message_inbox.id"),
												goqu.I("__message_inbox.topic"),
												goqu.I("__message_inbox.partition"),
												goqu.I("__message_inbox.offset"),
												goqu.I("__message_inbox.ts"),
												goqu.I("__message_inbox.high_water_mark"),
												goqu.I("__message_inbox.key"),
												goqu.I("__message_inbox.value"),
												goqu.I("__message_inbox.event_type"),
												goqu.I("__message_inbox.header_keys"),
												goqu.I("__message_inbox.header_values")).
		Where(goqu.I("__message_inbox.status").Neq(inboxMessageStatusProcessed), // Because of partial index.
			goqu.I("__message_inbox.id").In(possibleNext),
			goqu.Or(
				// Common case.
				goqu.I("__message_inbox.status").Eq(inboxMessageStatusPending),
			)).
		Order(goqu.I("__message_inbox.status_ts").Asc()).
		ForUpdate(exp.SkipLocked).
		Limit(1).ToSQL()
	if err != nil {
		return InboundMessage{}, false, meh.NewInternalErrFromErr(err, "next-query to sql", nil)
	}
	nextRows, err := tx.Query(ctx, nextQuery)
	if err != nil {
		return InboundMessage{}, false, mehpg.NewQueryDBErr(err, "exec next-query", nextQuery)
	}
	defer nextRows.Close()
	if !nextRows.Next() {
		return InboundMessage{}, false, nil
	}
	var m InboundMessage
	var mValue string
	var mHeaderKeys []string
	var mHeaderValues []string
	err = nextRows.Scan(&m.id,
		&m.Topic,
		&m.Partition,
		&m.Offset,
		&m.TS,
		&m.HighWaterMark,
		&m.Key,
		&mValue,
		&m.EventType,
		&mHeaderKeys,
		&mHeaderValues)
	if err != nil {
		return InboundMessage{}, false, mehpg.NewScanRowsErr(err, "scan possible-next-row", possibleNextQuery)
	}
	m.RawValue = json.RawMessage(mValue)
	if len(mHeaderKeys) != len(mHeaderValues) {
		return InboundMessage{}, false, meh.NewInternalErr("list length mismatch for header keys and values", meh.Details{
			"message_header_keys":    mHeaderKeys,
			"message_header_values":  mHeaderValues,
			"next_message_until_now": m,
			"query":                  possibleNextQuery,
		})
	}
	m.Headers = make([]MessageHeader, 0, len(mHeaderKeys))
	for i := range mHeaderKeys {
		m.Headers = append(m.Headers, MessageHeader{
			Key:   mHeaderKeys[i],
			Value: mHeaderValues[i],
		})
	}
	return m, true, nil
}

// setInboxMessageStatus updates the status and update timestamp for the given
// message in the inbox table in the database, not having
// inboxMessageStatusProcessed.
func (s *dbStore) setInboxMessageStatus(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messageID int, status inboxMessageStatus) error {
	q, _, err := s.dialect.Update(goqu.T("__message_inbox")).Set(goqu.Record{
		"status":    status,
		"status_ts": time.Now().UTC(),
		"status_by": instanceID,
	}).Where(goqu.C("status").Neq(inboxMessageStatusProcessed),
		goqu.C("id").Eq(messageID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("message not found", meh.Details{"query": q})
	}
	return nil
}

// addOutboxMessages inserts the given messages into the outbox table in the
// database.
func (s *dbStore) addOutboxMessages(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messages ...OutboundMessage) error {
	if len(messages) == 0 {
		return nil
	}
	rows := make([]any, 0, len(messages))
	for _, message := range messages {
		headerKeys := make([]string, 0, len(message.Headers))
		headerValues := make([]string, 0, len(message.Headers))
		for _, header := range message.Headers {
			headerKeys = append(headerKeys, header.Key)
			headerValues = append(headerValues, header.Value)
		}
		valueRaw, err := json.Marshal(message.Value)
		if err != nil {
			return meh.NewInternalErrFromErr(err, "marshal message value", nil)
		}
		rows = append(rows, goqu.Record{
			"topic":         message.Topic,
			"created":       time.Now().UTC(),
			"key":           message.Key,
			"value":         string(valueRaw),
			"event_type":    message.EventType,
			"header_keys":   pq.Array(headerKeys),
			"header_values": pq.Array(headerValues),
			"status":        inboxMessageStatusPending,
			"status_ts":     time.Now().UTC(),
			"status_by":     instanceID,
		})
	}
	q, _, err := s.dialect.Insert(goqu.T("__message_outbox")).
		Rows(rows...).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	_, err = tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	return nil
}

// nextOutboxMessage retrieves the next outbox message to send and locks it.
func (s *dbStore) nextOutboxMessage(ctx context.Context, tx pgx.Tx) (OutboundMessage, bool, error) {
	// Gather identifiers for the oldest unprocessed message per segment (topic,
	// key).
	possibleNextQuery, _, err := s.dialect.From(goqu.T("__message_outbox")).As("oldest_pending").
		Select(goqu.MIN(goqu.I("__message_outbox.id")).As("id")).
		Where(goqu.I("__message_outbox.status").Neq(outboxMessageStatusSent)).
		GroupBy(goqu.I("__message_outbox.topic"),
			goqu.I("__message_outbox.key")).ToSQL()
	if err != nil {
		return OutboundMessage{}, false, meh.NewInternalErrFromErr(err, "possible-next-query to sql", nil)
	}
	possibleNextRows, err := tx.Query(ctx, possibleNextQuery)
	if err != nil {
		return OutboundMessage{}, false, mehpg.NewQueryDBErr(err, "exec possible-next-query", possibleNextQuery)
	}
	defer possibleNextRows.Close()
	possibleNext := make([]int, 0)
	var possibleNextID int
	for possibleNextRows.Next() {
		err = possibleNextRows.Scan(&possibleNextID)
		if err != nil {
			return OutboundMessage{}, false, mehpg.NewScanRowsErr(err, "scan possible-next-row", possibleNextQuery)
		}
		possibleNext = append(possibleNext, possibleNextID)
	}
	possibleNextRows.Close()
	if len(possibleNext) == 0 {
		return OutboundMessage{}, false, nil
	}
	// Choose the oldest one that matches conditions.
	nextQuery, _, err := s.dialect.From(goqu.T("__message_outbox")).
		Select(goqu.I("__message_outbox.id"),
											goqu.I("__message_outbox.topic"),
											goqu.I("__message_outbox.key"),
											goqu.I("__message_outbox.value"),
											goqu.I("__message_outbox.event_type"),
											goqu.I("__message_outbox.header_keys"),
											goqu.I("__message_outbox.header_values")).
		Where(goqu.I("__message_outbox.status").Neq(outboxMessageStatusSent), // Because of partial index.
			goqu.I("__message_outbox.id").In(possibleNext),
			goqu.Or(
				// Common case.
				goqu.I("__message_outbox.status").Eq(outboxMessageStatusPending),
			)).
		Order(goqu.I("__message_outbox.id").Asc()).
		ForUpdate(exp.SkipLocked).
		Limit(1).ToSQL()
	if err != nil {
		return OutboundMessage{}, false, meh.NewInternalErrFromErr(err, "next-query to sql", nil)
	}
	// Retrieve.
	nextRows, err := tx.Query(ctx, nextQuery)
	if err != nil {
		return OutboundMessage{}, false, mehpg.NewQueryDBErr(err, "exec next-query", nextQuery)
	}
	defer nextRows.Close()
	if !nextRows.Next() {
		return OutboundMessage{}, false, nil
	}
	var m OutboundMessage
	var mValue string
	var mHeaderKeys []string
	var mHeaderValues []string
	err = nextRows.Scan(&m.id,
		&m.Topic,
		&m.Key,
		&mValue,
		&m.EventType,
		&mHeaderKeys,
		&mHeaderValues)
	if err != nil {
		return OutboundMessage{}, false, mehpg.NewScanRowsErr(err, "scan row", nextQuery)
	}
	m.Value = json.RawMessage(mValue)
	if len(mHeaderKeys) != len(mHeaderValues) {
		return OutboundMessage{}, false, meh.NewInternalErr("list length mismatch for header keys and values", meh.Details{
			"message_header_keys":    mHeaderKeys,
			"message_header_values":  mHeaderValues,
			"next_message_until_now": m,
			"query":                  nextQuery,
		})
	}
	m.Headers = make([]MessageHeader, 0, len(mHeaderKeys))
	for i := range mHeaderKeys {
		m.Headers = append(m.Headers, MessageHeader{
			Key:   mHeaderKeys[i],
			Value: mHeaderValues[i],
		})
	}
	return m, true, nil
}

// setOutboxMessageStatus updates the status and update timestamp for the given
// message in the outbox table in the database.
func (s *dbStore) setOutboxMessageStatus(ctx context.Context, tx pgx.Tx, instanceID uuid.UUID, messageID int, status outboxMessageStatus) error {
	q, _, err := s.dialect.Update(goqu.T("__message_outbox")).Set(goqu.Record{
		"status":    status,
		"status_ts": time.Now().UTC(),
		"status_by": instanceID,
	}).Where(goqu.C("id").Eq(messageID)).ToSQL()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "query to sql", nil)
	}
	result, err := tx.Exec(ctx, q)
	if err != nil {
		return mehpg.NewQueryDBErr(err, "exec query", q)
	}
	if result.RowsAffected() == 0 {
		return meh.NewNotFoundErr("message not found", meh.Details{"query": q})
	}
	return nil
}
