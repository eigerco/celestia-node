//go:build notracing

package core

import (
	"go.opentelemetry.io/otel/trace/noop"
)

var tracer = noop.NewTracerProvider().Tracer("core/listener")
