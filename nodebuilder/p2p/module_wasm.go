//go:build wasm

package p2p

import (
	"fmt"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/metrics"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/share/ipld"
)

var log = logging.Logger("module/p2p")

// ConstructModule collects all the components and services related to p2p.
func ConstructModule(tp node.Type, cfg *Config) fx.Option {
	// sanitize config values before constructing module
	cfgErr := cfg.Validate()

	baseComponents := fx.Options(
		fx.Supply(*cfg),
		fx.Error(cfgErr),
		fx.Provide(Key),
		fx.Provide(id),
		fx.Provide(peerStore),
		fx.Provide(connectionManager),
		fx.Provide(connectionGater),
		fx.Provide(host),
		fx.Provide(routedHost),
		fx.Provide(pubSub),
		fx.Provide(dataExchange),
		fx.Provide(ipld.NewBlockservice),
		fx.Provide(peerRouting),
		fx.Provide(contentRouting),
		fx.Provide(addrsFactory(cfg.AnnounceAddresses, cfg.NoAnnounceAddresses)),
		fx.Provide(metrics.NewBandwidthCounter),
		fx.Provide(newModule),
		fx.Provide(resourceManager),
	)

	switch tp {
	case node.Full, node.Bridge:
		panic(fmt.Sprintf("node type %q not supported for wasm", tp))
	case node.Light:
		return fx.Module(
			"p2p",
			baseComponents,
			fx.Provide(blockstoreFromDatastore),
			fx.Provide(autoscaleResources),
		)
	default:
		panic("invalid node type")
	}
}
