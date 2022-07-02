package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerProxySuite tests Controller.Proxy.
type ControllerProxySuite struct {
	suite.Suite
	ctrl              *ControllerMock
	sampleToken       string
	sampleUser        store.UserWithPass
	samplePermissions []permission.Permission
}

func (suite *ControllerProxySuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleToken = "birth"
	suite.sampleUser = store.UserWithPass{
		User: store.User{
			ID:       testutil.NewUUIDV4(),
			Username: "avoid",
			IsAdmin:  true,
		},
		Pass: []byte("gold"),
	}
	suite.samplePermissions = []permission.Permission{{
		Name:    permission.UpdatePermissionsPermissionName,
		Options: nulls.JSONRawMessage{},
	}}
}

func (suite *ControllerProxySuite) parseAuthToken(authTokenStr string) auth.Token {
	authToken, err := auth.ParseJWTToken(authTokenStr, suite.ctrl.Ctrl.AuthTokenSecret)
	suite.Require().NoError(err, "parse jwt token should not fail")
	return authToken
}

func (suite *ControllerProxySuite) TestEmptyToken() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.sampleToken = ""

	go func() {
		defer cancel()
		authTokenStr, err := suite.ctrl.Ctrl.Proxy(timeout, suite.sampleToken)
		suite.Require().NoError(err, "should not fail")
		authToken := suite.parseAuthToken(authTokenStr)
		suite.False(authToken.IsAuthenticated)
	}()

	wait()
}

func (suite *ControllerProxySuite) TestRetrieveUserIDFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UserIDBySessionToken", timeout, suite.ctrl.DB, suite.sampleToken).
		Return(uuid.Nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Proxy(timeout, suite.sampleToken)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerProxySuite) TestSessionTokenNotFound() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.Store.On("UserIDBySessionToken", timeout, suite.ctrl.DB, suite.sampleToken).
		Return(uuid.Nil, meh.NewNotFoundErr("not found", nil))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		authTokenStr, err := suite.ctrl.Ctrl.Proxy(timeout, suite.sampleToken)
		suite.Require().NoError(err, "should not fail")
		authToken := suite.parseAuthToken(authTokenStr)
		suite.False(authToken.IsAuthenticated)
	}()

	wait()
}

func (suite *ControllerProxySuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true
	suite.ctrl.Store.On("UserIDBySessionToken", timeout, suite.ctrl.DB, suite.sampleToken).
		Return(suite.sampleUser.ID, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Proxy(timeout, suite.sampleToken)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerProxySuite) TestRetrieveUserDetailsFromStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserIDBySessionToken", timeout, suite.ctrl.DB, suite.sampleToken).
		Return(suite.sampleUser.ID, nil)
	suite.ctrl.Store.On("UserWithPassByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser.ID).
		Return(store.UserWithPass{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Proxy(timeout, suite.sampleToken)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	wait()
}

func (suite *ControllerProxySuite) TestRetrievePermissionsFromStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserIDBySessionToken", timeout, suite.ctrl.DB, suite.sampleToken).
		Return(suite.sampleUser.ID, nil)
	suite.ctrl.Store.On("UserWithPassByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser.ID).
		Return(suite.sampleUser, nil)
	suite.ctrl.Store.On("PermissionsByUserID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser.ID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Proxy(timeout, suite.sampleToken)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	wait()
}

func (suite *ControllerProxySuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserIDBySessionToken", timeout, suite.ctrl.DB, suite.sampleToken).
		Return(suite.sampleUser.ID, nil)
	suite.ctrl.Store.On("UserWithPassByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser.ID).
		Return(suite.sampleUser, nil)
	suite.ctrl.Store.On("PermissionsByUserID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser.ID).
		Return(suite.samplePermissions, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		authTokenStr, err := suite.ctrl.Ctrl.Proxy(timeout, suite.sampleToken)
		suite.Require().NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
		authToken := suite.parseAuthToken(authTokenStr)
		suite.Equal(suite.sampleUser.ID, authToken.UserID, "should have set user id in auth token correctly")
		suite.Equal(suite.sampleUser.User.Username, authToken.Username, "should have set username in auth token correctly")
		suite.Equal(suite.sampleUser.IsAdmin, authToken.IsAdmin, "should have set is-admin in auth token correctly")
		suite.True(authToken.IsAuthenticated, "should have set is-authenticated in auth token correctly")
		suite.Equal(suite.samplePermissions, authToken.Permissions, "should have set permissions in auth token correctly")
		suite.NotEmpty(authToken.RandomSalt, "should have set random salt in auth token")
	}()

	wait()
}

func TestController_Proxy(t *testing.T) {
	suite.Run(t, new(ControllerProxySuite))
}
