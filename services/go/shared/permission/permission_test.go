package permission

import "github.com/stretchr/testify/suite"

// NameMatcherSuite tests matchers, that only rely on a permission Name being
// present in the list of granted permissions.
type NameMatcherSuite struct {
	suite.Suite
	// MatcherName is the expected matcher name.
	MatcherName string
	// Matcher is the actual Matcher to test.
	Matcher Matcher
	// Granted is the Name of the permission, that means granted.
	Granted Name
	// Others is a Name list of permissions, that are not considered.
	Others []Name
}

type NameMapBuilder struct {
	m map[Name]Permission
}

func NewNameMapBuilder(names ...Name) *NameMapBuilder {
	m := make(map[Name]Permission, len(names))
	for _, name := range names {
		m[name] = Permission{Name: name}
	}
	return &NameMapBuilder{m: m}
}

func (builder *NameMapBuilder) Add(names ...Name) *NameMapBuilder {
	for _, name := range names {
		builder.m[name] = Permission{Name: name}
	}
	return builder
}

func (builder *NameMapBuilder) Map() map[Name]Permission {
	return builder.m
}

func (suite *NameMatcherSuite) TestMatcherName() {
	suite.Equal(suite.MatcherName, suite.Matcher.Name, "should have correct matcher name")
}

func (suite *NameMatcherSuite) TestMatchFnOK() {
	suite.NotNil(suite.Matcher.MatchFn, "should provide matcher fn")
}

func (suite *NameMatcherSuite) TestEmpty() {
	granted, err := suite.Matcher.MatchFn(map[Name]Permission{})
	suite.Require().NoError(err, "should not fail")
	suite.False(granted, "should not have granted")
}

func (suite *NameMatcherSuite) TestGrantedSingle() {
	granted, err := suite.Matcher.MatchFn(map[Name]Permission{
		suite.Granted: {Name: suite.Granted},
	})
	suite.Require().NoError(err, "should not fail")
	suite.True(granted, "should have granted")
}

func (suite *NameMatcherSuite) TestGrantedMulti() {
	m := NewNameMapBuilder(suite.Granted).Add(suite.Others...).Map()
	granted, err := suite.Matcher.MatchFn(m)
	suite.Require().NoError(err, "should not fail")
	suite.True(granted, "should have granted")
}

func (suite *NameMatcherSuite) TestNotGranted() {
	m := NewNameMapBuilder(suite.Others...).Map()
	granted, err := suite.Matcher.MatchFn(m)
	suite.Require().NoError(err, "should not fail")
	suite.False(granted, "should have granted")
}
