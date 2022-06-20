package controller

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"math/rand"
)

// Proxy checks if the user is logged in (returned as second return value) and
// generates an internal authentication token, which will be passed with the
// forwarded request.
func (c *Controller) Proxy(ctx context.Context, publicToken string) (string, error) {
	authToken, err := c.gatherProxyToken(ctx, publicToken)
	if err != nil {
		return "", meh.Wrap(err, "gather proxy token", nil)
	}
	// Add random salt.
	randomSalt := make([]byte, 8)
	_, err = rand.Read(randomSalt)
	if err != nil {
		return "", meh.NewInternalErrFromErr(err, "read random salt", nil)
	}
	authToken.RandomSalt = randomSalt
	// Build internal token.
	signedAuthToken, err := auth.GenJWTToken(authToken, c.AuthTokenSecret)
	if err != nil {
		return "", meh.Wrap(err, "gen auth token", meh.Details{"auth_token": authToken})
	}
	return signedAuthToken, nil
}

// gatherProxyToken builds an auth.Token based on user details. Only the random
// salt needs to be set, and then it can be signed in Proxy. This is mainly for
// better code readability.
func (c *Controller) gatherProxyToken(ctx context.Context, publicToken string) (auth.Token, error) {
	var authToken auth.Token
	// If no token is provided, we have nothing to do.
	if publicToken == "" {
		return authToken, nil
	}
	// Retrieve user details.
	userID, err := c.Store.UserIDBySessionToken(ctx, c.DB, publicToken)
	if err != nil {
		if meh.ErrorCode(err) != meh.ErrNotFound {
			return auth.Token{}, meh.Wrap(err, "username by session token", meh.Details{"token": publicToken})
		}
		// Not found -> not authenticated.
		return authToken, nil
	}
	authToken.UserID = userID
	authToken.IsAuthenticated = true
	err = pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Retrieve user details.
		user, err := c.Store.UserWithPassByID(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "user by id", meh.Details{"user_id": userID})
		}
		authToken.Username = user.Username
		authToken.IsAdmin = user.IsAdmin
		// Retrieve permissions.
		permissions, err := c.Store.PermissionsByUserID(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "permissions by username", nil)
		}
		authToken.Permissions = permissions
		return nil
	})
	if err != nil {
		return auth.Token{}, meh.Wrap(err, "run in tx", nil)
	}
	return authToken, nil
}
