package share

import (
	"context"

	"github.com/ipfs/boxo/blockservice"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	routingdisc "github.com/libp2p/go-libp2p/p2p/discovery/routing"

	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/celestia-node/share/getters"
	"github.com/celestiaorg/celestia-node/share/ipld"
	disc "github.com/celestiaorg/celestia-node/share/p2p/discovery"
	"github.com/celestiaorg/celestia-node/share/p2p/peers"
)

const (
	// fullNodesTag is the tag used to identify full nodes in the discovery service.
	fullNodesTag = "full"
)

func newDiscovery(cfg *disc.Parameters,
) func(routing.ContentRouting, host.Host, *peers.Manager) (*disc.Discovery, error) {
	return func(
		r routing.ContentRouting,
		h host.Host,
		manager *peers.Manager,
	) (*disc.Discovery, error) {
		return disc.NewDiscovery(
			cfg,
			h,
			routingdisc.NewRoutingDiscovery(r),
			fullNodesTag,
			disc.WithOnPeersUpdate(manager.UpdateFullNodePool),
		)
	}
}

func newModule(getter share.Getter, avail share.Availability) Module {
	return &module{getter, avail}
}

// ensureEmptyEDSInBS checks if the given DAG contains an empty block data square.
// If it does not, it stores an empty block. This optimization exists to prevent
// redundant storing of empty block data so that it is only stored once and returned
// upon request for a block with an empty data square.
func ensureEmptyEDSInBS(ctx context.Context, bServ blockservice.BlockService) error {
	_, err := ipld.AddShares(ctx, share.EmptyBlockShares(), bServ)
	return err
}

func lightGetter(
	shrexGetter *getters.ShrexGetter,
	ipldGetter *getters.IPLDGetter,
	cfg Config,
) share.Getter {
	var cascade []share.Getter
	if cfg.UseShareExchange {
		cascade = append(cascade, shrexGetter)
	}
	cascade = append(cascade, ipldGetter)
	return getters.NewCascadeGetter(cascade)
}
