//go:build !wasm

package nodebuilder

import (
	"github.com/celestiaorg/celestia-node/nodebuilder/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/das"
	"github.com/celestiaorg/celestia-node/nodebuilder/gateway"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/nodebuilder/rpc"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
	"github.com/celestiaorg/celestia-node/nodebuilder/state"
)

// Config is main configuration structure for a Node.
// It combines configuration units for all Node subsystems.
type Config struct {
	Node    node.Config
	Core    core.Config
	State   state.Config
	P2P     p2p.Config
	RPC     rpc.Config
	Gateway gateway.Config
	Share   share.Config
	Header  header.Config
	DASer   das.Config `toml:",omitempty"`
}

// DefaultConfig provides a default Config for a given Node Type 'tp'.
// NOTE: Currently, configs are identical, but this will change.
func DefaultConfig(tp node.Type) *Config {
	commonConfig := &Config{
		Node:    node.DefaultConfig(tp),
		Core:    core.DefaultConfig(),
		State:   state.DefaultConfig(),
		P2P:     p2p.DefaultConfig(tp),
		RPC:     rpc.DefaultConfig(),
		Gateway: gateway.DefaultConfig(),
		Share:   share.DefaultConfig(tp),
		Header:  header.DefaultConfig(tp),
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
