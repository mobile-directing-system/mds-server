package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/stretchr/testify/mock"
	"net/http"
)

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) PickUpNextRadioDelivery(ctx context.Context, operationID uuid.UUID,
	by uuid.UUID) (store.AcceptedIntelDeliveryAttempt, bool, error) {
	args := m.Called(ctx, operationID, by)
	return args.Get(0).(store.AcceptedIntelDeliveryAttempt), args.Bool(1), args.Error(2)
}

func (m *StoreMock) FinishRadioDelivery(ctx context.Context, attemptID uuid.UUID, success bool, note string, limitToPickedUpBy uuid.NullUUID) error {
	return m.Called(ctx, attemptID, success, note, limitToPickedUpBy).Error(0)
}

func (m *StoreMock) ReleasePickedUpRadioDelivery(ctx context.Context, attemptID uuid.UUID, limitToPickedUpBy uuid.NullUUID) error {
	return m.Called(ctx, attemptID, limitToPickedUpBy).Error(0)
}

type wsHubStub struct {
}

func (m *wsHubStub) UpgradeHandler() httpendpoints.HandlerFunc {
	return func(c *gin.Context, token auth.Token) error {
		c.Status(http.StatusTeapot)
		return nil
	}
}
