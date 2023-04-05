package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

// Test_publicIntelTypeFromStore reads all constants of store.IntelType and
// assures that the mapper does not fail.
func Test_publicIntelDeliveryStatusFromStore(t *testing.T) {
	testutil.TestMapperWithConstExtraction(t, publicIntelDeliveryStatusFromStore, "../store/intel_delivery.go", nulls.String{})
}

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

// handleCreateIntelDeliveryAttemptForDeliverySuite tests
// handleCreateIntelDeliveryAttemptForDelivery.
type handleCreateIntelDeliveryAttemptForDeliverySuite struct {
	suite.Suite
	s                *StoreMock
	r                *gin.Engine
	tokenOK          auth.Token
	sampleDeliveryID uuid.UUID
	sampleChannelID  uuid.UUID
	sCreated         store.IntelDeliveryAttempt
	pCreated         publicIntelDeliveryAttempt
}

func (suite *handleCreateIntelDeliveryAttemptForDeliverySuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "memory",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions: []permission.Permission{
			{Name: permission.ManageIntelDeliveryPermissionName},
		},
	}
	suite.sampleDeliveryID = testutil.NewUUIDV4()
	suite.sampleChannelID = testutil.NewUUIDV4()
	suite.sCreated = store.IntelDeliveryAttempt{
		ID:        uuid.UUID{},
		Delivery:  suite.sampleDeliveryID,
		Channel:   suite.sampleChannelID,
		CreatedAt: testutil.NewRandomTime(),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusOpen,
		StatusTS:  testutil.NewRandomTime(),
		Note:      nulls.NewString("saucer"),
	}
	var err error
	suite.pCreated, err = publicIntelDeliveryAttemptFromStore(suite.sCreated)
	suite.Require().NoError(err, "converting created attempt to public should not fail")
}

func (suite *handleCreateIntelDeliveryAttemptForDeliverySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL: fmt.Sprintf("/intel-deliveries/%s/deliver/channel/%s",
			suite.sampleDeliveryID.String(), suite.sampleChannelID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelDeliveryAttemptForDeliverySuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL: fmt.Sprintf("/intel-deliveries/%s/deliver/channel/%s",
			suite.sampleDeliveryID.String(), suite.sampleChannelID.String()),
		Token: token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelDeliveryAttemptForDeliverySuite) TestMissingPermissions() {
	token := suite.tokenOK
	token.Permissions = nil

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL: fmt.Sprintf("/intel-deliveries/%s/deliver/channel/%s",
			suite.sampleDeliveryID.String(), suite.sampleChannelID.String()),
		Token: token,
	})

	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelDeliveryAttemptForDeliverySuite) TestInvalidDeliveryID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel-deliveries/abc/deliver/channel/%s", suite.sampleChannelID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelDeliveryAttemptForDeliverySuite) TestInvalidChannelID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/intel-deliveries/%s/deliver/channel/abc", suite.sampleDeliveryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelDeliveryAttemptForDeliverySuite) TestCreateFail() {
	suite.s.On("CreateIntelDeliveryAttempt", mock.Anything, suite.sampleDeliveryID, suite.sampleChannelID).
		Return(store.IntelDeliveryAttempt{}, errors.New("sad life")).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL: fmt.Sprintf("/intel-deliveries/%s/deliver/channel/%s",
			suite.sampleDeliveryID.String(), suite.sampleChannelID.String()),
		Token: suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateIntelDeliveryAttemptForDeliverySuite) TestOK() {
	suite.s.On("CreateIntelDeliveryAttempt", mock.Anything, suite.sampleDeliveryID, suite.sampleChannelID).
		Return(suite.sCreated, nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL: fmt.Sprintf("/intel-deliveries/%s/deliver/channel/%s",
			suite.sampleDeliveryID.String(), suite.sampleChannelID.String()),
		Token: suite.tokenOK,
	})

	suite.Require().Equal(http.StatusCreated, rr.Code, "should return correct code")
	var got publicIntelDeliveryAttempt
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.pCreated, got, "should return correct body")
}

func Test_handleCreateIntelDeliveryAttemptForDelivery(t *testing.T) {
	suite.Run(t, new(handleCreateIntelDeliveryAttemptForDeliverySuite))
}

// handleGetIntelDeliveryAttemptsByDeliverySuite tests
// handleGetIntelDeliveryAttemptsByDelivery.
type handleGetIntelDeliveryAttemptsByDeliverySuite struct {
	suite.Suite
	s                *StoreMock
	r                *gin.Engine
	tokenOK          auth.Token
	sampleDeliveryID uuid.UUID
	sAttempts        []store.IntelDeliveryAttempt
	pAttempts        []publicIntelDeliveryAttempt
}

func (suite *handleGetIntelDeliveryAttemptsByDeliverySuite) genAttempt(deliveryID uuid.UUID) store.IntelDeliveryAttempt {
	return store.IntelDeliveryAttempt{
		ID:        testutil.NewUUIDV4(),
		Delivery:  deliveryID,
		Channel:   testutil.NewUUIDV4(),
		CreatedAt: testutil.NewRandomTime(),
		IsActive:  true,
		Status:    store.IntelDeliveryStatusAwaitingAck,
		StatusTS:  testutil.NewRandomTime(),
		Note:      nulls.NewString("veil"),
	}
}

func (suite *handleGetIntelDeliveryAttemptsByDeliverySuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "memory",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions: []permission.Permission{
			{Name: permission.ManageIntelDeliveryPermissionName},
		},
	}
	suite.sampleDeliveryID = testutil.NewUUIDV4()
	suite.sAttempts = []store.IntelDeliveryAttempt{
		suite.genAttempt(suite.sampleDeliveryID),
		suite.genAttempt(suite.sampleDeliveryID),
		suite.genAttempt(suite.sampleDeliveryID),
	}
	suite.pAttempts = make([]publicIntelDeliveryAttempt, 0, len(suite.sAttempts))
	for _, sAttempt := range suite.sAttempts {
		pAttempt, err := publicIntelDeliveryAttemptFromStore(sAttempt)
		suite.Require().NoError(err, "converting created attempt to public should not fail")
		suite.pAttempts = append(suite.pAttempts, pAttempt)
	}
}

func (suite *handleGetIntelDeliveryAttemptsByDeliverySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel-deliveries/%s/attempts", suite.sampleDeliveryID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsByDeliverySuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel-deliveries/%s/attempts", suite.sampleDeliveryID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsByDeliverySuite) TestMissingPermissions() {
	token := suite.tokenOK
	token.Permissions = nil

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel-deliveries/%s/attempts", suite.sampleDeliveryID.String()),
		Token:  token,
	})

	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsByDeliverySuite) TestInvalidDeliveryID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/intel-deliveries/abc/attempts",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsByDeliverySuite) TestRetrieveFail() {
	suite.s.On("IntelDeliveryAttemptsByDelivery", mock.Anything, suite.sampleDeliveryID).
		Return(nil, errors.New("sad life")).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel-deliveries/%s/attempts", suite.sampleDeliveryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetIntelDeliveryAttemptsByDeliverySuite) TestOK() {
	suite.s.On("IntelDeliveryAttemptsByDelivery", mock.Anything, suite.sampleDeliveryID).
		Return(suite.sAttempts, nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/intel-deliveries/%s/attempts", suite.sampleDeliveryID.String()),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got []publicIntelDeliveryAttempt
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.pAttempts, got, "should return correct body")
}

func Test_handleGetIntelDeliveryAttemptsByDelivery(t *testing.T) {
	suite.Run(t, new(handleGetIntelDeliveryAttemptsByDeliverySuite))
}
