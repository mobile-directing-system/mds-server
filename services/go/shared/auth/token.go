package auth

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"reflect"
	"strings"
)

// Token are the internal token details for communicating with internal
// services.
type Token struct {
	// UserID is the id of the user that performed the request. Only set, if
	// IsAuthenticated.
	UserID uuid.UUID `json:"user_id"`
	// Username of the user that performed the request. Only set, if
	// IsAuthenticated.
	Username string `json:"username"`
	// IsAuthenticated describes whether the user is currently logged in.
	IsAuthenticated bool `json:"is_authenticated"`
	// IsAdmin describes whether the user is an admin.
	IsAdmin bool `json:"is_admin"`
	// Permissions the user was granted.
	Permissions []permission.Permission `json:"permissions"`
	// RandomSalt are just some random values.
	RandomSalt []byte `json:"randomSalt"`
}

// jwtClaimName is the name of the JWT claim, that contains the Token.
const jwtClaimName = "mds"

// GenJWTToken generates a signed JWT token with the given Token payload.
func GenJWTToken(token Token, secret string) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{jwtClaimName: token})
	// Return signed token.
	signedJWTToken, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return "", meh.NewInternalErrFromErr(err, "sign token", nil)
	}
	return signedJWTToken, nil
}

// ParseJWTToken validates and parses the given signed JWT token. Errors
// regarding parsing are returned as meh.ErrInternal because we expect the token
// to be valid as it is supplied by the API Gateway.
func ParseJWTToken(signedToken string, secret string) (Token, error) {
	// Parse JWT token.
	jwtToken, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, meh.NewInternalErr("unexpected signing method",
				meh.Details{"was": token.Header["alg"]})
		}
		return []byte(secret), nil
	})
	if err != nil {
		return Token{}, meh.NewInternalErrFromErr(err, "parse jwt token", nil)
	}
	// Extract claims.
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return Token{}, meh.NewInternalErr("cannot cast jwt token claims",
			meh.Details{"was": reflect.TypeOf(jwtToken.Claims)})
	}
	tokenRawMap, ok := claims[jwtClaimName]
	if !ok {
		return Token{}, meh.NewInternalErr("missing token payload", nil)
	}
	// We get the token as a raw go map with strings. Currently, we cannot unmarshal
	// it into struct directly, so we simply marshal it as JSOn and then unmarshal
	// again. If this poses to be a performance issue, we can still use third-party
	// libraries for that job.
	tokenRaw, err := json.Marshal(tokenRawMap)
	if err != nil {
		return Token{}, meh.NewInternalErrFromErr(err, "cannot marshal raw token map", nil)
	}
	// Parse token.
	var token Token
	err = json.Unmarshal(tokenRaw, &token)
	if err != nil {
		return Token{}, meh.NewInternalErrFromErr(err, "unmarshal token", nil)
	}
	return token, nil
}

// ParseJWTTokenFromHeader extracts the JWT token from the given gin.Context.
// This is meant for internal use, and we treat a non-existent header as error.
func ParseJWTTokenFromHeader(c *gin.Context, secret string) (Token, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return Token{}, meh.NewInternalErr("no token found in headers", nil)
	}
	jwtToken := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := ParseJWTToken(jwtToken, secret)
	if err != nil {
		return Token{}, meh.Wrap(err, "parse jwt token", meh.Details{"jwt_token": jwtToken})
	}
	return token, nil
}

// HasPermission matches the given Token against the permission.matcher.
// However, there are two special cases: If Token.IsAuthenticated is false, we
// always return false. If Token.IsAdmin is true, we always return true.
func HasPermission(token Token, toHave ...permission.Matcher) (bool, error) {
	if !token.IsAuthenticated {
		return false, nil
	}
	if token.IsAdmin {
		return true, nil
	}
	ok, err := permission.Has(token.Permissions, toHave...)
	if err != nil {
		return false, meh.Wrap(err, "match permissions", meh.Details{"permissions": token.Permissions})
	}
	return ok, nil
}

// AssurePermission acts similarly to HasPermission but the result is returned
// as error. Regardless of the given list of permission.Matcher, if not
// authenticated, an meh.ErrUnauthorized error is returned. If the token claims
// to be admin, no error will be returned. Otherwise, the returned error is nil,
// if all permissions are granted, and non-nil with error details if not.
func AssurePermission(token Token, toHave ...permission.Matcher) error {
	if !token.IsAuthenticated {
		return meh.NewUnauthorizedErr("not authenticated", nil)
	}
	if token.IsAdmin {
		return nil
	}
	err := permission.Assure(token.Permissions, toHave...)
	if err != nil {
		return meh.Wrap(err, "match permissions", meh.Details{"granted_permissions": token.Permissions})
	}
	return nil
}
