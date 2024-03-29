package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
	"time"
)

// handleCreateUserStoreMock mocks handleCreateUserStore.
type handleCreateUserStoreMock struct {
	mock.Mock
}

func (m *handleCreateUserStoreMock) CreateUser(ctx context.Context, user store.UserWithPass) (store.UserWithPass, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(store.UserWithPass), args.Error(1)
}

// handleCreateUserSuite tests handleCreateUser.
type handleCreateUserSuite struct {
	suite.Suite
	s          *handleCreateUserStoreMock
	r          *gin.Engine
	createUser createUserRequest
}

func (suite *handleCreateUserSuite) SetupTest() {
	suite.s = &handleCreateUserStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.POST("/", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleCreateUser(suite.s)))
	suite.createUser = createUserRequest{
		Username:  "olaf",
		FirstName: "snow",
		LastName:  "man",
		IsAdmin:   false,
		Pass:      "rio",
	}
}

func (suite *handleCreateUserSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.createUser)),
		Token: auth.Token{
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.CreateUserPermissionName}},
		},
		Secret: "woof",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateUserSuite) TestNotAuthenticated() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.createUser)),
		Token:  auth.Token{IsAuthenticated: false},
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleCreateUserSuite) TestMissingCreateUserPermission() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.createUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleCreateUserSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   strings.NewReader("{invalid"),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.CreateUserPermissionName}},
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleCreateUserSuite) TestMissingSetAdminPermission() {
	suite.createUser.IsAdmin = true
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.createUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.CreateUserPermissionName}},
		},
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleCreateUserSuite) TestInvalidUser() {
	suite.createUser.Username = ""
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.createUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.CreateUserPermissionName}},
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleCreateUserSuite) TestCreateUserFail() {
	suite.s.On("CreateUser", mock.Anything, mock.Anything).Return(store.UserWithPass{}, errors.New("meh"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.createUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.CreateUserPermissionName}},
		},
		Secret: "",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleCreateUserSuite) TestOKNoAdmin() {
	hashedPass, _ := auth.HashPassword(suite.createUser.Pass)
	created := store.UserWithPass{
		User: store.User{
			Username:  suite.createUser.Username,
			FirstName: suite.createUser.FirstName,
			LastName:  suite.createUser.LastName,
			IsAdmin:   suite.createUser.IsAdmin,
			IsActive:  true,
		},
		Pass: hashedPass,
	}

	created.ID = testutil.NewUUIDV4()
	suite.s.On("CreateUser", mock.Anything, mock.Anything).Return(created, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.createUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.CreateUserPermissionName}},
		},
		Secret: "",
	})

	suite.Require().Equal(http.StatusCreated, rr.Code, "should return correct code")
	var got createUserResponse
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "response body should be valid")
	suite.Equal(createUserResponse{
		ID:        created.ID,
		Username:  created.Username,
		FirstName: created.FirstName,
		LastName:  created.LastName,
		IsAdmin:   created.IsAdmin,
		IsActive:  created.IsActive,
	}, got, "response body should be correct")
}

func (suite *handleCreateUserSuite) TestOKAdmin() {
	suite.s.On("CreateUser", mock.Anything, mock.Anything).Return(store.UserWithPass{}, nil)
	defer suite.s.AssertExpectations(suite.T())
	suite.createUser.IsAdmin = true

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.createUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions: []permission.Permission{
				{Name: permission.CreateUserPermissionName},
				{Name: permission.SetAdminUserPermissionName},
			},
		},
		Secret: "",
	})

	suite.Equal(http.StatusCreated, rr.Code, "should return correct code")
}

func Test_handleCreateUser(t *testing.T) {
	suite.Run(t, new(handleCreateUserSuite))
}

// handleUpdateUserByIDStoreMock mocks handleUpdateUserByIDStore.
type handleUpdateUserByIDStoreMock struct {
	mock.Mock
}

func (m *handleUpdateUserByIDStoreMock) UpdateUser(ctx context.Context, user store.User, allowAdminChange bool, allowActiveStateChange bool) error {
	return m.Called(ctx, user, allowAdminChange, allowActiveStateChange).Error(0)
}

