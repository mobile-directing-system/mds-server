package endpoints

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

// handleMarkIntelDeliveryAsDeliveredSuite tests
// handleMarkIntelDeliveryAsDelivered.
type handleMarkIntelDeliveryAsDeliveredSuite struct {
	suite.Suite
	s                *StoreMock
	r                *gin.Engine
	tokenOK          auth.Token
	sampleDeliveryID uuid.UUID
}

func (suite *handleMarkIntelDeliveryAsDeliveredSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "realize",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     nil,
	}
	suite.sampleDeliveryID = testutil.NewUUIDV4()
}

func (suite *handleMarkIntelDeliveryAsDeliveredSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel-deliveries/%s/delivered", suite.sampleDeliveryID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleMarkIntelDeliveryAsDeliveredSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel-deliveries/%s/delivered", suite.sampleDeliveryID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleMarkIntelDeliveryAsDeliveredSuite) TestInvalidDeliveryID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/intel-deliveries/abc/delivered",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleMarkIntelDeliveryAsDeliveredSuite) TestMarkFail() {
	suite.s.On("MarkIntelDeliveryAsDelivered", mock.Anything, suite.sampleDeliveryID, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel-deliveries/%s/delivered", suite.sampleDeliveryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleMarkIntelDeliveryAsDeliveredSuite) TestOK() {
	suite.s.On("MarkIntelDeliveryAsDelivered", mock.Anything, suite.sampleDeliveryID, nulls.NewUUID(suite.tokenOK.UserID)).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel-deliveries/%s/delivered", suite.sampleDeliveryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleMarkIntelDeliveryAsDelivered(t *testing.T) {
	suite.Run(t, new(handleMarkIntelDeliveryAsDeliveredSuite))
}
