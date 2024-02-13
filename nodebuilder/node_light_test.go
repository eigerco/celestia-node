package nodebuilder_test

import (
	"context"
	"crypto/rand"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	testing2 "github.com/celestiaorg/celestia-node/nodebuilder/testing"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/nodebuilder/state"
)

func TestNewLightWithP2PKey(t *testing.T) {
	key, _, err := crypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)
	node := testing2.TestNode(t, node.Light, p2p.WithP2PKey(key))
	assert.True(t, node.Host.ID().MatchesPrivateKey(key))
}

func TestNewLightWithHost(t *testing.T) {
	nw, _ := mocknet.WithNPeers(1)
	node := testing2.TestNode(t, node.Light, p2p.WithHost(nw.Hosts()[0]))
	assert.Equal(t, nw.Peers()[0], node.Host.ID())
}

func TestLight_WithMutualPeers(t *testing.T) {
	peers := []string{
		"/ip6/100:0:114b:abc5:e13a:c32f:7a9e:f00a/tcp/2121/p2p/12D3KooWSRqDfpLsQxpyUhLC9oXHD2WuZ2y5FWzDri7LT4Dw9fSi",
		"/ip4/192.168.1.10/tcp/2121/p2p/12D3KooWSRqDfpLsQxpyUhLC9oXHD2WuZ2y5FWzDri7LT4Dw9fSi",
	}
	cfg := nodebuilder.DefaultConfig(node.Light)
	cfg.P2P.MutualPeers = peers
	node := testing2.TestNodeWithConfig(t, node.Light, cfg)

	require.NotNil(t, node)
	assert.Equal(t, node.Config.P2P.MutualPeers, peers)
}

func TestLight_WithNetwork(t *testing.T) {
	node := testing2.TestNode(t, node.Light)
	require.NotNil(t, node)
	assert.Equal(t, p2p.Private, node.Network)
}

// TestLight_WithStubbedCoreAccessor ensures that a node started without
// a core connection will return a stubbed StateModule.
func TestLight_WithStubbedCoreAccessor(t *testing.T) {
	node := testing2.TestNode(t, node.Light)
	_, err := node.StateServ.Balance(context.Background())
	assert.ErrorIs(t, state.ErrNoStateAccess, err)
}
