package auth

import (
	"github.com/golang-jwt/jwt"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/stretchr/testify/suite"
	"testing"
)

// TokenSuite tests ParseJWTToken.
type ParseJWTTokenSuite struct {
	suite.Suite
	secret         string
	jwtToken       *jwt.Token
	signedJWTToken string
	originalToken  Token
}

func (suite *ParseJWTTokenSuite) SetupTest() {
	var err error
	suite.secret = "meow"
	suite.originalToken = Token{
		Username: "admin",
		Permissions: []permission.Permission{
			{Name: "say.hello"},
			{Name: "be.nice"},
		},
	}
	suite.jwtToken = jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{jwtClaimName: suite.originalToken})
	suite.signedJWTToken, err = suite.jwtToken.SignedString([]byte(suite.secret))
	suite.Require().NoError(err, "generating signed jwt token should not fail")
}

func (suite *ParseJWTTokenSuite) TestSigningMethodMismatch() {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{jwtClaimName: suite.originalToken})
	signedJWTToken, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	suite.Require().NoError(err, "should not fail")

	_, err = ParseJWTToken(signedJWTToken, suite.secret)
	suite.Error(err, "should fail")
}

func (suite *ParseJWTTokenSuite) TestSecretMismatch() {
	_, err := ParseJWTToken(suite.signedJWTToken, "woof")
	suite.Error(err, "should fail")
}

func (suite *ParseJWTTokenSuite) TestMissingClaim() {
	jwtToken := jwt.New(jwt.SigningMethodHS512)
	signedJWTToken, err := jwtToken.SignedString([]byte(suite.secret))
	suite.Require().NoError(err, "signing token should not fail")

	_, err = ParseJWTToken(signedJWTToken, suite.secret)
	suite.Error(err, "should fail")
}

func (suite *ParseJWTTokenSuite) TestBadTokenPayload() {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{jwtClaimName: 1})
	signedJWTToken, err := jwtToken.SignedString([]byte(suite.secret))
	suite.Require().NoError(err, "signing token should not fail")

	_, err = ParseJWTToken(signedJWTToken, suite.secret)
	suite.Error(err, "should fail")
}

func (suite *ParseJWTTokenSuite) TestOK() {
	signed, err := GenJWTToken(suite.originalToken, suite.secret)
	suite.Require().NoError(err, "generate should not fail")

	parsedToken, err := ParseJWTToken(signed, suite.secret)

	suite.Require().NoError(err, "parse should not fail")
	suite.Equal(suite.originalToken, parsedToken, "parsed token should match original token")
}

func TestParseJWTToken(t *testing.T) {
	suite.Run(t, new(ParseJWTTokenSuite))
}

// HasPermissionsSuite tests HasPermission.
type HasPermissionSuite struct {
	suite.Suite
}

func (suite *HasPermissionSuite) TestNotAuthenticated1() {
	ok, err := HasPermission(Token{
		IsAuthenticated: false,
		Permissions:     []permission.Permission{{Name: "meow"}},
	}, permission.Matcher{MatchFn: func(_ map[permission.Name]permission.Permission) (bool, error) {
		return true, nil
	}})
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not have permission")
}

func (suite *HasPermissionSuite) TestNotAuthenticated2() {
	ok, err := HasPermission(Token{
		IsAuthenticated: false,
		IsAdmin:         true,
	})
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not have permission")
}

func (suite *HasPermissionSuite) TestAdmin() {
	ok, err := HasPermission(Token{
		IsAuthenticated: true,
		IsAdmin:         true,
	}, permission.Matcher{MatchFn: func(_ map[permission.Name]permission.Permission) (bool, error) {
		return false, nil
	}})
	suite.Require().NoError(err, "should not fail")
	suite.True(ok, "should have permission")
}

func (suite *HasPermissionSuite) TestNoPermission() {
	ok, err := HasPermission(Token{
		IsAuthenticated: true,
		Permissions:     []permission.Permission{{Name: "meow"}},
	}, permission.Matcher{MatchFn: func(_ map[permission.Name]permission.Permission) (bool, error) {
		return false, nil
	}})
	suite.Require().NoError(err, "should not fail")
	suite.False(ok, "should not have permission")
}

func (suite *HasPermissionSuite) TestOK() {
	ok, err := HasPermission(Token{
		IsAuthenticated: true,
		Permissions:     []permission.Permission{},
	}, permission.Matcher{MatchFn: func(_ map[permission.Name]permission.Permission) (bool, error) {
		return true, nil
	}})
	suite.Require().NoError(err, "should not fail")
	suite.True(ok, "should have permission")
}

func TestHasPermission(t *testing.T) {
	suite.Run(t, new(HasPermissionSuite))
}
