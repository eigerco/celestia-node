//go:build wasm

package nodebuilder

import (
	"github.com/celestiaorg/celestia-node/indexeddb"
	"github.com/celestiaorg/celestia-node/libs/fslock"
	"github.com/celestiaorg/celestia-node/libs/keystore"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/ipfs/go-datastore"
)

func newDataStore(path string) (datastore.Batching, error) {
	return &indexeddb.DataStore{}, nil
}

// OpenStore creates new FS Store under the given 'path'.
// To be opened the Store must be initialized first, otherwise ErrNotInited is thrown.
// OpenStore takes a file Lock on directory, hence only one Store can be opened at a time under the
// given 'path', otherwise ErrOpened is thrown.
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
	return ""
}

func (f *fsStore) Config() (*Config, error) {
	return &Config{}, nil
}

func (f *fsStore) PutConfig(cfg *Config) error {
	return nil
}

func (f *fsStore) Keystore() (_ keystore.Keystore, err error) {
	return nil, nil
}

func (f *fsStore) Datastore() (datastore.Batching, error) {
	return nil, nil
}

func (f *fsStore) Close() (err error) {
	return
}

func storePath(path string) (string, error) {
	return "", nil
}

func configPath(base string) string {
	return ""
}

func lockPath(base string) string {
	return ""
}

func keysPath(base string) string {
	return ""
}

func blocksPath(base string) string {
	return ""
}

func transientsPath(base string) string {
	// we don't actually use the transients directory anymore, but it could be populated from previous
	// versions.
	return ""
}

func indexPath(base string) string {
	return ""
}

func dataPath(base string) string {
	return ""
}
