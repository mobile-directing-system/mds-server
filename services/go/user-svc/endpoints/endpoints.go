package endpoints

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehgin"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"go.uber.org/zap"
	"net/http"
)

// Serve the endpoints via HTTP.
func Serve(lifetime context.Context, logger *zap.Logger, addr string, authSecret string) error {
	r := httpendpoints.NewEngine(logger)
	// TODO: REMOVE
	r.NoRoute(func(c *gin.Context) {
		authToken, err := auth.ParseJWTTokenFromHeader(c, authSecret)
		if err != nil {
			mehgin.LogAndRespondError(logger, c, meh.Wrap(err, "paarse jwt token from header", nil))
			return
		}
		fmt.Printf("%+v\n", authToken)
		c.Status(http.StatusTeapot)
	})
	err := httpendpoints.Serve(lifetime, r, addr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": addr})
	}
	return nil
}
