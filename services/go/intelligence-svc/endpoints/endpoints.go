package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"go.uber.org/zap"
)

// Store provides all dependencies for the handlers.
type Store interface {
	handleCreateIntelStore
	handleGetIntelByIDStore
	handleInvalidateIntelByIDStore
}

// Serve the endpoints via HTTP.
func Serve(lifetime context.Context, logger *zap.Logger, addr string, authSecret string, s Store) error {
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	r := httpendpoints.NewEngine(logger)
	populateRoutes(r, logger, authSecret, s)
	err := httpendpoints.Serve(lifetime, r, addr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": addr})
	}
	return nil
}

func populateRoutes(r *gin.Engine, logger *zap.Logger, secret string, s Store) {
	r.POST("/intel", httpendpoints.GinHandlerFunc(logger, secret, handleCreateIntel(s)))
	r.GET("/intel/:intelID", httpendpoints.GinHandlerFunc(logger, secret, handleGetIntelByID(s)))
	r.POST("/intel/:intelID/invalidate", httpendpoints.GinHandlerFunc(logger, secret, handleInvalidateIntelByID(s)))
}
