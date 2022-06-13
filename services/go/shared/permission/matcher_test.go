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

func (suite *HasSuite) TestMultipleEmpty() {
	ok, err := Has("meow", "woof")([]Permission{})
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not be ok")
}

func (suite *HasSuite) TestMultipleNone() {
	ok, err := Has("meow", "woof")([]Permission{"quack"})
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not be ok")
}

func (suite *HasSuite) TestMultiplePartly() {
	ok, err := Has("meow", "woof")([]Permission{"woof"})
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not be ok")
}

func (suite *HasSuite) TestOK1() {
	ok, err := Has("meow", "woof")([]Permission{"meow", "woof"})
	suite.Require().NoError(err, "should not fail")
	suite.True(ok, "should be ok")
}

func (suite *HasSuite) TestOK2() {
	ok, err := Has("meow", "woof")([]Permission{"quack", "woof", "meow"})
	suite.Require().NoError(err, "should not fail")
	suite.True(ok, "should be ok")
}

func (suite *HasSuite) TestOK3() {
	ok, err := Has("meow", "woof")([]Permission{"woof", "meow", "woof"})
	suite.Require().NoError(err, "should not fail")
	suite.True(ok, "should be ok")
}

func TestHas(t *testing.T) {
	suite.Run(t, new(HasSuite))
}
