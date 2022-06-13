package testutil

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"io"
	"net/http"
	"net/http/httptest"
)

// NewGinEngine creates a new gin.Engine for usage in unit tests.
func NewGinEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	return gin.New()
}

// HTTPRequestProps are the props used in DoHTTPRequestMust.
type HTTPRequestProps struct {
	// Server is the http.Handler that is used in order to serve the HTTP request.
	Server http.Handler
	// Method is the HTTP method to use.
	Method string
	// URL is the called URL.
	URL string
	// Body is the optional body. If not provided, set it to nil.
	Body io.Reader
	// Token is the auth.Token that will be added to the request in form of the
	// usual JWT token.
	Token auth.Token
	// Secret used for generating the JWT token from Token.
	Secret string
}

// DoHTTPRequestMust creates a new http.Request using the given properties and
// serves HTTP with the http.Handler. The response is recorded and returned
// using an httptest.ResponseRecorder.
func DoHTTPRequestMust(props HTTPRequestProps) *httptest.ResponseRecorder {
	req, err := http.NewRequest(props.Method, props.URL, props.Body)
	if err != nil {
		panic(meh.NewInternalErrFromErr(err, "new http request", nil).Error())
	}
	// Add token to request.
	jwtToken, err := auth.GenJWTToken(props.Token, props.Secret)
	if err != nil {
		panic(meh.Wrap(err, "gen jwt token", nil).Error())
	}
	req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", jwtToken)}
	// Do request.
	rr := httptest.NewRecorder()
	props.Server.ServeHTTP(rr, req)
	return rr
}

// MarshalJSONMust calls json.Marshal for the given data and panics in case of
// failure.
func MarshalJSONMust(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(meh.NewInternalErrFromErr(err, "marshal json", nil))
	}
	return b
}
