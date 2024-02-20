//go:build !wasm

package share

import (
	"github.com/celestiaorg/celestia-node/share/eds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexeds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexnd"
)

func WithStoreMetrics(s *eds.Store) error {
	return s.WithMetrics()
}

func WithShrexServerMetrics(edsServer *shrexeds.Server, ndServer *shrexnd.Server) error {
	err := edsServer.WithMetrics()
	if err != nil {
		return err
	}

	return ndServer.WithMetrics()
}
