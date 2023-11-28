//go:build light

package p2p

import "go.uber.org/fx"

func lightModule(baseComponents fx.Option) (fx.Option, error) {
	return fx.Module(
		"p2p",
		baseComponents,
		fx.Provide(blockstoreFromDatastore),
		fx.Provide(autoscaleResources),
	), nil
}
