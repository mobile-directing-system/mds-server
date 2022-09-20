package store

import (
	"encoding/json"
	"github.com/gofrs/uuid"
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
		CreatedBy: testutil.NewUUIDV4(),
		Operation: testutil.NewUUIDV4(),
		Type:      IntelTypePlaintextMessage,
		Content:   json.RawMessage(`{"text":"hello"}`),
		InitialDeliverTo: []uuid.UUID{
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
			testutil.NewUUIDV4(),
		},
	}
}

func (suite *CreateIntelValidateSuite) TestDuplicateInitialDeliverToEntries() {
	a := suite.sampleCreateIntel.InitialDeliverTo[0]
	suite.sampleCreateIntel.InitialDeliverTo = append(suite.sampleCreateIntel.InitialDeliverTo, a)

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
