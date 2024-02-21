package nodebuilder

import (
	"errors"
	"path/filepath"

	"github.com/ipfs/go-datastore"
	"github.com/mitchellh/go-homedir"

	"github.com/celestiaorg/celestia-node/libs/keystore"
)

var (
	// ErrOpened is thrown on attempt to open already open/in-use Store.
	ErrOpened = errors.New("node: store is in use")
	// ErrNotInited is thrown on attempt to open Store without initialization.
	ErrNotInited = errors.New("node: store is not initialized")
)

// Store encapsulates storage for the Node. Basically, it is the Store of all Stores.
// It provides access for the Node data stored in root directory e.g. '~/.celestia'.
type Store interface {
	// Path reports the FileSystem path of Store.
	Path() string

	// Keystore provides a Keystore to access keys.
	Keystore() (keystore.Keystore, error)

	// Datastore provides a Datastore - a KV store for arbitrary data to be stored on disk.
	Datastore() (datastore.Batching, error)

	// Config loads the stored Node config.
	Config() (*Config, error)

	// PutConfig alters the stored Node config.
	PutConfig(*Config) error

	// Close closes the Store freeing up acquired resources and locks.
	Close() error
}

func storePath(path string) (string, error) {
	return homedir.Expand(filepath.Clean(path))
}

func configPath(base string) string {
	return filepath.Join(base, "config.toml")
}

func lockPath(base string) string {
	return filepath.Join(base, "lock")
}

func keysPath(base string) string {
	return filepath.Join(base, "keys")
}

func blocksPath(base string) string {
	return filepath.Join(base, "blocks")
}

func transientsPath(base string) string {
	// we don't actually use the transients directory anymore, but it could be populated from previous
	// versions.
	return filepath.Join(base, "transients")
}

func indexPath(base string) string {
	return filepath.Join(base, "index")
}

func dataPath(base string) string {
	return filepath.Join(base, "data")
}
