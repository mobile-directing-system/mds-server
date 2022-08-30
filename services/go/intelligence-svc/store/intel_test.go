package store

import (
	"encoding/json"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// CreateIntelValidateSuite tests CreateIntel.Validate.
type CreateIntelValidateSuite struct {
	suite.Suite
	sampleCreateIntel CreateIntel
}

func (suite *CreateIntelValidateSuite) SetupTest() {
	suite.sampleCreateIntel = CreateIntel{
		CreatedBy:  testutil.NewUUIDV4(),
		Operation:  testutil.NewUUIDV4(),
		Type:       "test",
		Content:    json.RawMessage(`{"hello":"world"}`),
		SearchText: nulls.NewString("hello world"),
		Assignments: []IntelAssignment{
			{
				ID: testutil.NewUUIDV4(),
				To: testutil.NewUUIDV4(),
			},
			{
				ID: testutil.NewUUIDV4(),
				To: testutil.NewUUIDV4(),
			},
		},
	}
}

func (suite *CreateIntelValidateSuite) TestDuplicateAssignments() {
	a := suite.sampleCreateIntel.Assignments[0]
	suite.sampleCreateIntel.Assignments = append(suite.sampleCreateIntel.Assignments, a)

	report, err := suite.sampleCreateIntel.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "report should not be ok")
}

func (suite *CreateIntelValidateSuite) TestOK() {
	report, err := suite.sampleCreateIntel.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.True(report.IsOK(), "report should be ok")
}

func TestCreateIntel_Validate(t *testing.T) {
	suite.Run(t, new(CreateIntelValidateSuite))
}
