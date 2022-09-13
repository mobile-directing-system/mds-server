package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"go.uber.org/zap"
)

// Serve the endpoints via HTTP.
func Serve(lifetime context.Context, logger *zap.Logger, addr string, authSecret string, wsHub wsutil.Hub) error {
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	r := httpendpoints.NewEngine(logger)
	populateRoutes(r, logger, authSecret, wsHub)
	err := httpendpoints.Serve(lifetime, r, addr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": addr})
	}
	return nil
}

func populateRoutes(r *gin.Engine, logger *zap.Logger, secret string, wsHub wsutil.Hub) {
	r.GET("/ws", httpendpoints.GinHandlerFunc(logger, secret, wsHub.UpgradeHandler()))
}
