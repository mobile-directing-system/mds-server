package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateOperationSuite tests Controller.CreateOperation.
type ControllerCreateOperationSuite struct {
	suite.Suite
	ctrl            *ControllerMock
	sampleOperation uuid.UUID
}

func (suite *ControllerCreateOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleOperation = testutil.NewUUIDV4()
}

func (suite *ControllerCreateOperationSuite) TestCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateOperation", timeout, tx, suite.sampleOperation).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateOperation(timeout, tx, suite.sampleOperation)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("CreateOperation", timeout, tx, suite.sampleOperation).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateOperation(timeout, tx, suite.sampleOperation)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_CreateOperation(t *testing.T) {
	suite.Run(t, new(ControllerCreateOperationSuite))
}

// ControllerUpdateOperationMembersByOperationSuite tests Controller.UpdateOperationMembersByOperation.
type ControllerUpdateOperationMembersByOperationSuite struct {
	suite.Suite
	ctrl                     *ControllerMock
	sampleOperation          uuid.UUID
	sampleMembers            []uuid.UUID
	sampleOldMembersWithMore []uuid.UUID
	sampleRemovedMembers     []uuid.UUID
	sampleOperationGroups    pagination.Paginated[store.Group]
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleOperation = testutil.NewUUIDV4()
	suite.sampleMembers = make([]uuid.UUID, 16)
	for i := range suite.sampleMembers {
		suite.sampleMembers[i] = testutil.NewUUIDV4()
	}
	suite.sampleOldMembersWithMore = make([]uuid.UUID, 0, len(suite.sampleMembers))
	for _, member := range suite.sampleMembers {
		suite.sampleOldMembersWithMore = append(suite.sampleOldMembersWithMore, member)
	}
	suite.sampleRemovedMembers = make([]uuid.UUID, 4)
	for i := range suite.sampleRemovedMembers {
		suite.sampleRemovedMembers[i] = testutil.NewUUIDV4()
	}
	suite.sampleOldMembersWithMore = append(suite.sampleOldMembersWithMore, suite.sampleRemovedMembers...)
	groups := make([]store.Group, 0)
	// Add a group for each removed member.
	for _, member := range suite.sampleRemovedMembers {
		groups = append(groups, store.Group{
			ID:          testutil.NewUUIDV4(),
			Title:       "meow",
			Description: "ola",
			Operation:   nulls.NewUUID(suite.sampleOperation),
			Members:     []uuid.UUID{member},
		})
	}
	// Add a group with no removed members.
	groups = append(groups, store.Group{
		ID:        testutil.NewUUIDV4(),
		Title:     "woof",
		Operation: nulls.NewUUID(suite.sampleOperation),
		Members:   suite.sampleMembers,
	})
	// Add a group with all members.
	groups = append(groups, store.Group{
		ID:          testutil.NewUUIDV4(),
		Title:       "with all",
		Description: "",
		Operation:   nulls.NewUUID(suite.sampleOperation),
		Members:     suite.sampleOldMembersWithMore,
	})
	suite.sampleOperationGroups = pagination.NewPaginated(pagination.Params{}, groups, len(groups))
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestRetrieveOperationMembersFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, tx, suite.sampleOperation).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, tx, suite.sampleOperation, suite.sampleMembers)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestRetrieveGroupsFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, tx, suite.sampleOperation).
		Return(suite.sampleOldMembersWithMore, nil)
	suite.ctrl.Store.On("Groups", timeout, tx, mock.Anything, mock.Anything).
		Return(pagination.Paginated[store.Group]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, tx, suite.sampleOperation, suite.sampleMembers)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestUpdateGroupFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, tx, suite.sampleOperation).
		Return(suite.sampleOldMembersWithMore, nil)
	suite.ctrl.Store.On("Groups", timeout, tx, mock.Anything, mock.Anything).
		Return(suite.sampleOperationGroups, nil)
	suite.ctrl.Store.On("UpdateGroup", timeout, tx, mock.Anything).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, tx, suite.sampleOperation, suite.sampleMembers)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, tx, suite.sampleOperation).
		Return(suite.sampleMembers, nil)
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, tx, suite.sampleOperation, suite.sampleMembers).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, tx, suite.sampleOperation, suite.sampleMembers)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestNotifyUpdatedGroupFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, tx, suite.sampleOperation).
		Return(suite.sampleOldMembersWithMore, nil)
	suite.ctrl.Store.On("Groups", timeout, tx, mock.Anything, mock.Anything).
		Return(suite.sampleOperationGroups, nil)
	suite.ctrl.Store.On("UpdateGroup", timeout, tx, mock.Anything).
		Return(nil)
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, tx, suite.sampleOperation, suite.sampleMembers).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupUpdated", timeout, tx, mock.Anything).Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, tx, suite.sampleOperation, suite.sampleMembers)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestOKNoRemovedMembers() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, tx, suite.sampleOperation).
		Return(suite.sampleMembers, nil)
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, tx, suite.sampleOperation, suite.sampleMembers).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, tx, suite.sampleOperation, suite.sampleMembers)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestOKWithRemovedMembers() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, tx, suite.sampleOperation).
		Return(suite.sampleOldMembersWithMore, nil)
	suite.ctrl.Store.On("Groups", timeout, tx, mock.Anything, mock.Anything).
		Return(suite.sampleOperationGroups, nil)
	updatedGroups := make([]store.Group, 0)
	for _, group := range suite.sampleOperationGroups.Entries {
		// If the group does not contain removed ones, we can skip.
		membersWithoutRemoved := make([]uuid.UUID, 0, len(group.Members))
		for _, groupMember := range group.Members {
			removed := false
			for _, removedMember := range suite.sampleRemovedMembers {
				if removedMember == groupMember {
					removed = true
					break
				}
			}
			if !removed {
				membersWithoutRemoved = append(membersWithoutRemoved, groupMember)
			}
		}
		if len(membersWithoutRemoved) == len(group.Members) {
			continue
		}
		updatedGroup := group
		updatedGroup.Members = membersWithoutRemoved
		suite.ctrl.Store.On("UpdateGroup", timeout, tx, updatedGroup).
			Return(nil).Once()
		updatedGroups = append(updatedGroups, updatedGroup)
	}
	suite.ctrl.Store.On("UpdateOperationMembersByOperation", timeout, tx, suite.sampleOperation, suite.sampleMembers).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	for _, group := range updatedGroups {
		suite.ctrl.Notifier.On("NotifyGroupUpdated", timeout, tx, group).Return(nil).Once()
	}
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())
	suite.Require().NotEmpty(updatedGroups, "invalid test")

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateOperationMembersByOperation(timeout, tx, suite.sampleOperation, suite.sampleMembers)
		suite.NoError(err, "should fail")
	}()

	wait()
}

func TestController_UpdateOperationMembersByOperation(t *testing.T) {
	suite.Run(t, new(ControllerUpdateOperationMembersByOperationSuite))
}
