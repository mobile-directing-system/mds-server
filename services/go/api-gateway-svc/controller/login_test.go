package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

// Test_generatePublicSessionToken tests generatePublicSessionToken.
func Test_generatePublicSessionToken(t *testing.T) {
	token, err := generatePublicSessionToken("meow", "ola")
	require.NoError(t, err, "should not fail")
	assert.NotEmpty(t, token, "token should not be empty")
}

// ControllerLoginSuite tests Controller.Login.
type ControllerLoginSuite struct {
	suite.Suite
	ctrl *ControllerMock

	sampleUsername        string
	sampleUserPass        string
	sampleUserPassHashed  []byte
	sampleRequestMetadata AuthRequestMetadata
	sampleUser            store.UserWithPass
}

func (suite *ControllerLoginSuite) SetupSuite() {
	suite.sampleUsername = "meow"
	suite.sampleUserPass = "woof"
	var err error
	suite.sampleUserPassHashed, err = auth.HashPassword(suite.sampleUserPass)
	if err != nil {
		panic(err)
	}
	suite.sampleRequestMetadata = AuthRequestMetadata{
		Host:       "court",
		UserAgent:  "pad",
		RemoteAddr: "borrow",
	}
}

func (suite *ControllerLoginSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUser = store.UserWithPass{
		User: store.User{
			ID:       testutil.NewUUIDV4(),
			Username: suite.sampleUsername,
			IsAdmin:  false,
		},
		Pass: suite.sampleUserPassHashed,
	}
}

func (suite *ControllerLoginSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, _, _, err := suite.ctrl.Ctrl.Login(timeout, suite.sampleUsername, suite.sampleUserPass, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerLoginSuite) TestRetrieveUserFromStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserWithPassByUsername", timeout, suite.ctrl.DB.Tx[0], suite.sampleUsername).
		Return(store.UserWithPass{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, _, err := suite.ctrl.Ctrl.Login(timeout, suite.sampleUsername, suite.sampleUserPass, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerLoginSuite) TestPasswordCheckFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	user := suite.sampleUser
	user.Pass = []byte("meow")
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserWithPassByUsername", timeout, suite.ctrl.DB.Tx[0], suite.sampleUsername).
		Return(user, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, _, err := suite.ctrl.Ctrl.Login(timeout, suite.sampleUsername, "nonono", suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerLoginSuite) TestPasswordMismatch() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserWithPassByUsername", timeout, suite.ctrl.DB.Tx[0], suite.sampleUsername).
		Return(suite.sampleUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, ok, err := suite.ctrl.Ctrl.Login(timeout, suite.sampleUsername, "nonono", suite.sampleRequestMetadata)
		suite.Require().NoError(err, "should not fail")
		suite.False(ok, "should not return ok")
	}()

	wait()
}

func (suite *ControllerLoginSuite) TestStoreSessionTokenFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserWithPassByUsername", timeout, suite.ctrl.DB.Tx[0], suite.sampleUsername).
		Return(suite.sampleUser, nil)
	suite.ctrl.Store.On("StoreSessionTokenForUser", timeout, suite.ctrl.DB.Tx[0], mock.Anything, suite.sampleUser.ID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, _, err := suite.ctrl.Ctrl.Login(timeout, suite.sampleUsername, suite.sampleUserPass, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerLoginSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserWithPassByUsername", timeout, suite.ctrl.DB.Tx[0], suite.sampleUsername).
		Return(suite.sampleUser, nil)
	suite.ctrl.Store.On("StoreSessionTokenForUser", timeout, suite.ctrl.DB.Tx[0], mock.Anything, suite.sampleUser.ID).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserLoggedIn", suite.sampleUser.ID, suite.sampleUsername, suite.sampleRequestMetadata).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, _, _, err := suite.ctrl.Ctrl.Login(timeout, suite.sampleUsername, suite.sampleUserPass, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerLoginSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserWithPassByUsername", timeout, suite.ctrl.DB.Tx[0], suite.sampleUsername).
		Return(suite.sampleUser, nil)
	suite.ctrl.Store.On("StoreSessionTokenForUser", timeout, suite.ctrl.DB.Tx[0], mock.Anything, suite.sampleUser.ID).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserLoggedIn", suite.sampleUser.ID, suite.sampleUsername, suite.sampleRequestMetadata).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		userID, token, ok, err := suite.ctrl.Ctrl.Login(timeout, suite.sampleUsername, suite.sampleUserPass, suite.sampleRequestMetadata)
		suite.Require().NoError(err, "should not fail")
		suite.True(ok, "should return ok")
		suite.Equal(suite.sampleUser.ID, userID, "should return correct user id")
		suite.NotEmpty(token, "should return token")
	}()

	wait()
}

func TestController_Login(t *testing.T) {
	suite.Run(t, new(ControllerLoginSuite))
}

// ControllerLogoutSuite tests Controller.Logout.
type ControllerLogoutSuite struct {
	suite.Suite
	ctrl                  *ControllerMock
	sampleToken           string
	sampleRequestMetadata AuthRequestMetadata
	sampleUser            store.UserWithPass
}

func (suite *ControllerLogoutSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleToken = "xyz"
	suite.sampleRequestMetadata = AuthRequestMetadata{
		Host:       "meantime",
		UserAgent:  "behavior",
		RemoteAddr: "between",
	}
	suite.sampleUser = store.UserWithPass{
		User: store.User{
			ID:       testutil.NewUUIDV4(),
			Username: "caution",
			IsAdmin:  true,
		},
		Pass: []byte("ticket"),
	}
}

func (suite *ControllerLogoutSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.Logout(timeout, suite.sampleToken, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerLogoutSuite) TestDeleteUserIDBySessionTokenFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("GetAndDeleteUserIDBySessionToken", timeout, suite.ctrl.DB.Tx[0], suite.sampleToken).
		Return(uuid.Nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.Logout(timeout, suite.sampleToken, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	wait()
}

func (suite *ControllerLogoutSuite) TestUserWithPassByIDFromStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("GetAndDeleteUserIDBySessionToken", timeout, suite.ctrl.DB.Tx[0], suite.sampleToken).
		Return(suite.sampleUser.ID, nil)
	suite.ctrl.Store.On("UserWithPassByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser.ID).
		Return(store.UserWithPass{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.Logout(timeout, suite.sampleToken, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	wait()
}

func (suite *ControllerLogoutSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("GetAndDeleteUserIDBySessionToken", timeout, suite.ctrl.DB.Tx[0], suite.sampleToken).
		Return(suite.sampleUser.ID, nil)
	suite.ctrl.Store.On("UserWithPassByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser.ID).
		Return(suite.sampleUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserLoggedOut", suite.sampleUser.ID, suite.sampleUser.Username, suite.sampleRequestMetadata).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.Logout(timeout, suite.sampleToken, suite.sampleRequestMetadata)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	wait()
}

func (suite *ControllerLogoutSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("GetAndDeleteUserIDBySessionToken", timeout, suite.ctrl.DB.Tx[0], suite.sampleToken).
		Return(suite.sampleUser.ID, nil)
	suite.ctrl.Store.On("UserWithPassByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser.ID).
		Return(suite.sampleUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserLoggedOut", suite.sampleUser.ID, suite.sampleUser.Username, suite.sampleRequestMetadata).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.Logout(timeout, suite.sampleToken, suite.sampleRequestMetadata)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	wait()
}

func TestController_Logout(t *testing.T) {
	suite.Run(t, new(ControllerLogoutSuite))
}