// handleUpdateUserByIDSuite tests handleUpdateUserByID.
type handleUpdateUserByIDSuite struct {
	suite.Suite
	s          *handleUpdateUserByIDStoreMock
	r          *gin.Engine
	updateUser updateUserRequest
}

func (suite *handleUpdateUserByIDSuite) SetupTest() {
	suite.s = &handleUpdateUserByIDStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.PUT("/:userID", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleUpdateUserByID(suite.s)))
	suite.updateUser = updateUserRequest{
		ID:        testutil.NewUUIDV4(),
		Username:  "anyone",
		FirstName: "stay",
		LastName:  "smile",
		IsAdmin:   false,
		IsActive:  true,
	}
}

func (suite *handleUpdateUserByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.UpdateUserPermissionName}},
		},
		Secret: "woof",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestNotAuthenticated() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token:  auth.Token{IsAuthenticated: false},
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   strings.NewReader("{invalid"),
		Token: auth.Token{
			UserID:          suite.updateUser.ID,
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestIDMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    "/meow",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          suite.updateUser.ID,
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestMissingPermissionForForeignUser() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestInvalidUser() {
	suite.updateUser.Username = ""
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          suite.updateUser.ID,
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestUpdateUserFail() {
	suite.s.On("UpdateUser", mock.Anything, store.User{
		ID:        suite.updateUser.ID,
		Username:  suite.updateUser.Username,
		FirstName: suite.updateUser.FirstName,
		LastName:  suite.updateUser.LastName,
		IsAdmin:   suite.updateUser.IsAdmin,
		IsActive:  suite.updateUser.IsActive,
	}, false, false).Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          suite.updateUser.ID,
			IsAuthenticated: true,
		},
		Secret: "",
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestOKWithSelf() {
	suite.s.On("UpdateUser", mock.Anything, store.User{
		ID:        suite.updateUser.ID,
		Username:  suite.updateUser.Username,
		FirstName: suite.updateUser.FirstName,
		LastName:  suite.updateUser.LastName,
		IsAdmin:   suite.updateUser.IsAdmin,
		IsActive:  suite.updateUser.IsActive,
	}, false, false).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          suite.updateUser.ID,
			IsAuthenticated: true,
		},
		Secret: "",
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestOKWithActiveStateChange() {
	suite.s.On("UpdateUser", mock.Anything, store.User{
		ID:        suite.updateUser.ID,
		Username:  suite.updateUser.Username,
		FirstName: suite.updateUser.FirstName,
		LastName:  suite.updateUser.LastName,
		IsAdmin:   suite.updateUser.IsAdmin,
		IsActive:  suite.updateUser.IsActive,
	}, false, true).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          suite.updateUser.ID,
			IsAuthenticated: true,
			Permissions: []permission.Permission{
				{Name: permission.SetUserActiveStatePermission},
			},
		},
		Secret: "",
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestOKWithSelfAdminChange() {
	suite.s.On("UpdateUser", mock.Anything, store.User{
		ID:        suite.updateUser.ID,
		Username:  suite.updateUser.Username,
		FirstName: suite.updateUser.FirstName,
		LastName:  suite.updateUser.LastName,
		IsAdmin:   suite.updateUser.IsAdmin,
		IsActive:  suite.updateUser.IsActive,
	}, true, false).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          suite.updateUser.ID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.SetAdminUserPermissionName}},
		},
		Secret: "",
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestOKWithForeign() {
	suite.s.On("UpdateUser", mock.Anything, store.User{
		ID:        suite.updateUser.ID,
		Username:  suite.updateUser.Username,
		FirstName: suite.updateUser.FirstName,
		LastName:  suite.updateUser.LastName,
		IsAdmin:   suite.updateUser.IsAdmin,
		IsActive:  suite.updateUser.IsActive,
	}, false, false).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.UpdateUserPermissionName}},
		},
		Secret: "",
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestOKWithForeignAdminChange() {
	suite.s.On("UpdateUser", mock.Anything, store.User{
		ID:        suite.updateUser.ID,
		Username:  suite.updateUser.Username,
		FirstName: suite.updateUser.FirstName,
		LastName:  suite.updateUser.LastName,
		IsAdmin:   suite.updateUser.IsAdmin,
		IsActive:  suite.updateUser.IsActive,
	}, true, false).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions: []permission.Permission{
				{Name: permission.UpdateUserPermissionName},
				{Name: permission.SetAdminUserPermissionName},
			},
		},
		Secret: "",
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleUpdateUserByID(t *testing.T) {
	suite.Run(t, new(handleUpdateUserByIDSuite))
}

