package testutil

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestRouteInfoFromGin(t *testing.T) {
	r := NewGinEngine()
	r.POST("/post", gin.Logger())
	r.GET("/get", gin.Logger())
	r.PUT("/put", gin.Logger())
	assert.ElementsMatch(t, []RouteInfo{
		{
			Method: http.MethodPost,
			Path:   "/post",
		},
		{
			Method: http.MethodGet,
			Path:   "/get",
		},
		{
			Method: http.MethodPut,
			Path:   "/put",
		},
	}, RouteInfoFromGin(r.Routes()))
}
