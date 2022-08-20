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
	handleGetOperationsStore
	handleGetOperationByIDStore
	handleCreateOperationStore
	handleUpdateOperationStore
	handleGetOperationMembersByOperationStore
	handleUpdateOperationMembersByOperationStore
	handleSearchOperationsStore
	handleRebuildOperationSearchStore
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
	r.GET("/", httpendpoints.GinHandlerFunc(logger, secret, handleGetOperations(s)))
	r.GET("/search", httpendpoints.GinHandlerFunc(logger, secret, handleSearchOperations(s)))
	r.POST("/search/rebuild", httpendpoints.GinHandlerFunc(logger, secret, handleRebuildOperationSearch(s)))
	r.GET("/:operationID/members", httpendpoints.GinHandlerFunc(logger, secret, handleGetOperationMembersByOperation(s)))
	r.PUT("/:operationID/members", httpendpoints.GinHandlerFunc(logger, secret, handleUpdateOperationMembersByOperation(s)))
	r.GET("/:operationID", httpendpoints.GinHandlerFunc(logger, secret, handleGetOperationByID(s)))
	r.POST("/", httpendpoints.GinHandlerFunc(logger, secret, handleCreateOperation(s)))
	r.PUT("/:operationID", httpendpoints.GinHandlerFunc(logger, secret, handleUpdateOperation(s)))
}
