//go:build notracing

package eds

import "go.opentelemetry.io/otel/trace/noop"

var tracer = noop.NewTracerProvider().Tracer("share/eds")
