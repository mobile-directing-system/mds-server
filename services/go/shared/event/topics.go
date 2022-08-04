package event

// Topic is a Kafka topic. We only declare a different type for it in order to
// avoid common mistakes and misusage.
type Topic string

// AuthTopic is the Kafka topic for all events related to authentication.
const AuthTopic Topic = "core.auth.0"

// GroupsTopic is the Kafka topic for groups.
const GroupsTopic Topic = "orga.groups.0"

// AddressBookTopic is the Kafka topic for the address book.
const AddressBookTopic Topic = "logistics.address-book.0"

// OperationsTopic is the Kafka topic for all events related to operation
// management.
const OperationsTopic Topic = "operations.operations.0"

// PermissionsTopic is the Kafka topic for all events related to permissions.
const PermissionsTopic Topic = "core.permissions.0"

// UsersTopic is the Kafka topic to write user events to.
const UsersTopic Topic = "core.users.0"
