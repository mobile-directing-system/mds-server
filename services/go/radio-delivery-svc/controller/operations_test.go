package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

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

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
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

func (suite *ControllerUpdateOperationMembersByOperationSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
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

func TestController_UpdateOperationMembersByOperation(t *testing.T) {
	suite.Run(t, new(ControllerUpdateOperationMembersByOperationSuite))
}
