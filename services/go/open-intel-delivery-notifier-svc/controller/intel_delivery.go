package controller

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/pgutil"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	openIntelDeliveryWatcherPeriodicNotifyInterval = 5 * time.Second
	openIntelDeliveryWatcherNotifyDelay            = 100 * time.Millisecond
)

// openIntelDeliveriesNotifyHub is a hub for notifying about open intel
// deliveries. Use registerListener in order to subscribe to changes and feed for
// notifying about updated values.
type openIntelDeliveriesNotifyHub struct {
	// listeners holds the amount of all listeners that have subscribed to open intel
	// delivery notifications for the operation with operationID.
	listeners int
	// listenersMutex locks listeners.
	listenersMutex sync.Mutex
	// currentAt is the last timestamp when currentOpenDeliveries was updated.
	currentAt time.Time
	// currentOpenDeliveries holds all open intel deliveries at the moment of
	// currentAt.
	currentOpenDeliveries []store.OpenIntelDeliverySummary
	// currentCond locks currentAt and currentOpenDeliveries.
	currentCond sync.Cond
}

func newOpenIntelDeliveriesNotifierHub() *openIntelDeliveriesNotifyHub {
	return &openIntelDeliveriesNotifyHub{
		listeners:             0,
		currentAt:             time.Now(),
		currentOpenDeliveries: make([]store.OpenIntelDeliverySummary, 0),
		currentCond:           sync.Cond{L: &sync.Mutex{}},
	}
}

// copyOpenDeliveries creates a deep copy of the given
// store.OpenIntelDeliverySummary list.
func copyOpenDeliveries(openDeliveries []store.OpenIntelDeliverySummary) []store.OpenIntelDeliverySummary {
	copied := make([]store.OpenIntelDeliverySummary, 0, len(openDeliveries))
	for _, openDelivery := range openDeliveries {
		copied = append(copied, store.OpenIntelDeliverySummary{
			Delivery: store.ActiveIntelDelivery{
				ID:    openDelivery.Delivery.ID,
				Intel: openDelivery.Delivery.Intel,
				To:    openDelivery.Delivery.To,
				Note:  openDelivery.Delivery.Note,
			},
			Intel: store.Intel{
				ID:         openDelivery.Intel.ID,
				CreatedAt:  openDelivery.Intel.CreatedAt,
				CreatedBy:  openDelivery.Intel.CreatedBy,
				Operation:  openDelivery.Intel.Operation,
				Importance: openDelivery.Intel.Importance,
				IsValid:    openDelivery.Intel.IsValid,
			},
		})
	}
	return copied
}

// registerListener notifies via the given OpenIntelDeliveriesListener about
// updated open intel deliveries. The returned function cancels the subscription.
func (hub *openIntelDeliveriesNotifyHub) registerListener(listener OpenIntelDeliveriesListener) func() {
	lifetime, cancel := context.WithCancel(context.Background())
	running := atomic.NewBool(true)
	// Register.
	hub.listenersMutex.Lock()
	hub.listeners++
	hub.listenersMutex.Unlock()
	hub.currentCond.L.Lock()
	go func() {
		// Last timestamp of the update moment for the open deliveries we notified about.
		var lastNotifyForUpdateAt time.Time
		for {
			// Wait for updated deliveries.
			for !(lastNotifyForUpdateAt.Before(hub.currentAt) || !running.Load()) {
				hub.currentCond.Wait()
			}
			if !running.Load() {
				hub.currentCond.L.Unlock()
				return
			}
			// Copy everything and unlock in order to not block the hub for other listeners.
			currentOpenDeliveries := copyOpenDeliveries(hub.currentOpenDeliveries)
			currentOpenDeliveriesAt := hub.currentAt
			hub.currentCond.L.Unlock()
			// Notify.
			deliveryOK := listener.NotifyOpenIntelDeliveries(lifetime, currentOpenDeliveries)
			if deliveryOK {
				lastNotifyForUpdateAt = currentOpenDeliveriesAt
			}
			hub.currentCond.L.Lock()
		}
	}()

	return func() {
		cancel()
		running.Store(false)
		// Unregister.
		hub.listenersMutex.Lock()
		hub.listeners--
		hub.listenersMutex.Unlock()
	}
}