// handleUpdateUserPassByUserIDStoreMock mocks handleUpdateUserPassByUserIDStore.
type handleUpdateUserPassByUserIDStoreMock struct {
	mock.Mock
}

func (m *handleUpdateUserPassByUserIDStoreMock) UpdateUserPassByUserID(ctx context.Context, userID uuid.UUID, newPass []byte) error {
	return m.Called(ctx, userID, newPass).Error(0)
}

// handleUpdateUserPassByUserIDSuite tests handleUpdateUserPassByUserID.
type handleUpdateUserPassByUserIDSuite struct {
	suite.Suite
	s          *handleUpdateUserPassByUserIDStoreMock
	r          *gin.Engine
	updatePass updateUserPassByUserIDRequest
}

func (suite *handleUpdateUserPassByUserIDSuite) SetupTest() {
	suite.s = &handleUpdateUserPassByUserIDStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.PUT("/:userID/pass", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleUpdateUserPassByUserID(suite.s)))
	suite.updatePass = updateUserPassByUserIDRequest{
		UserID:  testutil.NewUUIDV4(),
		NewPass: "industry",
	}
}

func (suite *handleUpdateUserPassByUserIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/pass", suite.updatePass.UserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updatePass)),
		Token: auth.Token{
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.UpdateUserPermissionName}},
		},
		Secret: "woof",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserPassByUserIDSuite) TestNotAuthenticated() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/pass", suite.updatePass.UserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updatePass)),
		Token: auth.Token{
			IsAuthenticated: false,
		},
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserPassByUserIDSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/pass", suite.updatePass.UserID.String()),
		Body:   strings.NewReader("{invalid"),
		Token: auth.Token{
			UserID:          suite.updatePass.UserID,
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correcd code")
}

func (suite *handleUpdateUserPassByUserIDSuite) TestIDMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    "/meow/pass",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updatePass)),
		Token: auth.Token{
			UserID:          suite.updatePass.UserID,
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserPassByUserIDSuite) TestMissingPermissionForForeignUser() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/pass", suite.updatePass.UserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updatePass)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserPassByUserIDSuite) TestUpdateUserPassFail() {
	suite.s.On("UpdateUserPassByUserID", mock.Anything, suite.updatePass.UserID, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/pass", suite.updatePass.UserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updatePass)),
		Token: auth.Token{
			UserID:          suite.updatePass.UserID,
			IsAuthenticated: true,
		},
		Secret: "",
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserPassByUserIDSuite) TestOKSelf() {
	suite.s.On("UpdateUserPassByUserID", mock.Anything, suite.updatePass.UserID, mock.Anything).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/pass", suite.updatePass.UserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updatePass)),
		Token: auth.Token{
			UserID:          suite.updatePass.UserID,
			IsAuthenticated: true,
		},
		Secret: "",
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserPassByUserIDSuite) TestOKForeign() {
	suite.s.On("UpdateUserPassByUserID", mock.Anything, suite.updatePass.UserID, mock.Anything).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s/pass", suite.updatePass.UserID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updatePass)),
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.UpdateUserPassPermissionName}},
		},
		Secret: "",
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleUpdateUserPassByID(t *testing.T) {
	suite.Run(t, new(handleUpdateUserPassByUserIDSuite))
}

// handleDeleteUserByIDStoreMock mocks handleDeleteUserByIDStore.
type handleDeleteUserByIDStoreMock struct {
	mock.Mock
}

func (m *handleDeleteUserByIDStoreMock) SetUserInactiveByID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

// handleDeleteUserByIDSuite tests handleDeleteUserByID.
type handleDeleteUserByIDSuite struct {
	suite.Suite
	s      *handleDeleteUserByIDStoreMock
	r      *gin.Engine
	userID uuid.UUID
}

