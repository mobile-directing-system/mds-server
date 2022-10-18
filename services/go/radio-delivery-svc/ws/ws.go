package ws

import (
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/radio-delivery-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"go.uber.org/zap"
)

// ForwardListener is the listener to forward a created controller.Connection
// to.
type ForwardListener interface {
	AcceptNewConnection(connection controller.Connection)
}

// Gatekeeper is a ws.Gatekeeper, assuring that the auth.Token is authenticated.
func Gatekeeper() wsutil.Gatekeeper {
	return func(token auth.Token) error {
		if !token.IsAuthenticated {
			return meh.NewUnauthorizedErr("not authenticated", nil)
		}
		return nil
	}
}

// ConnListener is the listener for ws.ConnListener that forwards created and
// mapped connections to the given ForwardListener.
func ConnListener(logger *zap.Logger, forwardListener ForwardListener) wsutil.ConnListener {
	return func(conn wsutil.Connection) {
		if !conn.AuthToken().IsAuthenticated {
			mehlog.Log(logger, meh.NewInternalErr("websocket connection listener received unauthenticated connection", nil))
			return
		}
		forwardListener.AcceptNewConnection(newConnection(conn))
	}
}
