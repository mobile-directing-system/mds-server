package controller

import (
	"github.com/golang-jwt/jwt"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/store"
	"go.uber.org/zap"
)

// Controller manages all core operations of the gateway.
type Controller struct {
	Logger     *zap.Logger
	HMACSecret string
	Mall       *store.Mall
}

// Login tries to log in the user with the given username and password. If login
// fails, false is returned as second value. Otherwise, the first return value
// will be the assigned JWT-token.
func (c *Controller) Login(username string, pass string) (string, bool, error) {
	// TODO: check login.
	// Generate token.
	token := jwt.New(jwt.SigningMethodHS512)
	// Return signed token.
	signedToken, err := token.SignedString(c.HMACSecret)
	if err != nil {
		return "", false, meh.NewInternalErrFromErr(err, "sign token", nil)
	}
	return signedToken, true, nil
}
