package utils

import (
	"github.com/spf13/afero"
	"os"
)

// Exists checks whether file or directory exists under the given 'path' on the system.
func Exists(fs afero.Fs, path string) bool {
	_, err := fs.Stat(path)
	return !os.IsNotExist(err)
}
