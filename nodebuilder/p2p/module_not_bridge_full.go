//go:build !bridge_full

package p2p

import (
	"errors"

	"go.uber.org/fx"
)

func bridgeFullModule(baseComponents fx.Option) (fx.Option, error) {
	return nil, errors.New("the program was built without the bridge node functionality, needs to be built using `go build -tags=bridge_full`")
}
