package wasmfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"syscall/js"
	"time"

	"github.com/paralin/go-indexeddb"
	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"
)

func New() (afero.Fs, error) {
	ctx := context.Background()
	id := "wasmfs"
	name := "wasmfs"
	version := 3
	db, err := indexeddb.GlobalIndexedDB().Open(ctx, name, version, func(d *indexeddb.DatabaseUpdate, oldVersion, newVersion int) error {
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
	return &wasmfs{
		id: id,
		db: db,
	}, nil
}

type wasmfs struct {
	id string
	db *indexeddb.Database
}

func (w *wasmfs) Create(name string) (afero.File, error) {
	return w.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
}

func (w *wasmfs) Mkdir(name string, perm os.FileMode) error {
	return nil // there is no need to create directories, since keys can contain the path
}

func (w *wasmfs) MkdirAll(path string, perm os.FileMode) error {
	return nil // there is no need to create directories, since keys can contain the path
}

func (w *wasmfs) Open(name string) (afero.File, error) {
	return w.OpenFile(name, os.O_RDONLY, os.ModePerm)
}

func (w *wasmfs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	readOnly := flag&os.O_RDONLY > 0
	imode := indexeddb.READWRITE
	if readOnly {
		imode = indexeddb.READONLY
	}
	durTx, err := indexeddb.NewDurableTransaction(w.db, []string{w.id}, imode)
	if err != nil {
		return nil, fmt.Errorf("error getting durable transaction %w", err)
	}
	objStore, err := durTx.GetObjectStore(w.id)
	if err != nil {
		return nil, err
	}
	c, err := objStore.GetAllKeys(name)
	if err != nil {
		return nil, err
	}
	fileExists := c.Length() > 0
	if fileExists && (flag&os.O_EXCL > 0) {
		return nil, os.ErrExist
	}
	if !fileExists && (flag&os.O_CREATE == 0) {
		return nil, os.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	file := &File{
		objStore: objStore,
		mu:       &sync.Mutex{},
		name:     name,
		readOnly: readOnly,
	}
	if flag&os.O_APPEND > 0 {
		_, err = file.Seek(0, io.SeekEnd)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	if flag&os.O_TRUNC > 0 && flag&(os.O_RDWR|os.O_WRONLY) > 0 {
		err = file.Truncate(0)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	return file, nil
}

func (w *wasmfs) Remove(name string) error {
	durTx, err := indexeddb.NewDurableTransaction(w.db, []string{w.id}, indexeddb.READWRITE)
	if err != nil {
		return fmt.Errorf("error getting durable transaction %w", err)
	}
	kvtx, err := indexeddb.NewKvtxTx(durTx, w.id)
	if err != nil {
		return err
	}
	if err := kvtx.Delete([]byte(name)); err != nil {
		return err
	}
	return nil
}

func (w *wasmfs) RemoveAll(path string) error {
	return w.Remove(path)
}

func (w *wasmfs) Rename(oldname, newname string) error {
	return syscall.EPERM
}

func (w *wasmfs) Stat(name string) (os.FileInfo, error) {
	return &mem.FileInfo{FileData: mem.CreateFile(name)}, nil
}

func (w *wasmfs) Name() string {
	return "wasmfs"
}

func (w *wasmfs) Chmod(name string, mode os.FileMode) error {
	return syscall.EPERM
}

func (w *wasmfs) Chown(name string, uid, gid int) error {
	return syscall.EPERM
}

func (w *wasmfs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return syscall.EPERM
}

type File struct {
	objStore *indexeddb.DurableObjectStore
	mu       *sync.Mutex
	name     string
	at       int64
	closed   bool
	readOnly bool
}

func (f *File) Close() error {
	f.closed = true
	return nil
}

func (f *File) getBytes() ([]byte, error) {
	jsObj, err := f.objStore.Get(f.name)
	if err != nil {
		return nil, err
	}
	if !jsObj.Truthy() {
		return nil, nil
	}
	dlen := jsObj.Length()
	data := make([]byte, dlen)
	js.CopyBytesToGo(data, jsObj)
	return data, err
}

func (f *File) Read(b []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, afero.ErrFileClosed
	}
	data, err := f.getBytes()
	if err != nil {
		return 0, err
	}
	if len(b) > 0 && int(f.at) == len(data) {
		return 0, io.EOF
	}
	if int(f.at) > len(data) {
		return 0, io.ErrUnexpectedEOF
	}
	if len(data)-int(f.at) >= len(b) {
		n = len(b)
	} else {
		n = len(data) - int(f.at)
	}
	copy(b, data[f.at:f.at+int64(n)])
	atomic.AddInt64(&f.at, int64(n))
	return n, nil
}

func (f *File) Write(b []byte) (n int, err error) {
	cur := atomic.LoadInt64(&f.at)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, afero.ErrFileClosed
	}
	if f.readOnly {
		return 0, &os.PathError{Op: "write", Path: f.name, Err: errors.New("file handle is read only")}
	}
	data, err := f.getBytes()
	if err != nil {
		return 0, err
	}
	n = len(b)
	diff := cur - int64(len(data))
	var tail []byte
	if n+int(cur) < len(data) {
		tail = data[n+int(cur):]
	}
	if diff > 0 {
		data = append(data, append(bytes.Repeat([]byte{0o0}, int(diff)), b...)...)
		data = append(data, tail...)
	} else {
		data = append(data[:cur], b...)
		data = append(data, tail...)
	}
	if err := f.objStore.Put(data, f.name); err != nil {
		return 0, err
	}
	return n, nil
}

func (f *File) WriteString(s string) (n int, err error) {
	return f.Write([]byte(s))
}

func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	atomic.StoreInt64(&f.at, off)
	return f.Read(p)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		return 0, afero.ErrFileClosed
	}
	switch whence {
	case io.SeekStart:
		atomic.StoreInt64(&f.at, offset)
	case io.SeekCurrent:
		atomic.AddInt64(&f.at, offset)
	case io.SeekEnd:
		data, err := f.getBytes()
		if err != nil {
			return 0, err
		}
		atomic.StoreInt64(&f.at, int64(len(data))+offset)
	}
	return f.at, nil
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	atomic.StoreInt64(&f.at, off)
	return f.Write(b)
}

func (f *File) Name() string {
	return string(f.name)
}

func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	return nil, syscall.EPERM
}

