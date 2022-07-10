package store

import (
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// OperationValidateSuite tests Operation.Validate.
type OperationValidateSuite struct {
	suite.Suite
	o Operation
}

func (suite *OperationValidateSuite) SetupTest() {
	suite.o = Operation{
		ID:          testutil.NewUUIDV4(),
		Title:       "block",
		Description: "sight",
		Start:       time.Date(2022, 1, 2, 12, 0, 0, 0, time.UTC),
		End:         nulls.NewTime(time.Date(2022, 2, 2, 12, 0, 0, 0, time.UTC)),
		IsArchived:  true,
	}
}

func (suite *OperationValidateSuite) TestMissingTitle() {
	suite.o.Title = ""
	report, err := suite.o.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "should not be ok")
}

func (suite *OperationValidateSuite) TestEndBeforeStart() {
	suite.o.End = nulls.NewTime(time.Date(1990, 2, 2, 12, 0, 0, 0, time.UTC))
	report, err := suite.o.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "should not be ok")
}

func (suite *OperationValidateSuite) TestOK() {
	report, err := suite.o.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.True(report.IsOK(), "should be ok")
}

func TestOperation_Validate(t *testing.T) {
	suite.Run(t, new(OperationValidateSuite))
}
