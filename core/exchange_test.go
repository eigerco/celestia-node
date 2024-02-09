//go:build !wasm || !js

package core_test

import (
	"context"
	"testing"
	"time"

	"github.com/celestiaorg/celestia-app/test/util/testnode"
	"github.com/celestiaorg/celestia-node/core"
	testing2 "github.com/celestiaorg/celestia-node/core/testing"
	ds "github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share/eds"
)

func TestCoreExchange_RequestHeaders(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	fetcher, _ := createCoreFetcher(t, testing2.DefaultTestConfig())

	// generate 10 blocks
	generateBlocks(t, fetcher)

	store := createStore(t)

	ce := core.NewExchange(fetcher, store, header.MakeExtendedHeader)

	// initialize store with genesis block
	genHeight := int64(1)
	genBlock, err := fetcher.GetBlock(ctx, &genHeight)
	require.NoError(t, err)
	genHeader, err := ce.Get(ctx, genBlock.Header.Hash().Bytes())
	require.NoError(t, err)

	to := uint64(10)
	expectedFirstHeightInRange := genHeader.Height() + 1
	expectedLastHeightInRange := to - 1
	expectedLenHeaders := to - expectedFirstHeightInRange

	// request headers from height 1 to 10 [2:10)
	headers, err := ce.GetRangeByHeight(context.Background(), genHeader, to)
	require.NoError(t, err)

	assert.Len(t, headers, int(expectedLenHeaders))
	assert.Equal(t, expectedFirstHeightInRange, headers[0].Height())
	assert.Equal(t, expectedLastHeightInRange, headers[len(headers)-1].Height())
}

func createCoreFetcher(t *testing.T, cfg *testnode.Config) (*core.BlockFetcher, testnode.Context) {
	cctx := testing2.StartTestNodeWithConfig(t, cfg)
	// wait for height 2 in order to be able to start submitting txs (this prevents
	// flakiness with accessing account state)
	_, err := cctx.WaitForHeightWithTimeout(2, time.Second*2) // TODO @renaynay: configure?
	require.NoError(t, err)
	return core.NewBlockFetcher(cctx.Client), cctx
}

func createStore(t *testing.T) *eds.Store {
	t.Helper()

	storeCfg := eds.DefaultParameters()
	store, err := eds.NewStore(storeCfg, t.TempDir(), ds_sync.MutexWrap(ds.NewMapDatastore()))
	require.NoError(t, err)
	return store
}

func generateBlocks(t *testing.T, fetcher *core.BlockFetcher) {
	sub, err := fetcher.SubscribeNewBlockEvent(context.Background())
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		<-sub
	}
}
