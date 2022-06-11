package event

// Topic is a Kafka topic. We only declare a different type for it in order to
// avoid common mistakes and misusage.
type Topic string

// UsersTopic is the Kafka topic to write user events to.
const UsersTopic Topic = "core.users.0"

// AuthTopic is the Kafka topic for all events related to authentication.
const AuthTopic Topic = "core.auth.0"
