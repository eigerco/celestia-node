//go:build !wasm

package nodebuilder

import (
	"errors"
	"fmt"
	"sync"
	"time"

	dsbadger "github.com/celestiaorg/go-ds-badger4"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/ipfs/go-datastore"

	"github.com/celestiaorg/celestia-node/libs/fslock"
	"github.com/celestiaorg/celestia-node/libs/keystore"
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

// OpenStore creates new FS Store under the given 'path'.
// To be opened the Store must be initialized first, otherwise ErrNotInited is thrown.
// OpenStore takes a file Lock on directory, hence only one Store can be opened at a time under the
// given 'path', otherwise ErrOpened is thrown.
func OpenStore(path string, ring keyring.Keyring) (Store, error) {
	path, err := storePath(path)
	if err != nil {
		return nil, err
	}

	flock, err := fslock.Lock(lockPath(path))
	if err != nil {
		if err == fslock.ErrLocked {
			return nil, ErrOpened
		}
		return nil, err
	}

	ok := IsInit(path)
	if !ok {
		flock.Unlock() //nolint: errcheck
		return nil, ErrNotInited
	}

	ks, err := keystore.NewFSKeystore(keysPath(path), ring)
	if err != nil {
		return nil, err
	}

	return &fsStore{
		path:    path,
		dirLock: flock,
		keys:    ks,
	}, nil
}

func (f *fsStore) Path() string {
	return f.path
}

func (f *fsStore) Config() (*Config, error) {
	cfg, err := LoadConfig(configPath(f.path))
	if err != nil {
		return nil, fmt.Errorf("node: can't load Config: %w", err)
	}

	return cfg, nil
}

func (f *fsStore) PutConfig(cfg *Config) error {
	err := SaveConfig(configPath(f.path), cfg)
	if err != nil {
		return fmt.Errorf("node: can't save Config: %w", err)
	}

	return nil
}

func (f *fsStore) Keystore() (_ keystore.Keystore, err error) {
	if f.keys == nil {
		return nil, fmt.Errorf("node: no Keystore found")
	}
	return f.keys, nil
}

func (f *fsStore) Datastore() (datastore.Batching, error) {
	f.dataMu.Lock()
	defer f.dataMu.Unlock()
	if f.data != nil {
		return f.data, nil
	}

	opts := dsbadger.DefaultOptions // this should be copied
	opts.GcInterval = time.Minute * 10

	ds, err := dsbadger.NewDatastore(dataPath(f.path), &opts)
	if err != nil {
		return nil, fmt.Errorf("node: can't open Badger Datastore: %w", err)
	}

	f.data = ds
	return ds, nil
}

func (f *fsStore) Close() (err error) {
	err = errors.Join(err, f.dirLock.Unlock())
	f.dataMu.Lock()
	if f.data != nil {
		err = errors.Join(err, f.data.Close())
	}
	f.dataMu.Unlock()
	return
}

type fsStore struct {
	path string

	dataMu  sync.Mutex
	data    datastore.Batching
	keys    keystore.Keystore
	dirLock *fslock.Locker // protects directory
}
