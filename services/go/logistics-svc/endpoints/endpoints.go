package endpoints

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/httpendpoints"
	"go.uber.org/zap"
)

// Store are the handle dependencies.
type Store interface {
	handleGetAddressBookEntryByIDStore
	handleUpdateAddressBookEntryStore
	handleCreateAddressBookEntryStore
	handleUpdateChannelsByAddressBookEntryStore
	handleGetAllAddressBookEntriesStore
	handleDeleteAddressBookEntryByIDStore
	handleGetChannelsByAddressBookEntryStore
	handleSearchAddressBookEntriesStore
	handleRebuildAddressBookEntrySearchStore
	handleMarkIntelDeliveryAsDeliveredStore
	handleMarkIntelDeliveryAttemptAsDeliveredStore
	handleSearchIntelStore
	handleCreateIntelStore
	handleGetIntelByIDStore
	handleInvalidateIntelByIDStore
	handleRebuildIntelSearchStore
	handleGetAllIntelStore
	handleCreateIntelDeliveryAttemptForDeliveryStore
	handleGetIntelDeliveryAttemptsByDeliveryStore
	handleSetAddressBookEntriesWithAutoDeliveryEnabledStore
}

// Serve the endpoints via HTTP.
func Serve(lifetime context.Context, logger *zap.Logger, addr string, authSecret string, s Store) error {
	httpendpoints.ApplyDefaultErrorHTTPMapping()
	r := httpendpoints.NewEngine(logger)
	populateRoutes(r, logger, authSecret, s)
	err := httpendpoints.Serve(lifetime, r, addr)
	if err != nil {
		return meh.Wrap(err, "serve", meh.Details{"addr": addr})
	}
	return nil
}

func populateRoutes(r *gin.Engine, logger *zap.Logger, secret string, s Store) {
	r.PUT("/address-book/entries-with-auto-intel-delivery", httpendpoints.GinHandlerFunc(logger, secret, handleSetAddressBookEntriesWithAutoDeliveryEnabled(s)))
	r.GET("/address-book/entries/search", httpendpoints.GinHandlerFunc(logger, secret, handleSearchAddressBookEntries(s)))
	r.POST("/address-book/entries/search/rebuild", httpendpoints.GinHandlerFunc(logger, secret, handleRebuildAddressBookEntrySearch(s)))
	r.GET("/address-book/entries/:entryID", httpendpoints.GinHandlerFunc(logger, secret, handleGetAddressBookEntryByID(s)))
	r.GET("/address-book/entries/:entryID/channels", httpendpoints.GinHandlerFunc(logger, secret, handleGetChannelsByAddressBookEntry(s)))
	r.PUT("/address-book/entries/:entryID/channels", httpendpoints.GinHandlerFunc(logger, secret, handleUpdateChannelsByAddressBookEntry(s)))
	r.PUT("/address-book/entries/:entryID", httpendpoints.GinHandlerFunc(logger, secret, handleUpdateAddressBookEntry(s)))
	r.DELETE("/address-book/entries/:entryID", httpendpoints.GinHandlerFunc(logger, secret, handleDeleteAddressBookEntryByID(s)))
	r.GET("/address-book/entries", httpendpoints.GinHandlerFunc(logger, secret, handleGetAllAddressBookEntries(s)))
	r.POST("/address-book/entries", httpendpoints.GinHandlerFunc(logger, secret, handleCreateAddressBookEntry(s)))
	r.POST("/intel", httpendpoints.GinHandlerFunc(logger, secret, handleCreateIntel(s)))
	r.GET("/intel", httpendpoints.GinHandlerFunc(logger, secret, handleGetAllIntel(s)))
	r.GET("/intel/search", httpendpoints.GinHandlerFunc(logger, secret, handleSearchIntel(s)))
	r.POST("/intel/search/rebuild", httpendpoints.GinHandlerFunc(logger, secret, handleRebuildIntelSearch(s)))
	r.GET("/intel/:intelID", httpendpoints.GinHandlerFunc(logger, secret, handleGetIntelByID(s)))
	r.POST("/intel/:intelID/invalidate", httpendpoints.GinHandlerFunc(logger, secret, handleInvalidateIntelByID(s)))
	r.GET("/intel-deliveries/:deliveryID/attempts", httpendpoints.GinHandlerFunc(logger, secret, handleGetIntelDeliveryAttemptsByDelivery(s)))
	r.POST("/intel-deliveries/:deliveryID/delivered", httpendpoints.GinHandlerFunc(logger, secret, handleMarkIntelDeliveryAsDelivered(s)))
	r.POST("/intel-deliveries/:deliveryID/deliver/channel/:channelID", httpendpoints.GinHandlerFunc(logger, secret, handleCreateIntelDeliveryAttemptForDelivery(s)))
	r.POST("/intel-delivery-attempts/:attemptID/delivered", httpendpoints.GinHandlerFunc(logger, secret, handleMarkIntelDeliveryAttemptAsDelivered(s)))
}
