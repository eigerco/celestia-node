package indexeddb

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

func NewDataStore(ctx context.Context, id string) (*DataStore, error) {
	db, err := indexeddb.GlobalIndexedDB().Open(ctx, "test-db", 3, func(d *indexeddb.DatabaseUpdate, oldVersion, newVersion int) error {
		if !d.ContainsObjectStore(id) {
			if err := d.CreateObjectStore(id, nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &DataStore{
		id: id,
		db: db,
	}, nil
}

type DataStore struct {
	id string
	db *indexeddb.Database
}

func (d DataStore) tx() (value *indexeddb.Kvtx, err error) {
	durTx, err := indexeddb.NewDurableTransaction(d.db, []string{d.id}, indexeddb.READWRITE)
	if err != nil {
		return nil, fmt.Errorf("error getting durable transaction %w", err)
	}
	return indexeddb.NewKvtxTx(durTx, d.id)
}

func (d DataStore) Get(ctx context.Context, key datastore.Key) (value []byte, err error) {
	objStore, err := d.tx()
	if err != nil {
		return nil, err
	}

	data, _, err := objStore.Get(key.Bytes())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d DataStore) Has(ctx context.Context, key datastore.Key) (exists bool, err error) {
	objStore, err := d.tx()
	if err != nil {
		return false, err
	}

	return objStore.Exists(key.Bytes())
}

func (d DataStore) GetSize(ctx context.Context, key datastore.Key) (size int, err error) {
	data, err := d.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (d DataStore) Query(ctx context.Context, q query.Query) (query.Results, error) {
	objStore, err := d.tx()
	if err != nil {
		return nil, err
	}

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

	qrb := dsq.NewResultBuilder(q)
	qrb.Process.Go(func(worker process.Process) {
		//var closedEarly bool
		if q.KeysOnly {
			err = objStore.ScanPrefixKeys(prefixBytes, func(key []byte) error {
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
			err = objStore.ScanPrefix(prefixBytes, func(key, val []byte) error {
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
	return qrb.Results(), nil
}

func (d DataStore) Put(ctx context.Context, key datastore.Key, value []byte) error {
	objStore, err := d.tx()
	if err != nil {
		return err
	}

	return objStore.Set(key.Bytes(), value)
}

func (d DataStore) Delete(ctx context.Context, key datastore.Key) error {
	objStore, err := d.tx()
	if err != nil {
		return err
	}

	return objStore.Delete(key.Bytes())
}

func (d DataStore) Sync(ctx context.Context, prefix datastore.Key) error {
	//TODO implement me
	panic("implement me")
}

func (d DataStore) Close() error {
	d.db.Close()
	return nil
}

func (d DataStore) Batch(ctx context.Context) (datastore.Batch, error) {
	//TODO implement me
	panic("implement me")
}
