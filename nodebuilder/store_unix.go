//go:build (aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos) && go1.9

package nodebuilder

import (
	"fmt"
	dsbadger "github.com/celestiaorg/go-ds-badger4"
	"github.com/ipfs/go-datastore"
	"time"
)

func newDataStore(path string) (datastore.Batching, error) {
	opts := dsbadger.DefaultOptions // this should be copied
	opts.GcInterval = time.Minute * 10

	ds, err := dsbadger.NewDatastore(dataPath(path), &opts)
	if err != nil {
		return nil, fmt.Errorf("node: can't open Badger Datastore: %w", err)
	}
	return ds, nil
}
