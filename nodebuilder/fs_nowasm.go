//go:build !wasm

package nodebuilder

import "github.com/spf13/afero"

var fs = afero.NewOsFs()