func (f *File) Readdirnames(n int) ([]string, error) {
	return nil, syscall.EPERM
}

func (f *File) Stat() (os.FileInfo, error) {
	data, err := f.getBytes()
	if err != nil {
		return nil, err
	}
	return &fileInfo{
		name: f.Name(),
		size: int64(len(data)),
	}, nil
}

func (f *File) Sync() error {
	return nil
}

func (f *File) Truncate(size int64) error {
	if f.closed {
		return afero.ErrFileClosed
	}
	if size < 0 {
		return afero.ErrOutOfRange
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	data, err := f.getBytes()
	if err != nil {
		return err
	}
	if size > int64(len(data)) {
		diff := size - int64(len(data))
		data = append(data, bytes.Repeat([]byte{0o0}, int(diff))...)
	} else {
		data = data[0:size]
	}
	return f.objStore.Put(data, f.name)
}

type fileInfo struct {
	name string
	size int64
}

func (f *fileInfo) Name() string {
	return f.name
}

func (f *fileInfo) Size() int64 {
	return f.size
}

func (f *fileInfo) Mode() fs.FileMode {
	return os.ModeTemporary
}

func (f *fileInfo) ModTime() time.Time {
	return time.Time{} //TODO
}

func (f *fileInfo) IsDir() bool {
	return false
}

func (f *fileInfo) Sys() any {
	return nil
}
