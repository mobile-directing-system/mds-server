package endpoints

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/stretchr/testify/mock"
)

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) Groups(ctx context.Context, filters store.GroupFilters, params pagination.Params) (pagination.Paginated[store.Group], error) {
	args := m.Called(ctx, filters, params)
	return args.Get(0).(pagination.Paginated[store.Group]), args.Error(1)
}

func (m *StoreMock) GroupByID(ctx context.Context, groupID uuid.UUID) (store.Group, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).(store.Group), args.Error(1)
}

func (m *StoreMock) CreateGroup(ctx context.Context, create store.Group) (store.Group, error) {
	args := m.Called(ctx, create)
	return args.Get(0).(store.Group), args.Error(1)
}

func (m *StoreMock) UpdateGroup(ctx context.Context, update store.Group) error {
	return m.Called(ctx, update).Error(0)
}

func (m *StoreMock) DeleteGroupByID(ctx context.Context, groupID uuid.UUID) error {
	return m.Called(ctx, groupID).Error(0)
}
