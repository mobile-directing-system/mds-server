package controller

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
)

// ControllerCreateUserSuite tests Controller.CreateUser.
type ControllerCreateUserSuite struct {
	suite.Suite
	ctrl       *ControllerMock
	sampleUser uuid.UUID
}

func (suite *ControllerCreateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUser = testutil.NewUUIDV4()
}

func (suite *ControllerCreateUserSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, suite.sampleUser)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateUserSuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, suite.sampleUser)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateUserSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, suite.sampleUser)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_CreateUser(t *testing.T) {
	suite.Run(t, new(ControllerCreateUserSuite))
}

// ControllerDeleteUserByIDSuite tests Controller.DeleteUserByID.
type ControllerDeleteUserByIDSuite struct {
	suite.Suite
	ctrl         *ControllerMock
	sampleUser   uuid.UUID
	memberGroups []store.Group
}

func (suite *ControllerDeleteUserByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUser = testutil.NewUUIDV4()
	maxUsers := 16
	suite.memberGroups = make([]store.Group, 8)
	for i := range suite.memberGroups {
		group := store.Group{
			ID:          testutil.NewUUIDV4(),
			Title:       "where",
			Description: "thick",
			Operation:   uuid.NullUUID{},
			Members:     make([]uuid.UUID, rand.Intn(maxUsers)+1),
		}
		for member := range group.Members {
			group.Members[member] = testutil.NewUUIDV4()
		}
		group.Members[rand.Intn(len(group.Members))] = suite.sampleUser
		suite.memberGroups[i] = group
	}
	fmt.Println("n")
}

func (suite *ControllerDeleteUserByIDSuite) groupWithoutMember(group store.Group, without uuid.UUID) store.Group {
	// Deep copy members.
	newMembers := make([]uuid.UUID, 0, len(group.Members))
	for _, member := range group.Members {
		newMembers = append(newMembers, member)
	}
	for i, member := range newMembers {
		if member != without {
			continue
		}
		newMembers[i] = newMembers[len(newMembers)-1]
		newMembers = newMembers[:len(newMembers)-1]
	}
	group.Members = newMembers
	return group
}

func (suite *ControllerDeleteUserByIDSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUser)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestCurrentGroupsFromStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Groups", timeout, suite.ctrl.DB.Tx[0], store.GroupFilters{
		ByUser: uuid.NullUUID{UUID: suite.sampleUser, Valid: true},
	}, pagination.Params{}).Return(pagination.Paginated[store.Group]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUser)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestUpdateGroupFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Groups", timeout, suite.ctrl.DB.Tx[0], store.GroupFilters{
		ByUser: uuid.NullUUID{UUID: suite.sampleUser, Valid: true},
	}, pagination.Params{}).
		Return(pagination.NewPaginated(pagination.Params{}, suite.memberGroups, len(suite.memberGroups)), nil)
	// Fail at random call.
	fail := rand.Intn(len(suite.memberGroups))
	for i, group := range suite.memberGroups {
		var err error
		if i == fail {
			err = errors.New("sad life")
		}
		suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.groupWithoutMember(group, suite.sampleUser)).
			Return(err).Once()
		if i == fail {
			break
		}
	}
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUser)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestDeleteUserInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Groups", timeout, suite.ctrl.DB.Tx[0], store.GroupFilters{
		ByUser: uuid.NullUUID{UUID: suite.sampleUser, Valid: true},
	}, pagination.Params{}).
		Return(pagination.NewPaginated(pagination.Params{}, suite.memberGroups, len(suite.memberGroups)), nil)
	for _, group := range suite.memberGroups {
		suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.groupWithoutMember(group, suite.sampleUser)).
			Return(nil).Once()
	}
	suite.ctrl.Store.On("DeleteUserByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUser)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestNotifyUpdatedGroupsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Groups", timeout, suite.ctrl.DB.Tx[0], store.GroupFilters{
		ByUser: uuid.NullUUID{UUID: suite.sampleUser, Valid: true},
	}, pagination.Params{}).
		Return(pagination.NewPaginated(pagination.Params{}, suite.memberGroups, len(suite.memberGroups)), nil)
	for _, group := range suite.memberGroups {
		suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.groupWithoutMember(group, suite.sampleUser)).
			Return(nil).Once()
	}
	suite.ctrl.Store.On("DeleteUserByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	// Fail at random call.
	fail := rand.Intn(len(suite.memberGroups))
	for i, group := range suite.memberGroups {
		var err error
		if i == fail {
			err = errors.New("sad life")
		}
		suite.ctrl.Notifier.On("NotifyGroupUpdated", suite.groupWithoutMember(group, suite.sampleUser)).
			Return(err).Once()
	}
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUser)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Groups", timeout, suite.ctrl.DB.Tx[0], store.GroupFilters{
		ByUser: uuid.NullUUID{UUID: suite.sampleUser, Valid: true},
	}, pagination.Params{}).
		Return(pagination.NewPaginated(pagination.Params{}, suite.memberGroups, len(suite.memberGroups)), nil)
	for _, group := range suite.memberGroups {
		suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.groupWithoutMember(group, suite.sampleUser)).
			Return(nil).Once()
	}
	suite.ctrl.Store.On("DeleteUserByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	for _, group := range suite.memberGroups {
		suite.ctrl.Notifier.On("NotifyGroupUpdated", suite.groupWithoutMember(group, suite.sampleUser)).
			Return(nil).Once()
	}
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, suite.sampleUser)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_DeleteUserByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteUserByIDSuite))
}
