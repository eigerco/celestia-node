//go:build !metrics

package peers

import (
	"context"
	"fmt"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"time"
)

type metrics struct {
}

func (m metrics) validationObserver(validator shrexsub.ValidatorFn) shrexsub.ValidatorFn {
	return func(ctx context.Context, id peer.ID, n shrexsub.Notification) pubsub.ValidationResult {
		res := validator(ctx, id, n)

		var resStr string
		switch res {
		case pubsub.ValidationAccept:
			resStr = validationAccept
		case pubsub.ValidationReject:
			resStr = validationReject
		case pubsub.ValidationIgnore:
			resStr = validationIgnore
		default:
			resStr = "unknown"
		}

		if ctx.Err() != nil {
			ctx = context.Background()
		}
		fmt.Printf("Got the RESOLUTION STRING: %s \n", resStr)
		return res
	}
}

func (m metrics) observeGetPeer(ctx context.Context, source peerSource, size int, time time.Duration) {

}

func (m metrics) observeDoneResult(source peerSource, r result) {

}

func (m metrics) observeBlacklistPeers(reason blacklistPeerReason, i int) {

}

type blacklistPeerReason string

type peerStatus string

type poolStatus string

type peerSource string

const (
	isInstantKey  = "is_instant"
	doneResultKey = "done_result"

	sourceKey                  = "source"
	sourceShrexSub  peerSource = "shrexsub"
	sourceFullNodes peerSource = "full_nodes"

	blacklistPeerReasonKey                     = "blacklist_reason"
	reasonInvalidHash      blacklistPeerReason = "invalid_hash"
	reasonMisbehave        blacklistPeerReason = "misbehave"

	validationResultKey = "validation_result"
	validationAccept    = "accept"
	validationReject    = "reject"
	validationIgnore    = "ignore"

	peerStatusKey                 = "peer_status"
	peerStatusActive   peerStatus = "active"
	peerStatusCooldown peerStatus = "cooldown"

	poolStatusKey                    = "pool_status"
	poolStatusCreated     poolStatus = "created"
	poolStatusValidated   poolStatus = "validated"
	poolStatusSynced      poolStatus = "synced"
	poolStatusBlacklisted poolStatus = "blacklisted"
	// Pool status model:
	//        	created(unvalidated)
	//  	/						\
	//  validated(unsynced)  	  blacklisted
	//			|
	//  	  synced
)
