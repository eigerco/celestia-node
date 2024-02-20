//go:build wasm

package das

import (
	"context"
	"time"

	"github.com/celestiaorg/celestia-node/header"
)

type metrics struct {
}

func (m metrics) observeNewHead(ctx context.Context) {

}

func (m metrics) observeSample(ctx context.Context, h *header.ExtendedHeader, since time.Duration, t jobType, err error) {

}

func (m metrics) observeGetHeader(ctx context.Context, since time.Duration) {

}

func (d *DASer) InitMetrics() error {
	d.sampler.metrics = &metrics{}
	return nil
}
