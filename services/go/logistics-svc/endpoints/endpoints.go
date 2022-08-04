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
	r.GET("/address-book/entries/:entryID", httpendpoints.GinHandlerFunc(logger, secret, handleGetAddressBookEntryByID(s)))
	r.GET("/address-book/entries/:entryID/channels", httpendpoints.GinHandlerFunc(logger, secret, handleGetChannelsByAddressBookEntry(s)))
	r.PUT("/address-book/entries/:entryID/channels", httpendpoints.GinHandlerFunc(logger, secret, handleUpdateChannelsByAddressBookEntry(s)))
	r.PUT("/address-book/entries/:entryID", httpendpoints.GinHandlerFunc(logger, secret, handleUpdateAddressBookEntry(s)))
	r.DELETE("/address-book/entries/:entryID", httpendpoints.GinHandlerFunc(logger, secret, handleDeleteAddressBookEntryByID(s)))
	r.GET("/address-book/entries", httpendpoints.GinHandlerFunc(logger, secret, handleGetAllAddressBookEntries(s)))
	r.POST("/address-book/entries", httpendpoints.GinHandlerFunc(logger, secret, handleCreateAddressBookEntry(s)))

}
