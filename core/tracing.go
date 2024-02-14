//go:build !notracing

package core

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("core/listener")
