//go:build bridge_full

package p2p

import "go.uber.org/fx"

func bridgeFullModule(baseComponents fx.Option) (fx.Option, error) {
	return fx.Module(
		"p2p",
		baseComponents,
		fx.Provide(blockstoreFromEDSStore),
		fx.Provide(infiniteResources),
	), nil
}

func blockstoreFromEDSStore(ctx context.Context, store *eds.Store) (blockstore.Blockstore, error) {
	return blockstore.CachedBlockstore(
		ctx,
		store.Blockstore(),
		blockstore.CacheOpts{
			HasTwoQueueCacheSize: defaultARCCacheSize,
		},
	)
}
