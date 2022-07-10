package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
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
)

// groupFiltersFromRequestSuite tests groupFiltersFromRequest.
type groupFiltersFromRequestSuite struct {
	suite.Suite
}

func (suite *groupFiltersFromRequestSuite) genContext(queryParams map[string]string) *gin.Context {
	req, err := http.NewRequest(http.MethodGet, "http://meow", nil)
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	for k, v := range queryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	return &gin.Context{Request: req}
}

func (suite *groupFiltersFromRequestSuite) TestNone() {
	c := suite.genContext(map[string]string{})
	filters, err := groupFiltersFromRequest(c)
	suite.Require().NoError(err, "should not fail")
	suite.Equal(store.GroupFilters{}, filters, "should return correct filters")
}

func (suite *groupFiltersFromRequestSuite) TestInvalidForOperation() {
	c := suite.genContext(map[string]string{
		"for_operation": "meow",
	})
	_, err := groupFiltersFromRequest(c)
	suite.Error(err, "shoudl fail")
}

func (suite *groupFiltersFromRequestSuite) TestForOperationOK() {
	id := testutil.NewUUIDV4()
	c := suite.genContext(map[string]string{
		"for_operation": id.String(),
	})
	filters, err := groupFiltersFromRequest(c)
	suite.Require().NoError(err, "should not fail")
	suite.True(filters.ForOperation.Valid, "should return correct filters")
	suite.Equal(id, filters.ForOperation.UUID, "should return correct filters")
}

func (suite *groupFiltersFromRequestSuite) TestInvalidExcludeGlobal() {
	c := suite.genContext(map[string]string{
		"exclude_global": "meow",
	})
	_, err := groupFiltersFromRequest(c)
	suite.Error(err, "shoudl fail")
}

func (suite *groupFiltersFromRequestSuite) TestExcludeGlobalOK1() {
	c := suite.genContext(map[string]string{
		"exclude_global": "true",
	})
	filters, err := groupFiltersFromRequest(c)
	suite.Require().NoError(err, "should not fail")
	suite.True(filters.ExcludeGlobal, "should return correct filters")
}

func (suite *groupFiltersFromRequestSuite) TestExcludeGlobalOK2() {
	c := suite.genContext(map[string]string{
		"exclude_global": "false",
	})
	filters, err := groupFiltersFromRequest(c)
	suite.Require().NoError(err, "should not fail")
	suite.False(filters.ExcludeGlobal, "should return correct filters")
}

func (suite *groupFiltersFromRequestSuite) TestInvalidByUser() {
	c := suite.genContext(map[string]string{
		"by_user": "meow",
	})
	_, err := groupFiltersFromRequest(c)
	suite.Error(err, "shoudl fail")
}

func (suite *groupFiltersFromRequestSuite) TestByUserOK() {
	id := testutil.NewUUIDV4()
	c := suite.genContext(map[string]string{
		"by_user": id.String(),
	})
	filters, err := groupFiltersFromRequest(c)
	suite.Require().NoError(err, "should not fail")
	suite.True(filters.ByUser.Valid, "should return correct filters")
	suite.Equal(id, filters.ByUser.UUID, "should return correct filters")
}

func (suite *groupFiltersFromRequestSuite) TestAllOK() {
	byUser := testutil.NewUUIDV4()
	forOperation := testutil.NewUUIDV4()
	c := suite.genContext(map[string]string{
		"by_user":        byUser.String(),
		"for_operation":  forOperation.String(),
		"exclude_global": "true",
	})
	filters, err := groupFiltersFromRequest(c)
	suite.Require().NoError(err, "should not fail")
	suite.Equal(nulls.NewUUID(byUser), filters.ByUser, "should return correct filters")
	suite.Equal(nulls.NewUUID(forOperation), filters.ForOperation, "should return correct filters")
	suite.True(filters.ExcludeGlobal, "should return correct filters")
}

func Test_groupFiltersFromRequest(t *testing.T) {
	suite.Run(t, new(groupFiltersFromRequestSuite))
}

// handleGetGroupsSuite tests handleGetGroups.
type handleGetGroupsSuite struct {
	suite.Suite
	s                      *StoreMock
	r                      *gin.Engine
	tokenOK                auth.Token
	sampleGroupFilters     store.GroupFilters
	samplePaginationParams pagination.Params
	sampleGroups           pagination.Paginated[store.Group]
	samplePublicGroups     pagination.Paginated[publicGroup]
}

