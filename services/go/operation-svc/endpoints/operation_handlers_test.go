package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
	"time"
)

// handleGetOperationsSuite tests handleGetOperations.
type handleGetOperationsSuite struct {
	suite.Suite
	s                      *StoreMock
	r                      *gin.Engine
	tokenOK                auth.Token
	sampleParams           pagination.Params
	sampleOperations       pagination.Paginated[store.Operation]
	samplePublicOperations pagination.Paginated[publicOperation]
}

func (suite *handleGetOperationsSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          uuid.New(),
		Username:        "stop",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ViewAnyOperationPermissionName}},
	}
	suite.sampleParams = pagination.Params{
		Limit:          1,
		Offset:         2,
		OrderBy:        nulls.NewString("meow"),
		OrderDirection: pagination.OrderDirDesc,
	}
	suite.sampleOperations = pagination.NewPaginated(suite.sampleParams, []store.Operation{
		{
			ID:          uuid.New(),
			Title:       "roast",
			Description: "ashamed",
			Start:       time.Date(2022, 3, 4, 12, 0, 0, 0, time.UTC),
			End:         nulls.NewTime(time.Date(2022, 11, 1, 12, 0, 0, 0, time.UTC)),
			IsArchived:  true,
		},
		{
			ID:          uuid.New(),
			Title:       "west",
			Description: "",
			Start:       time.Date(2020, 1, 2, 4, 0, 12, 0, time.UTC),
			End:         nulls.Time{},
			IsArchived:  false,
		},
	}, 54)
	suite.samplePublicOperations = pagination.MapPaginated(suite.sampleOperations, publicOperationFromStore)
}

func (suite *handleGetOperationsSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetOperationsSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetOperationsSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleGetOperationsSuite) TestInvalidPaginationParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/?%s=abc", pagination.LimitQueryParam),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetOperationsSuite) TestStoreRetrievalFail() {
	suite.s.On("Operations", mock.Anything, suite.sampleParams).
		Return(pagination.Paginated[store.Operation]{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/?%s", pagination.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetOperationsSuite) TestOK() {
	suite.s.On("Operations", mock.Anything, suite.sampleParams).Return(suite.sampleOperations, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/?%s", pagination.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got pagination.Paginated[publicOperation]
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicOperations, got, "should return correct body")
}

func Test_handleGetOperations(t *testing.T) {
	suite.Run(t, new(handleGetOperationsSuite))
}

// handleGetOperationByIDSuite tests handleGetOperationByID.
type handleGetOperationByIDSuite struct {
	suite.Suite
	s                     *StoreMock
	r                     *gin.Engine
	tokenOK               auth.Token
	sampleOperationID     uuid.UUID
	sampleOperation       store.Operation
	samplePublicOperation publicOperation
}

func (suite *handleGetOperationByIDSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          uuid.New(),
		Username:        "civilize",
		IsAuthenticated: true,
		IsAdmin:         false,
	}
	suite.sampleOperationID = uuid.New()
	suite.sampleOperation = store.Operation{
		ID:          suite.sampleOperationID,
		Title:       "steer",
		Description: "money",
		Start:       time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
		End:         nulls.NewTime(time.Date(2021, 4, 3, 2, 0, 0, 0, time.UTC)),
		IsArchived:  true,
	}
	suite.samplePublicOperation = publicOperationFromStore(suite.sampleOperation)
}

func (suite *handleGetOperationByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.sampleOperationID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetOperationByIDSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.sampleOperationID.String()),
		Token:  token,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetOperationByIDSuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/meow",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetOperationByIDSuite) TestStoreRetrievalFail() {
	suite.s.On("OperationByID", mock.Anything, suite.sampleOperationID).
		Return(store.Operation{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.sampleOperationID.String()),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetOperationByIDSuite) TestOK() {
	suite.s.On("OperationByID", mock.Anything, suite.sampleOperationID).
		Return(suite.sampleOperation, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.sampleOperationID.String()),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got publicOperation
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicOperation, got, "should return correct body")
}

func Test_handleGetOperationByID(t *testing.T) {
	suite.Run(t, new(handleGetOperationByIDSuite))
}

// handleCreateOperationSuite tests handleCreateOperation.
type handleCreateOperationSuite struct {
	suite.Suite
	s                  *StoreMock
	r                  *gin.Engine
	sampleCreate       store.Operation
	samplePublicCreate publicOperation
	tokenOK            auth.Token
}

func (suite *handleCreateOperationSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.sampleCreate = store.Operation{
		Title:       "step",
		Description: "loan",
		Start:       time.Date(2003, 1, 23, 15, 0, 0, 0, time.UTC),
		End:         nulls.NewTime(time.Date(2004, 12, 1, 12, 0, 0, 0, time.UTC)),
		IsArchived:  true,
	}
	suite.samplePublicCreate = publicOperationFromStore(suite.sampleCreate)
	suite.tokenOK = auth.Token{
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: permission.CreateOperationPermissionName}},
	}
}

func (suite *handleCreateOperationSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateOperationSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleCreateOperationSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleCreateOperationSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   strings.NewReader("{invalid"),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleCreateOperationSuite) TestCreateFail() {
	suite.s.On("CreateOperation", mock.Anything, suite.sampleCreate).
		Return(store.Operation{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateOperationSuite) TestOK() {
	created := suite.sampleCreate
	created.ID = uuid.New()
	suite.s.On("CreateOperation", mock.Anything, suite.sampleCreate).
		Return(created, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
		Token:  suite.tokenOK,
		Secret: "",
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got publicOperation
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(publicOperationFromStore(created), got, "should return correct body")
}

func Test_handleCreateOperation(t *testing.T) {
	suite.Run(t, new(handleCreateOperationSuite))
}

// handleUpdateOperationSuite tests handleUpdateOperation.
type handleUpdateOperationSuite struct {
	suite.Suite
	s                  *StoreMock
	r                  *gin.Engine
	sampleUpdateID     uuid.UUID
	sampleUpdate       store.Operation
	samplePublicUpdate publicOperation
	tokenOK            auth.Token
}

func (suite *handleUpdateOperationSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.sampleUpdateID = uuid.New()
	suite.sampleUpdate = store.Operation{
		ID:          suite.sampleUpdateID,
		Title:       "step",
		Description: "loan",
		Start:       time.Date(2003, 1, 23, 15, 0, 0, 0, time.UTC),
		End:         nulls.NewTime(time.Date(2004, 12, 1, 12, 0, 0, 0, time.UTC)),
		IsArchived:  true,
	}
	suite.samplePublicUpdate = publicOperationFromStore(suite.sampleUpdate)
	suite.tokenOK = auth.Token{
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: permission.UpdateOperationPermissionName}},
	}
}

func (suite *handleUpdateOperationSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.sampleUpdateID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.sampleUpdateID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.sampleUpdateID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.sampleUpdateID.String()),
		Body:   strings.NewReader("{invalid"),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationSuite) TestIDMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", uuid.New()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationSuite) TestUpdateFail() {
	suite.s.On("UpdateOperation", mock.Anything, suite.sampleUpdate).Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.sampleUpdateID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationSuite) TestOK() {
	updated := suite.sampleUpdate
	updated.ID = uuid.New()
	suite.s.On("UpdateOperation", mock.Anything, suite.sampleUpdate).Return(nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.sampleUpdateID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleUpdateOperation(t *testing.T) {
	suite.Run(t, new(handleUpdateOperationSuite))
}
