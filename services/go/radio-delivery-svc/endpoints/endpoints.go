package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"go.uber.org/zap"
)

// Store are the handle dependencies.
type Store interface {
	handleGetNextRadioDeliveryStore
	handleFinishRadioDeliveryStore
	handleReleasePickedUpRadioDeliveryStore
}

// Serve the endpoints via HTTP.
func Serve(lifetime context.Context, logger *zap.Logger, addr string, authSecret string, s Store, wsHub wsutil.Hub) error {
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	r := httpendpoints.NewEngine(logger)
	populateRoutes(r, logger, authSecret, s, wsHub)
	err := httpendpoints.Serve(lifetime, r, addr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": addr})
	}
	return nil
}

func populateRoutes(r *gin.Engine, logger *zap.Logger, secret string, s Store, wsHub wsutil.Hub) {
	r.GET("/ws", httpendpoints.GinHandlerFunc(logger, secret, wsHub.UpgradeHandler()))
	r.GET("/operations/:operationID/next", httpendpoints.GinHandlerFunc(logger, secret, handleGetNextRadioDelivery(s)))
	r.POST("/:attemptID/release", httpendpoints.GinHandlerFunc(logger, secret, handleReleasePickedUpRadioDelivery(s)))
	r.POST("/:attemptID/finish", httpendpoints.GinHandlerFunc(logger, secret, handleFinishRadioDelivery(s)))
}
