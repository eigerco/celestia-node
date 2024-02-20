//go:build !notracing

package eds

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("share/eds")
