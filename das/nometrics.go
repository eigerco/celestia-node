//go:build nometrics

package das

import (
	"context"
	"github.com/celestiaorg/celestia-node/header"
	"time"
)

type metrics struct {
}

func (m metrics) observeNewHead(ctx context.Context) {

}

func (m metrics) observeSample(ctx context.Context, h *header.ExtendedHeader, since time.Duration, t jobType, err error) {

}

func (m metrics) observeGetHeader(ctx context.Context, since time.Duration) {

}
