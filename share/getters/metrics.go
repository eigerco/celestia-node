//go:build !wasm

package getters

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("shrex/getter")

type metrics struct {
	edsAttempts metric.Int64Histogram
	ndAttempts  metric.Int64Histogram
}

func (m *metrics) recordEDSAttempt(ctx context.Context, attemptCount int, success bool) {
	if m == nil {
		return
	}
	if ctx.Err() != nil {
		ctx = context.Background()
	}
	m.edsAttempts.Record(ctx, int64(attemptCount),
		metric.WithAttributes(
			attribute.Bool("success", success)))
}

func (m *metrics) recordNDAttempt(ctx context.Context, attemptCount int, success bool) {
	if m == nil {
		return
	}
	if ctx.Err() != nil {
		ctx = context.Background()
	}
	m.ndAttempts.Record(ctx, int64(attemptCount),
		metric.WithAttributes(
			attribute.Bool("success", success)))
}

func (sg *ShrexGetter) WithMetrics() error {
	edsAttemptHistogram, err := meter.Int64Histogram(
		"getters_shrex_eds_attempts_per_request",
		metric.WithDescription("Number of attempts per shrex/eds request"),
	)
	if err != nil {
		return err
	}

	ndAttemptHistogram, err := meter.Int64Histogram(
		"getters_shrex_nd_attempts_per_request",
		metric.WithDescription("Number of attempts per shrex/nd request"),
	)
	if err != nil {
		return err
	}

	sg.metrics = &metrics{
		edsAttempts: edsAttemptHistogram,
		ndAttempts:  ndAttemptHistogram,
	}
	return nil
}
