//go:build wasm

package nodebuilder

import (
	"github.com/celestiaorg/celestia-node/indexeddb"
	"github.com/ipfs/go-datastore"
)

func newDataStore(path string) (datastore.Batching, error) {
	return &indexeddb.DataStore{}, nil
}
