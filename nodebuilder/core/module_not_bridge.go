//go:build !bridge_full

package core

import (
	"errors"

	"go.uber.org/fx"
)

func moduleBridge() (fx.Option, error) {
	return nil, errors.New("the program was built without the bridge node functionality, needs to be built using `go build -tags=bridge_full`")
}
