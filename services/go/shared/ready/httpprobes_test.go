package ready

import (
	"context"
	"errors"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

// ServerSuite tests Server.
type ServerSuite struct {
	suite.Suite
}

func (suite *ServerSuite) TestLivenessStartUpIncomplete() {
	s, _ := NewServer(zap.NewNop())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: s.r,
		Method: http.MethodGet,
		URL:    "/livez",
		Body:   nil,
		Token:  auth.Token{},
		Secret: "",
	})
	suite.Equal(rr.Code, http.StatusOK, "should return correct code")
}

func (suite *ServerSuite) TestLivenessStartUpComplete() {
	s, completed := NewServer(zap.NewNop())
	completed(func(_ context.Context) error {
		suite.T().Errorf("ready-call not expected")
		return errors.New("sad life")
	})
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: s.r,
		Method: http.MethodGet,
		URL:    "/livez",
		Body:   nil,
		Token:  auth.Token{},
		Secret: "",
	})
	suite.Equal(rr.Code, http.StatusOK, "should return correct code")
}

func (suite *ServerSuite) TestReadinessStartUpIncomplete() {
	s, _ := NewServer(zap.NewNop())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: s.r,
		Method: http.MethodGet,
		URL:    "/readyz",
		Body:   nil,
		Token:  auth.Token{},
		Secret: "",
	})
	suite.Equal(rr.Code, http.StatusServiceUnavailable, "should return correct code")
}

func (suite *ServerSuite) TestReadinessFail() {
	s, completed := NewServer(zap.NewNop())
	completed(func(_ context.Context) error {
		return errors.New("sad life")
	})
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: s.r,
		Method: http.MethodGet,
		URL:    "/readyz",
		Body:   nil,
		Token:  auth.Token{},
		Secret: "",
	})
	suite.Equal(rr.Code, http.StatusServiceUnavailable, "should return correct code")
}

func (suite *ServerSuite) TestReadinessReady() {
	s, completed := NewServer(zap.NewNop())
	completed(func(_ context.Context) error {
		return nil
	})
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: s.r,
		Method: http.MethodGet,
		URL:    "/readyz",
		Body:   nil,
		Token:  auth.Token{},
		Secret: "",
	})
	suite.Equal(rr.Code, http.StatusOK, "should return correct code")
}

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}
