package endpoints

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Store holds all dependencise for handlers.
type Store interface {
	handleLoginStore
	handleLogoutStore
	handleProxyController
}

// Serve endpoints over HTTP.
func Serve(lifetime context.Context, logger *zap.Logger, serveAddr string, forwardAddr string, ctrl *controller.Controller) error {
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	router := httpendpoints.NewEngine(logger)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	populateAPIV1Routes(router, logger.Named("api-v1"), ctrl, forwardAddr)
	err := httpendpoints.Serve(lifetime, router, serveAddr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": serveAddr})
	}
	return nil
}

func populateAPIV1Routes(router *gin.Engine, logger *zap.Logger, s Store, forwardAddr string) {
	router.POST("/login", handleLogin(logger, s))
	router.POST("/logout", handleLogout(logger, s))
	router.NoRoute(handleProxy(logger, s, forwardAddr))
}

// ServeInternal endpoints over HTTP.
func ServeInternal(lifetime context.Context, logger *zap.Logger, serveAddr string, ctrl *controller.Controller) error {
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	router := httpendpoints.NewEngine(logger)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	populateInternalAPIV1Routes(router, logger.Named("api-v1"), ctrl)
	err := httpendpoints.Serve(lifetime, router, serveAddr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": serveAddr})
	}
	return nil
}

func populateInternalAPIV1Routes(router *gin.Engine, logger *zap.Logger, s Store) {
	router.POST("/tokens/resolve-public", handleResolvePublicToken(logger, s))
}
