package ready

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

// DefaultServeAddr to use in Server.Serve.
const DefaultServeAddr = ":31234"

// StartUpCompleteFn is returned by NewServer and needs to be called in order to
// signal, that start-up has completed.
type StartUpCompleteFn func(isReadyFn CheckFn)

// Server allows providing health-information for Kubernetes. It exposes
// endpoints /livez and /readyz for liveness and readiness probes. Create one
// using NewServer and serve using Serve.
type Server struct {
	logger *zap.Logger
	r      *gin.Engine
	// isReadyFn checks if we are currently ready. If not set, we assume start-up
	// not having completed, yet.
	isReadyFn CheckFn
	// isReadyFnMutex locks isReadyFn.
	isReadyFnMutex sync.RWMutex
}

// NewServer creates a Server and returns a StartUpCompleteFn that needs to be
// called after start-up is complete. When the function is called, readiness
// probes will call the given CheckFn for checking ready-state. Serve the
// returned Server using Server.Serve.
func NewServer(logger *zap.Logger) (*Server, StartUpCompleteFn) {
	gin.SetMode(gin.ReleaseMode)
	s := &Server{
		logger: logger,
		r:      gin.New(),
	}
	s.r.GET("/livez", s.handleLivenessProbe())
	s.r.GET("/readyz", s.handleReadinessProbe())
	startUpCompleteFn := func(isReadyFn CheckFn) {
		s.isReadyFnMutex.Lock()
		defer s.isReadyFnMutex.Unlock()
		s.isReadyFn = isReadyFn
	}
	return s, startUpCompleteFn
}

// Serve an HTTP server for probes.
func (s *Server) Serve(lifetime context.Context, addr string) error {
	err := httpendpoints.Serve(lifetime, s.r, addr)
	if err != nil {
		return meh.Wrap(err, "serve http", meh.Details{"addr": addr})
	}
	return nil
}

// handleLivenessProbe simply return http.StatusOK.
func (s *Server) handleLivenessProbe() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
}

// handleReadinessProbe runs the ready-check if start-up has completed and
// reports the status.
func (s *Server) handleReadinessProbe() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If start-up not complete, we can skip the ready-check.
		s.isReadyFnMutex.RLock()
		defer s.isReadyFnMutex.RUnlock()
		if s.isReadyFn == nil {
			s.logger.Debug("skipping is-ready-check because of start-up not being completed")
			c.Status(http.StatusServiceUnavailable)
			return
		}
		err := s.isReadyFn(c.Request.Context())
		if err != nil {
			mehlog.LogToLevel(s.logger, zap.DebugLevel, meh.Wrap(err, "call is-ready-fn", nil))
			c.Status(http.StatusServiceUnavailable)
			return
		}
		c.Status(http.StatusOK)
	}
}
