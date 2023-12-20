package main

import (
	"context"
	"fmt"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

func startPeer(ctx context.Context, log func(msg string, level string)) error {
	log("Initializing new P2P package instance...", "debug")

	// Create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New()
	if err != nil {
		return fmt.Errorf("failed to start libp2p: %s", err)
	}

	peerID := h.ID()
	log(fmt.Sprintf("Created new P2P instance with Peer ID: %s", peerID.String()), "debug")

	info, err := peer.AddrInfoFromString("/ip4/127.0.0.1/udp/53688/quic-v1/webtransport/certhash/uEiBv_tYyk1kMv9irNyvmGOF1ovI4akRptbKu2edtCroKRA/certhash/uEiAGeihPco5ClWTN6ZInkdfa6qzrZSFFTziUpuo7lMlPhg/p2p/12D3KooWKxzPvdZWrLDsqX1FQ9KcLCzQCdbbh1iVh8aLz6Fjmd5F")
	if err != nil {
		return fmt.Errorf("parsing peer address: %w", err)
	}

	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log("P2P client ping connectivity tester exited", "warn")
				return
			default:
				startTime := time.Now() // Record the start time
				pingServer()
				elapsedTime := time.Since(startTime)
				log(fmt.Sprintf("Client ping took: %s", elapsedTime), "debug")
				time.Sleep(5 * time.Second)
			}
		}
	}()

	s, err := h.NewStream(ctx, info.ID, "/multistream/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open stream: %s", err)
	}

	_ = s

	// Log the listening addresses
	for _, addr := range h.Addrs() {
		fullAddr := fmt.Sprintf("%s/p2p/%s", addr, peerID.String())
		log(fmt.Sprintf("Listening on: %s", fullAddr), "info")
	}

	log("I am here....", "info")

	select {
	case <-ctx.Done():
		log("P2P peer exited", "warn")
		return nil
	}
}

func pingServer() {
	// Simulate pinging the server here
	fmt.Printf("Pinging client ...")
}

func handleStream(log func(msg string, level string)) func(s network.Stream) {
	return func(s network.Stream) {
		log("New stream opened", "info")
	}
}
