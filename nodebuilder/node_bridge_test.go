package nodebuilder_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	coretesting "github.com/celestiaorg/celestia-node/core/testing"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	coremodule "github.com/celestiaorg/celestia-node/nodebuilder/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	nodetesting "github.com/celestiaorg/celestia-node/nodebuilder/testing"
)

func TestBridge_WithMockedCoreClient(t *testing.T) {
	t.Skip("skipping") // consult https://github.com/celestiaorg/celestia-core/issues/667 for reasoning
	repo := nodetesting.MockStore(t, nodebuilder.DefaultConfig(node.Bridge))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	client := coretesting.StartTestNode(t).Client
	node, err := nodebuilder.New(node.Bridge, p2p.Private, repo, coremodule.WithClient(client))
	require.NoError(t, err)
	require.NotNil(t, node)
	err = node.Start(ctx)
	require.NoError(t, err)

	err = node.Stop(ctx)
	require.NoError(t, err)
}
