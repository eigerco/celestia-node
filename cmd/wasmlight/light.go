//go:build wasm && js

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall/js"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
)

var (
	nd     *nodebuilder.Node
	cancel context.CancelFunc
)

func log(msg string, level string) {
	js.Global().Get("appendLog").Invoke(msg, level)
}

func main() {
	logging.SetupLogging(logging.Config{
		Stderr: true,
	})

	if err := os.Setenv("CELESTIA_ENABLE_QUIC", "true"); err != nil {
		panic(err)
		return
	}

	if err := os.Setenv("CELESTIA_HOME", "test"); err != nil {
		panic(err)
		return
	}
	if err := os.Setenv("HOME", "test"); err != nil {
		panic(err)
		return
	}

	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	js.Global().Set("startNode", js.FuncOf(func(this js.Value, args []js.Value) any {
		bootstrapAddressesStr := args[0].String()
		cfg := nodebuilder.DefaultConfig(node.Light)
		bootstrapAddresses := strings.Split(bootstrapAddressesStr, "\n")
		for _, addr := range bootstrapAddresses {
			addr := strings.TrimSpace(addr)
			if len(addr) > 0 {
				cfg.P2P.BootstrapAddresses = append(cfg.P2P.BootstrapAddresses, addr)
				cfg.Header.TrustedPeers = append(cfg.P2P.BootstrapAddresses, addr)
			}
		}
		go start(ctx, cfg)
		return nil
	}))

	js.Global().Set("stopNode", js.FuncOf(func(this js.Value, args []js.Value) any {
		go stop(ctx)
		return nil
	}))

	select {
	case <-ctx.Done():
		log("Node exited", "warn")
		return
	}
}

func start(ctx context.Context, cfg *nodebuilder.Config) {
	store, err := nodebuilder.NewIndexedDBStore(ctx, cfg)
	if err != nil {
		log(fmt.Sprintf("Failed to init indexeddb store: %s", err), "error")
		return
	}
	defer store.Close()
	log("Store opened successfully!", "debug")

	ks, _ := store.Keystore() // we know for sure there is no error
	if err := nodebuilder.GenerateKeys(ks.Keyring()); err != nil {
		log(fmt.Sprintf("Failed to generate keys: %s", err), "error")
		return
	}
	log("Keys generated successfully!", "debug")

	nd, err = nodebuilder.NewWithConfig(node.Light, p2p.Mainnet, store, cfg, nodebuilder.WithMetrics())
	if err != nil {
		log(fmt.Sprintf("Failed to create new node: %s", err), "error")
		return
	}

	log("New node created successfully!", "debug")

	log("Starting node", "info")
	if err := nd.Start(ctx); err != nil {
		log(fmt.Sprintf("Failed to start node: %s", err), "error")
		return
	}

	log("Node started successfully!", "info")

	addrs, err := peer.AddrInfoToP2pAddrs(host.InfoFromHost(nd.Host))
	if err != nil {
		log(fmt.Sprintf("Retrieving multiaddress information error: %s", err), "error")
		return
	}

	// Call a JavaScript function and pass the Peer ID
	// We use this peer ids to display which peer current running node is using.
	for _, addr := range addrs {
		js.Global().Call("setPeerID", addr.String())
	}

	js.Global().Call("startedNode")

	<-ctx.Done()
	return
}

func stop(ctx context.Context) {
	if nd == nil {
		log("Node is not running", "warn")
		return
	}

	if err := nd.Stop(ctx); err != nil {
		log(fmt.Sprintf("Failed to stop node: %s", err), "error")
		return
	}

	log("Node stopped successfully", "info")
	if cancel != nil {
		cancel()
	}
	return
}
