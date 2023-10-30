package eds

import (
	"context"
	"errors"
	"fmt"

	"github.com/filecoin-project/dagstore/index"
	"github.com/filecoin-project/dagstore/shard"
	ds "github.com/ipfs/go-datastore"
	"github.com/multiformats/go-multihash"
)

const invertedIndexPath = "/inverted_index/"

// ErrNotFoundInIndex is returned instead of ErrNotFound if the multihash doesn't exist in the index
var ErrNotFoundInIndex = fmt.Errorf("does not exist in index")

// simpleInvertedIndex is an inverted index that only stores a single shard key per multihash. Its
// implementation is modified from the default upstream implementation in dagstore/index.
type simpleInvertedIndex struct {
	ds ds.Batching
}

func (s *simpleInvertedIndex) AddMultihashesForShard(
	ctx context.Context,
	mhIter index.MultihashIterator,
	sk shard.Key,
) error {
	// in the original implementation, a mutex is used here to prevent unnecessary updates to the
	// key. The amount of extra data produced by this is negligible, and the performance benefits
	// from removing the lock are significant (indexing is a hot path during sync).
	batch, err := s.ds.Batch(ctx)
	if err != nil {
		return fmt.Errorf("failed to create ds batch: %w", err)
	}

	err = mhIter.ForEach(func(mh multihash.Multihash) error {
		key := ds.NewKey(string(mh))
		if err := batch.Put(ctx, key, []byte(sk.String())); err != nil {
			return fmt.Errorf("failed to put mh=%s, err=%w", mh, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add index entry: %w", err)
	}

	if err := batch.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit batch: %w", err)
	}
	return nil
}

func (s *simpleInvertedIndex) GetShardsForMultihash(ctx context.Context, mh multihash.Multihash) ([]shard.Key, error) {
	key := ds.NewKey(string(mh))
	sbz, err := s.ds.Get(ctx, key)
	if err != nil {
		return nil, errors.Join(ErrNotFoundInIndex, err)
	}

	return []shard.Key{shard.KeyFromString(string(sbz))}, nil
}

func (s *simpleInvertedIndex) close() error {
	return s.ds.Close()
}
