//go:build wasm

package nodebuilder

import (
	"github.com/celestiaorg/celestia-node/libs/wasmfs"
	"github.com/spf13/afero"
)

var fs afero.Fs

func init() {
	var err error
	fs, err = wasmfs.New()
	if err != nil {
		return
	}
	return
}