func (suite *handleGetGroupsSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "note",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ViewGroupPermissionName}},
	}
	suite.sampleGroupFilters = store.GroupFilters{
		ByUser:        nulls.NewUUID(testutil.NewUUIDV4()),
		ForOperation:  nulls.NewUUID(testutil.NewUUIDV4()),
		ExcludeGlobal: true,
	}
	suite.samplePaginationParams = pagination.Params{
		Limit:          2,
		Offset:         3,
		OrderBy:        nulls.NewString("meow"),
		OrderDirection: pagination.OrderDirDesc,
	}
	suite.sampleGroups = pagination.NewPaginated(suite.samplePaginationParams, []store.Group{
		{
			ID:          testutil.NewUUIDV4(),
			Title:       "sleep",
			Description: "surprise",
			Operation:   uuid.NullUUID{},
			Members:     []uuid.UUID{},
		},
		{
			ID:          testutil.NewUUIDV4(),
			Title:       "avenue",
			Description: "",
			Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
			Members: []uuid.UUID{
				testutil.NewUUIDV4(),
				testutil.NewUUIDV4(),
				testutil.NewUUIDV4(),
			},
		},
	}, 933)
	suite.samplePublicGroups = pagination.MapPaginated(suite.sampleGroups, publicGroupFromStore)
}

func (suite *handleGetGroupsSuite) queryString() string {
	paginationParams := pagination.ParamsToQueryString(suite.samplePaginationParams)
	return fmt.Sprintf("%s&by_user=%s&for_operation=%s&exclude_global=%t", paginationParams,
		suite.sampleGroupFilters.ByUser.UUID.String(),
		suite.sampleGroupFilters.ForOperation.UUID.String(),
		suite.sampleGroupFilters.ExcludeGlobal)
}

func (suite *handleGetGroupsSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetGroupsSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetGroupsSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleGetGroupsSuite) TestInvalidGroupFilters() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/?for_operation=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetGroupsSuite) TestInvalidPaginationParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/?%s=abc", pagination.LimitQueryParam),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetGroupsSuite) TestStoreRetrievalFail() {
	suite.s.On("Groups", mock.Anything, suite.sampleGroupFilters, suite.samplePaginationParams).
		Return(pagination.Paginated[store.Group]{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/?%s", suite.queryString()),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetGroupsSuite) TestOK() {
	suite.s.On("Groups", mock.Anything, suite.sampleGroupFilters, suite.samplePaginationParams).
		Return(suite.sampleGroups, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/?%s", suite.queryString()),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got pagination.Paginated[publicGroup]
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicGroups, got, "should return correct body")
}

func Test_handleGetGroups(t *testing.T) {
	suite.Run(t, new(handleGetGroupsSuite))
}

// handleGetGroupByIDSuite tests handleGetGroupByID.
type handleGetGroupByIDSuite struct {
	suite.Suite
	s                 *StoreMock
	r                 *gin.Engine
	tokenOK           auth.Token
	sampleGroupID     uuid.UUID
	sampleGroup       store.Group
	samplePublicGroup publicGroup
}

func (suite *handleGetGroupByIDSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "civilize",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ViewGroupPermissionName}},
	}
	suite.sampleGroupID = testutil.NewUUIDV4()
	suite.sampleGroup = store.Group{
		ID:          suite.sampleGroupID,
		Title:       "steer",
		Description: "money",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	suite.samplePublicGroup = publicGroupFromStore(suite.sampleGroup)
}

func (suite *handleGetGroupByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.sampleGroupID.String()),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetGroupByIDSuite) TestNotAuthenticated() {
	token := suite.tokenOK
	token.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.sampleGroupID.String()),
		Token:  token,
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetGroupByIDSuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/meow",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetGroupByIDSuite) TestStoreRetrievalFail() {
	suite.s.On("GroupByID", mock.Anything, suite.sampleGroupID).
		Return(store.Group{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.sampleGroupID.String()),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetGroupByIDSuite) TestOK() {
	suite.s.On("GroupByID", mock.Anything, suite.sampleGroupID).
		Return(suite.sampleGroup, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.sampleGroupID.String()),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got publicGroup
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicGroup, got, "should return correct body")
}

func Test_handleGetGroupByID(t *testing.T) {
	suite.Run(t, new(handleGetGroupByIDSuite))
}

// handleCreateGroupSuite tests handleCreateGroup.
type handleCreateGroupSuite struct {
	suite.Suite
	s                  *StoreMock
	r                  *gin.Engine
	sampleCreate       store.Group
	samplePublicCreate publicGroup
	tokenOK            auth.Token
}

