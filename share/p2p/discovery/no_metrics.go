//go:build !metrics

package discovery

import "context"

type metrics struct {
}

func (m metrics) observeAdvertise(ctx context.Context, err error) {

}

func (m metrics) observeFindPeers(ctx context.Context, peers bool) {

}

func (m metrics) observeHandlePeer(ctx context.Context, self interface{}) {

}

type handlePeerResult string

const (
	discoveryEnoughPeersKey = "enough_peers"

	handlePeerResultKey                    = "result"
	handlePeerSkipSelf    handlePeerResult = "skip_self"
	handlePeerEnoughPeers handlePeerResult = "skip_enough_peers"
	handlePeerBackoff     handlePeerResult = "skip_backoff"
	handlePeerConnected   handlePeerResult = "connected"
	handlePeerConnErr     handlePeerResult = "conn_err"
	handlePeerInSet       handlePeerResult = "in_set"

	advertiseFailedKey = "failed"
)
