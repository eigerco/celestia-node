//go:build !notracing

package getters

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("share/getters")
