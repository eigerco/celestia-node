//go:build bridge_full

package share

import (
	"github.com/celestiaorg/celestia-node/share/p2p/shrexeds"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/share/availability/full"
	"github.com/celestiaorg/celestia-node/share/getters"
)

// ensureEmptyCARExists adds an empty EDS to the provided EDS store.
func ensureEmptyCARExists(ctx context.Context, store *eds.Store) error {
	emptyEDS := share.EmptyExtendedDataSquare()
	emptyDAH, err := da.NewDataAvailabilityHeader(emptyEDS)
	if err != nil {
		return err
	}

	err = store.Put(ctx, emptyDAH.Hash(), emptyEDS)
	if errors.Is(err, dagstore.ErrShardExists) {
		return nil
	}
	return err
}

func bridgeAndFullComponents(cfg *Config) fx.Option {
	return fx.Options(
		fx.Provide(getters.NewStoreGetter),
		fx.Invoke(func(edsSrv *shrexeds.Server, ndSrc *shrexnd.Server) {}),
		fx.Provide(fx.Annotate(
			func(host host.Host, store *eds.Store, network modp2p.Network) (*shrexeds.Server, error) {
				cfg.ShrExEDSParams.WithNetworkID(network.String())
				return shrexeds.NewServer(cfg.ShrExEDSParams, host, store)
			},
			fx.OnStart(func(ctx context.Context, server *shrexeds.Server) error {
				return server.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, server *shrexeds.Server) error {
				return server.Stop(ctx)
			}),
		)),
		fx.Provide(fx.Annotate(
			func(
				host host.Host,
				store *eds.Store,
				network modp2p.Network,
			) (*shrexnd.Server, error) {
				cfg.ShrExNDParams.WithNetworkID(network.String())
				return shrexnd.NewServer(cfg.ShrExNDParams, host, store)
			},
			fx.OnStart(func(ctx context.Context, server *shrexnd.Server) error {
				return server.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, server *shrexnd.Server) error {
				return server.Stop(ctx)
			}),
		)),
		fx.Provide(fx.Annotate(
			func(path node.StorePath, ds datastore.Batching) (*eds.Store, error) {
				return eds.NewStore(cfg.EDSStoreParams, string(path), ds)
			},
			fx.OnStart(func(ctx context.Context, store *eds.Store) error {
				err := store.Start(ctx)
				if err != nil {
					return err
				}
				return ensureEmptyCARExists(ctx, store)
			}),
			fx.OnStop(func(ctx context.Context, store *eds.Store) error {
				return store.Stop(ctx)
			}),
		)),
		fx.Provide(fx.Annotate(
			full.NewShareAvailability,
			fx.OnStart(func(ctx context.Context, avail *full.ShareAvailability) error {
				return avail.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, avail *full.ShareAvailability) error {
				return avail.Stop(ctx)
			}),
		)),
		fx.Provide(func(avail *full.ShareAvailability) share.Availability {
			return avail
		}),
		fx.Provide(func(shrexSub *shrexsub.PubSub) shrexsub.BroadcastFn {
			return shrexSub.Broadcast
		}),
	)
}

func fullModule(cfg *Config, peerManagerWithShrexPools, baseComponents, shrexGetterComponents fx.Option) (fx.Option, error) {
	return fx.Module(
		"share",
		peerManagerWithShrexPools,
		baseComponents,
		bridgeAndFullComponents(cfg),
		shrexGetterComponents,
		fx.Provide(getters.NewIPLDGetter),
		fx.Provide(fullGetter),
	), nil
}

func bridgeModule(cfg *Config, baseComponents, shrexGetterComponents fx.Option) (fx.Option, error) {
	return fx.Module(
		"share",
		baseComponents,
		fx.Provide(peers.NewManager),
		bridgeAndFullComponents(cfg),
		shrexGetterComponents,
		fx.Provide(bridgeGetter),
		fx.Invoke(func(lc fx.Lifecycle, sub *shrexsub.PubSub) error {
			lc.Append(fx.Hook{
				OnStart: sub.Start,
				OnStop:  sub.Stop,
			})
			return nil
		}),
	), nil
}

// ShrexGetter is added to bridge nodes for the case that a shard is removed
// after detected shard corruption. This ensures the block is fetched and stored
// by shrex the next time the data is retrieved (meaning shard recovery is
// manual after corruption is detected).
func bridgeGetter(
	storeGetter *getters.StoreGetter,
	shrexGetter *getters.ShrexGetter,
	cfg Config,
) share.Getter {
	var cascade []share.Getter
	cascade = append(cascade, storeGetter)
	if cfg.UseShareExchange {
		cascade = append(cascade, shrexGetter)
	}
	return getters.NewCascadeGetter(cascade)
}

func fullGetter(
	storeGetter *getters.StoreGetter,
	shrexGetter *getters.ShrexGetter,
	ipldGetter *getters.IPLDGetter,
	cfg Config,
) share.Getter {
	var cascade []share.Getter
	cascade = append(cascade, storeGetter)
	if cfg.UseShareExchange {
		cascade = append(cascade, shrexGetter)
	}
	cascade = append(cascade, ipldGetter)
	return getters.NewCascadeGetter(cascade)
}