// feed the given updated list of open intel deliveries.
func (hub *openIntelDeliveriesNotifyHub) feed(openDeliveries []store.OpenIntelDeliverySummary) {
	hub.currentCond.L.Lock()
	defer hub.currentCond.L.Unlock()
	hub.currentOpenDeliveries = openDeliveries
	hub.currentAt = time.Now()
	hub.currentCond.Broadcast()
}

type openIntelDeliveryWatcher struct {
	logger    *zap.Logger
	isRunning *atomic.Bool
	// notifierHub takes the latest state of open intel deliveries and manages
	// propagating updates to subscribers.
	notifierHub *openIntelDeliveriesNotifyHub
	// doNotify describes whether an update notification is wanted. This is a field
	// because if runNew detects this as true, it will wait for a given delay before
	// performing notifications and then set this to false.
	doNotify bool
	// doNotifyCond locks doNotify.
	doNotifyCond *sync.Cond
}

// notifyIntelDeliveryChanged notifies that an intel delivery, that is observed
// by the watcher, has changed.
func (w *openIntelDeliveryWatcher) notifyIntelDeliveryChanged() {
	w.doNotifyCond.L.Lock()
	defer w.doNotifyCond.L.Unlock()
	w.doNotify = true
	w.doNotifyCond.Broadcast()
}

// runNewOpenIntelDeliveryWatcher the watcher. Notifications will be performed in
// the given interval or if notifyIntelDeliveryChanged is called. Then, a delay
// is added as given in order to avoid burst updates. The given function is used
// for retrieving current open intel deliveries that are being propagated to
// listening listeners.
func runNewOpenIntelDeliveryWatcher(logger *zap.Logger, periodicNotificationInterval time.Duration,
	notifyDelay time.Duration, retrieveCurrentOpenIntelDeliveries func() ([]store.OpenIntelDeliverySummary, error)) *openIntelDeliveryWatcher {
	watcher := &openIntelDeliveryWatcher{
		isRunning:    atomic.NewBool(true),
		logger:       logger,
		notifierHub:  newOpenIntelDeliveriesNotifierHub(),
		doNotify:     true,
		doNotifyCond: sync.NewCond(&sync.Mutex{}),
	}
	// Periodic updates and lifetime watcher.
	go func() {
		for {
			if !watcher.isRunning.Load() {
				return
			}
			watcher.doNotifyCond.L.Lock()
			watcher.doNotify = true
			watcher.doNotifyCond.Broadcast()
			watcher.doNotifyCond.L.Unlock()
			<-time.After(periodicNotificationInterval)
		}
	}()
	// Wait until notification is desired and notify.
	go func() {
		for {
			watcher.doNotifyCond.L.Lock()
			for !(watcher.doNotify || !watcher.isRunning.Load()) {
				watcher.doNotifyCond.Wait()
			}
			watcher.doNotifyCond.L.Unlock()
			if !watcher.isRunning.Load() {
				return
			}
			// Wait for given delay.
			<-time.After(notifyDelay)
			if !watcher.isRunning.Load() {
				return
			}
			watcher.doNotifyCond.L.Lock()
			watcher.doNotify = false
			watcher.doNotifyCond.L.Unlock()
			openIntelDeliveries, err := retrieveCurrentOpenIntelDeliveries()
			// Update.
			if err != nil {
				mehlog.Log(watcher.logger, meh.Wrap(err, "retrieve current open intel deliveries", nil))
				continue
			}
			watcher.notifierHub.feed(openIntelDeliveries)
		}
	}()
	return watcher
}

func (w *openIntelDeliveryWatcher) shutdown() {
	w.isRunning.Store(false)
}

