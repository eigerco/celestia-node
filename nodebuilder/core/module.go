package core

import (
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"go.uber.org/fx"
)

// ConstructModule collects all the components and services related to managing the relationship
// with the Core node.
func ConstructModule(tp node.Type, cfg *Config, options ...fx.Option) (fx.Option, error) {
	// sanitize config values before constructing module
	cfgErr := cfg.Validate()

	baseComponents := fx.Options(
		fx.Supply(*cfg),
		fx.Error(cfgErr),
		fx.Options(options...),
	)

	switch tp {
	case node.Light, node.Full:
		return fx.Module("core", baseComponents), nil
	case node.Bridge:
		return moduleBridge()
	default:
		panic("invalid node type")
	}
}
