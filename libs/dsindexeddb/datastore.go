//go:build wasm && js

package dsindexeddb

import (
	"context"
	"errors"
	"fmt"

	"github.com/ipfs/go-datastore"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	dsq "github.com/ipfs/go-datastore/query"
	process "github.com/jbenet/goprocess"
	"github.com/paralin/go-indexeddb"
)

func NewDataStore(db *indexeddb.Database, id string) (*DataStore, error) {
	dss := &DataStore{
		db: db,
		id: id,
	}
	return dss, nil
}

type DataStore struct {
	db *indexeddb.Database
	id string
}

func (d *DataStore) kvtx() (*indexeddb.Kvtx, error) {
	durTx, err := indexeddb.NewDurableTransaction(d.db, []string{d.id}, indexeddb.READWRITE)
	if err != nil {
		return nil, fmt.Errorf("error getting durable transaction %w", err)
	}
	kvtx, err := indexeddb.NewKvtxTx(durTx, d.id)
	if err != nil {
		return nil, err
	}
	return kvtx, nil
}

func (d *DataStore) Get(ctx context.Context, key datastore.Key) (value []byte, err error) {
	kvtx, err := d.kvtx()
	if err != nil {
		return nil, err
	}
	data, found, err := kvtx.Get(key.Bytes())
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, datastore.ErrNotFound
	}

	return data, nil
}

func (d *DataStore) Has(ctx context.Context, key datastore.Key) (exists bool, err error) {
	kvtx, err := d.kvtx()
	if err != nil {
		return false, err
	}
	return kvtx.Exists(key.Bytes())
}

func (d *DataStore) GetSize(ctx context.Context, key datastore.Key) (size int, err error) {
	data, err := d.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (d *DataStore) Query(ctx context.Context, q query.Query) (_ query.Results, err error) {
	var prefixBytes []byte
	if prefix := ds.NewKey(q.Prefix).String(); prefix != "/" {
		prefixBytes = []byte(prefix + "/")
	}

	// Handle ordering
	if len(q.Orders) > 0 {
		switch q.Orders[0].(type) {
		case dsq.OrderByKey, *dsq.OrderByKey:
		// We order by key by default.
		case dsq.OrderByKeyDescending, *dsq.OrderByKeyDescending:
			// Reverse order by key
			//opt.Reverse = true
		default:
			return nil, errors.New("unsupported order type")
		}
	}

	kvtx, err := d.kvtx()
	if err != nil {
		return nil, err
	}
	qrb := dsq.NewResultBuilder(q)
	qrb.Process.Go(func(worker process.Process) {
		//var closedEarly bool
		if q.KeysOnly {
			err = kvtx.ScanPrefixKeys(prefixBytes, func(key []byte) error {
				select {
				case qrb.Output <- dsq.Result{
					Entry: dsq.Entry{
						Key: string(key),
					}}:
					//sent++
				//case <-t.ds.closing: // datastore closing.
				//	closedEarly = true
				//	return
				case <-worker.Closing(): // client told us to close early
					return nil
				}
				return nil
			})
		} else {
			err = kvtx.ScanPrefix(prefixBytes, func(key, val []byte) error {
				select {
				case qrb.Output <- dsq.Result{
					Entry: dsq.Entry{
						Key:   string(key),
						Value: val,
					}}:
				case <-worker.Closing(): // client told us to close early
					return nil
				}
				return nil
			})
		}
		if err != nil {
			select {
			case qrb.Output <- dsq.Result{
				Error: err,
			}:
			case <-qrb.Process.Closing():
			}
		}
	})

	go qrb.Process.CloseAfterChildren()

	return qrb.Results(), nil
}

func (d *DataStore) Put(ctx context.Context, key datastore.Key, value []byte) error {
	kvtx, err := d.kvtx()
	if err != nil {
		return err
	}
	return kvtx.Set(key.Bytes(), value)
}

func (d *DataStore) Delete(ctx context.Context, key datastore.Key) error {
	kvtx, err := d.kvtx()
	if err != nil {
		return err
	}
	return kvtx.Delete(key.Bytes())
}

func (d *DataStore) Sync(ctx context.Context, prefix datastore.Key) error {
	//TODO implement me
	panic("implement me datastore")
}

func (d *DataStore) Close() error {
	d.db.Close()
	return nil
}

func (d *DataStore) Batch(ctx context.Context) (datastore.Batch, error) {
	return ds.NewBasicBatch(d), nil
}
