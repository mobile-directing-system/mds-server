package kafkautil

import (
	"context"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/event"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"time"
)

// AwaitTopics waits for each given event.Topic to exist. Once, all topics are
// available, the waiting time will be logged to the given zap.Logger.
func AwaitTopics(ctx context.Context, logger *zap.Logger, kafkaAddr string, topics ...event.Topic) error {
	start := time.Now()
	awaiter := newTopicAwaiter(topics...)
	for {
		select {
		case <-ctx.Done():
			return meh.NewInternalErrFromErr(ctx.Err(), "await topics", meh.Details{
				"remaining_expected": awaiter.remainingExpected(),
				"topics":             "topics",
				"waiting_since":      time.Since(start),
			})
		default:
		}
		// Check if done.
		if awaiter.remainingExpected() == 0 {
			logger.Info("topics available",
				zap.Any("topics", topics),
				zap.Duration("took", time.Since(start)))
			return nil
		}
		// Read and mark as available.
		partitions, err := readPartitions(ctx, kafkaAddr, topics)
		if err != nil {
			return meh.Wrap(err, "read partitions", meh.Details{
				"kafka_addr": kafkaAddr,
				"topics":     topics,
			})
		}
		for _, partition := range partitions {
			awaiter.markTopicAsAvailable(event.Topic(partition.Topic))
		}
	}
}

// readPartitions reads partitions from the Kafka broker at the given address.
func readPartitions(ctx context.Context, kafkaAddr string, topics []event.Topic) ([]kafka.Partition, error) {
	conn, err := kafka.DialContext(ctx, "tcp", kafkaAddr)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "dial kafka", meh.Details{"kafka_addr": kafkaAddr})
	}
	defer func() { _ = conn.Close() }()
	topicsStr := make([]string, 0, len(topics))
	for _, topic := range topics {
		topicsStr = append(topicsStr, string(topic))
	}
	partitions, err := conn.ReadPartitions(topicsStr...)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "read partitions", meh.Details{"topics": topicsStr})
	}
	return partitions, nil
}

// topicAwaiter manager an event.Topic map with expected topics and provides
// markTopicAsAvailable for removing the given event.Topic from remaining
// expected ones. The amount of remaining expected ones is provided via
// remainingExpected.
type topicAwaiter struct {
	expected map[event.Topic]struct{}
}

// newTopicAwaiter creates a new topicAwaiter that expects the given event.Topic
// list.
func newTopicAwaiter(expectedTopics ...event.Topic) *topicAwaiter {
	expected := make(map[event.Topic]struct{})
	for _, topic := range expectedTopics {
		expected[topic] = struct{}{}
	}
	return &topicAwaiter{
		expected: expected,
	}
}

// markTopicAsAvailable deletes the given event.Topic from remaining expected
// ones.
func (awaiter *topicAwaiter) markTopicAsAvailable(topic event.Topic) {
	delete(awaiter.expected, topic)
}

// remainingExpected returns the amount of remaining expected topics.
func (awaiter *topicAwaiter) remainingExpected() int {
	return len(awaiter.expected)
}
