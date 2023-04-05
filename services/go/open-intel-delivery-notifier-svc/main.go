package main

import (
	"context"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/mobile-directing-system/mds-server/services/go/open-intel-delivery-notifier-svc/app"
	"github.com/mobile-directing-system/mds-server/services/go/shared/waitforterminate"
	"log"
)

func main() {
	err := waitforterminate.Run(func(ctx context.Context) error {
		return app.Run(ctx)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
