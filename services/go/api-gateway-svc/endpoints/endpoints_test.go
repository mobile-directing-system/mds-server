package endpoints

import (
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

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
