package controller

import (
	"context"
	"crypto/rand"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
)

// AuthRequestMetadata holds metadata for login- and logout-requests for logging
// purposes.
type AuthRequestMetadata struct {
	Host       string
	UserAgent  string
	RemoteAddr string
}

// Login tries to log in the user with the given username and password. If login
// fails, false is returned as second value. Otherwise, the first return value
// will be the user id and the second one the assigned JWT-token.
func (c *Controller) Login(ctx context.Context, username string, pass string, requestMetadata AuthRequestMetadata) (uuid.UUID, string, bool, error) {
	var ok bool
	var userID uuid.UUID
	var token string
	var err error
	// Load actual password for username.
	err = pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		user, err := c.Store.UserWithPassByUsername(ctx, tx, username)
		if err != nil {
			return meh.Wrap(err, "user by username", meh.Details{"username": username})
		}
		userID = user.ID
		// Check password.
		passOK, err := auth.PasswordOK(user.Pass, pass)
		if err != nil {
			return meh.Wrap(err, "check if password ok", nil)
		}
		// If password wrong, we are done.
		if !passOK {
			ok = false
			return nil
		}
		ok = true
		// Generate public session token.
		token, err = generatePublicSessionToken(username, c.PublicAuthTokenSecret)
		if err != nil {
			return meh.Wrap(err, "generate public session token", meh.Details{"username": username})
		}
		// Store public session token.
		err = c.Store.StoreSessionTokenForUser(ctx, tx, token, user.ID)
		if err != nil {
			return meh.Wrap(err, "store username by session token", meh.Details{
				"session_token": token,
				"username":      username,
			})
		}
		// Notify.
		err = c.Notifier.NotifyUserLoggedIn(ctx, tx, user.ID, user.Username, requestMetadata)
		if err != nil {
			return meh.Wrap(err, "notify user logged in", meh.Details{
				"user_id":          user.ID,
				"username":         user.Username,
				"request_metadata": requestMetadata,
			})
		}
		return nil
	})
	if err != nil {
		return uuid.Nil, "", false, meh.Wrap(err, "run in tx", nil)
	}
	return userID, token, ok, nil
}

// generatePublicSessionToken generates and signs the JWT token, that will be
// sent to the client.
func generatePublicSessionToken(username string, secret string) (string, error) {
	// Generate random salt.
	randomSalt := make([]byte, 512)
	_, err := rand.Read(randomSalt)
	if err != nil {
		return "", meh.NewInternalErrFromErr(err, "read random salt", nil)
	}
	// Generate and sign JWT token.
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"username":    username,
		"random_salt": randomSalt,
	})
	jwtTokenSigned, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return "", meh.NewInternalErrFromErr(err, "sign jwt token", nil)
	}
	return jwtTokenSigned, nil
}

// Logout is called when a client wants to log out.
func (c *Controller) Logout(ctx context.Context, publicToken string, requestMetadata AuthRequestMetadata) error {
	err := pgutil.RunInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Delete session.
		userID, err := c.Store.GetAndDeleteUserIDBySessionToken(ctx, tx, publicToken)
		if err != nil {
			return meh.Wrap(err, "get and delete id by session token", meh.Details{"token": publicToken})
		}
		// Gather some data for more detailed logging.
		user, err := c.Store.UserWithPassByID(ctx, tx, userID)
		if err != nil {
			return meh.Wrap(err, "user by id", meh.Details{"user_id": userID})
		}
		// Notify.
		err = c.Notifier.NotifyUserLoggedOut(ctx, tx, userID, user.Username, requestMetadata)
		if err != nil {
			return meh.Wrap(err, "notify user logged out", nil)
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	return nil
}
