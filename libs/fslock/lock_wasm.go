//go:build wasm

package fslock

// TODO: Implement for wasm if necessary
func (l *Locker) lock() (err error) {
	return nil
}

// TODO: Implement for wasm if necessary
func (l *Locker) unlock() error {
	return nil
}
