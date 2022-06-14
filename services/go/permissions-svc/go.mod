module github.com/mobile-directing-system/mds-server/services/go/permissions-svc

go 1.18

replace github.com/mobile-directing-system/mds-server/services/go/shared => ../shared

require (
	github.com/lefinal/meh v1.5.1
	github.com/mobile-directing-system/mds-server/services/go/shared v0.0.0-00010101000000-000000000000
)

require (
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
)
