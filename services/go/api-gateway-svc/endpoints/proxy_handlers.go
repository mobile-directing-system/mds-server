package endpoints

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"net/http/httputil"
)

func handleProxy(logger *zap.Logger, forwardAddr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("proxy call lolololololololololololol")
		logger.Info("forwarding to " + forwardAddr)

		director := func(req *http.Request) {
			r := c.Request

			req.URL.Scheme = "http"
			req.URL.Host = forwardAddr
			// req.URL.Path = r.URL.Path
			req.URL.Path = "/users"
			req.Header["my-header"] = []string{r.Header.Get("my-header")}
			// Golang camelcases headers
			delete(req.Header, "My-Header")
			fmt.Printf("was %s and now forwarding to url%s\n", c.Request.URL.String(), req.URL.String())
		}
		proxy := &httputil.ReverseProxy{Director: director,
			ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
				fmt.Println("THE ERROR: ", err.Error())
			}}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
