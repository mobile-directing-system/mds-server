package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"go.uber.org/zap"
)

// Serve the endpoints via HTTP.
func Serve(lifetime context.Context, logger *zap.Logger, addr string, authSecret string, ctrl *controller.Controller) error {
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	r := httpendpoints.NewEngine(logger)
	populateRoutes(r, logger, authSecret, ctrl)
	err := httpendpoints.Serve(lifetime, r, addr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": addr})
	}
	return nil
}

func populateRoutes(r *gin.Engine, logger *zap.Logger, secret string, ctrl *controller.Controller) {
	r.GET("/user/:userID", httpendpoints.GinHandlerFunc(logger, secret, handleGetPermissionsByUser(ctrl)))
	r.PUT("/user/:userID", httpendpoints.GinHandlerFunc(logger, secret, handleUpdatePermissionsByUser(ctrl)))
}
