package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
)

// handleGetPermissionsByUserStoreMock mocks handleUpdatePermissionsByUserStore.
type handleGetPermissionsByUserStoreMock struct {
	mock.Mock
}

func (m *handleGetPermissionsByUserStoreMock) PermissionsByUser(ctx context.Context, userID uuid.UUID) ([]store.Permission, error) {
	args := m.Called(ctx, userID)
	var p []store.Permission
	if argsPermissions := args.Get(0); argsPermissions != nil {
		p = argsPermissions.([]store.Permission)
	}
	return p, args.Error(1)
}

// handleGetPermissionsByUserSuite tests handleGetPermissionsByUser.
type handleGetPermissionsByUserSuite struct {
	suite.Suite
	s                       *handleGetPermissionsByUserStoreMock
	r                       *gin.Engine
	sampleUserID            uuid.UUID
	tokenOK                 auth.Token
	samplePermissions       []store.Permission
	samplePublicPermissions []publicPermission
}

func (suite *handleGetPermissionsByUserSuite) SetupTest() {
	suite.s = &handleGetPermissionsByUserStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.GET("/user/:userID", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleGetPermissionsByUser(suite.s)))
	suite.tokenOK = auth.Token{
		UserID:          suite.sampleUserID,
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: permission.ViewPermissionsPermissionName}},
	}
	suite.samplePermissions = []store.Permission{
		{
			Name: "Hello",
		},
		{
			Name:    "World",
			Options: nulls.NewJSONRawMessage([]byte("1")),
		},
		{
			Name: "!",
		},
	}
	suite.samplePublicPermissions = make([]publicPermission, 0, len(suite.samplePermissions))
	for _, samplePermission := range suite.samplePermissions {
		suite.samplePublicPermissions = append(suite.samplePublicPermissions, publicPermissionFromPermission(samplePermission))
	}
}

func (suite *handleGetPermissionsByUserSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   nil,
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetPermissionsByUserSuite) TestNotAuthenticated() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   nil,
		Token:  auth.Token{IsAuthenticated: false},
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetPermissionsByUserSuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/user/meow",
		Body:   nil,
		Token:  suite.tokenOK,
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetPermissionsByUserSuite) TestMissingPermission() {
	token := suite.tokenOK
	token.Permissions = nil
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/user/%s", uuid.New().String()),
		Body:   nil,
		Token:  token,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code", nil)
}

func (suite *handleGetPermissionsByUserSuite) TestStoreRetrievalFail() {
	suite.s.On("PermissionsByUser", mock.Anything, suite.sampleUserID).Return(nil, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   nil,
		Token:  suite.tokenOK,
		Secret: "",
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetPermissionsByUserSuite) TestOKSelf() {
	suite.s.On("PermissionsByUser", mock.Anything, suite.sampleUserID).Return(suite.samplePermissions, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   nil,
		Token:  suite.tokenOK,
		Secret: "",
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got []publicPermission
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicPermissions, got, "should return correct body")
}

func Test_handleGetPermissionsByUser(t *testing.T) {
	suite.Run(t, new(handleGetPermissionsByUserSuite))
}

// handleUpdatePermissionsByUserStoreMock mocks
// handleUpdatePermissionsByUserStore.
type handleUpdatePermissionsByUserStoreMock struct {
	mock.Mock
}

func (m *handleUpdatePermissionsByUserStoreMock) UpdatePermissionsByUser(ctx context.Context, userID uuid.UUID, permissions []store.Permission) error {
	return m.Called(ctx, userID, permissions).Error(0)
}

// handleUpdatePermissionsByUserSuite tests handleUpdatePermissionsByUser.
type handleUpdatePermissionsByUserSuite struct {
	suite.Suite
	s                              *handleUpdatePermissionsByUserStoreMock
	r                              *gin.Engine
	sampleUpdatedPermissions       []store.Permission
	sampleUpdatedPublicPermissions []publicPermission
	sampleUserID                   uuid.UUID
	tokenOK                        auth.Token
}

func (suite *handleUpdatePermissionsByUserSuite) SetupTest() {
	suite.s = &handleUpdatePermissionsByUserStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.PUT("/user/:userID", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleUpdatePermissionsByUser(suite.s)))
	suite.sampleUpdatedPermissions = []store.Permission{
		{
			Name: "Hello",
		},
		{
			Name:    "World",
			Options: nulls.NewJSONRawMessage([]byte(`{"meow":"woof"}`)),
		},
		{
			Name: "!",
		},
	}
	suite.sampleUpdatedPublicPermissions = make([]publicPermission, 0, len(suite.sampleUpdatedPermissions))
	for _, p := range suite.sampleUpdatedPermissions {
		suite.sampleUpdatedPublicPermissions = append(suite.sampleUpdatedPublicPermissions, publicPermissionFromPermission(p))
	}
	suite.sampleUserID = uuid.New()
	suite.tokenOK = auth.Token{
		UserID:          suite.sampleUserID,
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: permission.UpdatePermissionsPermissionName}},
	}
}

func (suite *handleUpdatePermissionsByUserSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleUpdatedPublicPermissions)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdatePermissionsByUserSuite) TestNotAuthenticated() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleUpdatedPublicPermissions)),
		Token:  auth.Token{IsAuthenticated: false},
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleUpdatePermissionsByUserSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   strings.NewReader("{invalid"),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdatePermissionsByUserSuite) TestMissingPermission() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/user/%s", uuid.New().String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleUpdatedPublicPermissions)),
		Token:  token,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleUpdatePermissionsByUserSuite) TestSelfWithoutPermission() {
	token := suite.tokenOK
	token.Permissions = []permission.Permission{}
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleUpdatedPublicPermissions)),
		Token:  token,
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleUpdatePermissionsByUserSuite) TestStoreUpdateFail() {
	suite.s.On("UpdatePermissionsByUser", mock.Anything, suite.sampleUserID, suite.sampleUpdatedPermissions).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleUpdatedPublicPermissions)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdatePermissionsByUserSuite) TestOK() {
	suite.s.On("UpdatePermissionsByUser", mock.Anything, suite.sampleUserID, suite.sampleUpdatedPermissions).
		Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/user/%s", suite.sampleUserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleUpdatedPublicPermissions)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleUpdatePermissionsByUser(t *testing.T) {
	suite.Run(t, new(handleUpdatePermissionsByUserSuite))
}
