package controller

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/lefinal/zaprec"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/search"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/mobile-directing-system/mds-server/services/go/user-svc/store"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
	"testing"
)

// ControllerCreateUserSuite tests Controller.CreateUser.
type ControllerCreateUserSuite struct {
	suite.Suite
	ctrl       *ControllerMock
	createUser store.UserWithPass
}

func (suite *ControllerCreateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.createUser = store.UserWithPass{
		User: store.User{
			ID:        testutil.NewUUIDV4(),
			Username:  "duty",
			FirstName: "song",
			LastName:  "twist",
			IsAdmin:   true,
		},
		Pass: []byte("meow"),
	}
}

func (suite *ControllerCreateUserSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateUser(timeout, suite.createUser)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerCreateUserSuite) TestStoreCreateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateUser", timeout, suite.ctrl.DB.Tx[0], suite.createUser).
		Return(store.User{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateUser(timeout, suite.createUser)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerCreateUserSuite) TestNotifyFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	created := suite.createUser
	created.ID = testutil.NewUUIDV4()
	suite.ctrl.Store.On("CreateUser", timeout, suite.ctrl.DB.Tx[0], suite.createUser).Return(created.User, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserCreated", timeout, suite.ctrl.DB.Tx[0], created).Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateUser(timeout, suite.createUser)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerCreateUserSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	created := suite.createUser
	created.ID = testutil.NewUUIDV4()
	suite.ctrl.Store.On("CreateUser", timeout, suite.ctrl.DB.Tx[0], suite.createUser).Return(created.User, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserCreated", timeout, suite.ctrl.DB.Tx[0], created).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.CreateUser(timeout, suite.createUser)
		suite.Require().NoError(err, "should fail")
		suite.Equal(created, got, "should return correct value")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestController_CreateUser(t *testing.T) {
	suite.Run(t, new(ControllerCreateUserSuite))
}

// ControllerUpdateUserSuite tests Controller.UpdateUser.
type ControllerUpdateUserSuite struct {
	suite.Suite
	ctrl       *ControllerMock
	updateUser store.User
}

func (suite *ControllerUpdateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.updateUser = store.User{
		ID:        testutil.NewUUIDV4(),
		Username:  "cook",
		FirstName: "high",
		LastName:  "preserve",
		IsAdmin:   false,
		IsActive:  true,
	}
}

func (suite *ControllerUpdateUserSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserSuite) TestRetrieveFromStoreFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).
		Return(store.User{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserSuite) TestAdminChangeNotAllowed() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	originalUser := suite.updateUser
	originalUser.IsAdmin = false
	suite.updateUser.IsAdmin = true
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserSuite) TestForbiddenAdminUsernameChange() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	originalUser := suite.updateUser
	originalUser.Username = adminUsername
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not timeout")
}

func (suite *ControllerUpdateUserSuite) TestForbiddenAdminUsernameChangeWithAllow() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	originalUser := suite.updateUser
	originalUser.Username = adminUsername
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, true, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not timeout")
}

func (suite *ControllerUpdateUserSuite) TestForbiddenAdminChange() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.updateUser.Username = adminUsername
	suite.updateUser.IsAdmin = false
	originalUser := suite.updateUser
	originalUser.IsAdmin = true
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserSuite) TestForbiddenSetAdminToNonAdminWithAllow() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.updateUser.Username = adminUsername
	suite.updateUser.IsAdmin = false
	originalUser := suite.updateUser
	originalUser.IsAdmin = true
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, true, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserSuite) TestForbiddenAdminDisable() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.updateUser.Username = adminUsername
	suite.updateUser.IsActive = false
	suite.updateUser.IsAdmin = true
	originalUser := suite.updateUser
	originalUser.IsAdmin = true
	originalUser.IsActive = true
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserSuite) TestForbiddenAdminDisableWithAllow() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.updateUser.Username = adminUsername
	suite.updateUser.IsActive = false
	suite.updateUser.IsAdmin = true
	originalUser := suite.updateUser
	originalUser.IsAdmin = true
	originalUser.IsActive = true
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, true, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserSuite) TestForbiddenActiveStateChange() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	originalUser := suite.updateUser
	originalUser.IsActive = !suite.updateUser.IsActive
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not timeout")
}

func (suite *ControllerUpdateUserSuite) TestStoreUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	originalUser := suite.updateUser
	originalUser.Username = "force"
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserSuite) TestNotifyFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	originalUser := suite.updateUser
	originalUser.Username = "force"
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserUpdated", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserSuite) TestOKSetAdmin() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.updateUser.IsAdmin = true
	originalUser := suite.updateUser
	originalUser.IsAdmin = false
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserUpdated", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, true, false)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not fail")
}

func (suite *ControllerUpdateUserSuite) TestOKSetActiveState() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	originalUser := suite.updateUser
	originalUser.IsActive = !suite.updateUser.IsActive
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserUpdated", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, true)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not fail")
}