func (suite *handleDeleteUserByIDSuite) SetupTest() {
	suite.s = &handleDeleteUserByIDStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.DELETE("/:userID", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleDeleteUserByID(suite.s)))
	suite.userID = testutil.NewUUIDV4()
}

func (suite *handleDeleteUserByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.SetUserActiveStatePermission}},
		},
		Secret: "woof",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleDeleteUserByIDSuite) TestNotAuthenticated() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			IsAuthenticated: false,
		},
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleDeleteUserByIDSuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    "/meow",
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.SetUserActiveStatePermission}},
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleDeleteUserByIDSuite) TestMissingPermission() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleDeleteUserByIDSuite) TestDeleteFail() {
	suite.s.On("SetUserInactiveByID", mock.Anything, suite.userID).Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.SetUserActiveStatePermission}},
		},
		Secret: "",
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleDeleteUserByIDSuite) TestOK() {
	suite.s.On("SetUserInactiveByID", mock.Anything, suite.userID).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.SetUserActiveStatePermission}},
		},
		Secret: "",
	})

	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleDeleteUserByID(t *testing.T) {
	suite.Run(t, new(handleDeleteUserByIDSuite))
}

// handleGetUserByIDStoreMock mocks handleGetUserByID.
type handleGetUserByIDStoreMock struct {
	mock.Mock
}

func (m *handleGetUserByIDStoreMock) UserByID(ctx context.Context, userID uuid.UUID) (store.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(store.User), args.Error(1)
}

// handleGetUserByID tests handleGetUserByID.
type handleGetUserByIDSuite struct {
	suite.Suite
	s      *handleGetUserByIDStoreMock
	r      *gin.Engine
	userID uuid.UUID
}

func (suite *handleGetUserByIDSuite) SetupTest() {
	suite.s = &handleGetUserByIDStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.GET("/:userID", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleGetUserByID(suite.s)))
	suite.userID = testutil.NewUUIDV4()
}

func (suite *handleGetUserByIDSuite) TestNotAuthenticated() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			IsAuthenticated: false,
		},
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetUserByIDSuite) TestInvalidID() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/meow",
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.ViewUserPermissionName}},
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetUserByIDSuite) TestMissingPermission() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%v", testutil.NewUUIDV4()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleGetUserByIDSuite) TestRetrieveFromStoreFail() {
	suite.s.On("UserByID", mock.Anything, suite.userID).Return(store.User{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.ViewUserPermissionName}},
		},
		Secret: "",
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetUserByIDSuite) TestOKSelf() {
	user := store.User{
		ID:        suite.userID,
		Username:  "white",
		FirstName: "ring",
		LastName:  "weave",
		IsAdmin:   true,
	}
	suite.s.On("UserByID", mock.Anything, suite.userID).Return(user, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
		},
		Secret: "",
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got getUserResponse
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid response")
	suite.Equal(getUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsAdmin:   user.IsAdmin,
	}, got, "should return correct response")
}

func (suite *handleGetUserByIDSuite) TestOKForeign() {
	user := store.User{
		ID:        testutil.NewUUIDV4(),
		Username:  "white",
		FirstName: "ring",
		LastName:  "weave",
		IsAdmin:   true,
	}
	suite.s.On("UserByID", mock.Anything, user.ID).Return(user, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%s", user.ID.String()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.ViewUserPermissionName}},
		},
		Secret: "",
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got getUserResponse
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid response")
	suite.Equal(getUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsAdmin:   user.IsAdmin,
	}, got, "should return correct response")
}

func Test_handleGetUserByID(t *testing.T) {
	suite.Run(t, new(handleGetUserByIDSuite))
}

// handleGetUsersStoreMock mocks handleGetUsersStore.
type handleGetUsersStoreMock struct {
	mock.Mock
}

func (m *handleGetUsersStoreMock) Users(ctx context.Context, filters store.UserFilters,
	params pagination.Params) (pagination.Paginated[store.User], error) {
	args := m.Called(ctx, filters, params)
	return args.Get(0).(pagination.Paginated[store.User]), args.Error(1)
}

// handleGetUsersSuite tests handleGetUsers.
type handleGetUsersSuite struct {
	suite.Suite
	s *handleGetUsersStoreMock
	r *gin.Engine
}

