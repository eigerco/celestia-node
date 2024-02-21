//go:build wasm

package nodebuilder

import (
	"github.com/celestiaorg/celestia-node/nodebuilder/das"
	"github.com/ipfs/boxo/blockservice"
	"github.com/ipfs/boxo/exchange"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/conngater"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/nodebuilder/fraud"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
)

// Node represents the core structure of a Celestia node. It keeps references to all
// Celestia-specific components and services in one place and provides flexibility to run a
// Celestia light node. Only the light node is supported for wasm for now.
type Node struct {
	fx.In `ignore-unexported:"true"`

	Type          node.Type
	Network       p2p.Network
	Bootstrappers p2p.Bootstrappers
	Config        *Config

	// p2p components
	Host         host.Host
	ConnGater    *conngater.BasicConnectionGater
	Routing      routing.PeerRouting
	DataExchange exchange.Interface
	BlockService blockservice.BlockService
	// p2p protocols
	PubSub *pubsub.PubSub
	// services
	ShareServ  share.Module
	HeaderServ header.Module
	DASer      das.Module // not optional
	FraudServ  fraud.Module

	// start and stop control ref internal fx.App lifecycle funcs to be called from Start and Stop
	start, stop lifecycleFunc
}

// newNode creates a new Node from given DI options.
// DI options allow initializing the Node with a customized set of components and services.
// NOTE: newNode is currently meant to be used privately to create various custom Node types e.g.
// Light, unless we decide to give package users the ability to create custom node types themselves.
func newNode(opts ...fx.Option) (*Node, error) {
	toReturn := new(Node)
	toReturn.Type = node.Light // TODO figure this shit out - should not be here...

	log.Infow("Node memory footprint initialized @ node.newNode()...")

	app := fx.New(
		/* 		fx.WithLogger(func() fxevent.Logger { TODO introduce the logger?
			zl := &fxevent.ZapLogger{Logger: log.Desugar()}
			zl.UseLogLevel(zapcore.DebugLevel)
			return zl
		}), */
		fx.Populate(toReturn),
		fx.Options(opts...),
	)

	if err := app.Err(); err != nil {
		log.Errorf("error creating %s Node: %s", toReturn.Type, err)
		return nil, err
	}

	log.Infow("Node created @ node.newNode()...")

	toReturn.start, toReturn.stop = app.Start, app.Stop
	return toReturn, nil
}
