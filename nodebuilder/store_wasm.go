//go:build wasm

package nodebuilder

import (
	"context"

	"github.com/celestiaorg/celestia-node/indexeddb"
	"github.com/ipfs/go-datastore"
)

func newDataStore(path string) (datastore.Batching, error) {
	return indexeddb.NewDataStore(context.Background(), path)
}
