package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/zap"
)

// Controller manages all operations regarding intelligence.
type Controller struct {
	Logger   *zap.Logger
	DB       pgutil.DBTxSupplier
	Store    Store
	Notifier Notifier
}

// Store for Controller.
type Store interface {
	// CreateUser creates the user with the given id.
	CreateUser(ctx context.Context, tx pgx.Tx, userID store.User) error
	// UpdateUser updates the given store.user, identified by its id.
	UpdateUser(ctx context.Context, tx pgx.Tx, user store.User) error
	// CreateOperation creates the given store.Operation.
	CreateOperation(ctx context.Context, tx pgx.Tx, create store.Operation) error
	// UpdateOperation updates the given store.Operation.
	UpdateOperation(ctx context.Context, tx pgx.Tx, update store.Operation) error
	// UpdateOperationMembersByOperation updates the operation members for the given
	// operation.
	UpdateOperationMembersByOperation(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, newMembers []uuid.UUID) error
	// CreateAddressBookEntry creates the given store.AddressBookEntry.
	CreateAddressBookEntry(ctx context.Context, tx pgx.Tx, create store.AddressBookEntry) error
	// UpdateAddressBookEntry updates the given store.AddressBookEntry.
	UpdateAddressBookEntry(ctx context.Context, tx pgx.Tx, update store.AddressBookEntry) error
	// DeleteAddressBookEntryByID deletes the address book entry with the given id.
	DeleteAddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) error
	// CreateIntel creates the given intel and returns it with its assigned id.
	CreateIntel(ctx context.Context, tx pgx.Tx, create store.CreateIntel) (store.Intel, error)
	// InvalidateIntelByID marks the intel with the given id as invalid.
	InvalidateIntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) error
	// IntelByID retrieves a store.Intel by its id.
	IntelByID(ctx context.Context, tx pgx.Tx, intelID uuid.UUID) (store.Intel, error)
	// IsUserOperationMember checks if the user with the given id is member of the
	// give operation.
	IsUserOperationMember(ctx context.Context, tx pgx.Tx, userID uuid.UUID, operationID uuid.UUID) (bool, error)
	// AddressBookEntryByID retrieves the store.AddressBookEntry with the given id.
	AddressBookEntryByID(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) (store.AddressBookEntry, error)
}

// Notifier for Controller.
type Notifier interface {
	NotifyIntelCreated(ctx context.Context, tx pgx.Tx, created store.Intel) error
	NotifyIntelInvalidated(ctx context.Context, tx pgx.Tx, intelID uuid.UUID, by uuid.UUID) error
}
