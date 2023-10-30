package indexeddb

import (
	"context"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

type DataStore struct {
}

func (d DataStore) Get(ctx context.Context, key datastore.Key) (value []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (d DataStore) Has(ctx context.Context, key datastore.Key) (exists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (d DataStore) GetSize(ctx context.Context, key datastore.Key) (size int, err error) {
	//TODO implement me
	panic("implement me")
}

func (d DataStore) Query(ctx context.Context, q query.Query) (query.Results, error) {
	//TODO implement me
	panic("implement me")
}

func (d DataStore) Put(ctx context.Context, key datastore.Key, value []byte) error {
	//TODO implement me
	panic("implement me")
}

func (d DataStore) Delete(ctx context.Context, key datastore.Key) error {
	//TODO implement me
	panic("implement me")
}

func (d DataStore) Sync(ctx context.Context, prefix datastore.Key) error {
	//TODO implement me
	panic("implement me")
}

func (d DataStore) Close() error {
	//TODO implement me
	panic("implement me")
}

func (d DataStore) Batch(ctx context.Context) (datastore.Batch, error) {
	//TODO implement me
	panic("implement me")
}
