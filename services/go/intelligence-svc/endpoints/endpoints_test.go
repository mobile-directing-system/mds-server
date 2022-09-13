package endpoints

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/stretchr/testify/mock"
)

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) CreateIntel(ctx context.Context, create store.CreateIntel) (store.Intel, error) {
	args := m.Called(ctx, create)
	return args.Get(0).(store.Intel), args.Error(1)
}

func (m *StoreMock) IntelByID(ctx context.Context, intelID uuid.UUID, limitToAssignedUser uuid.NullUUID) (store.Intel, error) {
	args := m.Called(ctx, intelID, limitToAssignedUser)
	return args.Get(0).(store.Intel), args.Error(1)
}

func (m *StoreMock) InvalidateIntelByID(ctx context.Context, intelID uuid.UUID, by uuid.UUID) error {
	return m.Called(ctx, intelID, by).Error(0)
}