// ServeOpenIntelDeliveriesListener notifies the given
// OpenIntelDeliveriesListener about updates open intel deliveries for the intel,
// associated with the operation with the given id, until the given context is
// done.
func (c *Controller) ServeOpenIntelDeliveriesListener(lifetime context.Context, operationID uuid.UUID, notifier OpenIntelDeliveriesListener) {
	// Register at watcher or create and run new watcher if not present.
	c.openIntelDeliveryWatchersByOperationMutex.Lock()
	watcher, ok := c.openIntelDeliveryWatchersByOperation[operationID]
	if !ok {
		watcherLogger := c.logger.Named("open-intel-delivery-watcher").Named("operation").Named(operationID.String())
		watcherLogger.Debug("spawning")
		watcher = runNewOpenIntelDeliveryWatcher(watcherLogger, openIntelDeliveryWatcherPeriodicNotifyInterval, openIntelDeliveryWatcherNotifyDelay,
			func() ([]store.OpenIntelDeliverySummary, error) {
				var summaries []store.OpenIntelDeliverySummary
				err := pgutil.RunInTx(context.Background(), c.db, func(ctx context.Context, tx pgx.Tx) error {
					var err error
					summaries, err = c.store.OpenIntelDeliveriesByOperation(ctx, tx, operationID)
					if err != nil {
						return meh.Wrap(err, "open intel deliveries by operation from store", meh.Details{"operation_id": operationID})
					}
					return nil
				})
				if err != nil {
					return nil, meh.Wrap(err, "run in tx", nil)
				}
				return summaries, nil
			})
	}
	cancelListener := watcher.notifierHub.registerListener(notifier)
	c.openIntelDeliveryWatchersByOperation[operationID] = watcher
	c.openIntelDeliveryWatchersByOperationMutex.Unlock()
	// Wait until lifetime done and remove watcher if no listeners are listening
	// anymore.
	<-lifetime.Done()
	c.openIntelDeliveryWatchersByOperationMutex.Lock()
	defer c.openIntelDeliveryWatchersByOperationMutex.Unlock()
	cancelListener()
	watcher.notifierHub.listenersMutex.Lock()
	defer watcher.notifierHub.listenersMutex.Unlock()
	if watcher.notifierHub.listeners > 1 {
		return
	}
	// No more listeners, remove the watcher.
	watcher.shutdown()
	delete(c.openIntelDeliveryWatchersByOperation, operationID)
}

// notifyIntelDeliveryChanged notifies potential watchers for the operation with
// the given id that intel deliveries have changed.
func (c *Controller) notifyIntelDeliveryChanged(operationID uuid.UUID) {
	c.openIntelDeliveryWatchersByOperationMutex.RLock()
	defer c.openIntelDeliveryWatchersByOperationMutex.RUnlock()
	if watcher, ok := c.openIntelDeliveryWatchersByOperation[operationID]; ok {
		watcher.notifyIntelDeliveryChanged()
	}
}

// CreateActiveIntelDelivery creates the given store.ActiveIntelDelivery in the
// store.
func (c *Controller) CreateActiveIntelDelivery(ctx context.Context, create store.ActiveIntelDelivery) error {
	var operationID uuid.UUID
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		err := c.store.CreateActiveIntelDelivery(ctx, tx, create)
		if err != nil {
			return meh.Wrap(err, "create active intel delivery in store", meh.Details{"create": create})
		}
		intel, err := c.store.IntelByID(ctx, tx, create.Intel)
		if err != nil {
			return meh.Wrap(err, "intel by id from store", meh.Details{"intel_id": create.Intel})
		}
		operationID = intel.Operation
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	c.notifyIntelDeliveryChanged(operationID)
	return nil
}

