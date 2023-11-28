//go:build metrics

package node

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/celestiaorg/celestia-node/nodebuilder"
)

var meter = otel.Meter("node")

var (
	timeStarted time.Time
	nodeStarted bool
)

// WithMetrics registers node metrics.
func WithMetrics() error {
	nodeStartTS, err := meter.Int64ObservableGauge(
		"node_start_ts",
		metric.WithDescription("timestamp when the node was started"),
	)
	if err != nil {
		return err
	}

	totalNodeRunTime, err := meter.Float64ObservableCounter(
		"node_runtime_counter_in_seconds",
		metric.WithDescription("total time the node has been running"),
	)
	if err != nil {
		return err
	}

	buildInfoGauge, err := meter.Float64ObservableGauge(
		"build_info",
		metric.WithDescription("Celestia Node build information"),
	)
	if err != nil {
		return err
	}

	callback := func(ctx context.Context, observer metric.Observer) error {
		if !nodeStarted {
			// Observe node start timestamp
			timeStarted = time.Now()
			observer.ObserveInt64(nodeStartTS, timeStarted.Unix())
			nodeStarted = true
		}

		observer.ObserveFloat64(totalNodeRunTime, time.Since(timeStarted).Seconds())

		// Observe build info with labels
		labels := metric.WithAttributes(
			attribute.String("build_time", nodebuilder.buildTime),
			attribute.String("last_commit", nodebuilder.lastCommit),
			attribute.String("semantic_version", nodebuilder.semanticVersion),
			attribute.String("system_version", nodebuilder.systemVersion),
			attribute.String("golang_version", nodebuilder.golangVersion),
		)

		observer.ObserveFloat64(buildInfoGauge, 1, labels)

		return nil
	}

	_, err = meter.RegisterCallback(callback, nodeStartTS, totalNodeRunTime, buildInfoGauge)

	return err
}
