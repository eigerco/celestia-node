//go:build wasm

package nodebuilder

import (
	"context"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/paralin/go-indexeddb"

	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-node/libs/codec"
	"github.com/celestiaorg/celestia-node/libs/dsindexeddb"
	"github.com/celestiaorg/celestia-node/libs/keystore"
	"github.com/celestiaorg/celestia-node/libs/krindexeddb"
)

const (
	databaseName    = "celestia"
	version         = 1
	keyringPassword = "testpassword" // TODO secure the keyring password
	keystoreID      = "keystore"
	keyringID       = "keyring"
	datastoreID     = "datastore"
)

type indexeddbStore struct {
	keys keystore.Keystore
	data datastore.Batching
	cfg  *Config
	cfgL sync.Mutex
	db   *indexeddb.Database
}

// NewIndexedDBStore opens a indexeddb database and initializes the datastore and keyring while keeping the config in-mem
func NewIndexedDBStore(ctx context.Context, cfg *Config) (Store, error) {
	db, err := indexeddb.GlobalIndexedDB().Open(ctx, databaseName, version, func(d *indexeddb.DatabaseUpdate, oldVersion, newVersion int) error {
		for _, id := range []string{
			keystoreID,
			keyringID,
			datastoreID,
		} {
			if !d.ContainsObjectStore(id) {
				if err := d.CreateObjectStore(id, nil); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	encConf := encoding.MakeConfig(codec.ModuleEncodingRegisters...)
	ring, err := krindexeddb.NewKeyring(db, keyringID, encConf.Codec, keyringPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	keys, err := keystore.NewIDBKeystore(db, keystoreID, ring)
	if err != nil {
		return nil, fmt.Errorf("failed to init keystore: %w", err)
	}
	data, err := dsindexeddb.NewDataStore(db, datastoreID)
	if err != nil {
		return nil, fmt.Errorf("failed to init datastore: %w", err)
	}
	return &indexeddbStore{
		cfg:  cfg,
		keys: keys,
		data: data,
		db:   db,
	}, nil
}

func (m *indexeddbStore) Keystore() (keystore.Keystore, error) {
	return m.keys, nil
}

func (m *indexeddbStore) Datastore() (datastore.Batching, error) {
	return m.data, nil
}

func (m *indexeddbStore) Config() (*Config, error) {
	m.cfgL.Lock()
	defer m.cfgL.Unlock()
	return m.cfg, nil
}

func (m *indexeddbStore) PutConfig(cfg *Config) error {
	m.cfgL.Lock()
	defer m.cfgL.Unlock()
	m.cfg = cfg
	return nil
}

func (m *indexeddbStore) Path() string {
	return ""
}

func (m *indexeddbStore) Close() error {
	m.db.Close()
	return nil
}
