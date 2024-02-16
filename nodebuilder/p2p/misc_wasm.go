//go:build wasm

package p2p

import (
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"time"
)

// defaultConnManagerConfigWasm returns defaults for ConnManagerConfig used in WASM mode.
func defaultConnManagerConfigWasm(tp node.Type) connManagerConfig {
	switch tp {
	case node.Light:
		return connManagerConfig{
			Low:         1,
			High:        10,
			GracePeriod: time.Minute,
		}
	default:
		panic("unknown wasm (browser) supported node type")
	}
}
