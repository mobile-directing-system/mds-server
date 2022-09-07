package store

import (
	"fmt"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

// IntelTypePlaintextMessageContentValidateSuite tests
// IntelTypePlaintextMessageContent.Validate.
type IntelTypePlaintextMessageContentValidateSuite struct {
	suite.Suite
	contentOK IntelTypePlaintextMessageContent
}

func (suite *IntelTypePlaintextMessageContentValidateSuite) SetupTest() {
	suite.contentOK = IntelTypePlaintextMessageContent{
		Text: "thunder",
	}
}

func (suite *IntelTypePlaintextMessageContentValidateSuite) TestMissingTest() {
	content := suite.contentOK
	content.Text = ""
	report, err := content.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "report should not be ok")
}

func (suite *IntelTypePlaintextMessageContentValidateSuite) TestOK() {
	report, err := suite.contentOK.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.True(report.IsOK(), "report should be ok")
}

func TestIntelTypePlaintextMessageContent_Validate(t *testing.T) {
	suite.Run(t, new(IntelTypePlaintextMessageContentValidateSuite))
}

func Test_validateCreateIntelTypeAndContent(t *testing.T) {
	unsupportedIntelTypeErrMessage := "unsupported intel-type"
	t.Run("TestAssureUnsupportedIntelError", func(t *testing.T) {
		unknownIntelType := IntelType(testutil.NewUUIDV4().String())
		report, err := validateCreateIntelTypeAndContent(unknownIntelType, []byte(`{}`))
		require.NoError(t, err, "should not fail")
		require.NotEmpty(t, report.Errors, "should return validation errors")
		assert.Contains(t, report.Errors[0], unsupportedIntelTypeErrMessage, "should contain correct error message")
	})
	testutil.TestMapperWithConstExtraction(t, func(from IntelType) (string, error) {
		report, err := validateCreateIntelTypeAndContent(from, []byte(`{}`))
		if err != nil {
			return "", err
		}
		if report.IsOK() {
			return "", nil
		}
		// If validation errors, assure that this is not because of bad content but
		// unsupported intel-type.
		for _, s := range report.Errors {
			if strings.Contains(s, unsupportedIntelTypeErrMessage) {
				return "", fmt.Errorf("unsupported intel-type: %v", from)
			}
		}
		return "", nil
	}, "./intel_content.go", nulls.String{})
}
