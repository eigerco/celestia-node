//go:build wasm && js

package main

import (
	"os"

	logging "github.com/ipfs/go-log/v2"

	"github.com/celestiaorg/celestia-node/libs/jsnode"
)

func main() {
	logging.SetupLogging(logging.Config{
		Stderr: true,
	})

	if err := os.Setenv("CELESTIA_ENABLE_QUIC", "true"); err != nil {
		panic(err)
		return
	}

	if err := os.Setenv("CELESTIA_HOME", "test"); err != nil {
		panic(err)
		return
	}
	if err := os.Setenv("HOME", "test"); err != nil {
		panic(err)
		return
	}

	jsnode.RegisterJSFunctions()

	select {}
}
