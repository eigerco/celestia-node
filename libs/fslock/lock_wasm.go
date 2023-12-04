//go:build wasm

package fslock

import (
	"errors"
	"log"
)

func (l *Locker) lock() (err error) {
	log.Println("TODO: lock() implement me")
	return errors.New("TODO: lock() implement me")
}

func (l *Locker) unlock() error {
	log.Println("TODO: unlock() implement me")
	return errors.New("TODO: unlock() implement me")
}
