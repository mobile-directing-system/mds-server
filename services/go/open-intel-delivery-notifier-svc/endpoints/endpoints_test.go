package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wstest"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

type handleWSSuite struct {
	suite.Suite
	r   *gin.Engine
	hub *wstest.HubMock
}

func (suite *handleWSSuite) SetupTest() {
	suite.hub = &wstest.HubMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.hub)
}

func (suite *handleWSSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/ws",
		Body:   nil,
		Token:  auth.Token{},
		Secret: "band",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct status code")
}

func (suite *handleWSSuite) TestUpgradeFail() {
	suite.hub.UpgradeFail = true
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/ws",
		Body:   nil,
		Token:  auth.Token{},
		Secret: "",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct status code")
	suite.Equal(1, suite.hub.UpgradeCalled, "should have called upgrade")
}

func (suite *handleWSSuite) TestOK() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/ws",
		Body:   nil,
		Token:  auth.Token{},
		Secret: "",
	})
	suite.Equal(http.StatusOK, rr.Code, "should return correct status code")
	suite.Equal(1, suite.hub.UpgradeCalled, "should have called upgrade")
}

func Test_handleWS(t *testing.T) {
	suite.Run(t, new(handleWSSuite))
}
