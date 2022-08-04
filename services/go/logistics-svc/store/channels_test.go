package store

import (
	"errors"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// channelDetailsMock mocks ChannelDetails.
type channelDetailsMock struct {
	mock.Mock
}

func (m *channelDetailsMock) Validate() (entityvalidation.Report, error) {
	args := m.Called()
	return args.Get(0).(entityvalidation.Report), args.Error(1)
}

// ChannelValidateSuite tests Channel.Validate.
type ChannelValidateSuite struct {
	suite.Suite
	ok      Channel
	details *channelDetailsMock
}

func (suite *ChannelValidateSuite) SetupTest() {
	suite.details = &channelDetailsMock{}
	suite.ok = Channel{
		ID:            testutil.NewUUIDV4(),
		Entry:         testutil.NewUUIDV4(),
		Label:         "", // Can be empty.
		Type:          ChannelTypeDirect,
		Priority:      12,
		MinImportance: 24,
		Details:       suite.details,
		Timeout:       10 * time.Minute,
	}
}

func (suite *ChannelValidateSuite) TestUnknownType() {
	suite.details.On("Validate").Return(entityvalidation.NewReport(), nil)
	defer suite.details.AssertExpectations(suite.T())
	suite.ok.Type = "y6dO7YdF"

	report, err := suite.ok.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "report should not be ok")
}

func (suite *ChannelValidateSuite) TestDetailsValidationFail() {
	suite.details.On("Validate").Return(entityvalidation.NewReport(), errors.New("sad life"))
	defer suite.details.AssertExpectations(suite.T())

	_, err := suite.ok.Validate()
	suite.Error(err, "should fail")
}

func (suite *ChannelValidateSuite) TestInvalidDetails() {
	detailsReport := entityvalidation.NewReport()
	detailsReport.AddError("sad life")
	suite.details.On("Validate").Return(detailsReport, nil)
	defer suite.details.AssertExpectations(suite.T())

	report, err := suite.ok.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "report should not be ok")
}

func (suite *ChannelValidateSuite) TestMissingTimeout() {
	suite.details.On("Validate").Return(entityvalidation.NewReport(), nil)
	defer suite.details.AssertExpectations(suite.T())
	suite.ok.Timeout = 0

	report, err := suite.ok.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "report should not be ok")
}

func (suite *ChannelValidateSuite) TestNegativeTimeout() {
	suite.details.On("Validate").Return(entityvalidation.NewReport(), nil)
	defer suite.details.AssertExpectations(suite.T())
	suite.ok.Timeout = -10 * time.Minute

	report, err := suite.ok.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "report should not be ok")
}

func (suite *ChannelValidateSuite) TestOK() {
	suite.details.On("Validate").Return(entityvalidation.NewReport(), nil)
	defer suite.details.AssertExpectations(suite.T())

	report, err := suite.ok.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.True(report.IsOK(), "report should be ok")
}

func TestChannel_Validate(t *testing.T) {
	suite.Run(t, new(ChannelValidateSuite))
}
