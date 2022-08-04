package main

import (
	"context"
	"github.com/mobile-directing-system/mds-server/services/go/logistics-svc/app"
	"github.com/mobile-directing-system/mds-server/services/go/shared/waitforterminate"
	"log"
	_ "net/http/pprof"
)

func main() {
	err := waitforterminate.Run(func(ctx context.Context) error {
		return app.Run(ctx)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
