package auth

import (
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

// PasswordOKSuite tests PasswordOK.
type PasswordOKSuite struct {
	suite.Suite
}

func (suite *PasswordOKSuite) hash(str string) []byte {
	hashed, err := bcrypt.GenerateFromPassword([]byte(str), BCryptHashCost)
	if err != nil {
		panic(err)
	}
	return hashed
}

func (suite *PasswordOKSuite) TestOK() {
	hashed := suite.hash("meow")
	ok, err := PasswordOK(hashed, "meow")
	suite.Require().NoError(err, "should not fail")
	suite.True(ok, "should be ok")
}

func (suite *PasswordOKSuite) TestMismatch() {
	hashed := suite.hash("meow")
	ok, err := PasswordOK(hashed, "woof")
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not be ok")
}

func (suite *PasswordOKSuite) TestFail() {
	_, err := PasswordOK([]byte("meow"), "meow")
	suite.Error(err, "should fail")
}

func TestPasswordOK(t *testing.T) {
	suite.Run(t, new(PasswordOKSuite))
}
