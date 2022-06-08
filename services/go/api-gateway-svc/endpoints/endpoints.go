package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"go.uber.org/zap"
)

// Serve endpoints over HTTP.
func Serve(lifetime context.Context, logger *zap.Logger, serveAddr string, forwardAddr string, ctrl *controller.Controller) error {
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	router := httpendpoints.NewEngine(logger)
	populateAPIV1Routes(router, logger.Named("api-v1"), ctrl, forwardAddr)
	err := httpendpoints.Serve(lifetime, router, serveAddr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": serveAddr})
	}
	return nil
}

func populateAPIV1Routes(router *gin.Engine, logger *zap.Logger, ctrl *controller.Controller, forwardAddr string) {
	router.POST("/login", handleLogin(logger, ctrl))
	router.POST("/logout", handleLogout(logger, ctrl))
	router.NoRoute(handleProxy(logger, ctrl, forwardAddr))
}
