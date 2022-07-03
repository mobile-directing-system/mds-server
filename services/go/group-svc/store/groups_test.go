package store

import (
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

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
