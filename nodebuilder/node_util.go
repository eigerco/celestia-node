package nodebuilder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
)

var (
	log   = logging.Logger("node")
	fxLog = logging.Logger("fx")
)

// New assembles a new Node with the given type 'tp' over Store 'store'.
func New(tp node.Type, network p2p.Network, store Store, options ...fx.Option) (*Node, error) {
	cfg, err := store.Config()
	if err != nil {
		return nil, err
	}

	return NewWithConfig(tp, network, store, cfg, options...)
}

// NewWithConfig assembles a new Node with the given type 'tp' over Store 'store' and a custom
// config.
func NewWithConfig(tp node.Type, network p2p.Network, store Store, cfg *Config, options ...fx.Option) (*Node, error) {
	mod, err := ConstructModule(tp, network, cfg, store)
	if err != nil {
		return nil, err
	}

	log.Infow("Module construction complete @ node.NewWithConfig()...")

	opts := append([]fx.Option{mod}, options...)
	return newNode(opts...)
}

// Start launches the Node and all its components and services.
func (n *Node) Start(ctx context.Context) error {
	//to := n.Config.Node.StartupTimeout
	to := time.Second * 120 // TODO hardcoded
	ctx, cancel := context.WithTimeout(ctx, to)
	defer cancel()

	err := n.start(ctx)
	if err != nil {
		log.Debugf("error starting %s Node: %s", n.Type, err)
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("node: failed to start within timeout(%s): %w", to, err)
		}
		return fmt.Errorf("node: failed to start: %w", err)
	}

	log.Infof("\n\n/_____/  /_____/  /_____/  /_____/  /_____/ \n\nStarted celestia DA node \nnode "+
		"type: 	%s\nnetwork: 	%s\n\n/_____/  /_____/  /_____/  /_____/  /_____/ \n", strings.ToLower(n.Type.String()),
		n.Network)

	addrs, err := peer.AddrInfoToP2pAddrs(host.InfoFromHost(n.Host))
	if err != nil {
		log.Errorw("Retrieving multiaddress information", "err", err)
		return err
	}
	fmt.Println("The p2p host is listening on:")
	for _, addr := range addrs {
		fmt.Println("* ", addr.String())
	}
	fmt.Println()
	return nil
}

// Run is a Start which blocks on the given context 'ctx' until it is canceled.
// If canceled, the Node is still in the running state and should be gracefully stopped via Stop.
func (n *Node) Run(ctx context.Context) error {
	err := n.Start(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return ctx.Err()
}

// Stop shuts down the Node, all its running Modules/Services and returns.
// Canceling the given context earlier 'ctx' unblocks the Stop and aborts graceful shutdown forcing
// remaining Modules/Services to close immediately.
func (n *Node) Stop(ctx context.Context) error {
	//to := n.Config.Node.ShutdownTimeout
	to := time.Second * 20 // TODO hardcoded
	ctx, cancel := context.WithTimeout(ctx, to)
	defer cancel()

	err := n.stop(ctx)
	if err != nil {
		log.Debugf("error stopping %s Node: %s", n.Type, err)
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("node: failed to stop within timeout(%s): %w", to, err)
		}
		return fmt.Errorf("node: failed to stop: %w", err)
	}

	log.Debugf("stopped %s Node", n.Type)
	return nil
}

// lifecycleFunc defines a type for common lifecycle funcs.
type lifecycleFunc func(context.Context) error
