package ws

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/permission"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"go.uber.org/zap"
	"sync"
)

// ControllerSink is the abstraction of controller.Controller that is used for
// serving open intel delivery listeners.
type ControllerSink interface {
	ServeOpenIntelDeliveriesListener(lifetime context.Context, operationID uuid.UUID, notifier controller.OpenIntelDeliveriesListener)
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

// ConnListener is the listener for ws.ConnListener.
func ConnListener(logger *zap.Logger, sink ControllerSink) wsutil.ConnListener {
	return func(wsRawConn wsutil.RawConnection) {
		wsConn := wsutil.NewAutoParserConnection(wsRawConn)
		if !wsConn.AuthToken().IsAuthenticated {
			mehlog.Log(logger, meh.NewInternalErr("websocket connection listener received unauthenticated connection", nil))
			return
		}
		conn := newConnection(logger.Named("wsconn").Named(wsConn.ID().String()), wsConn, sink)
		for receivedMessage := range wsConn.Receive() {
			err := conn.handleReceivedMessage(receivedMessage)
			if err != nil {
				err = meh.Wrap(err, "handle received message", meh.Details{"message": receivedMessage})
				mehlog.Log(conn.logger, err)
				wsConn.SendErr(context.Background(), err)
				continue
			}
		}
		conn.cancelSubscriptionsAndWait()
	}
}

type connection struct {
	logger *zap.Logger
	// accept is true when new subscriptions, etc. are accepted. A call to
	// cancelSubscriptionsAndWait will set it to false.
	accept         bool
	wsConn         wsutil.BaseConnection
	controllerSink ControllerSink
	// subscriptions is a wait group that is done when all subscriptions have
	// finished.
	subscriptions sync.WaitGroup
	// openIntelDeliverySubscriptionsByOperation holds cancel funcs for subscriptions
	// for open intel deliveries by operation id.
	openIntelDeliverySubscriptionsByOperation map[uuid.UUID]context.CancelFunc
	// m locks running and openIntelDeliverySubscriptionsByOperation.
	m sync.Mutex
}

// openIntelDeliveriesListener is the controller.OpenIntelDeliveriesListener that
// adds operation information and calls notify on values.
type openIntelDeliveriesListener struct {
	operationID uuid.UUID
	notify      func(ctx context.Context, operationID uuid.UUID, openDeliveries []store.OpenIntelDeliverySummary) bool
}

func (l openIntelDeliveriesListener) NotifyOpenIntelDeliveries(ctx context.Context, openDeliveries []store.OpenIntelDeliverySummary) bool {
	return l.notify(ctx, l.operationID, openDeliveries)
}

func newConnection(logger *zap.Logger, wsConn wsutil.BaseConnection, controllerSink ControllerSink) *connection {
	return &connection{
		logger:         logger,
		accept:         true,
		wsConn:         wsConn,
		controllerSink: controllerSink,
		openIntelDeliverySubscriptionsByOperation: make(map[uuid.UUID]context.CancelFunc, 0),
	}
}

func (conn *connection) handleReceivedMessage(receivedMessage wsutil.Message) error {
	switch receivedMessage.Type {
	case messageTypeSubscribeOpenIntelDeliveries:
		err := wsutil.ParseAndHandle(receivedMessage, func(message messageSubscribeOpenIntelDeliveries) error {
			err := conn.subscribeOpenIntelDeliveries(message.Operation)
			if err != nil {
				return meh.Wrap(err, "subscribe open intel deliveries", meh.Details{"operation": message.Operation})
			}
			return nil
		})
		if err != nil {
			return meh.Wrap(err, "handle subscribe open intel deliveries message", nil)
		}
	case messageTypeUnsubscribeOpenIntelDeliveries:
		err := wsutil.ParseAndHandle(receivedMessage, func(message messageUnsubscribeOpenIntelDeliveries) error {
			conn.unsubscribeOpenIntelDeliveries(message.Operation)
			return nil
		})
		if err != nil {
			return meh.Wrap(err, "handle unsubscribe open intel deliveries message", nil)
		}
	default:
		return meh.NewBadInputErr("unsupported message type", meh.Details{"message_type": receivedMessage.Type})
	}
	return nil
}

func (conn *connection) subscribeOpenIntelDeliveries(operationID uuid.UUID) error {
	// Assure sufficient permissions.
	err := auth.AssurePermission(conn.wsConn.AuthToken(), permission.ManageIntelDelivery())
	if err != nil {
		return meh.Wrap(err, "assure permission", nil)
	}
	conn.m.Lock()
	defer conn.m.Unlock()
	if !conn.accept {
		return meh.NewBadInputErr("not accepting anymore", nil)
	}
	// Assure not already subscribed.
	if _, ok := conn.openIntelDeliverySubscriptionsByOperation[operationID]; ok {
		return nil
	}
	subscriptionLifetime, cancelSubscription := context.WithCancel(context.Background())
	go func() {
		defer cancelSubscription()
		select {
		case <-conn.wsConn.Lifetime().Done():
			return
		case <-subscriptionLifetime.Done():
			conn.unsubscribeOpenIntelDeliveries(operationID)
			err := conn.notifySubscribedOpenIntelDeliveries(conn.wsConn.Lifetime())
			if err != nil {
				mehlog.Log(conn.logger, meh.Wrap(err, "notify subscribed open intel deliveries after subscription done", meh.Details{
					"operation_id": operationID,
				}))
				return
			}
		}
	}()
	conn.openIntelDeliverySubscriptionsByOperation[operationID] = cancelSubscription
	// Subscribe and forward new open intel deliveries via message.
	conn.subscriptions.Add(1)
	go func() {
		defer conn.subscriptions.Done()
		err := conn.notifySubscribedOpenIntelDeliveries(subscriptionLifetime)
		if err != nil {
			mehlog.Log(conn.logger, meh.Wrap(err, "notify subscribed open intel deliveries", nil))
		}
		conn.controllerSink.ServeOpenIntelDeliveriesListener(subscriptionLifetime, operationID, openIntelDeliveriesListener{
			operationID: operationID,
			notify: func(ctx context.Context, operationID uuid.UUID, openDeliveries []store.OpenIntelDeliverySummary) bool {
				notifyMessage := messageOpenIntelDeliveries{
					Operation:           operationID,
					OpenIntelDeliveries: make([]publicOpenIntelDeliverySummary, 0, len(openDeliveries)),
				}
				for _, openDelivery := range openDeliveries {
					notifyMessage.OpenIntelDeliveries = append(notifyMessage.OpenIntelDeliveries, publicOpenIntelDeliverySummaryFromStore(openDelivery))
				}
				err := conn.wsConn.Send(ctx, messageTypeOpenIntelDeliveries, notifyMessage)
				if err != nil {
					mehlog.Log(conn.logger, meh.Wrap(err, "send open intel deliveries message", meh.Details{"message": notifyMessage}))
					return false
				}
				return true
			},
		})
	}()
	return nil
}

func (conn *connection) notifySubscribedOpenIntelDeliveries(ctx context.Context) error {
	conn.m.Lock()
	defer conn.m.Unlock()
	subscribedOperations := make([]uuid.UUID, 0)
	for operationID := range conn.openIntelDeliverySubscriptionsByOperation {
		subscribedOperations = append(subscribedOperations, operationID)
	}
	err := conn.wsConn.Send(ctx, messageTypeSubscribedOpenIntelDeliveries, messageSubscribedOpenIntelDeliveries{
		Operations: subscribedOperations,
	})
	if err != nil {
		return meh.Wrap(err, "send subscribed open intel deliveries message",
			meh.Details{"subscribed_operations": subscribedOperations})
	}
	return nil
}

func (conn *connection) unsubscribeOpenIntelDeliveries(operationID uuid.UUID) {
	conn.m.Lock()
	defer conn.m.Unlock()
	if cancelSubscription, ok := conn.openIntelDeliverySubscriptionsByOperation[operationID]; ok {
		cancelSubscription()
		delete(conn.openIntelDeliverySubscriptionsByOperation, operationID)
	}
}

// cancelSubscriptionsAndWait the connection by unsubscribing everything and
// waiting until everything has properly finished.
func (conn *connection) cancelSubscriptionsAndWait() {
	// Set to not running and gather all operations with active subscriptions in
	// order to release the lock and then unsubscribe each.
	conn.m.Lock()
	conn.accept = false
	unsubscribe := make([]uuid.UUID, 0, len(conn.openIntelDeliverySubscriptionsByOperation))
	for operationID := range conn.openIntelDeliverySubscriptionsByOperation {
		unsubscribe = append(unsubscribe, operationID)
	}
	conn.m.Unlock()
	// Unsubscribe each and wait until all subscriptions are done.
	for _, operationID := range unsubscribe {
		conn.unsubscribeOpenIntelDeliveries(operationID)
	}
	conn.subscriptions.Wait()
}
