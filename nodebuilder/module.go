//go:build !wasm

package nodebuilder

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/nodebuilder/core"
	modhead "github.com/celestiaorg/celestia-node/nodebuilder/header"
	nodemodule "github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
)

func ConstructModule(tp nodemodule.Type, network p2p.Network, cfg *Config, store Store) (fx.Option, error) {
	log.Infow("Accessing keyring @ construct module...")

	/* 	ks, err := store.Keystore()
	   	if err != nil {
	   		fx.Error(err)
	   	}
	*/
	log.Infow("Keystore created @ construct module...")

	/* 	signer, err := KeyringSigner("", "memory", ks, network) // TODO hardcoded must be changed
	   	if err != nil {
	   		fx.Error(err)
	   	}
	*/
	log.Infow("Keystore signer retrieved @ construct module...")

	//shareModule, err := share.ConstructModule(tp, &cfg.Share)
	//if err != nil {
	//	return nil, err
	//}
	coreModule, err := core.ConstructModule(tp, &cfg.Core)
	if err != nil {
		return nil, err
	}

	log.Infow("Keystore core module discovered @ construct module...")

	p2pModule, err := p2p.ConstructModule(tp, &cfg.P2P)
	if err != nil {
		return nil, err
	}

	log.Infow("Keystore P2P module @ construct module...")

	cfg.Header = modhead.DefaultConfig(tp)

	fmt.Printf("Got header cfg: %+v", cfg.Header)

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
		//fx.Supply(signer),
		// modules provided by the node
		p2pModule,
		//state.ConstructModule(tp, &cfg.State, &cfg.Core),
		modhead.ConstructModule[*header.ExtendedHeader](tp, &cfg.Header),
		//shareModule,
		//rpc.ConstructModule(tp, &cfg.RPC),
		//gateway.ConstructModule(tp, &cfg.Gateway), TODO
		coreModule,
		//das.ConstructModule(tp, &cfg.DASer),
		//fraud.ConstructModule(tp),
		//blob.ConstructModule(),
		//nodemodule.ConstructModule(tp), admin
	)

	log.Infow("Keystore base components constructed @ construct module...")

	return fx.Module(
		"node",
		baseComponents,
	), nil
}
