//go:build !wasm

package node

import (
	"github.com/cristalhq/jwt"
	"go.uber.org/fx"
)

func ConstructModule(tp Type) fx.Option {
	return fx.Module(
		"node",
		fx.Provide(func(secret jwt.Signer) Module {
			return NewModule(tp, secret)
		}),
		fx.Provide(Secret),
	)
}
