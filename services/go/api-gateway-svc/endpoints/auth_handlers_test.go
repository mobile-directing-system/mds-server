package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"testing"
)

// handleLoginSuite tests handleLogin.
type handleLoginSuite struct {
	suite.Suite
	s             *StoreMock
	r             *gin.Engine
	sampleRequest loginPayload
}

func (suite *handleLoginSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateAPIV1Routes(suite.r, zap.NewNop(), suite.s, "")
	suite.sampleRequest = loginPayload{
		Username: "sweat",
		Pass:     "bind",
	}
}

func (suite *handleLoginSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/login",
		Body:   strings.NewReader("{invalid"),
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleLoginSuite) TestLoginFail() {
	suite.s.On("Login", mock.Anything, suite.sampleRequest.Username, suite.sampleRequest.Pass, mock.Anything).
		Return(uuid.Nil, "", false, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/login",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleRequest)),
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleLoginSuite) TestBadLogin() {
	suite.s.On("Login", mock.Anything, suite.sampleRequest.Username, suite.sampleRequest.Pass, mock.Anything).
		Return(uuid.Nil, "", false, nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/login",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleRequest)),
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleLoginSuite) TestOK() {
	userID := testutil.NewUUIDV4()
	token := "feed"
	suite.s.On("Login", mock.Anything, suite.sampleRequest.Username, suite.sampleRequest.Pass, mock.Anything).
		Return(userID, token, true, nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/login",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleRequest)),
	})
	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got loginResponse
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid response")
	suite.Equal(loginResponse{
		UserID:      userID,
		AccessToken: token,
		TokenType:   "Bearer",
	}, got, "should return correct response")
}

func Test_handleLogin(t *testing.T) {
	suite.Run(t, new(handleLoginSuite))
}

func Test_extractAuthRequestMetadataFromRequest(t *testing.T) {
	host := "sheet"
	userAgent := "yes"
	remoteAddr := "ocean"
	req := &http.Request{
		Host:       host,
		RemoteAddr: remoteAddr,
		Header: http.Header{
			"User-Agent": []string{userAgent},
		},
	}
	got := extractAuthRequestMetadataFromRequest(req)
	assert.Equal(t, controller.AuthRequestMetadata{
		Host:       host,
		UserAgent:  userAgent,
		RemoteAddr: remoteAddr,
	}, got, "should extract correct metadata")
}

// handleLogoutSuite tests handleLogout.
type handleLogoutSuite struct {
	suite.Suite
	s              *StoreMock
	r              *gin.Engine
	sampleToken    auth.Token
	sampleTokenStr string
}

func (suite *handleLogoutSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateAPIV1Routes(suite.r, zap.NewNop(), suite.s, "")
	suite.sampleToken = auth.Token{
		UserID: testutil.NewUUIDV4(),
	}
	var err error
	suite.sampleTokenStr, err = auth.GenJWTToken(suite.sampleToken, "")
	if err != nil {
		panic(err)
	}
}

func (suite *handleLogoutSuite) TestLogoutFail() {
	suite.s.On("Logout", mock.Anything, suite.sampleTokenStr, mock.Anything).Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/logout",
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleLogoutSuite) TestOK() {
	suite.s.On("Logout", mock.Anything, suite.sampleTokenStr, mock.Anything).Return(nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/logout",
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleLogout(t *testing.T) {
	suite.Run(t, new(handleLogoutSuite))
}

// handleResolvePublicTokenSuite tests handleResolvePublicToken.
type handleResolvePublicTokenSuite struct {
	suite.Suite
	r                   *gin.Engine
	s                   *StoreMock
	samplePublicToken   string
	sampleResolvedToken string
}

func (suite *handleResolvePublicTokenSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateInternalAPIV1Routes(suite.r, zap.NewNop(), suite.s)
	suite.samplePublicToken = "servant"
	suite.sampleResolvedToken = "swear"
}

func (suite *handleResolvePublicTokenSuite) TestProxyFail() {
	suite.s.On("Proxy", mock.Anything, mock.Anything).
		Return("", errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/tokens/resolve-public",
		Body:   bytes.NewReader([]byte(suite.samplePublicToken)),
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct status code")
}

func (suite *handleResolvePublicTokenSuite) TestOK() {
	suite.s.On("Proxy", mock.Anything, mock.Anything).
		Return(suite.sampleResolvedToken, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/tokens/resolve-public",
		Body:   bytes.NewReader([]byte(suite.samplePublicToken)),
	})
	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct status code")
	got, err := io.ReadAll(rr.Body)
	suite.Require().NoError(err, "read body should not fail")
	suite.EqualValues(suite.sampleResolvedToken, got, "should return correct token")
}

func Test_handleResolvePublicToken(t *testing.T) {
	suite.Run(t, new(handleResolvePublicTokenSuite))
}
