package endpoints

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehgin"
	"github.com/lefinal/meh/mehlog"
	"go.uber.org/zap"
	"net/http"
	"net/http/httputil"
	"strings"
)

// handleProxyController are the dependencies needed for handleProxy.
type handleProxyController interface {
	// Proxy collects user information and retrieves it as signed JWT token as first
	// return value. This also includes, whether the user is authenticated or not.
	Proxy(ctx context.Context, token string) (string, error)
}

// handleProxy proxies requests.
func handleProxy(logger *zap.Logger, ctrl handleProxyController, forwardAddr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract public token from request.
		authHeader := c.GetHeader("Authorization")
		publicToken := strings.TrimPrefix(authHeader, "Bearer ")
		internalSignedToken, err := ctrl.Proxy(c.Request.Context(), publicToken)
		if err != nil {
			mehgin.LogAndRespondError(logger, c, meh.Wrap(err, "proxy", meh.Details{"public_token": publicToken}))
			return
		}
		director := func(req *http.Request) {
			r := c.Request
			req.URL.Scheme = "http"
			req.URL.Host = forwardAddr
			req.URL.Path = r.URL.Path
			req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", internalSignedToken)}
		}
		proxy := &httputil.ReverseProxy{Director: director,
			ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
				mehlog.Log(logger, meh.NewInternalErrFromErr(err, "reverse proxy", meh.Details{
					"url": c.Request.URL.String(),
				}))
			}}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
