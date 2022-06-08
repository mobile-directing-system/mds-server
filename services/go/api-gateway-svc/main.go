package main

import (
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/app"
	"github.com/mobile-directing-system/mds-server/services/go/shared/waitforterminate"
	"go.uber.org/zap"
)

func main() {
	err := waitforterminate.Run(app.Run)
	if err != nil {
		logger, _ := zap.NewProduction()
		mehlog.Log(logger, meh.Wrap(err, "run", nil))
	}
}
