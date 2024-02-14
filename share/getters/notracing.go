//go:build notracing

package getters

import "go.opentelemetry.io/otel/trace/noop"

var tracer = noop.NewTracerProvider().Tracer("share/getters")
