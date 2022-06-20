package permission

import (
	"errors"
	"github.com/lefinal/nulls"
	"github.com/stretchr/testify/suite"
	"testing"
)

// HasSuite tests Has.
type HasSuite struct {
	suite.Suite
}

var testPermissionOK = func() Matcher {
	return Matcher{
		Name: "ok",
		MatchFn: func(_ map[Name]Permission) (bool, error) {
			return true, nil
		},
	}
}

var testPermissionFail = func() Matcher {
	return Matcher{
		Name: "fail",
		MatchFn: func(_ map[Name]Permission) (bool, error) {
			return false, errors.New("sad life")
		},
	}
}
var testPermissionNotOK = func() Matcher {
	return Matcher{
		Name: "not-ok",
		MatchFn: func(_ map[Name]Permission) (bool, error) {
			return false, nil
		},
	}
}
var testPermissionMatchName = func(name Name) Matcher {
	return Matcher{
		Name: "match-name",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[name]
			return ok, nil
		},
	}
}

func (suite *HasSuite) TestOK() {
	ok, err := Has([]Permission{
		{
			Name:    "hello",
			Options: nulls.JSONRawMessage{},
		},
		{
			Name:    "world",
			Options: nulls.JSONRawMessage{},
		},
		{
			Name:    "!",
			Options: nulls.JSONRawMessage{},
		},
	}, testPermissionMatchName("hello"), testPermissionMatchName("!"))
	suite.Require().NoError(err, "should not fail")
	suite.True(ok, "should be ok")
}

func (suite *HasSuite) TestNotGranted() {
	ok, err := Has([]Permission{{Name: "meow"}}, testPermissionOK(), testPermissionNotOK())
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not be ok")
}

func (suite *HasSuite) TestEmptyList() {
	ok, err := Has([]Permission{}, testPermissionOK())
	suite.Require().NoError(err, "should not fail")
	suite.True(ok, "should be ok")
}

func (suite *HasSuite) TestMultipleEmpty() {
	ok, err := Has([]Permission{}, testPermissionOK(), testPermissionNotOK())
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not be ok")
}

func (suite *HasSuite) TestFail() {
	_, err := Has([]Permission{{Name: "hello"}}, testPermissionOK(), testPermissionMatchName("hello"), testPermissionFail())
	suite.Error(err, "should fail")
}

func TestHas(t *testing.T) {
	suite.Run(t, new(HasSuite))
}
