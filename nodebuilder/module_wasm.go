//go:build wasm

package nodebuilder

import (
	"context"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/nodebuilder/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/das"
	"github.com/celestiaorg/celestia-node/nodebuilder/fraud"
	modhead "github.com/celestiaorg/celestia-node/nodebuilder/header"
	nodemodule "github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
)

func ConstructModule(tp nodemodule.Type, network p2p.Network, cfg *Config, store Store) fx.Option {
	coreModule := core.ConstructModule(tp, &cfg.Core)
	p2pModule := p2p.ConstructModule(tp, &cfg.P2P)

	// TODO this should not be here. Instead header should be set when config is loaded.
	// Apparently, it's not.
	cfg.Header = modhead.DefaultConfig(tp)
	cfg.DASer = das.DefaultConfig(tp)
	cfg.Share = share.DefaultConfig(tp)

	shareModule := share.ConstructModule(tp, &cfg.Share)

	baseComponents := fx.Options(
		fx.Supply(tp),
		fx.Supply(network),
		fx.Supply(cfg.P2P.BootstrapAddresses),
		fx.Provide(BootstrappersFor),
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
		modhead.ConstructModule[*header.ExtendedHeader](tp, &cfg.Header),
		// TODO Share module is necessary for light node - gets the data from bitswap, is needed for the daser
		shareModule,
		coreModule,
		das.ConstructModule(tp, &cfg.DASer),
		fraud.ConstructModule(tp),
	)

	log.Infow("Node builder base components constructed @ nodebuilder/module_wasm.ConstructModule ...")

	return fx.Module(
		"node",
		baseComponents,
	)
}

func BootstrappersFor(network p2p.Network, bootstrapAddresses p2p.BootstrapAddresses) (p2p.Bootstrappers, error) {
	networkBootstrappers, err := p2p.BootstrappersFor(network)
	if err != nil {
		return nil, err
	}
	addressBootstrappers, err := p2p.BootstrappersAddresses(bootstrapAddresses)
	if err != nil {
		return nil, err
	}

	return append(networkBootstrappers, addressBootstrappers...), nil
}