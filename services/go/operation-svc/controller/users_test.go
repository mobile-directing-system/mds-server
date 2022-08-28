package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/operation-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
	"time"
)

// ControllerCreateUserSuite tests Controller.CreateUser.
type ControllerCreateUserSuite struct {
	suite.Suite
	ctrl       *ControllerMock
	sampleUser store.User
}

func (suite *ControllerCreateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUser = store.User{
		ID:        testutil.NewUUIDV4(),
		Username:  "arrow",
		FirstName: "tire",
		LastName:  "provide",
	}
}

func (suite *ControllerCreateUserSuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, &testutil.DBTx{}, suite.sampleUser)
		suite.Error(err, "should fail")
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
		err := suite.ctrl.Ctrl.CreateUser(timeout, &testutil.DBTx{}, suite.sampleUser)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_CreateUser(t *testing.T) {
	suite.Run(t, new(ControllerCreateUserSuite))
}

// ControllerUpdateUserSuite tests Controller.UpdateUser.
type ControllerUpdateUserSuite struct {
	suite.Suite
	ctrl       *ControllerMock
	sampleUser store.User
}

func (suite *ControllerUpdateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUser = store.User{
		ID:        testutil.NewUUIDV4(),
		Username:  "arrow",
		FirstName: "tire",
		LastName:  "provide",
	}
}

func (suite *ControllerUpdateUserSuite) TestUpdateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, &testutil.DBTx{}, suite.sampleUser)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateUserSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdateUser", timeout, suite.ctrl.DB.Tx[0], suite.sampleUser).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateUser(timeout, &testutil.DBTx{}, suite.sampleUser)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_UpdateUser(t *testing.T) {
	suite.Run(t, new(ControllerUpdateUserSuite))
}

// ControllerDeleteUserByIDSuite tests Controller.DeleteUserByID.
type ControllerDeleteUserByIDSuite struct {
	suite.Suite
	ctrl                     *ControllerMock
	sampleUserID             uuid.UUID
	sampleOperations         []store.Operation
	sampleMembersByOperation map[uuid.UUID][]store.User
}

func (suite *ControllerDeleteUserByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUserID = testutil.NewUUIDV4()
	suite.sampleOperations = make([]store.Operation, 16)
	rand.Seed(0)
	suite.sampleMembersByOperation = make(map[uuid.UUID][]store.User)
	for operationIndex := range suite.sampleOperations {
		operation := store.Operation{
			ID:          testutil.NewUUIDV4(),
			Title:       "rejoice",
			Description: "pig",
			Start:       time.Time{},
			End:         nulls.Time{},
			IsArchived:  false,
		}
		suite.sampleOperations[operationIndex] = operation
		members := make([]store.User, 16)
		userPos := rand.Intn(len(members))
		for memberIndex := range members {
			member := store.User{
				ID:        testutil.NewUUIDV4(),
				Username:  "justice",
				FirstName: "song",
				LastName:  "year",
			}
			if memberIndex == userPos {
				member.ID = suite.sampleUserID
			}
			members[memberIndex] = member
		}
		suite.sampleMembersByOperation[operation.ID] = members
	}
}

func (suite *ControllerDeleteUserByIDSuite) TestRetrieveCurrentOperationsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, &testutil.DBTx{}, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestRetrieveOperationMembersFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(suite.sampleOperations, nil)
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, &testutil.DBTx{}, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestUpdateOperationMembersFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(suite.sampleOperations, nil)
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(suite.sampleMembersByOperation[suite.sampleOperations[0].ID], nil)
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, &testutil.DBTx{}, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestDeleteUserFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(suite.sampleOperations, nil)
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(suite.sampleMembersByOperation[suite.sampleOperations[0].ID], nil)
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, &testutil.DBTx{}, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestNotifyUpdatedMembersFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(suite.sampleOperations, nil)
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(suite.sampleMembersByOperation[suite.sampleOperations[0].ID], nil)
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyOperationMembersUpdated", timeout, suite.ctrl.DB.Tx[0], mock.Anything, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, &testutil.DBTx{}, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationsByMember", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(suite.sampleOperations, nil)
	// Expectations for all operations to update.
	for i := range suite.sampleOperations {
		operation := suite.sampleOperations[i]
		// Store stuff.
		suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], operation.ID).
			Return(suite.sampleMembersByOperation[operation.ID], nil).Once()
		newMembers := make([]uuid.UUID, 0)
		for _, member := range suite.sampleMembersByOperation[operation.ID] {
			if member.ID != suite.sampleUserID {
				newMembers = append(newMembers, member.ID)
			}
		}
		suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], operation.ID, newMembers).
			Return(nil).Once()
		// Notification.
		suite.ctrl.Notifier.On("NotifyOperationMembersUpdated", timeout, suite.ctrl.DB.Tx[0], operation.ID, newMembers).
			Return(nil).Once()
	}
	suite.ctrl.Store.On("DeleteUserByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleUserID).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, &testutil.DBTx{}, suite.sampleUserID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_DeleteUserByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteUserByIDSuite))
}
