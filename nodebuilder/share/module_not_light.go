//go:build !light

package share

import (
	"errors"

	"go.uber.org/fx"
)

func lightModule(baseComponents, peerManagerWithShrexPools, shrexGetterComponents fx.Option, sampleAmount uint) (fx.Option, error) {
	return nil, errors.New("the program was built without the light node functionality, needs to be built using `go build -tags=light`")
}
