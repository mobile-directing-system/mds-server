package store

import (
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// UserValidateSuite tests User.Validate.
type UserValidateSuite struct {
	suite.Suite
	u User
}

func (suite *UserValidateSuite) SetupTest() {
	suite.u = User{
		ID:        testutil.NewUUIDV4(),
		Username:  "curve",
		FirstName: "basic",
		LastName:  "such",
		IsAdmin:   true,
	}
}

func (suite *UserValidateSuite) TestMissingUsername() {
	suite.u.Username = ""
	report, err := suite.u.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "should not be ok")
}

func (suite *UserValidateSuite) TestMissingFirstName() {
	suite.u.FirstName = ""
	report, err := suite.u.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "should not be ok")
}

func (suite *UserValidateSuite) TestMissingLastName() {
	suite.u.LastName = ""
	report, err := suite.u.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "should not be ok")
}

func (suite *UserValidateSuite) TestOK() {
	report, err := suite.u.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.True(report.IsOK(), "should be ok")
}

func TestUser_Validate(t *testing.T) {
	suite.Run(t, new(UserValidateSuite))
}

// UserWithPassValidateSuite test UserWithPass.Validate.
type UserWithPassValidateSuite struct {
	suite.Suite
	u UserWithPass
}

func (suite *UserWithPassValidateSuite) SetupTest() {
	suite.u = UserWithPass{
		User: User{
			ID:        testutil.NewUUIDV4(),
			Username:  "curve",
			FirstName: "basic",
			LastName:  "such",
			IsAdmin:   true,
		},
		Pass: []byte(`Hello World!`),
	}
}

func (suite *UserWithPassValidateSuite) TestInvalidBase() {
	suite.u.FirstName = ""
	report, err := suite.u.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "should not be ok")
}

func (suite *UserWithPassValidateSuite) TestEmptyPass() {
	suite.u.Pass = []byte{}
	report, err := suite.u.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.False(report.IsOK(), "should not be ok")
}

func (suite *UserWithPassValidateSuite) TestOK() {
	report, err := suite.u.Validate()
	suite.Require().NoError(err, "should not fail")
	suite.True(report.IsOK(), "should be ok")
}

func TestUserWithPass_Validate(t *testing.T) {
	suite.Run(t, new(UserWithPassValidateSuite))
}
