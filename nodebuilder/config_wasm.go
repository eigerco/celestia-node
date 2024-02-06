//go:build js && wasm

package nodebuilder

import (
	"github.com/celestiaorg/celestia-node/nodebuilder/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/das"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
)

// Config is main configuration structure for a Node.
// It combines configuration units for all Node subsystems.
type Config struct {
	Core   core.Config
	P2P    p2p.Config
	Share  share.Config
	Header header.Config
	DASer  das.Config `toml:",omitempty"`
}

// DefaultConfig provides a default Config for a given Node Type 'tp'.
// NOTE: Currently, configs are identical, but this will change.
func DefaultConfig(tp node.Type) *Config {
	commonConfig := &Config{
		Core:   core.DefaultConfig(),
		P2P:    p2p.DefaultConfig(tp),
		Share:  share.DefaultConfig(tp),
		Header: header.DefaultConfig(tp),
	}

	switch tp {
	case node.Bridge:
		return commonConfig
	case node.Light, node.Full:
		commonConfig.DASer = das.DefaultConfig(tp)
		return commonConfig
	default:
		panic("node: invalid node type")
	}
}
