package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"

	"github.com/ipfs/go-datastore"
	connmgri "github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds" //nolint:staticcheck
	"github.com/libp2p/go-libp2p/p2p/net/conngater"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

// connManagerConfig configures connection manager.
type connManagerConfig struct {
	// Low and High are watermarks governing the number of connections that'll be maintained.
	Low, High int
	// GracePeriod is the amount of time a newly opened connection is given before it becomes subject
	// to pruning.
	GracePeriod time.Duration
}

// defaultConnManagerConfig returns defaults for ConnManagerConfig.
func defaultConnManagerConfig(tp node.Type) connManagerConfig {
	switch tp {
	case node.Light:
		return connManagerConfig{
			Low:         50,
			High:        100,
			GracePeriod: time.Minute,
		}
	case node.Bridge, node.Full:
		return connManagerConfig{
			Low:         800,
			High:        1000,
			GracePeriod: time.Minute,
		}
	default:
		panic("unknown node type")
	}
}

// connectionManager provides a constructor for ConnectionManager.
func connectionManager(cfg Config, bpeers Bootstrappers) (connmgri.ConnManager, error) {
	fmt.Println("CONNECTION MANAGER LOADING...")
	fpeers, err := cfg.mutualPeers()
	if err != nil {
		return nil, err
	}
	cm, err := connmgr.NewConnManager(
		cfg.ConnManager.Low,
		cfg.ConnManager.High,
		connmgr.WithGracePeriod(cfg.ConnManager.GracePeriod),
	)
	if err != nil {
		return nil, err
	}
	for _, info := range fpeers {
		cm.Protect(info.ID, "protected-mutual")
	}
	for _, info := range bpeers {
		cm.Protect(info.ID, "protected-bootstrap")
	}

	fmt.Println("CONNECTION MANAGER LOADED")

	return cm, nil
}

// connectionGater constructs a BasicConnectionGater.
func connectionGater(ds datastore.Batching) (*conngater.BasicConnectionGater, error) {
	fmt.Println("CONNECTION GATER LOADING....")
	toReturn, err := conngater.NewBasicConnectionGater(ds)
	fmt.Println("AM I HERE?")
	fmt.Println("CONNECTION GATER LOADED")
	return toReturn, err
}

// peerStore constructs an on-disk PeerStore.
func peerStore(ctx context.Context, ds datastore.Batching) (peerstore.Peerstore, error) {
	fmt.Println("PEERSTORE LOADING....")
	toReturn, err := pstoreds.NewPeerstore(ctx, ds, pstoreds.DefaultOpts())
	fmt.Println("PEERSTORE LOADED")
	return toReturn, err
}
