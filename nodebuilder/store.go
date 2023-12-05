package nodebuilder

import (
	"errors"
	"sync"

	"github.com/ipfs/go-datastore"

	"github.com/celestiaorg/celestia-node/libs/fslock"
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

type fsStore struct {
	path string

	dataMu  sync.Mutex
	data    datastore.Batching
	keys    keystore.Keystore
	dirLock *fslock.Locker // protects directory
}
