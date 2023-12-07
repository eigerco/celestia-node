package wasmfs

import (
	"github.com/spf13/afero"
	"os"
	"time"
)

func New() afero.Fs {
	return &wasmfs{}
}

type wasmfs struct{}

func (w *wasmfs) Create(name string) (afero.File, error) {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) Mkdir(name string, perm os.FileMode) error {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) MkdirAll(path string, perm os.FileMode) error {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) Open(name string) (afero.File, error) {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) Remove(name string) error {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) RemoveAll(path string) error {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) Rename(oldname, newname string) error {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) Stat(name string) (os.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) Name() string {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) Chmod(name string, mode os.FileMode) error {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) Chown(name string, uid, gid int) error {
	//TODO implement me
	panic("implement me")
}

func (w *wasmfs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	//TODO implement me
	panic("implement me")
}
