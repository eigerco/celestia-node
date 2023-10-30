//go:build (aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos) && go1.9

package eds

import (
	"fmt"

	dsbadger "github.com/celestiaorg/go-ds-badger4"
)

// newSimpleInvertedIndex returns a new inverted index that only stores a single shard key per
// multihash. This is because we use badger as a storage backend, so updates are expensive, and we
// don't care which shard is used to serve a cid.
func newSimpleInvertedIndex(storePath string) (*simpleInvertedIndex, error) {
	opts := dsbadger.DefaultOptions // this should be copied
	// turn off value log GC
	opts.GcInterval = 0
	// 20 compactors show to have no hangups on put operation up to 40k blocks with eds size 128.
	opts.NumCompactors = 20
	// use minimum amount of NumLevelZeroTables to trigger L0 compaction faster
	opts.NumLevelZeroTables = 1
	// MaxLevels = 8 will allow the db to grow to ~11.1 TiB
	opts.MaxLevels = 8

	ds, err := dsbadger.NewDatastore(storePath+invertedIndexPath, &opts)
	if err != nil {
		return nil, fmt.Errorf("can't open Badger Datastore: %w", err)
	}

	return &simpleInvertedIndex{ds: ds}, nil
}
