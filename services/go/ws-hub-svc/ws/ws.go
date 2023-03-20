package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/logging"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wshub"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strings"
	"sync"
	"time"
)

const controlChannel = "_control"

// Gate holds information regarding connection fan-out.
type Gate struct {
	// Name of the gate.
	Name string
	// Channels for the gate.
	Channels map[wshub.Channel]Channel
}

// Channel holds configuration regarding a channel in a Gate.
type Channel struct {
	// URL to which the websocket request is sent to.
	URL string
}

// Hub provides Serve for serving WebSocket requests.
type Hub interface {
	// Serve the WebSocket request for the gate with the given id.
	Serve(clientWriter http.ResponseWriter, clientRequest *http.Request, gate string, requestHeader http.Header) error
}

// hub is the implementation of Hub.
type hub struct {
	// connLifetime of all connections.
	connLifetime context.Context
	// logger for any information regarding WebSocket connections.
	logger *zap.Logger
	// dialer is used for connecting to channels.
	dialer Dialer
	// gateConfigs are the configs to use for retrieving channel-information, etc.
	gateConfigs map[string]Gate
	// upgrader to upgrade to WebSocket connections.
	upgrader websocket.Upgrader
}

// Dialer is an abstraction of websocket.Dialer. See websocket.Dialer for
// documentation.
type Dialer interface {
	DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)
}

// NewNetHub creates a new Hub with the given config and connection lifetime.
// When it is done, all active clients will be disconnected.
func NewNetHub(connLifetime context.Context, logger *zap.Logger, gateConfigs map[string]Gate) Hub {
	return &hub{
		connLifetime: connLifetime,
		logger:       logger,
		dialer:       websocket.DefaultDialer,
		gateConfigs:  gateConfigs,
		upgrader: websocket.Upgrader{
			ReadBufferSize:   wsutil.ReadBufferSize,
			WriteBufferSize:  wsutil.WriteBufferSize,
			HandshakeTimeout: 2 * time.Second,
		},
	}
}

// session represents a client connection and its fan-out-connections.
type session struct {
	logger     *zap.Logger
	gateConfig Gate
	client     *websocket.Conn
	channels   map[wshub.Channel]*websocket.Conn
}

var headersToRemoveLower = map[string]struct{}{
	"upgrade":                  {},
	"connection":               {},
	"sec-websocket-key":        {},
	"sec-websocket-version":    {},
	"sec-websocket-extensions": {},
	"sec-websocket-protocol":   {},
}

func headersWithoutWS(h http.Header) http.Header {
	newHeaders := make(http.Header)
	for k, v := range h {
		if _, ok := headersToRemoveLower[strings.ToLower(k)]; ok {
			continue
		}
		newHeaders[k] = v
	}
	return newHeaders
}

// connectChannels connects to all channels in the given Gate. If one fails, all
// connections are closed. Otherwise, a map with the connections by channel-name
// is returned. The cancel-func closes all channel-connections.
func (h *hub) connectChannels(ctx context.Context, sessionLogger *zap.Logger, channels map[wshub.Channel]Channel,
	requestHeader http.Header) (map[wshub.Channel]*websocket.Conn, error) {
	connections := make(map[wshub.Channel]*websocket.Conn, len(channels))
	var connectionsMutex sync.Mutex
	eg, egCtx := errgroup.WithContext(ctx)
	for channelLabel, channel := range channels {
		channelLabel := channelLabel
		channel := channel
		eg.Go(func() error {
			// Remove duplicate header for WebSocket stuff.
			requestHeader := headersWithoutWS(requestHeader)
			conn, _, err := h.dialer.DialContext(egCtx, channel.URL, requestHeader)
			if err != nil {
				return meh.NewErrFromErr(err, wsutil.ErrWSCommunication, "dial", meh.Details{"url": channel.URL})
			}
			sessionLogger.Debug("channel up",
				zap.Any("channel_label", channelLabel),
				zap.Any("channel_url", channel.URL))
			connectionsMutex.Lock()
			defer connectionsMutex.Unlock()
			// No duplicates should be possible.
			connections[channelLabel] = conn
			return nil
		})
	}

	err := eg.Wait()
	if err != nil {
		// Close all existing connections.
		for _, conn := range connections {
			_ = conn.Close()
		}
		return nil, err
	}
	return connections, nil
}

// Serve the WebSocket connection for the gate with the given id.
func (h *hub) Serve(clientWriter http.ResponseWriter, clientRequest *http.Request, gate string, requestHeader http.Header) error {
	gateConfig, ok := h.gateConfigs[gate]
	if !ok {
		return meh.NewNotFoundErr("gate not found", meh.Details{"gate": gate})
	}
	sessionLifetime, closeSession := context.WithCancel(h.connLifetime)
	defer closeSession()
	gateInstance, _ := uuid.NewV4()
	sessionLogger := h.logger.Named("gate").Named(gate).Named("session").Named(gateInstance.String())
	start := time.Now()
	sessionLogger.Debug("opening gate")
	defer func() {
		h.logger.Debug("closing gate", zap.Any("was_open", time.Since(start)))
	}()
	// Connect all channels.
	channelConns, err := h.connectChannels(sessionLifetime, sessionLogger, gateConfig.Channels, requestHeader)
	if err != nil {
		return meh.Wrap(err, "connect channels", nil)
	}
	defer func() {
		for _, conn := range channelConns {
			_ = conn.Close()
		}
	}()
	// Now, we can upgrade the client's connection.
	clientConn, err := h.upgrader.Upgrade(clientWriter, clientRequest, nil)
	if err != nil {
		return meh.NewErrFromErr(err, wsutil.ErrWSCommunication, "upgrade client connection", nil)
	}
	defer func() {
		_ = clientConn.Close()
	}()
	sessionLogger.Debug("client connection ready")
	// Route.
	session := session{
		logger:     sessionLogger,
		gateConfig: gateConfig,
		client:     clientConn,
		channels:   channelConns,
	}
	err = route(sessionLifetime, sessionLogger, session)
	if err != nil {
		mehlog.Log(sessionLogger, meh.Wrap(err, "route", nil))
	}
	return nil
}

