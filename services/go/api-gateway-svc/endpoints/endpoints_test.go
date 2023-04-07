package endpoints

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const timeout = 5 * time.Second

// StoreMock mocks Store.
type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) Login(ctx context.Context, username string, pass string, requestMetadata controller.AuthRequestMetadata) (uuid.UUID, string, bool, error) {
	args := m.Called(ctx, username, pass, requestMetadata)
	return args.Get(0).(uuid.UUID), args.String(1), args.Bool(2), args.Error(3)
}

func (m *StoreMock) Logout(ctx context.Context, publicToken string, requestMetadata controller.AuthRequestMetadata) error {
	return m.Called(ctx, publicToken, requestMetadata).Error(0)
}

func (m *StoreMock) Proxy(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func Test_populateAPIV1Routes(t *testing.T) {
	r := testutil.NewGinEngine()
	ctrl := &controller.Controller{}
	logger := zap.NewNop()
	forwardAddr := "sun"
	populateAPIV1Routes(r, logger, ctrl, forwardAddr)
	assert.ElementsMatch(t, []testutil.RouteInfo{
		{
			Method: http.MethodPost,
			Path:   "/login",
		},
		{
			Method: http.MethodPost,
			Path:   "/logout",
		},
	}, testutil.RouteInfoFromGin(r.Routes()))
}

func TestCORS(t *testing.T) {
	s := httptest.NewServer(testutil.NewGinEngine())
	listenAddr := s.Listener.Addr().String()
	serverURL := s.URL
	s.Close()

	timeout, cancel, wait := testutil.NewTimeout(testutil.TestFailerFromT(t), timeout)
	defer cancel()
	runCtx, cancelRun := context.WithCancel(timeout)
	defer cancelRun()

	go func() {
		defer cancel()
		ctrl := &controller.Controller{}
		err := Serve(runCtx, zap.NewNop(), listenAddr, "", ctrl)
		require.NoError(t, err, "serving should not fail")
	}()

	c := &http.Client{}
tryRequests:
	for {
		req, err := http.NewRequestWithContext(timeout, http.MethodGet, serverURL, nil)
		require.NoError(t, err, "building request should not fail")
		req.Header.Set("Origin", "tQgK9U5")
		result, err := c.Do(req)
		if err == nil {
			assert.Equal(t, "*", result.Header.Get("Access-Control-Allow-Origin"),
				"should set origin header correctly")
			assert.Equal(t, "true", result.Header.Get("Access-Control-Allow-Credentials"),
				"should set credentials header correctly")
			cancelRun()
			break tryRequests
		}
		select {
		case <-timeout.Done():
			t.Error("timeout during http request cooldown")
			break tryRequests
		case <-time.After(100 * time.Millisecond):
		}
	}

	wait()
}
