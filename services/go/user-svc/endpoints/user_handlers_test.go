package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
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
			Permissions:     []permission.Permission{permission.CreateUser},
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
			UserID:          uuid.New(),
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
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.CreateUser},
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
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.CreateUser},
		},
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
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
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.CreateUser},
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
		},
		Pass: hashedPass,
	}

	created.ID, _ = uuid.NewUUID()
	suite.s.On("CreateUser", mock.Anything, mock.Anything).Return(created, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.createUser)),
		Token: auth.Token{
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.CreateUser},
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
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.CreateUser, permission.SetAdminUser},
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

func (m *handleUpdateUserByIDStoreMock) UpdateUser(ctx context.Context, user store.User, allowAdminChange bool) error {
	return m.Called(ctx, user, allowAdminChange).Error(0)
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
		ID:        uuid.New(),
		Username:  "anyone",
		FirstName: "stay",
		LastName:  "smile",
		IsAdmin:   false,
	}
}

func (suite *handleUpdateUserByIDSuite) TestSecretMismatch() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.UpdateUser},
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
			UserID:          uuid.New(),
			IsAuthenticated: true,
		},
		Secret: "",
	})
	suite.Equal(http.StatusForbidden, rr.Code, "should return correct code")
}

func (suite *handleUpdateUserByIDSuite) TestUpdateUserFail() {
	suite.s.On("UpdateUser", mock.Anything, store.User{
		ID:        suite.updateUser.ID,
		Username:  suite.updateUser.Username,
		FirstName: suite.updateUser.FirstName,
		LastName:  suite.updateUser.LastName,
		IsAdmin:   suite.updateUser.IsAdmin,
	}, false).Return(errors.New("sad life"))
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
	}, false).Return(nil)
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

func (suite *handleUpdateUserByIDSuite) TestOKWithSelfAdminChange() {
	suite.s.On("UpdateUser", mock.Anything, store.User{
		ID:        suite.updateUser.ID,
		Username:  suite.updateUser.Username,
		FirstName: suite.updateUser.FirstName,
		LastName:  suite.updateUser.LastName,
		IsAdmin:   suite.updateUser.IsAdmin,
	}, true).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          suite.updateUser.ID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.SetAdminUser},
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
	}, false).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.UpdateUser},
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
	}, true).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/%s", suite.updateUser.ID.String()),
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.updateUser)),
		Token: auth.Token{
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.UpdateUser, permission.SetAdminUser},
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
		UserID:  uuid.New(),
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
			Permissions:     []permission.Permission{permission.UpdateUser},
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
			UserID:          uuid.New(),
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
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.UpdateUserPass},
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

func (m *handleDeleteUserByIDStoreMock) DeleteUserByID(ctx context.Context, userID uuid.UUID) error {
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
	suite.userID = uuid.New()
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
			Permissions:     []permission.Permission{permission.DeleteUser},
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
			Permissions:     []permission.Permission{permission.DeleteUser},
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
	suite.s.On("DeleteUserByID", mock.Anything, suite.userID).Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.DeleteUser},
		},
		Secret: "",
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleDeleteUserByIDSuite) TestOK() {
	suite.s.On("DeleteUserByID", mock.Anything, suite.userID).Return(nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/%s", suite.userID.String()),
		Body:   nil,
		Token: auth.Token{
			UserID:          suite.userID,
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.DeleteUser},
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
	suite.userID = uuid.New()
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
			Permissions:     []permission.Permission{permission.ViewUser},
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetUserByIDSuite) TestMissingPermission() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/%v", uuid.New()),
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
			Permissions:     []permission.Permission{permission.ViewUser},
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
		ID:        uuid.New(),
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
			Permissions:     []permission.Permission{permission.ViewUser},
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

func (m *handleGetUsersStoreMock) Users(ctx context.Context, params pagination.Params) (pagination.Paginated[store.User], error) {
	args := m.Called(ctx, params)
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
			Permissions:     []permission.Permission{permission.ViewUser},
		},
		Secret: "",
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleGetUsersSuite) TestStoreRetrievalFail() {
	suite.s.On("Users", mock.Anything, mock.Anything).
		Return(pagination.Paginated[store.User]{}, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    "/",
		Body:   nil,
		Token: auth.Token{
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.ViewUser},
		},
		Secret: "",
	})

	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleGetUsersSuite) TestOK() {
	params := pagination.Params{Limit: 7}
	paginated := pagination.NewPaginated(params, []store.User{
		{
			ID:        uuid.New(),
			Username:  "rabbit",
			FirstName: "scarce",
			LastName:  "sudden",
			IsAdmin:   false,
		},
		{
			ID:        uuid.New(),
			Username:  "content",
			FirstName: "parcel",
			LastName:  "discover",
			IsAdmin:   true,
		},
	}, 28)
	suite.s.On("Users", mock.Anything, pagination.Params{Limit: 7, OrderDirection: "asc"}).
		Return(paginated, nil)
	defer suite.s.AssertExpectations(suite.T())

	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/?%s=7", pagination.LimitQueryParam),
		Body:   nil,
		Token: auth.Token{
			UserID:          uuid.New(),
			IsAuthenticated: true,
			Permissions:     []permission.Permission{permission.ViewUser},
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