// TODO: Add global flag for omitting error details for security reasons.

func route(lifetime context.Context, logger *zap.Logger, session session) error {
	// Create client's client (lol).
	clientClient := wsutil.NewClient(lifetime, logger.Named("client"), auth.Token{}, session.client)
	// Create channel clients.
	channelClients := make(map[wshub.Channel]wsutil.Client, len(session.channels))
	for channelLabel, channelConn := range session.channels {
		channelClients[channelLabel] = wsutil.NewClient(lifetime, logger.Named("channel").Named(string(channelLabel)), auth.Token{}, channelConn)
	}
	// Run all.
	eg, egCtx := errgroup.WithContext(lifetime)
	eg.Go(func() error {
		defer logger.Debug("client listener down")
		for {
			select {
			case <-egCtx.Done():
				return nil
			case messageRaw := <-clientClient.RawConnection().ReceiveRaw():
				// Parse message.
				var mc wshub.MessageContainer
				err := json.Unmarshal(messageRaw, &mc)
				if err != nil {
					return meh.NewBadInputErrFromErr(err, "parse as message-container", meh.Details{"was": string(messageRaw)})
				}
				// Forward to channel.
				channel, ok := channelClients[mc.Channel]
				if !ok {
					sendErrOverControlChannel(egCtx, clientClient.RawConnection(), meh.NewBadInputErr("unknown channel", meh.Details{"channel_label": mc.Channel}))
					continue
				}
				session.logger.Debug("forward message from client to channel",
					zap.String("channel", string(mc.Channel)),
					zap.String("message", string(mc.Payload)))
				select {
				case <-egCtx.Done():
					session.logger.Debug("dropping received message from client due to context done", zap.Any("message", string(messageRaw)))
					return nil
				case channel.RawConnection().SendRaw() <- mc.Payload:
				}
			}
		}
	})
	for channelLabel, channelClient := range channelClients {
		channelLabel := channelLabel
		channelClient := channelClient
		eg.Go(func() error {
			defer logger.Debug("channel listener down", zap.String("channel", string(channelLabel)))
			for {
				select {
				case <-egCtx.Done():
					return nil
				case message, more := <-channelClient.RawConnection().ReceiveRaw():
					if !more {
						return nil
					}
					// Generate message for client.
					messageRaw, err := json.Marshal(wshub.MessageContainer{
						Channel: channelLabel,
						Payload: message,
					})
					if err != nil {
						return meh.NewBadInputErrFromErr(err, "marshal message container", nil)
					}
					// Forward to client.
					logger.Debug("forward message from channel to client",
						zap.String("channel", string(channelLabel)),
						zap.String("message", string(messageRaw)))
					select {
					case <-egCtx.Done():
						session.logger.Debug("dropping outgoing message from channel due to context done",
							zap.Any("message", string(messageRaw)))
						return nil
					case clientClient.RawConnection().SendRaw() <- messageRaw:
					}
				}
			}
		})
		eg.Go(func() error {
			<-egCtx.Done()
			channelClient.Close()
			return nil
		})
		eg.Go(func() error {
			return meh.NilOrWrap(channelClient.RunAndClose(), "run and close channel-client",
				meh.Details{"channel_label": channelLabel})
		})
	}
	// Launch client.
	eg.Go(func() error {
		<-egCtx.Done()
		clientClient.Close()
		return nil
	})
	eg.Go(func() error {
		return meh.NilOrWrap(clientClient.RunAndClose(), "run and close client-client", nil)
	})
	return eg.Wait()
}

func sendErrOverControlChannel(ctx context.Context, c wsutil.RawConnection, err error) {
	errorMessageContentRaw, err := json.Marshal(wsutil.ErrorMessageFromErr(err))
	if err != nil {
		mehlog.Log(logging.DebugLogger(), meh.NewInternalErrFromErr(err, "marshal error message content",
			meh.Details{"error_message_from_err": fmt.Sprintf("%+v", wsutil.ErrorMessageFromErr(err))}))
	}
	payloadRaw, err := json.Marshal(wsutil.Message{
		Type:    wsutil.TypeError,
		Payload: errorMessageContentRaw,
	})
	if err != nil {
		mehlog.Log(logging.DebugLogger(), meh.NewInternalErrFromErr(err, "marshal message payload",
			meh.Details{"raw_error_message_from_err": string(errorMessageContentRaw)}))
	}
	finalMessageRaw, err := json.Marshal(wshub.MessageContainer{
		Channel: controlChannel,
		Payload: payloadRaw,
	})
	if err != nil {
		mehlog.Log(logging.DebugLogger(), meh.NewInternalErrFromErr(err, "marshal final message",
			meh.Details{"channel": controlChannel, "payload": string(payloadRaw)}))
	}
	// Forward to client.
	select {
	case <-ctx.Done():
		return
	case c.SendRaw() <- finalMessageRaw:
	}
}
