//go:build !bridge_full

package share

import (
	"errors"
	"go.uber.org/fx"
)

func fullModule(cfg *Config, peerManagerWithShrexPools, baseComponents, shrexGetterComponents fx.Option) (fx.Option, error) {
	return nil, errors.New("the program was built without the full node functionality, needs to be built using `go build -tags=bridge_full`")
}

func bridgeModule(cfg *Config, baseComponents, shrexGetterComponents fx.Option) (fx.Option, error) {
	return nil, errors.New("the program was built without the bridge node functionality, needs to be built using `go build -tags=bridge_full`")
}
