package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
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
		UserID:          testutil.NewUUIDV4(),
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
			ID:          testutil.NewUUIDV4(),
			Title:       "roast",
			Description: "ashamed",
			Start:       time.Date(2022, 3, 4, 12, 0, 0, 0, time.UTC),
			End:         nulls.NewTime(time.Date(2022, 11, 1, 12, 0, 0, 0, time.UTC)),
			IsArchived:  true,
		},
		{
			ID:          testutil.NewUUIDV4(),
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
	suite.s.On("Operations", mock.Anything, store.OperationRetrievalFilters{
		OnlyOngoing:     false,
		IncludeArchived: false,
		ForUser:         nulls.NewUUID(suite.tokenOK.UserID),
	}, mock.Anything).Return(suite.sampleOperations, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleGetOperationsSuite) TestInvalidFilterParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/?include_archived=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
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
	suite.s.On("Operations", mock.Anything, mock.Anything, suite.sampleParams).
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
	forUser := testutil.NewUUIDV4()
	suite.s.On("Operations", mock.Anything, store.OperationRetrievalFilters{
		OnlyOngoing:     true,
		IncludeArchived: true,
		ForUser:         nulls.NewUUID(forUser),
	}, suite.sampleParams).Return(suite.sampleOperations, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL: fmt.Sprintf("/?only_ongoing=true&include_archived=true&for_user=%s&%s",
			forUser.String(), pagination.ParamsToQueryString(suite.sampleParams)),
		Token: suite.tokenOK,
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
		UserID:          testutil.NewUUIDV4(),
		Username:        "civilize",
		IsAuthenticated: true,
		IsAdmin:         false,
	}
	suite.sampleOperationID = testutil.NewUUIDV4()
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

func (suite *handleCreateOperationSuite) TestInvalidOperation() {
	suite.samplePublicCreate.Title = ""
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicCreate)),
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
	created.ID = testutil.NewUUIDV4()
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
	suite.sampleUpdateID = testutil.NewUUIDV4()
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
		URL:    fmt.Sprintf("/%s", testutil.NewUUIDV4()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.samplePublicUpdate)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationSuite) TestInvalidOperation() {
	suite.samplePublicUpdate.Title = ""
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.sampleUpdateID.String()),
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
	updated.ID = testutil.NewUUIDV4()
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

// handleGetOperationMembersByOperationSuite tests
// handleGetOperationMembersByOperation.
type handleGetOperationMembersByOperationSuite struct {
	suite.Suite
	s                      *StoreMock
	r                      *gin.Engine
	tokenOK                auth.Token
	sampleOperationID      uuid.UUID
	samplePaginationParams pagination.Params
	sampleMembers          []store.User
	samplePublicMembers    []publicUser
}

func (suite *handleGetOperationMembersByOperationSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "bound",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ViewOperationMembersPermissionName}},
	}
	suite.samplePaginationParams = pagination.Params{
		Limit:          2,
		Offset:         3,
		OrderBy:        nulls.NewString("meow"),
		OrderDirection: pagination.OrderDirDesc,
	}
	suite.sampleMembers = []store.User{
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "spring",
			FirstName: "cautious",
			LastName:  "hit",
		},
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "piece",
			FirstName: "since",
			LastName:  "stand",
		},
	}
	suite.samplePublicMembers = make([]publicUser, 0, len(suite.sampleMembers))
	for _, sMember := range suite.sampleMembers {
		suite.samplePublicMembers = append(suite.samplePublicMembers, publicUser{
			ID:        sMember.ID,
			Username:  sMember.Username,
			FirstName: sMember.FirstName,
			LastName:  sMember.LastName,
		})
	}
}

func (suite *handleGetOperationMembersByOperationSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL: fmt.Sprintf("/%s/members?%s", suite.sampleOperationID.String(),
			pagination.ParamsToQueryString(suite.samplePaginationParams)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetOperationMembersByOperationSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL: fmt.Sprintf("/%s/members?%s", suite.sampleOperationID.String(),
			pagination.ParamsToQueryString(suite.samplePaginationParams)),
		Token: suite.tokenOK,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetOperationMembersByOperationSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL: fmt.Sprintf("/%s/members?%s", suite.sampleOperationID.String(),
			pagination.ParamsToQueryString(suite.samplePaginationParams)),
		Token: suite.tokenOK,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleGetOperationMembersByOperationSuite) TestStoreRetrievalFail() {
	suite.s.On("OperationMembersByOperation", mock.Anything, suite.sampleOperationID).
		Return(nil, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL: fmt.Sprintf("/%s/members?%s", suite.sampleOperationID.String(),
			pagination.ParamsToQueryString(suite.samplePaginationParams)),
		Token: suite.tokenOK,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetOperationMembersByOperationSuite) TestOK() {
	suite.s.On("OperationMembersByOperation", mock.Anything, suite.sampleOperationID).
		Return(suite.sampleMembers, nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL: fmt.Sprintf("/%s/members?%s", suite.sampleOperationID.String(),
			pagination.ParamsToQueryString(suite.samplePaginationParams)),
		Token: suite.tokenOK,
	})
	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got []publicUser
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicMembers, got, "should return correct body")
}

func Test_handleGetOperationMembersByOperation(t *testing.T) {
	suite.Run(t, new(handleGetOperationMembersByOperationSuite))
}

// handleUpdateOperationMembersByOperationSuite tests
// handleUpdateOperationMembersByOperation.
type handleUpdateOperationMembersByOperationSuite struct {
	suite.Suite
	s                 *StoreMock
	r                 *gin.Engine
	sampleOperationID uuid.UUID
	sampleMembers     []uuid.UUID
	tokenOK           auth.Token
}

func (suite *handleUpdateOperationMembersByOperationSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.sampleOperationID = testutil.NewUUIDV4()
	suite.sampleMembers = make([]uuid.UUID, 16)
	for i := range suite.sampleMembers {
		suite.sampleMembers[i] = testutil.NewUUIDV4()
	}
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "great",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.UpdateOperationMembersPermissionName}},
	}
}

func (suite *handleUpdateOperationMembersByOperationSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/members", suite.sampleOperationID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleMembers)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationMembersByOperationSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/members", suite.sampleOperationID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleMembers)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationMembersByOperationSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/members", suite.sampleOperationID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleMembers)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationMembersByOperationSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/members", suite.sampleOperationID.String()),
		Body:   strings.NewReader("{invalid"),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationMembersByOperationSuite) TestUpdateFail() {
	suite.s.On("UpdateOperationMembersByOperation", mock.Anything, suite.sampleOperationID, suite.sampleMembers).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/members", suite.sampleOperationID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleMembers)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateOperationMembersByOperationSuite) TestOK() {
	suite.s.On("UpdateOperationMembersByOperation", mock.Anything, suite.sampleOperationID, suite.sampleMembers).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/members", suite.sampleOperationID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleMembers)),
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleUpdateOperationMembersByOperation(t *testing.T) {
	suite.Run(t, new(handleUpdateOperationMembersByOperationSuite))
}

// handleSearchOperationsSuite tests handleSearchOperations.
type handleSearchOperationsSuite struct {
	suite.Suite
	s                  *StoreMock
	r                  *gin.Engine
	tokenOK            auth.Token
	sampleResult       search.Result[store.Operation]
	samplePublicResult search.Result[publicOperation]
	sampleParams       search.Params
}

func (suite *handleSearchOperationsSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.GET("/search", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleSearchOperations(suite.s)))
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "fame",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ViewAnyOperationPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleResult = search.Result[store.Operation]{
		Hits: []store.Operation{
			{
				ID:          testutil.NewUUIDV4(),
				Title:       "hook",
				Description: "war",
				Start:       time.Date(2022, 8, 20, 9, 0, 53, 1, time.UTC),
				End:         nulls.Time{},
				IsArchived:  false,
			},
			{
				ID:          testutil.NewUUIDV4(),
				Title:       "marriage",
				Description: "join",
				Start:       time.Date(2022, 8, 21, 9, 0, 53, 1, time.UTC),
				End:         nulls.NewTime(time.Date(2022, 7, 3, 1, 0, 3, 0, time.UTC)),
				IsArchived:  true,
			},
		},
		EstimatedTotalHits: 43,
		Offset:             278,
		Limit:              362,
		ProcessingTime:     388 * time.Millisecond,
		Query:              "these",
	}
	suite.samplePublicResult = search.ResultFromResult(suite.sampleResult, []publicOperation{
		{
			ID:          suite.sampleResult.Hits[0].ID,
			Title:       suite.sampleResult.Hits[0].Title,
			Description: suite.sampleResult.Hits[0].Description,
			Start:       suite.sampleResult.Hits[0].Start,
			End:         suite.sampleResult.Hits[0].End,
			IsArchived:  suite.sampleResult.Hits[0].IsArchived,
		},
		{
			ID:          suite.sampleResult.Hits[1].ID,
			Title:       suite.sampleResult.Hits[1].Title,
			Description: suite.sampleResult.Hits[1].Description,
			Start:       suite.sampleResult.Hits[1].Start,
			End:         suite.sampleResult.Hits[1].End,
			IsArchived:  suite.sampleResult.Hits[1].IsArchived,
		},
	})
	suite.sampleParams = search.Params{
		Query:  "elastic",
		Offset: 8129,
		Limit:  71,
	}
}

