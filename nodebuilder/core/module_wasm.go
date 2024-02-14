//go:build wasm

package core

import (
	"fmt"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

// ConstructModule collects all the components and services related to managing the relationship
// with the Core node.
func ConstructModule(tp node.Type, cfg *Config, options ...fx.Option) fx.Option {
	// sanitize config values before constructing module
	cfgErr := cfg.Validate()

	baseComponents := fx.Options(
		fx.Supply(*cfg),
		fx.Error(cfgErr),
		fx.Options(options...),
	)

	switch tp {
	case node.Light:
		return fx.Module("core", baseComponents)
	case node.Bridge, node.Full:
		panic(fmt.Sprintf("node type %q not supported for war wasm", tp))
	default:
		panic("invalid node type")
	}
}