func (suite *handleGetUsersSuite) SetupTest() {
	suite.s = &handleGetUsersStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.GET("/", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleGetUsers(suite.s)))
}

func (suite *handleGetUsersSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Body:   nil,
		Token: auth.Token{
			IsAuthenticated: true,
		},
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetUsersSuite) TestNotAuthenticated() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Body:   nil,
		Token: auth.Token{
			IsAuthenticated: false,
		},
		Secret: "",
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleGetUsersSuite) TestMissingPermission() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Body:   nil,
		Token: auth.Token{
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleGetUsersSuite) TestInvalidPagination() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/?%v=woof", pagination.LimitQueryParam),
		Body:   nil,
		Token: auth.Token{
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.ViewUserPermissionName}},
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetUsersSuite) TestInvalidFilterParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/?include_inactive=abc",
		Body:   nil,
		Token: auth.Token{
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.ViewUserPermissionName}},
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetUsersSuite) TestStoreRetrievalFail() {
	suite.s.On("Users", mock.Anything, mock.Anything, mock.Anything).
		Return(pagination.Paginated[store.User]{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Body:   nil,
		Token: auth.Token{
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.ViewUserPermissionName}},
		},
		Secret: "",
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetUsersSuite) TestOK() {
	params := pagination.Params{Limit: 7}
	paginated := pagination.NewPaginated(params, []store.User{
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "rabbit",
			FirstName: "scarce",
			LastName:  "sudden",
			IsAdmin:   false,
		},
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "content",
			FirstName: "parcel",
			LastName:  "discover",
			IsAdmin:   true,
		},
	}, 28)
	suite.s.On("Users", mock.Anything, store.UserFilters{
		IncludeInactive: true,
	}, pagination.Params{Limit: 7, OrderDirection: "asc"}).
		Return(paginated, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/?%s=7&include_inactive=true", pagination.LimitQueryParam),
		Body:   nil,
		Token: auth.Token{
			UserID:          testutil.NewUUIDV4(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{{Name: permission.ViewUserPermissionName}},
		},
		Secret: "",
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got pagination.Paginated[getUsersResponseUser]
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid response")
	suite.Equal(pagination.MapPaginated(paginated, func(from store.User) getUsersResponseUser {
		return getUsersResponseUser{
			ID:        from.ID,
			Username:  from.Username,
			FirstName: from.FirstName,
			LastName:  from.LastName,
			IsAdmin:   from.IsAdmin,
		}
	}), got, "should return correct response")
}

func Test_handleGetUsers(t *testing.T) {
	suite.Run(t, new(handleGetUsersSuite))
}

// handleSearchUsersStoreMock mocks handleSearchUsersStore.
type handleSearchUsersStoreMock struct {
	mock.Mock
}

func (m *handleSearchUsersStoreMock) SearchUsers(ctx context.Context, filters store.UserFilters, searchParams search.Params) (search.Result[store.User], error) {
	args := m.Called(ctx, filters, searchParams)
	return args.Get(0).(search.Result[store.User]), args.Error(1)
}

// handleSearchUsersSuite tests handleSearchUsers.
type handleSearchUsersSuite struct {
	suite.Suite
	s                  *handleSearchUsersStoreMock
	r                  *gin.Engine
	tokenOK            auth.Token
	sampleResult       search.Result[store.User]
	samplePublicResult search.Result[getUsersResponseUser]
	sampleParams       search.Params
}

func (suite *handleSearchUsersSuite) SetupTest() {
	suite.s = &handleSearchUsersStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.GET("/search", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleSearchUsers(suite.s)))
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "fame",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.ViewUserPermissionName}},
		RandomSalt:      nil,
	}
	suite.sampleResult = search.Result[store.User]{
		Hits: []store.User{
			{
				ID:        testutil.NewUUIDV4(),
				Username:  "hook",
				FirstName: "build",
				LastName:  "canal",
				IsAdmin:   true,
			},
			{
				ID:        testutil.NewUUIDV4(),
				Username:  "field",
				FirstName: "burn",
				LastName:  "paw",
				IsAdmin:   false,
			},
		},
		EstimatedTotalHits: 43,
		Offset:             278,
		Limit:              362,
		ProcessingTime:     388 * time.Millisecond,
		Query:              "over",
	}
	suite.samplePublicResult = search.ResultFromResult(suite.sampleResult, []getUsersResponseUser{
		{
			ID:        suite.sampleResult.Hits[0].ID,
			Username:  suite.sampleResult.Hits[0].Username,
			FirstName: suite.sampleResult.Hits[0].FirstName,
			LastName:  suite.sampleResult.Hits[0].LastName,
			IsAdmin:   suite.sampleResult.Hits[0].IsAdmin,
		},
		{
			ID:        suite.sampleResult.Hits[1].ID,
			Username:  suite.sampleResult.Hits[1].Username,
			FirstName: suite.sampleResult.Hits[1].FirstName,
			LastName:  suite.sampleResult.Hits[1].LastName,
			IsAdmin:   suite.sampleResult.Hits[1].IsAdmin,
		},
	})
	suite.sampleParams = search.Params{
		Query:  "needle",
		Offset: 819,
		Limit:  7,
	}
}

