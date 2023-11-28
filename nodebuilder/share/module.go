package share

import (
	"context"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/net/conngater"
	"go.uber.org/fx"

	libhead "github.com/celestiaorg/go-header"
	"github.com/celestiaorg/go-header/sync"

	"github.com/celestiaorg/celestia-node/header"
	modp2p "github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/share/getters"
	disc "github.com/celestiaorg/celestia-node/share/p2p/discovery"
	"github.com/celestiaorg/celestia-node/share/p2p/peers"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexeds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexnd"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
)

func ConstructModule(tp node.Type, cfg *Config, options ...fx.Option) (fx.Option, error) {
	// sanitize config values before constructing module
	cfgErr := cfg.Validate(tp)

	baseComponents := fx.Options(
		fx.Supply(*cfg),
		fx.Error(cfgErr),
		fx.Options(options...),
		fx.Provide(newModule),
		fx.Invoke(func(disc *disc.Discovery) {}),
		fx.Provide(fx.Annotate(
			newDiscovery(cfg.Discovery),
			fx.OnStart(func(ctx context.Context, d *disc.Discovery) error {
				return d.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, d *disc.Discovery) error {
				return d.Stop(ctx)
			}),
		)),
		fx.Provide(
			func(ctx context.Context, h host.Host, network modp2p.Network) (*shrexsub.PubSub, error) {
				return shrexsub.NewPubSub(ctx, h, network.String())
			},
		),
	)

	shrexGetterComponents := fx.Options(
		fx.Provide(func() peers.Parameters {
			return cfg.PeerManagerParams
		}),
		fx.Provide(
			func(host host.Host, network modp2p.Network) (*shrexnd.Client, error) {
				cfg.ShrExNDParams.WithNetworkID(network.String())
				return shrexnd.NewClient(cfg.ShrExNDParams, host)
			},
		),
		fx.Provide(
			func(host host.Host, network modp2p.Network) (*shrexeds.Client, error) {
				cfg.ShrExEDSParams.WithNetworkID(network.String())
				return shrexeds.NewClient(cfg.ShrExEDSParams, host)
			},
		),
		fx.Provide(fx.Annotate(
			getters.NewShrexGetter,
			fx.OnStart(func(ctx context.Context, getter *getters.ShrexGetter) error {
				return getter.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, getter *getters.ShrexGetter) error {
				return getter.Stop(ctx)
			}),
		)),
	)

	peerManagerWithShrexPools := fx.Options(
		fx.Provide(
			func(
				params peers.Parameters,
				host host.Host,
				connGater *conngater.BasicConnectionGater,
				shrexSub *shrexsub.PubSub,
				headerSub libhead.Subscriber[*header.ExtendedHeader],
				// we must ensure Syncer is started before PeerManager
				// so that Syncer registers header validator before PeerManager subscribes to headers
				_ *sync.Syncer[*header.ExtendedHeader],
			) (*peers.Manager, error) {
				return peers.NewManager(
					params,
					host,
					connGater,
					peers.WithShrexSubPools(shrexSub, headerSub),
				)
			},
		),
	)

	switch tp {
	case node.Bridge:
		return bridgeModule(cfg, baseComponents, shrexGetterComponents)
	case node.Full:
		return fullModule(cfg, peerManagerWithShrexPools, baseComponents, shrexGetterComponents)
	case node.Light:
		return lightModule(baseComponents, peerManagerWithShrexPools, shrexGetterComponents, cfg.LightAvailability.SampleAmount)
	default:
		panic("invalid node type")
	}
}
