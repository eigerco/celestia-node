//go:build !light

package p2p

import (
	"errors"

	"go.uber.org/fx"
)

func lightModule(baseComponents fx.Option) (fx.Option, error) {
	return nil, errors.New("the program was built without the bridge node functionality, needs to be built using `go build -tags=light`")
}
