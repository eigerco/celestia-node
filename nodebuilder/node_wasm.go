//go:build wasm

package nodebuilder

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

// newNode creates a new Node from given DI options.
// DI options allow initializing the Node with a customized set of components and services.
// NOTE: newNode is currently meant to be used privately to create various custom Node types e.g.
// Light, unless we decide to give package users the ability to create custom node types themselves.
func newNode(opts ...fx.Option) (*Node, error) {
	toReturn := new(Node)
	toReturn.Type = node.Light // TODO figure this shit out - should not be here...

	log.Infow("Node memory footprint initialized @ node.newNode()...")

	app := fx.New(
		/* 		fx.WithLogger(func() fxevent.Logger {
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
