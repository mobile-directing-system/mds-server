package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
)

// handleGetNextRadioDeliverySuite tests handleGetNextRadioDelivery.
type handleGetNextRadioDeliverySuite struct {
	suite.Suite
	s                        *StoreMock
	r                        *gin.Engine
	sampleToken              auth.Token
	sampleStoreNextDelivery  store.AcceptedIntelDeliveryAttempt
	samplePublicNextDelivery publicAcceptedIntelDeliveryAttempt
}

func (suite *handleGetNextRadioDeliverySuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s, &wsHubStub{})
	suite.sampleToken = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "organize",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.DeliverAnyRadioDeliveryPermissionName}},
	}
	suite.sampleStoreNextDelivery = store.AcceptedIntelDeliveryAttempt{
		ID:              testutil.NewUUIDV4(),
		Intel:           testutil.NewUUIDV4(),
		IntelOperation:  testutil.NewUUIDV4(),
		IntelImportance: 884,
		AssignedTo:      testutil.NewUUIDV4(),
		AssignedToLabel: "baggage",
		AssignedToUser:  nulls.NewUUID(testutil.NewUUIDV4()),
		Delivery:        testutil.NewUUIDV4(),
		Channel:         testutil.NewUUIDV4(),
		CreatedAt:       testutil.NewRandomTime(),
		IsActive:        true,
		StatusTS:        testutil.NewRandomTime(),
		Note:            nulls.NewString("front"),
		AcceptedAt:      testutil.NewRandomTime(),
	}
	suite.samplePublicNextDelivery = publicAcceptedIntelDeliveryAttemptFromStore(suite.sampleStoreNextDelivery)
}

func (suite *handleGetNextRadioDeliverySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/operations/%s/next", suite.sampleStoreNextDelivery.IntelOperation.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicNextDelivery)),
		Token:  suite.sampleToken,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetNextRadioDeliverySuite) TestNotAuthenticated() {
	suite.sampleToken.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/operations/%s/next", suite.sampleStoreNextDelivery.IntelOperation.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicNextDelivery)),
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetNextRadioDeliverySuite) TestInvalidOperationID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/operations/abc/next",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicNextDelivery)),
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetNextRadioDeliverySuite) TestMissingPermission() {
	suite.sampleToken.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/operations/%s/next", suite.sampleStoreNextDelivery.IntelOperation.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicNextDelivery)),
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleGetNextRadioDeliverySuite) TestPickUpNextFail() {
	suite.s.On("PickUpNextRadioDelivery", mock.Anything, suite.samplePublicNextDelivery.IntelOperation, suite.sampleToken.UserID).
		Return(store.AcceptedIntelDeliveryAttempt{}, false, errors.New("sad life")).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/operations/%s/next", suite.sampleStoreNextDelivery.IntelOperation.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicNextDelivery)),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetNextRadioDeliverySuite) TestNoNext() {
	suite.s.On("PickUpNextRadioDelivery", mock.Anything, suite.samplePublicNextDelivery.IntelOperation, suite.sampleToken.UserID).
		Return(store.AcceptedIntelDeliveryAttempt{}, false, nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/operations/%s/next", suite.sampleStoreNextDelivery.IntelOperation.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicNextDelivery)),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusNoContent, rr.Code, "should return correct code")
}

func (suite *handleGetNextRadioDeliverySuite) TestOK() {
	suite.s.On("PickUpNextRadioDelivery", mock.Anything, suite.samplePublicNextDelivery.IntelOperation, suite.sampleToken.UserID).
		Return(suite.sampleStoreNextDelivery, true, nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/operations/%s/next", suite.sampleStoreNextDelivery.IntelOperation.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicNextDelivery)),
		Token:  suite.sampleToken,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got publicAcceptedIntelDeliveryAttempt
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicNextDelivery, got, "should return correct body")
}

func Test_handleGetNextRadioDelivery(t *testing.T) {
	suite.Run(t, new(handleGetNextRadioDeliverySuite))
}

// handleReleasePickedUpRadioDeliverySuite tests
// handleReleasePickedUpRadioDelivery.
type handleReleasePickedUpRadioDeliverySuite struct {
	suite.Suite
	s               *StoreMock
	r               *gin.Engine
	sampleToken     auth.Token
	sampleAttemptID uuid.UUID
}

func (suite *handleReleasePickedUpRadioDeliverySuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s, &wsHubStub{})
	suite.sampleToken = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "organize",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.DeliverAnyRadioDeliveryPermissionName}},
	}
	suite.sampleAttemptID = testutil.NewUUIDV4()
}