func (suite *handleSearchOperationsSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s", search.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleSearchOperationsSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s", search.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleSearchOperationsSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}
	suite.s.On("SearchOperations", mock.Anything, store.OperationRetrievalFilters{
		OnlyOngoing:     false,
		IncludeArchived: false,
		ForUser:         nulls.NewUUID(suite.tokenOK.UserID),
	}, mock.Anything).Return(suite.sampleResult, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?for_user=%s&%s", testutil.NewUUIDV4().String(), search.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleSearchOperationsSuite) TestInvalidFilterParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/search?for_user=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleSearchOperationsSuite) TestInvalidSearchParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s=abc", search.QueryParamOffset),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleSearchOperationsSuite) TestStoreRetrievalFail() {
	suite.s.On("SearchOperations", mock.Anything, mock.Anything, suite.sampleParams).
		Return(search.Result[store.Operation]{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s", search.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleSearchOperationsSuite) TestOK() {
	forUser := testutil.NewUUIDV4()
	suite.s.On("SearchOperations", mock.Anything, store.OperationRetrievalFilters{
		OnlyOngoing:     true,
		IncludeArchived: true,
		ForUser:         nulls.NewUUID(forUser),
	},
		suite.sampleParams).
		Return(suite.sampleResult, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL: fmt.Sprintf("/search?only_ongoing=true&include_archived=true&for_user=%s&%s",
			forUser.String(), search.ParamsToQueryString(suite.sampleParams)),
		Token: suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got search.Result[publicOperation]
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicResult, got, "should return correct body")
}

func Test_handleSearchOperations(t *testing.T) {
	suite.Run(t, new(handleSearchOperationsSuite))
}

// handleRebuildOperationSearchSuite tests handleRebuildOperationSearch.
type handleRebuildOperationSearchSuite struct {
	suite.Suite
	s       *StoreMock
	r       *gin.Engine
	tokenOK auth.Token
}

func (suite *handleRebuildOperationSearchSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.POST("/search/rebuild", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleRebuildOperationSearch(suite.s)))
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "organ",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.RebuildSearchIndexPermissionName}},
		RandomSalt:      nil,
	}
}

func (suite *handleRebuildOperationSearchSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleRebuildOperationSearchSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleRebuildOperationSearchSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleRebuildOperationSearchSuite) TestOK() {
	_, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.s.On("RebuildOperationSearch", mock.Anything).Run(func(_ mock.Arguments) {
		cancel()
	}).Once()
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
	wait()
}

func Test_handleRebuildOperationSearch(t *testing.T) {
	suite.Run(t, new(handleRebuildOperationSearchSuite))
}
