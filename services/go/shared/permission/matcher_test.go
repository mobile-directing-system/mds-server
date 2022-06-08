package permission

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

// HasSuite tests Has.
type HasSuite struct {
	suite.Suite
}

func (suite *HasSuite) TestOK() {
	ok, err := Has("woof")([]Permission{"meow", "woof"})
	suite.Require().NoError(err, "should not fail")
	suite.True(ok, "should be ok")
}

func (suite *HasSuite) TestNotFound() {
	ok, err := Has("quack")([]Permission{"meow", "woof"})
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not be ok")
}

func (suite *HasSuite) TestEmptyList() {
	ok, err := Has("meow")([]Permission{})
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not be ok")
}

func TestHas(t *testing.T) {
	suite.Run(t, new(HasSuite))
}