func (suite *ControllerUpdateUserSuite) TestOKSetNonAdmin() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.updateUser.IsAdmin = false
	originalUser := suite.updateUser
	originalUser.IsAdmin = true
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserUpdated", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, true, false)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not fail")
}

func (suite *ControllerUpdateUserSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	originalUser := suite.updateUser
	originalUser.Username = "faith"
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.updateUser.ID).Return(originalUser, nil)
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserUpdated", timeout, suite.ctrl.DB.Tx[0], suite.updateUser).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, suite.updateUser, false, false)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not fail")
}

func TestController_UpdateUser(t *testing.T) {
	suite.Run(t, new(ControllerUpdateUserSuite))
}

// ControllerUpdateUserPassByUserIDSuite tests
// Controller.UpdateUserPassByUserID.
type ControllerUpdateUserPassByUserIDSuite struct {
	suite.Suite
	ctrl    *ControllerMock
	userID  uuid.UUID
	newPass []byte
}

func (suite *ControllerUpdateUserPassByUserIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.userID = testutil.NewUUIDV4()
	suite.newPass = []byte("meow")
}

func (suite *ControllerUpdateUserPassByUserIDSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUserPassByUserID(timeout, suite.userID, suite.newPass)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserPassByUserIDSuite) TestStoreUpdateFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateUserPassByUserID", timeout, suite.ctrl.DB.Tx[0], suite.userID, suite.newPass).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUserPassByUserID(timeout, suite.userID, suite.newPass)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserPassByUserIDSuite) TestNotifyFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateUserPassByUserID", timeout, suite.ctrl.DB.Tx[0], suite.userID, suite.newPass).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserPassUpdated", timeout, suite.ctrl.DB.Tx[0], suite.userID, suite.newPass).Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUserPassByUserID(timeout, suite.userID, suite.newPass)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUpdateUserPassByUserIDSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateUserPassByUserID", timeout, suite.ctrl.DB.Tx[0], suite.userID, suite.newPass).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserPassUpdated", timeout, suite.ctrl.DB.Tx[0], suite.userID, suite.newPass).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUserPassByUserID(timeout, suite.userID, suite.newPass)
		suite.NoError(err, "should fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestController_UpdateUserPassByUserID(t *testing.T) {
	suite.Run(t, new(ControllerUpdateUserPassByUserIDSuite))
}

// ControllerDeleteUserByIDSuite tests Controller.DeleteUserByID.
type ControllerDeleteUserByIDSuite struct {
	suite.Suite
	ctrl              *ControllerMock
	userID            uuid.UUID
	sampleUser        store.User
	sampleUpdatedUser store.User
}

func (suite *ControllerDeleteUserByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.userID = testutil.NewUUIDV4()
	suite.sampleUser = store.User{
		ID:        testutil.NewUUIDV4(),
		Username:  "nothing",
		FirstName: "overflow",
		LastName:  "sadden",
		IsAdmin:   false,
		IsActive:  true,
	}
	suite.sampleUpdatedUser = suite.sampleUser
	suite.sampleUpdatedUser.IsActive = false
}

func (suite *ControllerDeleteUserByIDSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.SetUserInactiveByID(timeout, suite.userID)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerDeleteUserByIDSuite) TestStoreRetrievalFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.userID).
		Return(store.User{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.SetUserInactiveByID(timeout, suite.userID)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

// TestDeleteAdmin assures that the admin user is not deleted.
func (suite *ControllerDeleteUserByIDSuite) TestDeleteAdmin() {
	original := suite.sampleUpdatedUser
	original.Username = adminUsername
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.userID).
		Return(original, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.SetUserInactiveByID(timeout, suite.userID)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerDeleteUserByIDSuite) TestStoreDeleteFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.userID).
		Return(suite.sampleUpdatedUser, nil)
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUpdatedUser).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.SetUserInactiveByID(timeout, suite.userID)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerDeleteUserByIDSuite) TestNotifyFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.userID).
		Return(suite.sampleUpdatedUser, nil)
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUpdatedUser).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleUpdatedUser).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.SetUserInactiveByID(timeout, suite.userID)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerDeleteUserByIDSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.userID).
		Return(suite.sampleUpdatedUser, nil)
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUpdatedUser).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyUserUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleUpdatedUser).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.SetUserInactiveByID(timeout, suite.userID)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestController_DeleteUserByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteUserByIDSuite))
}

// ControllerUserByIDSuite tests Controller.UserByID.
type ControllerUserByIDSuite struct {
	suite.Suite
	ctrl *ControllerMock
	user store.User
}

func (suite *ControllerUserByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.user = store.User{
		ID:        testutil.NewUUIDV4(),
		Username:  "fond",
		FirstName: "shop",
		LastName:  "defend",
		IsAdmin:   true,
	}
}

