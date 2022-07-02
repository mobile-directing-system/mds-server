package endpoints

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/stretchr/testify/mock"
)

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) OperationByID(ctx context.Context, operationID uuid.UUID) (store.Operation, error) {
	args := m.Called(ctx, operationID)
	return args.Get(0).(store.Operation), args.Error(1)
}

func (m *StoreMock) Operations(ctx context.Context, params pagination.Params) (pagination.Paginated[store.Operation], error) {
	args := m.Called(ctx, params)
	return args.Get(0).(pagination.Paginated[store.Operation]), args.Error(1)
}

func (m *StoreMock) CreateOperation(ctx context.Context, operation store.Operation) (store.Operation, error) {
	args := m.Called(ctx, operation)
	return args.Get(0).(store.Operation), args.Error(1)
}

func (m *StoreMock) UpdateOperation(ctx context.Context, operation store.Operation) error {
	return m.Called(ctx, operation).Error(0)
}
