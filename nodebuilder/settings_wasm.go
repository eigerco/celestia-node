//go:build wasm

package nodebuilder

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/nodebuilder/share"
)

// WithMetrics set dummy metrics
func WithMetrics() fx.Option {
	return fx.Options(
		fx.Options(
			fx.Invoke(share.WithDiscoveryMetrics),
		),
		fx.Options(
			fx.Invoke(share.WithPeerManagerMetrics),
			fx.Invoke(share.WithShrexClientMetrics),
			fx.Invoke(share.WithShrexGetterMetrics),
		),
	)
}
