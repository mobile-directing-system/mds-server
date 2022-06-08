package httpendpoints

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehhttp"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

// NewEngine returns a gin.Engine with preconfigured request-debug-logger.
func NewEngine(logger *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(requestDebugLogger(logger.Named("request")))
	return r
}

// requestDebugLogger logs requests on zap.DebugLevel to the given zap.Logger.
// The idea is based on gin.Logger.
func requestDebugLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// Process request.
		c.Next()
		// Log results.
		logger.Debug("request",
			zap.Time("timestamp", start),
			zap.Duration("took", time.Now().Sub(start)),
			zap.String("path", c.Request.URL.Path),
			zap.String("raw_query", c.Request.URL.RawQuery),
			zap.String("client_ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.Int("status_code", c.Writer.Status()),
			zap.String("error_message", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Int("body_size", c.Writer.Size()),
			zap.String("user_agent", c.Request.UserAgent()))
	}
}

// Serve the given gin.Engine until the context is done.
func Serve(lifetime context.Context, engine *gin.Engine, addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}
	go func() {
		<-lifetime.Done()
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := srv.Shutdown(timeout)
		if err != nil {
			log.Fatalf("shutdown http server: %s\n", err.Error())
		}
	}()
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return meh.NewInternalErrFromErr(err, "listen and serve", meh.Details{"addr": addr})
	}
	return nil
}

// ApplyDefaultErrorHTTPMapping applies a default mapping of meh.Code to HTTP
// status codes.
func ApplyDefaultErrorHTTPMapping() {
	mehhttp.SetHTTPStatusCodeMapping(func(code meh.Code) int {
		switch code {
		case meh.ErrBadInput:
			return http.StatusBadRequest
		case meh.ErrNotFound:
			return http.StatusNotFound
		case meh.ErrUnauthorized:
			return http.StatusUnauthorized
		default:
			return http.StatusInternalServerError
		}
	})
}
