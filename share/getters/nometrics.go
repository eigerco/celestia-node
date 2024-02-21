//go:build wasm

package getters

import "context"

type metrics struct{}

func (m metrics) recordEDSAttempt(ctx context.Context, attempt int, b bool) {}

func (m metrics) recordNDAttempt(ctx context.Context, attempt int, b bool) {}

func (sg *ShrexGetter) WithMetrics() error {
	sg.metrics = &metrics{}
	return nil
}
