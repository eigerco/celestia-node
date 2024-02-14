//go:build nometrics

package p2p

import (
	"context"
)

type status string

const (
	StatusBadRequest  status = "bad_request"
	StatusSendRespErr status = "send_resp_err"
	StatusSendReqErr  status = "send_req_err"
	StatusReadRespErr status = "read_resp_err"
	StatusInternalErr status = "internal_err"
	StatusNotFound    status = "not_found"
	StatusTimeout     status = "timeout"
	StatusSuccess     status = "success"
	StatusRateLimited status = "rate_limited"
)

type Metrics struct {
}

// ObserveRequests increments the total number of requests sent with the given status as an
// attribute.
func (m *Metrics) ObserveRequests(ctx context.Context, count int64, status status) {
}

func InitClientMetrics(protocol string) (*Metrics, error) {
	return &Metrics{}, nil
}

func InitServerMetrics(protocol string) (*Metrics, error) {
	return &Metrics{}, nil
}
