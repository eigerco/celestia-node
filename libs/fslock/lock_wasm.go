//go:build wasm

package fslock

func (l *Locker) lock() (err error) {
	panic("TODO: lock() implement me")
}

func (l *Locker) unlock() error {
	panic("TODO: unlock() implement me")
}