func (suite *handleCreateGroupSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.sampleCreate = store.Group{
		Title:       "steer",
		Description: "money",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	suite.samplePublicCreate = publicGroupFromStore(suite.sampleCreate)
	suite.tokenOK = auth.Token{
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: permission.CreateGroupPermissionName}},
	}
}

func (suite *handleCreateGroupSuite) TestSecretMismatch() {
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

func (suite *handleCreateGroupSuite) TestNotAuthenticated() {
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

func (suite *handleCreateGroupSuite) TestMissingPermission() {
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

func (suite *handleCreateGroupSuite) TestInvalidBody() {
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

func (suite *handleCreateGroupSuite) TestInvalidGroup() {
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

func (suite *handleCreateGroupSuite) TestCreateFail() {
	suite.s.On("CreateGroup", mock.Anything, suite.sampleCreate).
		Return(store.Group{}, errors.New("sad life"))
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

func (suite *handleCreateGroupSuite) TestOK() {
	created := suite.sampleCreate
	created.ID = testutil.NewUUIDV4()
	suite.s.On("CreateGroup", mock.Anything, suite.sampleCreate).
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
	var got publicGroup
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(publicGroupFromStore(created), got, "should return correct body")
}

func Test_handleCreateGroup(t *testing.T) {
	suite.Run(t, new(handleCreateGroupSuite))
}

// handleUpdateGroupSuite tests handleUpdateGroup.
type handleUpdateGroupSuite struct {
	suite.Suite
	s                  *StoreMock
	r                  *gin.Engine
	sampleUpdateID     uuid.UUID
	sampleUpdate       store.Group
	samplePublicUpdate publicGroup
	tokenOK            auth.Token
}

func (suite *handleUpdateGroupSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.sampleUpdateID = testutil.NewUUIDV4()
	suite.sampleUpdate = store.Group{
		ID:          suite.sampleUpdateID,
		Title:       "steer",
		Description: "money",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
	suite.samplePublicUpdate = publicGroupFromStore(suite.sampleUpdate)
	suite.tokenOK = auth.Token{
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: permission.UpdateGroupPermissionName}},
	}
}

func (suite *handleUpdateGroupSuite) TestSecretMismatch() {
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

func (suite *handleUpdateGroupSuite) TestNotAuthenticated() {
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

func (suite *handleUpdateGroupSuite) TestMissingPermission() {
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

func (suite *handleUpdateGroupSuite) TestInvalidBody() {
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

func (suite *handleUpdateGroupSuite) TestIDMismatch() {
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

func (suite *handleUpdateGroupSuite) TestInvalidGroup() {
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

func (suite *handleUpdateGroupSuite) TestUpdateFail() {
	suite.s.On("UpdateGroup", mock.Anything, suite.sampleUpdate).Return(errors.New("sad life"))
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

func (suite *handleUpdateGroupSuite) TestOK() {
	updated := suite.sampleUpdate
	updated.ID = testutil.NewUUIDV4()
	suite.s.On("UpdateGroup", mock.Anything, suite.sampleUpdate).Return(nil)
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

func Test_handleUpdateGroup(t *testing.T) {
	suite.Run(t, new(handleUpdateGroupSuite))
}

// handleDeleteGroupByIDSuite tests handleDeleteGroupByID.
type handleDeleteGroupByIDSuite struct {
	suite.Suite
	s              *StoreMock
	r              *gin.Engine
	sampleDeleteID uuid.UUID
	tokenOK        auth.Token
}

func (suite *handleDeleteGroupByIDSuite) SetupTest() {
	suite.s = &StoreMock{}
	suite.r = testutil.NewGinEngine()
	populateRoutes(suite.r, zap.NewNop(), "", suite.s)
	suite.sampleDeleteID = testutil.NewUUIDV4()
	suite.tokenOK = auth.Token{
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: permission.DeleteGroupPermissionName}},
	}
}

func (suite *handleDeleteGroupByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.sampleDeleteID.String()),
		Body:   nil,
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleDeleteGroupByIDSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.sampleDeleteID.String()),
		Body:   nil,
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleDeleteGroupByIDSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.sampleDeleteID.String()),
		Body:   nil,
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleDeleteGroupByIDSuite) TestDeleteFail() {
	suite.s.On("DeleteGroupByID", mock.Anything, suite.sampleDeleteID).Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.sampleDeleteID.String()),
		Body:   nil,
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleDeleteGroupByIDSuite) TestOK() {
	suite.s.On("DeleteGroupByID", mock.Anything, suite.sampleDeleteID).Return(nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.sampleDeleteID.String()),
		Body:   nil,
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleDeleteGroupByID(t *testing.T) {
	suite.Run(t, new(handleDeleteGroupByIDSuite))
}
