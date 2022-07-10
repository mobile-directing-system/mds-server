package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehgin"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"go.uber.org/zap"
	"net/http"
	"strings"
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
	UserID      uuid.UUID `json:"user_id"`
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
}

// tokenType is the type of the access token.
const tokenType = "Bearer"

// handleLoginStore are the dependencies needed for handleLogin.
type handleLoginStore interface {
	// Login returns on success the user id, token as string as well as a boolean
	// flag describing whether login was successful.
	Login(ctx context.Context, username string, pass string, requestMetadata controller.AuthRequestMetadata) (uuid.UUID, string, bool, error)
}

// handleLogin handles a login-request.
func handleLogin(logger *zap.Logger, s handleLoginStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse payload.
		var payload loginPayload
		err := c.BindJSON(&payload)
		if err != nil {
			mehgin.LogAndRespondError(logger, c, meh.NewBadInputErrFromErr(err, "invalid body", nil))
			return
		}
		// Login.
		requestMetadata := extractAuthRequestMetadataFromRequest(c.Request)
		userID, token, ok, err := s.Login(c.Request.Context(), payload.Username, payload.Pass, requestMetadata)
		if err != nil {
			mehgin.LogAndRespondError(logger, c, meh.Wrap(err, "login", meh.Details{
				"username":         payload.Username,
				"request_metadata": requestMetadata,
			}))
			return
		}
		if !ok {
			c.Status(http.StatusUnauthorized)
			return
		}
		// Respond with token.
		c.JSON(http.StatusOK, loginResponse{
			UserID:      userID,
			AccessToken: token,
			TokenType:   tokenType,
		})
	}
}

// extractAuthRequestMetadataFromRequest extracts
// controller.AuthRequestMetadata from the given http.Request.
func extractAuthRequestMetadataFromRequest(r *http.Request) controller.AuthRequestMetadata {
	return controller.AuthRequestMetadata{
		Host:       r.Host,
		UserAgent:  r.UserAgent(),
		RemoteAddr: r.RemoteAddr,
	}
}

// handleLogoutStore are the dependencies for handleLogout.
type handleLogoutStore interface {
	Logout(ctx context.Context, publicToken string, requestMetadata controller.AuthRequestMetadata) error
}

// handleLogout handles a logout-request.
func handleLogout(logger *zap.Logger, s handleLogoutStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract public token from request.
		authHeader := c.GetHeader("Authorization")
		publicToken := strings.TrimPrefix(authHeader, "Bearer ")
		// Logout.
		err := s.Logout(c.Request.Context(), publicToken, extractAuthRequestMetadataFromRequest(c.Request))
		if err != nil {
			mehgin.LogAndRespondError(logger, c, meh.Wrap(err, "logout", nil))
			return
		}
		c.Status(http.StatusOK)
	}
}
