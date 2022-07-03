package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"go.uber.org/zap"
)

// Store are the handle dependencies.
type Store interface {
	handleGetGroupsStore
	handleGetGroupByIDStore
	handleCreateGroupStore
	handleUpdateGroupStore
	handleDeleteGroupByIDStore
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
	r.GET("/", httpendpoints.GinHandlerFunc(logger, secret, handleGetGroups(s)))
	r.GET("/:groupID", httpendpoints.GinHandlerFunc(logger, secret, handleGetGroupByID(s)))
	r.POST("/", httpendpoints.GinHandlerFunc(logger, secret, handleCreateGroup(s)))
	r.PUT("/:groupID", httpendpoints.GinHandlerFunc(logger, secret, handleUpdateGroup(s)))
	r.DELETE("/:groupID", httpendpoints.GinHandlerFunc(logger, secret, handleDeleteGroupByID(s)))
}
