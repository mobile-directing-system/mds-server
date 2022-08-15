package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/group-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pagination"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateGroupSuite tests Controller.CreateGroup.
type ControllerCreateGroupSuite struct {
	suite.Suite
	ctrl                     *ControllerMock
	sampleGroup              store.Group
	sampleOperationMembersOK []uuid.UUID
}

func (suite *ControllerCreateGroupSuite) SetupTest() {
	suite.ctrl = NewMockController()
	members := make([]uuid.UUID, 16)
	suite.sampleOperationMembersOK = make([]uuid.UUID, 0, len(members))
	for i := range members {
		members[i] = testutil.NewUUIDV4()
		suite.sampleOperationMembersOK = append(suite.sampleOperationMembersOK, members[i])
		// Random user.
		suite.sampleOperationMembersOK = append(suite.sampleOperationMembersOK, testutil.NewUUIDV4())
	}
	suite.sampleGroup = store.Group{
		Title:       "command",
		Description: "ready",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members:     members,
	}
}

func (suite *ControllerCreateGroupSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateGroupSuite) TestOperationMemberCheckFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateGroupSuite) TestGroupMemberNotPartOfOperation() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(suite.sampleGroup.Members[:len(suite.sampleGroup.Members)-1], nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateGroupSuite) TestCreateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(suite.sampleOperationMembersOK, nil)
	suite.ctrl.Store.On("CreateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(store.Group{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateGroupSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	created := suite.sampleGroup
	created.ID = testutil.NewUUIDV4()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(suite.sampleOperationMembersOK, nil)
	suite.ctrl.Store.On("CreateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(created, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupCreated", timeout, suite.ctrl.DB.Tx[0], created).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerCreateGroupSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	created := suite.sampleGroup
	created.ID = testutil.NewUUIDV4()
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(suite.sampleOperationMembersOK, nil)
	suite.ctrl.Store.On("CreateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(created, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupCreated", timeout, suite.ctrl.DB.Tx[0], created).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(created, got, "should return correct value")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_CreateGroup(t *testing.T) {
	suite.Run(t, new(ControllerCreateGroupSuite))
}

// ControllerUpdateGroupSuite tests Controller.UpdateGroup.
type ControllerUpdateGroupSuite struct {
	suite.Suite
	ctrl                     *ControllerMock
	sampleGroup              store.Group
	sampleOperationMembersOK []uuid.UUID
}

func (suite *ControllerUpdateGroupSuite) SetupTest() {
	suite.ctrl = NewMockController()
	members := make([]uuid.UUID, 16)
	suite.sampleOperationMembersOK = make([]uuid.UUID, 0, len(members))
	for i := range members {
		members[i] = testutil.NewUUIDV4()
		suite.sampleOperationMembersOK = append(suite.sampleOperationMembersOK, members[i])
		// Random user.
		suite.sampleOperationMembersOK = append(suite.sampleOperationMembersOK, testutil.NewUUIDV4())
	}
	suite.sampleGroup = store.Group{
		ID:          testutil.NewUUIDV4(),
		Title:       "command",
		Description: "ready",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members:     members,
	}
}

func (suite *ControllerUpdateGroupSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerUpdateGroupSuite) TestOperationMemberCheckFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(nil, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateGroupSuite) TestGroupMemberNotPartOfOperation() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(suite.sampleGroup.Members[:len(suite.sampleGroup.Members)-1], nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.CreateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateGroupSuite) TestUpdateInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(suite.sampleOperationMembersOK, nil)
	suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateGroupSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(suite.sampleOperationMembersOK, nil)
	suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerUpdateGroupSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("OperationMembersByOperation", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.Operation.UUID).
		Return(suite.sampleOperationMembersOK, nil)
	suite.ctrl.Store.On("UpdateGroup", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupUpdated", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.UpdateGroup(timeout, suite.sampleGroup)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_UpdateGroup(t *testing.T) {
	suite.Run(t, new(ControllerUpdateGroupSuite))
}

// ControllerDeleteGroupByIDSuite tests Controller.DeleteGroupByID.
type ControllerDeleteGroupByIDSuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleGroup uuid.UUID
}

func (suite *ControllerDeleteGroupByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleGroup = testutil.NewUUIDV4()
}

func (suite *ControllerDeleteGroupByIDSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestDeleteInStoreFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("DeleteGroupByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("DeleteGroupByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupDeleted", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, suite.sampleGroup)
		suite.Error(err, "should fail")
		suite.False(suite.ctrl.DB.Tx[0].IsCommitted, "should not commit tx")
	}()

	wait()
}

func (suite *ControllerDeleteGroupByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("DeleteGroupByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyGroupDeleted", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup).
		Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteGroupByID(timeout, suite.sampleGroup)
		suite.NoError(err, "should not fail")
		suite.True(suite.ctrl.DB.Tx[0].IsCommitted, "should commit tx")
	}()

	wait()
}

func TestController_DeleteGroupByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteGroupByIDSuite))
}

// ControllerGroupByIDSuite tests Controller.GroupByID.
type ControllerGroupByIDSuite struct {
	suite.Suite
	ctrl        *ControllerMock
	sampleGroup store.Group
}

func (suite *ControllerGroupByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	members := make([]uuid.UUID, 16)
	for i := range members {
		members[i] = testutil.NewUUIDV4()
	}
	suite.sampleGroup = store.Group{
		ID:          testutil.NewUUIDV4(),
		Title:       "marry",
		Description: "stand",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members:     members,
	}
}

func (suite *ControllerGroupByIDSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.GroupByID(timeout, suite.sampleGroup.ID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerGroupByIDSuite) TestRetrieveFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("GroupByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.ID).
		Return(store.Group{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.GroupByID(timeout, suite.sampleGroup.ID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerGroupByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("GroupByID", timeout, suite.ctrl.DB.Tx[0], suite.sampleGroup.ID).
		Return(suite.sampleGroup, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.GroupByID(timeout, suite.sampleGroup.ID)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleGroup, got, "should return correct value")
	}()

	wait()
}

func TestController_GroupByID(t *testing.T) {
	suite.Run(t, new(ControllerGroupByIDSuite))
}

// ControllerGroupsSuite tests Controller.Groups.
type ControllerGroupsSuite struct {
	suite.Suite
	ctrl          *ControllerMock
	sampleFilters store.GroupFilters
	sampleParams  pagination.Params
	sampleGroups  pagination.Paginated[store.Group]
}

func (suite *ControllerGroupsSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleFilters = store.GroupFilters{
		ByUser:        nulls.NewUUID(testutil.NewUUIDV4()),
		ForOperation:  nulls.NewUUID(testutil.NewUUIDV4()),
		ExcludeGlobal: true,
	}
	suite.sampleParams = pagination.Params{
		Limit:          923,
		Offset:         209,
		OrderBy:        nulls.NewString("brick"),
		OrderDirection: "desc",
	}
	suite.sampleGroups = pagination.NewPaginated(suite.sampleParams, []store.Group{
		{
			ID:          testutil.NewUUIDV4(),
			Title:       "sympathy",
			Description: "civilize",
			Operation:   uuid.NullUUID{},
			Members: []uuid.UUID{
				testutil.NewUUIDV4(),
				testutil.NewUUIDV4(),
				testutil.NewUUIDV4(),
			},
		},
		{
			ID:          testutil.NewUUIDV4(),
			Title:       "regard",
			Description: "throw",
			Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
			Members: []uuid.UUID{
				testutil.NewUUIDV4(),
			},
		},
	}, 9313)
}

func (suite *ControllerGroupsSuite) TestTxFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.BeginFail = true

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Groups(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerGroupsSuite) TestRetrieveFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Groups", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(pagination.Paginated[store.Group]{}, errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		_, err := suite.ctrl.Ctrl.Groups(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerGroupsSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("Groups", timeout, suite.ctrl.DB.Tx[0], suite.sampleFilters, suite.sampleParams).
		Return(suite.sampleGroups, nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		got, err := suite.ctrl.Ctrl.Groups(timeout, suite.sampleFilters, suite.sampleParams)
		suite.Require().NoError(err, "should not fail")
		suite.Equal(suite.sampleGroups, got, "should return correct value")
	}()

	wait()
}

func TestController_Groups(t *testing.T) {
	suite.Run(t, new(ControllerGroupsSuite))
}
