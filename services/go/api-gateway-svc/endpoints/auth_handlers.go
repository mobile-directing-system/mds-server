package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehgin"
	"go.uber.org/zap"
	"net/http"
)

// loginPayload is the payload of a login-request in handleLogin.
type loginPayload struct {
	// Username for logging in.
	Username string `json:"username"`
	// Pass in plaintext format.
	Pass string `json:"pass"`
}

// loginResponse is the response in handleLogin when login was successful.
type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// tokenType is the type of the access token.
const tokenType = "Bearer"

// handleLoginStore are the dependencies needed for handleLogin.
type handleLoginStore interface {
	Login(username string, pass string) (string, bool, error)
}

// handleLogin handles a login-request.
func handleLogin(logger *zap.Logger, store handleLoginStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse payload.
		var payload loginPayload
		err := c.BindJSON(&payload)
		if err != nil {
			mehgin.LogAndRespondError(logger, c, meh.NewBadInputErrFromErr(err, "bad request", nil))
			return
		}
		// Login.
		token, ok, err := store.Login(payload.Username, payload.Pass)
		if err != nil {
			mehgin.LogAndRespondError(logger, c, meh.Wrap(err, "login", meh.Details{"username": payload.Username}))
			return
		}
		if !ok {
			c.Status(http.StatusUnauthorized)
			return
		}
		// Respond with token.
		c.JSON(http.StatusOK, loginResponse{
			AccessToken: token,
			TokenType:   tokenType,
		})
	}
}