func (suite *handleSearchUsersSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s", search.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleSearchUsersSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s", search.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleSearchUsersSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s", search.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleSearchUsersSuite) TestInvalidSearchParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s=abc", search.QueryParamOffset),
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleSearchUsersSuite) TestInvalidFilterParams() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/search?include_inactive=abc",
		Token:  suite.tokenOK,
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleSearchUsersSuite) TestStoreRetrievalFail() {
	suite.s.On("SearchUsers", mock.Anything, mock.Anything, suite.sampleParams).
		Return(search.Result[store.User]{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s", search.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleSearchUsersSuite) TestOK() {
	suite.s.On("SearchUsers", mock.Anything, store.UserFilters{IncludeInactive: true},
		suite.sampleParams).
		Return(suite.sampleResult, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/search?%s&include_inactive=true", search.ParamsToQueryString(suite.sampleParams)),
		Token:  suite.tokenOK,
	})

	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got search.Result[getUsersResponseUser]
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid body")
	suite.Equal(suite.samplePublicResult, got, "should return correct body")
}

func Test_handleSearchUsers(t *testing.T) {
	suite.Run(t, new(handleSearchUsersSuite))
}

// handleRebuildUserSearchStoreMocks mocks handleRebuildUserSearchStore.
type handleRebuildUserSearchStoreMock struct {
	mock.Mock
}

func (m *handleRebuildUserSearchStoreMock) RebuildUserSearch(ctx context.Context) {
	m.Called(ctx)
}

// handleRebuildUserSearchSuite tests handleRebuildUserSearch.
type handleRebuildUserSearchSuite struct {
	suite.Suite
	s       *handleRebuildUserSearchStoreMock
	r       *gin.Engine
	tokenOK auth.Token
}

func (suite *handleRebuildUserSearchSuite) SetupTest() {
	suite.s = &handleRebuildUserSearchStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.POST("/search/rebuild", httpendpoints.GinHandlerFunc(zap.NewNop(), "", handleRebuildUserSearch(suite.s)))
	suite.tokenOK = auth.Token{
		UserID:          testutil.NewUUIDV4(),
		Username:        "organ",
		IsAuthenticated: true,
		IsAdmin:         false,
		Permissions:     []permission.Permission{{Name: permission.RebuildSearchIndexPermissionName}},
		RandomSalt:      nil,
	}
}

func (suite *handleRebuildUserSearchSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
		Secret: "meow",
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleRebuildUserSearchSuite) TestNotAuthenticated() {
	suite.tokenOK.IsAuthenticated = false

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleRebuildUserSearchSuite) TestMissingPermission() {
	suite.tokenOK.Permissions = []permission.Permission{}

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/search/rebuild",
		Token:  suite.tokenOK,
	})

	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleRebuildUserSearchSuite) TestOK() {
	_, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.s.On("RebuildUserSearch", mock.Anything).Run(func(_ mock.Arguments) {
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

func Test_handleRebuildUserSearch(t *testing.T) {
	suite.Run(t, new(handleRebuildUserSearchSuite))
}
