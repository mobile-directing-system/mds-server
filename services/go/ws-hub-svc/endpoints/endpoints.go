package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehhttp"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/ws-hub-svc/ws"
	"go.uber.org/zap"
	"net/http"
)

// Serve via HTTP.
func Serve(lifetime context.Context, logger *zap.Logger, addr string, wsHub ws.Hub) error {
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	r := httpendpoints.NewEngine(logger)
	populateRoutes(r, logger, wsHub)
	err := httpendpoints.Serve(lifetime, r, addr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": addr})
	}
	return nil
}

func populateRoutes(r *gin.Engine, logger *zap.Logger, wsHub ws.Hub) {
	r.GET("/ws/:gate", func(c *gin.Context) {
		gate := c.Param("gate")
		err := wsHub.Serve(c.Writer, c.Request, gate, c.Request.Header)
		if err != nil {
			mehhttp.LogAndRespondError(logger, c.Writer, c.Request, meh.Wrap(err, "serve with websocket hub", meh.Details{"gate": gate}))
			return
		}
		c.Status(http.StatusOK)
	})
}
