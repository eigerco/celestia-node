//go:build wasm

package keystore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/codec"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/paralin/go-indexeddb"
)

const (
	WasmKsName    = "celestia-wasm-ks"
	WasmKsId      = "wasm-ks"
	WasmKsVersion = 1
)

func OpenIndexedDB(cdc codec.Codec, password string, opts ...cosmoskeyring.Option) (cosmoskeyring.Keyring, error) {
	db, err := indexeddb.GlobalIndexedDB().Open(context.Background(), WasmKsName, WasmKsVersion, func(d *indexeddb.DatabaseUpdate, oldVersion, newVersion int) error {
		if !d.ContainsObjectStore(WasmKsId) {
			if err := d.CreateObjectStore(WasmKsId, nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	durTx, err := indexeddb.NewDurableTransaction(db, []string{WasmKsId}, indexeddb.READWRITE)
	if err != nil {
		return nil, fmt.Errorf("error getting durable transaction %w", err)
	}
	kvtx, err := indexeddb.NewKvtxTx(durTx, WasmKsId)
	if err != nil {
		return nil, err
	}
	// TODO should we add indexeddb as another backend or keep it as in memory
	return cosmoskeyring.NewInMemoryWithKeyring(&indexeddbKeyring{
		kvtx:     kvtx,
		password: password,
	}, cdc, opts...), nil
}

type indexeddbKeyring struct {
	kvtx     *indexeddb.Kvtx
	password string
}

func (k *indexeddbKeyring) Get(key string) (keyring.Item, error) {
	bytes, found, err := k.kvtx.Get([]byte(key))
	if err != nil {
		return keyring.Item{}, err
	}
	if !found {
		return keyring.Item{}, keyring.ErrKeyNotFound
	} else if err != nil {
		return keyring.Item{}, err
	}

	payload, _, err := jose.Decode(string(bytes), k.password)
	if err != nil {
		return keyring.Item{}, err
	}

	var decoded keyring.Item
	err = json.Unmarshal([]byte(payload), &decoded)

	return decoded, err
}

func (k *indexeddbKeyring) GetMetadata(key string) (keyring.Metadata, error) {
	//stat, err := k.fs.Stat(filename)
	//if os.IsNotExist(err) {
	//	return keyring.Metadata{}, keyring.ErrKeyNotFound
	//} else if err != nil {
	//	return keyring.Metadata{}, err
	//}
	//
	//return keyring.Metadata{
	//	ModificationTime: stat.ModTime(),
	//}, nil
	return keyring.Metadata{
		ModificationTime: time.Time{}, //TODO get the mod time metadata???
	}, nil
}

func (k *indexeddbKeyring) Set(i keyring.Item) error {
	bytes, err := json.Marshal(i)
	if err != nil {
		return err
	}

	token, err := jose.Encrypt(string(bytes), jose.PBES2_HS256_A128KW, jose.A256GCM, k.password,
		jose.Headers(map[string]interface{}{
			"created": time.Now().String(),
		}))
	if err != nil {
		return err
	}

	return k.kvtx.Set([]byte(i.Key), []byte(token))
}

func (k *indexeddbKeyring) Remove(key string) error {
	return k.kvtx.Delete([]byte(key))
}

func (k *indexeddbKeyring) Keys() ([]string, error) {
	var keys = []string{}
	k.kvtx.ScanPrefixKeys([]byte(""), func(key []byte) error {
		keys = append(keys, string(key))
		return nil
	})
	return keys, nil
}
