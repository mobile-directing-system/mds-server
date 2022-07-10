package store

import (
	"github.com/gofrs/uuid"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// GroupValidateSuite tests Group.Validate.
type GroupValidateSuite struct {
	suite.Suite
	g Group
}

func (suite *GroupValidateSuite) SetupTest() {
	suite.g = Group{
		ID:          testutil.NewUUIDV4(),
		Title:       "Hello",
		Description: "World",
		Operation:   nulls.NewUUID(testutil.NewUUIDV4()),
		Members:     []uuid.UUID{testutil.NewUUIDV4(), testutil.NewUUIDV4()},
	}
}

func (suite *GroupValidateSuite) TestMissingTitle() {
	suite.g.Title = ""
	report, err := suite.g.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "should not be ok")
}

func (suite *GroupValidateSuite) TestOK() {
	report, err := suite.g.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.True(report.IsOK(), "should be ok")
}

func TestGroup_Validate(t *testing.T) {
	suite.Run(t, new(GroupValidateSuite))
}

// hasDuplicateMembersSuite tests hasDuplicateMembers.
type hasDuplicateMembersSuite struct {
	suite.Suite
}

func (suite *hasDuplicateMembersSuite) TestEmpty() {
	suite.False(hasDuplicateMembers([]uuid.UUID{}))
}

func (suite *hasDuplicateMembersSuite) TestSingle() {
	suite.False(hasDuplicateMembers([]uuid.UUID{testutil.NewUUIDV4()}))
}

func (suite *hasDuplicateMembersSuite) TestOK() {
	ids := make([]uuid.UUID, 64)
	for i := range ids {
		ids[i] = testutil.NewUUIDV4()
	}
	suite.False(hasDuplicateMembers(ids))
}

func (suite *hasDuplicateMembersSuite) TestDuplicates1() {
	duplicate := testutil.NewUUIDV4()
	ids := make([]uuid.UUID, 64)
	for i := range ids {
		ids[i] = testutil.NewUUIDV4()
	}
	ids[0] = duplicate
	ids[63] = duplicate
	suite.True(hasDuplicateMembers(ids))
}

func (suite *hasDuplicateMembersSuite) TestDuplicates2() {
	duplicate := testutil.NewUUIDV4()
	ids := make([]uuid.UUID, 64)
	for i := range ids {
		ids[i] = testutil.NewUUIDV4()
	}
	ids[0] = duplicate
	ids[1] = duplicate
	suite.True(hasDuplicateMembers(ids))
}

func (suite *hasDuplicateMembersSuite) TestDuplicates3() {
	duplicate := testutil.NewUUIDV4()
	ids := make([]uuid.UUID, 64)
	for i := range ids {
		ids[i] = testutil.NewUUIDV4()
	}
	ids[23] = duplicate
	ids[54] = duplicate
	ids[60] = duplicate
	suite.True(hasDuplicateMembers(ids))
}

func Test_hasDuplicateMembers(t *testing.T) {
	suite.Run(t, new(hasDuplicateMembersSuite))
}
