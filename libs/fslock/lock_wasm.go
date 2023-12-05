//go:build wasm

package fslock

import (
	"log"
)

func (l *Locker) lock() (err error) {
	log.Println("TODO: lock() implement me")
	//return errors.New("TODO: lock() implement me")
	return nil
}

func (l *Locker) unlock() error {
	log.Println("TODO: unlock() implement me")
	//return errors.New("TODO: unlock() implement me")
	return nil
}
