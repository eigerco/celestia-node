//go:build wasm

package rpc

import (
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"go.uber.org/fx"
)

func ConstructModule(tp node.Type, cfg *Config) fx.Option {
	return fx.Module("rpc")
}
