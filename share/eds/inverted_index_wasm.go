//go:build wasm

package eds

import "github.com/celestiaorg/celestia-node/indexeddb"

func newSimpleInvertedIndex(storePath string) (*simpleInvertedIndex, error) {
	return &simpleInvertedIndex{ds: &indexeddb.DataStore{}}, nil
}
