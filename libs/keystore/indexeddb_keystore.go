//go:build wasm && js

package keystore

import (
	"fmt"
	"sync"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/paralin/go-indexeddb"
)

// indexeddbKeystore is an indexeddb Keystore implementation for wasm usage.
type indexeddbKeystore struct {
	kvtx   *indexeddb.Kvtx
	keysLk sync.Mutex
	ring   keyring.Keyring
}

// NewIDBKeystore constructs indexeddb Keystore.
func NewIDBKeystore(db *indexeddb.Database, id string, ring keyring.Keyring) (Keystore, error) {
	durTx, err := indexeddb.NewDurableTransaction(db, []string{id}, indexeddb.READWRITE)
	if err != nil {
		return nil, fmt.Errorf("error getting durable transaction %w", err)
	}
	kvtx, err := indexeddb.NewKvtxTx(durTx, id)
	if err != nil {
		return nil, err
	}
	return &indexeddbKeystore{
		kvtx: kvtx,
		ring: ring,
	}, nil
}

func (m *indexeddbKeystore) Put(n KeyName, k PrivKey) error {
	m.keysLk.Lock()
	defer m.keysLk.Unlock()

	ok, err := m.kvtx.Exists([]byte(n))
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("keystore: key '%s' already exists", n)
	}

	return m.kvtx.Set([]byte(n), k.Body)
}

func (m *indexeddbKeystore) Get(n KeyName) (PrivKey, error) {
	m.keysLk.Lock()
	defer m.keysLk.Unlock()

	b, found, err := m.kvtx.Get([]byte(n))
	if err != nil {
	}
	if !found {
		return PrivKey{}, fmt.Errorf("%w: %s", ErrNotFound, n)
	}

	return PrivKey{Body: b}, nil
}

func (m *indexeddbKeystore) Delete(n KeyName) error {
	m.keysLk.Lock()
	defer m.keysLk.Unlock()

	ok, err := m.kvtx.Exists([]byte(n))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("keystore: key '%s' not found", n)
	}

	return m.kvtx.Delete([]byte(n))
}

func (m *indexeddbKeystore) List() (keys []KeyName, _ error) {
	m.keysLk.Lock()
	defer m.keysLk.Unlock()

	err := m.kvtx.ScanPrefixKeys([]byte(""), func(key []byte) error {
		keys = append(keys, KeyName(key))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (m *indexeddbKeystore) Path() string {
	return ""
}

func (m *indexeddbKeystore) Keyring() keyring.Keyring {
	return m.ring
}
