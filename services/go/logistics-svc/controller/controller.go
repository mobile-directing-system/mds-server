package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
)

// Controller manages all operations regarding logistics.
type Controller struct {
	Logger   *zap.Logger
	DB       pgutil.DBTxSupplier
	Store    Store
	Notifier Notifier
}

// Store for persistence.
type Store interface {
	// ChannelsByAddressBookEntry retrieves all channels for the address book entry
	// with the given id.
	ChannelsByAddressBookEntry(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) ([]store.Channel, error)
	// AssureAddressBookEntryExists makes sure that the address book entry with the
	// given id exists.
	AssureAddressBookEntryExists(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error
	// DeleteChannelWithDetailsByID deletes the channel with the given id and type.
	// This is meant to be used as a "shortcut" for clearing channel details as
	// well. This is why we expect the store.ChannelType as well without querying it
	// again.
	DeleteChannelWithDetailsByID(ctx context.Context, tx pgx.Tx, channelID uuid.UUID, channelType store.ChannelType) error
	// CreateChannelWithDetails creates the given store.Channel with its details.
	//
	// Warning: No entry existence checks are performed!
	CreateChannelWithDetails(ctx context.Context, tx pgx.Tx, channel store.Channel) error
	// UpdateChannelWithDetails updates the given store.Channel with its details.
	//
	// Warning: No entry existence checks are performed!
	UpdateChannelWithDetails(ctx context.Context, tx pgx.Tx, channel store.Channel) error
	// AddressBookEntries retrieves a paginated store.AddressBookEntryDetailed list
	// using the given store.AddressBookEntryFilters and pagination.Params.
	AddressBookEntries(ctx context.Context, tx pgx.Tx, filters store.AddressBookEntryFilters,
		paginationParams pagination.Params) (pagination.Paginated[store.AddressBookEntryDetailed], error)
	// AddressBookEntryByID retrieves the store.AddressBookEntryDetailed with the
	// given id. If visible-by is given, an meh.ErrNotFound will be returned, if the
	// entry is associated with a user, that is not part of any operation, the
	// client (visible-by) is part of.
	AddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID, visibleBy uuid.NullUUID) (store.AddressBookEntryDetailed, error)
	// CreateAddressBookEntry creates the given store.AddressBookEntry.
	CreateAddressBookEntry(ctx context.Context, tx pgx.Tx, entry store.AddressBookEntry) (store.AddressBookEntryDetailed, error)
	// UpdateAddressBookEntry updates the given store.AddressBookEntry, identified
	// by its id.
	UpdateAddressBookEntry(ctx context.Context, tx pgx.Tx, entry store.AddressBookEntry) error
	// DeleteAddressBookEntryByID deletes the address book entry with the given id.
	DeleteAddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error
	// CreateGroup creates the given store.Group.
	CreateGroup(ctx context.Context, tx pgx.Tx, create store.Group) error
	// UpdateGroup updates the given store.Group, identified by its id.
	UpdateGroup(ctx context.Context, tx pgx.Tx, update store.Group) error
	// DeleteGroupByID deletes the group with the given id.
	DeleteGroupByID(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error
	// CreateUser adds the given store.User to the store.
	CreateUser(ctx context.Context, tx pgx.Tx, create store.User) error
	// UpdateUser updates the given store.User, identified by its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, update store.User) error
	// DeleteUserByID deletes the user with the given id.
	DeleteUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	// CreateOperation creates the given store.Operation.
	CreateOperation(ctx context.Context, tx pgx.Tx, create store.Operation) error
	// UpdateOperation updates the given store.Operation.
	UpdateOperation(ctx context.Context, tx pgx.Tx, update store.Operation) error
	// UpdateOperationMembersByOperation updates the operation members for the given
	// operation.
	UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error
	// DeleteForwardToGroupChannelsByGroup deletes all channels with channel type
	// store.ChannelTypeForwardToGroup, that forward to the group with the given id.
	// It returns the list of affected address book entries.
	DeleteForwardToGroupChannelsByGroup(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) ([]uuid.UUID, error)
	// DeleteForwardToUserChannelsByUser deletes all channels with channel type
	// store.ChannelTypeForwardToUser, that forward to the user with the given id.
	// It returns the list of affected address book entries.
	DeleteForwardToUserChannelsByUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]uuid.UUID, error)
}

// Notifier sends event messages.
type Notifier interface {
	// NotifyAddressBookEntryCreated emits an event.TypeAddressBookEntryCreated
	// event.
	NotifyAddressBookEntryCreated(entry store.AddressBookEntry) error
	// NotifyAddressBookEntryUpdated emits an event.TypeAddressBookEntryUpdated
	// event.
	NotifyAddressBookEntryUpdated(entry store.AddressBookEntry) error
	// NotifyAddressBookEntryDeleted emits an event.TypeAddressBookEntryDeleted
	// event.
	NotifyAddressBookEntryDeleted(entryID uuid.UUID) error
	// NotifyAddressBookEntryChannelsUpdated emits an
	// event.TypeAddressBookEntryChannelsUpdated event.
	NotifyAddressBookEntryChannelsUpdated(entryID uuid.UUID, channels []store.Channel) error
}
