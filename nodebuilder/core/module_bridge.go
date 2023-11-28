//go:build bridge_full

package core

import (
	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/share/eds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
	libhead "github.com/celestiaorg/go-header"
	"go.uber.org/fx"
)

func moduleBridge() fx.Option {
	return fx.Module("core",
		baseComponents,
		fx.Provide(core.NewBlockFetcher),
		fxutil.ProvideAs(core.NewExchange, new(libhead.Exchange[*header.ExtendedHeader])),
		fx.Invoke(fx.Annotate(
			func(
				bcast libhead.Broadcaster[*header.ExtendedHeader],
				fetcher *core.BlockFetcher,
				pubsub *shrexsub.PubSub,
				construct header.ConstructFn,
				store *eds.Store,
			) *core.Listener {
				return core.NewListener(bcast, fetcher, pubsub.Broadcast, construct, store, p2p.BlockTime)
			},
			fx.OnStart(func(ctx context.Context, listener *core.Listener) error {
				return listener.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, listener *core.Listener) error {
				return listener.Stop(ctx)
			}),
		)),
		fx.Provide(fx.Annotate(
			remote,
			fx.OnStart(func(ctx context.Context, client core.Client) error {
				return client.Start()
			}),
			fx.OnStop(func(ctx context.Context, client core.Client) error {
				return client.Stop()
			}),
		)),
	)
}
