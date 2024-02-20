//go:build wasm

package share

import (
	"go.uber.org/fx"
)

func bridgeModule(cfg *Config, baseComponents, shrexGetterComponents fx.Option) fx.Option {
	panic("bridge node not supported for wasm")
}

func fullModule(cfg *Config, baseComponents, peerManagerWithShrexPools, shrexGetterComponents fx.Option) fx.Option {
	panic("full node not supported for wasm")
}
