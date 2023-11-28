//go:build tracing

package attribute

import (
	"go.opentelemetry.io/otel/attribute"
)

var (
	Int64  = attribute.Int64
	Int    = attribute.Int
	String = attribute.String
)