// DeleteActiveIntelDeliveryByID deletes the intel delivery with the given id
// from the store.
func (c *Controller) DeleteActiveIntelDeliveryByID(ctx context.Context, deliveryID uuid.UUID) error {
	var operationID uuid.UUID
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		operationID, err = c.store.IntelOperationByDelivery(ctx, tx, deliveryID)
		if err != nil {
			return meh.Wrap(err, "intel operation by delivery", meh.Details{"delivery_id": deliveryID})
		}
		err = c.store.DeleteActiveIntelDeliveryByID(ctx, tx, deliveryID)
		if err != nil {
			return meh.Wrap(err, "delete active intel delivery in store", meh.Details{"delivery_id": deliveryID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	c.notifyIntelDeliveryChanged(operationID)
	return nil
}

// CreateActiveIntelDeliveryAttempt creates the given
// store.ActiveIntelDeliveryAttempt.
func (c *Controller) CreateActiveIntelDeliveryAttempt(ctx context.Context, create store.ActiveIntelDeliveryAttempt) error {
	var operationID uuid.UUID
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		err := c.store.CreateActiveIntelDeliveryAttempt(ctx, tx, create)
		if err != nil {
			return meh.Wrap(err, "create active intel delivery attempt in store", meh.Details{"create": create})
		}
		operationID, err = c.store.IntelOperationByDeliveryAttempt(ctx, tx, create.ID)
		if err != nil {
			// Only log the error for better error recovery as retrieving the operation id
			// involves relying on many other parts to exist and notification is not urgent.
			operationID = uuid.Nil
			mehlog.Log(c.logger, meh.Wrap(err, "get intel operation by created intel delivery attempt",
				meh.Details{"created_attempt": create}))
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	if !operationID.IsNil() {
		c.notifyIntelDeliveryChanged(operationID)
	}
	return nil
}

// DeleteActiveIntelDeliveryAttemptByID deletes the
// store.ActiveIntelDeliveryAttempt with the given id.
func (c *Controller) DeleteActiveIntelDeliveryAttemptByID(ctx context.Context, attemptID uuid.UUID) error {
	var operationID uuid.UUID
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		operationID, err = c.store.IntelOperationByDeliveryAttempt(ctx, tx, attemptID)
		if err != nil {
			// Only log the error for better error recovery as retrieving the operation id
			// involves relying on many other parts to exist and notification is not urgent.
			operationID = uuid.Nil
			mehlog.Log(c.logger, meh.Wrap(err, "get intel operation by intel delivery attempt to be deleted",
				meh.Details{"attempt_id": attemptID}))
		}
		err = c.store.DeleteActiveIntelDeliveryAttemptByID(ctx, tx, attemptID)
		if err != nil {
			return meh.Wrap(err, "delete active intel delivery attempt in store", meh.Details{"attempt_id": attemptID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	if !operationID.IsNil() {
		c.notifyIntelDeliveryChanged(operationID)
	}
	return nil
}

// SetAutoIntelDeliveryEnabledForAddressBookEntry sets the auto-intel-delivery
// flag for the address book entry with the given id.
func (c *Controller) SetAutoIntelDeliveryEnabledForAddressBookEntry(ctx context.Context, entryID uuid.UUID, enabled bool) error {
	var affectedOperations []uuid.UUID
	err := pgutil.RunInTx(ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
		err := c.store.SetAutoIntelDeliveryEnabledForEntry(ctx, tx, entryID, enabled)
		if err != nil {
			return meh.Wrap(err, "set auto-intel-delivery enabled for entry in store", meh.Details{
				"entry_id": entryID,
				"enabled":  enabled,
			})
		}
		affectedOperations, err = c.store.IntelOperationsByActiveIntelDeliveryRecipient(ctx, tx, entryID)
		if err != nil {
			return meh.Wrap(err, "affected intel operations", meh.Details{"entry_id": entryID})
		}
		return nil
	})
	if err != nil {
		return meh.Wrap(err, "run in tx", nil)
	}
	for _, affectedOperation := range affectedOperations {
		c.notifyIntelDeliveryChanged(affectedOperation)
	}
	return nil
}
