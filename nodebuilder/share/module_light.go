//go:build light

package share

import (
	"context"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/celestia-node/share/availability/light"
	"github.com/celestiaorg/celestia-node/share/getters"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
)

func lightModule(baseComponents, peerManagerWithShrexPools, shrexGetterComponents fx.Option, sampleAmount uint) (fx.Option, error) {
	return fx.Module(
		"share",
		baseComponents,
		fx.Provide(func() []light.Option {
			return []light.Option{
				light.WithSampleAmount(sampleAmount),
			}
		}),
		peerManagerWithShrexPools,
		shrexGetterComponents,
		fx.Invoke(ensureEmptyEDSInBS),
		fx.Provide(getters.NewIPLDGetter),
		fx.Provide(lightGetter),
		// shrexsub broadcaster stub for daser
		fx.Provide(func() shrexsub.BroadcastFn {
			return func(context.Context, shrexsub.Notification) error {
				return nil
			}
		}),
		fx.Provide(fx.Annotate(
			light.NewShareAvailability,
			fx.OnStop(func(ctx context.Context, la *light.ShareAvailability) error {
				return la.Close(ctx)
			}),
		)),
		fx.Provide(func(avail *light.ShareAvailability) share.Availability {
			return avail
		}),
	), nil
}
