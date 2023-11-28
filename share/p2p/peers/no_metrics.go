//go:build !metrics

package peers

import (
	"context"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
	"time"
)

type metrics struct {
}

func (m metrics) validationObserver(validate shrexsub.ValidatorFn) shrexsub.ValidatorFn {
	return validate
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