func (suite *handleReleasePickedUpRadioDeliverySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/release", suite.sampleAttemptID),
		Token:  suite.sampleToken,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleReleasePickedUpRadioDeliverySuite) TestNotAuthenticated() {
	suite.sampleToken.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/release", suite.sampleAttemptID),
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleReleasePickedUpRadioDeliverySuite) TestInvalidAttemptID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/abc/release",
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleReleasePickedUpRadioDeliverySuite) TestMissingPermission() {
	suite.sampleToken.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/release", suite.sampleAttemptID),
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleReleasePickedUpRadioDeliverySuite) TestReleaseFail() {
	suite.s.On("ReleasePickedUpRadioDelivery", mock.Anything, suite.sampleAttemptID, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/release", suite.sampleAttemptID),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleReleasePickedUpRadioDeliverySuite) TestOK() {
	suite.s.On("ReleasePickedUpRadioDelivery", mock.Anything, suite.sampleAttemptID, mock.Anything).
		Return(nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/release", suite.sampleAttemptID),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleReleasePickedUpRadioDeliverySuite) TestOnlyDeliverPermission() {
	suite.sampleToken.Permissions = []permission.Permission{{Name: permission.DeliverAnyRadioDeliveryPermissionName}}
	suite.s.On("ReleasePickedUpRadioDelivery", mock.Anything, suite.sampleAttemptID, nulls.NewUUID(suite.sampleToken.UserID)).
		Return(nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/release", suite.sampleAttemptID),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleReleasePickedUpRadioDeliverySuite) TestOnlyManagePermission() {
	suite.sampleToken.Permissions = []permission.Permission{{Name: permission.ManageAnyRadioDeliveryPermissionName}}
	suite.s.On("ReleasePickedUpRadioDelivery", mock.Anything, suite.sampleAttemptID, uuid.NullUUID{}).
		Return(nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/release", suite.sampleAttemptID),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleReleasePickedUpRadioDeliverySuite) TestDeliverAndManagePermissions() {
	suite.sampleToken.Permissions = []permission.Permission{
		{Name: permission.DeliverAnyRadioDeliveryPermissionName},
		{Name: permission.ManageAnyRadioDeliveryPermissionName},
	}
	suite.s.On("ReleasePickedUpRadioDelivery", mock.Anything, suite.sampleAttemptID, uuid.NullUUID{}).
		Return(nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/release", suite.sampleAttemptID),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleReleasePickedUpRadioDelivery(t *testing.T) {
	suite.Run(t, new(handleReleasePickedUpRadioDeliverySuite))
}

// handleFinishRadioDeliverySuite tests
// handleFinishRadioDelivery.
type handleFinishRadioDeliverySuite struct {
	suite.Suite
	s                         *StoreMock
	r                         *gin.Engine
	sampleToken               auth.Token
	sampleAttemptID           uuid.UUID
	samplePublicFinishDetails publicFinishRadioDeliveryDetails
}

func (suite *handleFinishRadioDeliverySuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s, &wsHubStub{})
	suite.sampleToken = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "organize",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.DeliverAnyRadioDeliveryPermissionName}},
	}
	suite.sampleAttemptID = testutil.NewUUIDV4()
	suite.samplePublicFinishDetails = publicFinishRadioDeliveryDetails{
		Success: true,
		Note:    "seat",
	}
}

func (suite *handleFinishRadioDeliverySuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/finish", suite.sampleAttemptID),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicFinishDetails)),
		Token:  suite.sampleToken,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleFinishRadioDeliverySuite) TestNotAuthenticated() {
	suite.sampleToken.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/finish", suite.sampleAttemptID),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicFinishDetails)),
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleFinishRadioDeliverySuite) TestInvalidAttemptID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/abc/finish",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicFinishDetails)),
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleFinishRadioDeliverySuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/finish", suite.sampleAttemptID),
		Body:   strings.NewReader(`{invalid`),
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleFinishRadioDeliverySuite) TestMissingPermission() {
	suite.sampleToken.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/finish", suite.sampleAttemptID),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicFinishDetails)),
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleFinishRadioDeliverySuite) TestFinishFail() {
	suite.s.On("FinishRadioDelivery", mock.Anything, suite.sampleAttemptID, suite.samplePublicFinishDetails.Success,
		suite.samplePublicFinishDetails.Note, mock.Anything).
		Return(errors.New("sad life")).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/finish", suite.sampleAttemptID),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicFinishDetails)),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleFinishRadioDeliverySuite) TestOK() {
	suite.s.On("FinishRadioDelivery", mock.Anything, suite.sampleAttemptID, suite.samplePublicFinishDetails.Success,
		suite.samplePublicFinishDetails.Note, mock.Anything).
		Return(nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/finish", suite.sampleAttemptID),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicFinishDetails)),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleFinishRadioDeliverySuite) TestOnlyDeliverPermission() {
	suite.sampleToken.Permissions = []permission.Permission{{Name: permission.DeliverAnyRadioDeliveryPermissionName}}
	suite.s.On("FinishRadioDelivery", mock.Anything, suite.sampleAttemptID, suite.samplePublicFinishDetails.Success,
		suite.samplePublicFinishDetails.Note, nulls.NewUUID(suite.sampleToken.UserID)).
		Return(nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/finish", suite.sampleAttemptID),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicFinishDetails)),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleFinishRadioDeliverySuite) TestOnlyManagePermission() {
	suite.sampleToken.Permissions = []permission.Permission{{Name: permission.ManageAnyRadioDeliveryPermissionName}}
	suite.s.On("FinishRadioDelivery", mock.Anything, suite.sampleAttemptID, suite.samplePublicFinishDetails.Success,
		suite.samplePublicFinishDetails.Note, uuid.NullUUID{}).
		Return(nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/finish", suite.sampleAttemptID),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicFinishDetails)),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleFinishRadioDeliverySuite) TestDeliverAndManagePermissions() {
	suite.sampleToken.Permissions = []permission.Permission{
		{Name: permission.DeliverAnyRadioDeliveryPermissionName},
		{Name: permission.ManageAnyRadioDeliveryPermissionName},
	}
	suite.s.On("FinishRadioDelivery", mock.Anything, suite.sampleAttemptID, suite.samplePublicFinishDetails.Success,
		suite.samplePublicFinishDetails.Note, uuid.NullUUID{}).
		Return(nil).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/%s/finish", suite.sampleAttemptID),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicFinishDetails)),
		Token:  suite.sampleToken,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleFinishRadioDelivery(t *testing.T) {
	suite.Run(t, new(handleFinishRadioDeliverySuite))
}