func (suite *ControllerUserByIDSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.UserByID(timeout, suite.user.ID)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUserByIDSuite) TestStoreRetrievalFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.user.ID).
		Return(store.User{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.UserByID(timeout, suite.user.ID)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUserByIDSuite) TestOK() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UserByID", timeout, suite.ctrl.DB.Tx[0], suite.user.ID).Return(suite.user, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.UserByID(timeout, suite.user.ID)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.user, got, "should return correct user")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestController_UserByID(t *testing.T) {
	suite.Run(t, new(ControllerUserByIDSuite))
}

// ControllerUsersSuite tests Controller.Users.
type ControllerUsersSuite struct {
	suite.Suite
	ctrl          *ControllerMock
	sampleParams  pagination.Params
	sampleFilters store.UserFilters
}

func (suite *ControllerUsersSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleFilters = store.UserFilters{
		IncludeInactive: true,
	}
	suite.sampleParams = pagination.Params{
		Offset:         89,
		OrderBy:        nulls.NewString("hi"),
		OrderDirection: "desc",
	}
}

func (suite *ControllerUsersSuite) TestTxFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Users(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUsersSuite) TestStoreRetrievalFail() {
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Users", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(pagination.Paginated[store.User]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Users(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func (suite *ControllerUsersSuite) TestOK() {
	paginated := pagination.NewPaginated(suite.sampleParams, []store.User{
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "society",
			FirstName: "belief",
			LastName:  "wall",
			IsAdmin:   false,
		},
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "basket",
			FirstName: "letter",
			LastName:  "flag",
			IsAdmin:   true,
		},
	}, 27)
	timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Users", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).Return(paginated, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.Users(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(paginated, got, "should return correct result")
	}()

	<-timeout.Done()
	suite.NotEqual(context.DeadlineExceeded, timeout.Err(), "should not time out")
}

func TestController_Users(t *testing.T) {
	suite.Run(t, new(ControllerUsersSuite))
}

// ControllerSearchUsersSuite tests Controller.SearchUsers.
type ControllerSearchUsersSuite struct {
	suite.Suite
	ctrl          *ControllerMock
	sampleFilters store.UserFilters
	sampleParams  search.Params
	sampleUsers   []store.User
}

func (suite *ControllerSearchUsersSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleFilters = store.UserFilters{
		IncludeInactive: true,
	}
	suite.sampleParams = search.Params{
		Query:  "freedom",
		Offset: 734,
		Limit:  389,
	}
	suite.sampleUsers = []store.User{
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "possess",
			FirstName: "among",
			LastName:  "center",
			IsAdmin:   true,
		},
		{
			ID:        testutil.NewUUIDV4(),
			Username:  "anybody",
			FirstName: "bitter",
			LastName:  "law",
			IsAdmin:   false,
		},
	}
}

func (suite *ControllerSearchUsersSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.SearchUsers(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerSearchUsersSuite) TestStoreSearchFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("SearchUsers", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(search.Result[store.User]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.SearchUsers(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerSearchUsersSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("SearchUsers", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(search.Result[store.User]{Hits: suite.sampleUsers}, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.SearchUsers(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleUsers, got.Hits, "should return correct result")
	}()

	wait()
}

func TestController_SearchUsers(t *testing.T) {
	suite.Run(t, new(ControllerSearchUsersSuite))
}

// ControllerRebuildUserSearchSuite tests Controller.RebuildUserSearch.
type ControllerRebuildUserSearchSuite struct {
	suite.Suite
	ctrl     *ControllerMock
	recorder *zaprec.RecordStore
}

func (suite *ControllerRebuildUserSearchSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.ctrl.Logger, suite.recorder = zaprec.NewRecorder(zapcore.ErrorLevel)
	suite.ctrl.Ctrl.Logger = suite.ctrl.Logger
}

func (suite *ControllerRebuildUserSearchSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildUserSearch(timeout)
		suite.Len(suite.recorder.Records(), 1, "should have logged error")
	}()

	wait()
}

func (suite *ControllerRebuildUserSearchSuite) TestStoreRebuildFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("RebuildUserSearch", timeout, suite.ctrl.DB.Tx[0]).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildUserSearch(timeout)
		suite.Len(suite.recorder.Records(), 1, "should have logged error")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not have committed tx")
	}()

	wait()
}

func (suite *ControllerRebuildUserSearchSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("RebuildUserSearch", timeout, suite.ctrl.DB.Tx[0]).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		suite.ctrl.Ctrl.RebuildUserSearch(timeout)
		suite.Len(suite.recorder.Records(), 0, "should not have logged error")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should have committed tx")
	}()

	wait()
}

func TestController_RebuildUserSearch(t *testing.T) {
	suite.Run(t, new(ControllerRebuildUserSearchSuite))
}
