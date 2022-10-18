package event

// Topic is a Kafka topic. We only declare a different type for it in order to
// avoid common mistakes and misusage.
type Topic string

const (
	// AddressBookTopic is the Kafka topic for the address book.
	AddressBookTopic Topic = "logistics.address-book.0"
	// AuthTopic is the Kafka topic for all events related to authentication.
	AuthTopic Topic = "core.auth.0"
	// GroupsTopic is the Kafka topic for groups.
	GroupsTopic Topic = "orga.groups.0"
	// InAppNotificationsTopic is the Kafka topic for in-app-notifications.
	InAppNotificationsTopic Topic = "notifications.in-app.0"
	// IntelDeliveriesTopic is the Kafka topic for delivering intel.
	IntelDeliveriesTopic Topic = "logistics.intel-delivery.0"
	// IntelTopic is the Kafka topic for intel.
	IntelTopic = "intelligence.intel.0"
	// OperationsTopic is the Kafka topic for all events related to operation
	// management.
	OperationsTopic Topic = "operations.operations.0"
	// PermissionsTopic is the Kafka topic for all events related to permissions.
	PermissionsTopic Topic = "core.permissions.0"
	// RadioDeliveriesTopic is the Kafka topic for delivering intel over radio.
	RadioDeliveriesTopic Topic = "delivery.radio.0"
	// UsersTopic is the Kafka topic to write user events to.
	UsersTopic Topic = "core.users.0"
)
