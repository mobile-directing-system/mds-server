package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/controller"
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
	r.GET("/", httpendpoints.GinHandlerFunc(logger, secret, handleGetUsers(ctrl)))
	r.POST("/", httpendpoints.GinHandlerFunc(logger, secret, handleCreateUser(ctrl)))
	r.GET("/search", httpendpoints.GinHandlerFunc(logger, secret, handleSearchUsers(ctrl)))
	r.POST("/search/rebuild", httpendpoints.GinHandlerFunc(logger, secret, handleRebuildUserSearch(ctrl)))
	r.GET("/:userID", httpendpoints.GinHandlerFunc(logger, secret, handleGetUserByID(ctrl)))
	r.PUT("/:userID", httpendpoints.GinHandlerFunc(logger, secret, handleUpdateUserByID(ctrl)))
	r.PUT("/:userID/pass", httpendpoints.GinHandlerFunc(logger, secret, handleUpdateUserPassByUserID(ctrl)))
	r.DELETE("/:userID", httpendpoints.GinHandlerFunc(logger, secret, handleDeleteUserByID(ctrl)))
}
