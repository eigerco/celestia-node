//go:build wasm

package nodebuilder

import (
	"context"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/nodebuilder/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/das"
	"github.com/celestiaorg/celestia-node/nodebuilder/fraud"
	modhead "github.com/celestiaorg/celestia-node/nodebuilder/header"
	nodemodule "github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
	"go.uber.org/fx"
)

func ConstructModule(tp nodemodule.Type, network p2p.Network, cfg *Config, store Store) (fx.Option, error) {
	coreModule, err := core.ConstructModule(tp, &cfg.Core)
	if err != nil {
		return nil, err
	}

	p2pModule, err := p2p.ConstructModule(tp, &cfg.P2P)
	if err != nil {
		return nil, err
	}

	// TODO this should not be here. Instead header should be set when config is loaded.
	// Apparently, it's not.
	cfg.Header = modhead.DefaultConfig(tp)
	cfg.DASer = das.DefaultConfig(tp)
	//cfg.State = state.DefaultConfig()
	cfg.Share = share.DefaultConfig(tp)

	shareModule, err := share.ConstructModule(tp, &cfg.Share)
	if err != nil {
		return nil, err
	}

	baseComponents := fx.Options(
		fx.Supply(tp),
		fx.Supply(network),
		fx.Supply(cfg.P2P.BootstrapAddresses),
		fx.Provide(p2p.BootstrappersFor),
		fx.Provide(func(lc fx.Lifecycle) context.Context {
			return fxutil.WithLifecycle(context.Background(), lc)
		}),
		fx.Supply(cfg),
		fx.Supply(store.Config),
		fx.Provide(store.Datastore),
		fx.Provide(store.Keystore),
		fx.Supply(nodemodule.StorePath(store.Path())),
		// modules provided by the node
		p2pModule,
		//state.ConstructModule(tp, &cfg.State, &cfg.Core),
		modhead.ConstructModule[*header.ExtendedHeader](tp, &cfg.Header),
		// TODO Share module is necessary for light node - gets the data from bitswap, is needed for the daser
		shareModule,
		//rpc.ConstructModule(tp, &cfg.RPC),
		//gateway.ConstructModule(tp, &cfg.Gateway), TODO
		coreModule,
		das.ConstructModule(tp, &cfg.DASer),
		fraud.ConstructModule(tp),
		//blob.ConstructModule(),
		// nodemodule.ConstructModule(tp),
		// admin,
	)

	log.Infow("Node builder base components constructed @ nodebuilder/module_wasm.ConstructModule ...")

	return fx.Module(
		"node",
		baseComponents,
	), nil
}
