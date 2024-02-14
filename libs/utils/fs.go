package utils

import (
	"os"

	"github.com/spf13/afero"
)

// Exists checks whether file or directory exists under the given 'path' on the system.
func Exists(fs afero.Fs, path string) bool {
	_, err := fs.Stat(path)
	return !os.IsNotExist(err)
}
