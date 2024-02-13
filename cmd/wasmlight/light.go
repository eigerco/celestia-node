//go:build wasm && js

package main

import (
	"context"
	"fmt"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-node/libs/codec"
	"github.com/celestiaorg/celestia-node/libs/keystore"
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

const (
	basePath        = ".celestia-light-arabica-10"
	keyringPassword = "testpassword" //TODO
)

var (
	nd     *nodebuilder.Node
	cancel context.CancelFunc
)

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
	appendLog := js.Global().Get("appendLog")

	log := func(msg string, level string) {
		appendLog.Invoke(msg, level)
	}

	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	/*
		js.Global().Set("initNode", js.FuncOf(func(this js.Value, args []js.Value) any {
			bootstrapAddressesStr := args[0].String()
			return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				resolve := args[0]
				reject := args[1]

				go func() {
					cfg := nodebuilder.DefaultConfig(node.Light)
					bootstrapAddresses := strings.Split(bootstrapAddressesStr, "\n")
					for _, addr := range bootstrapAddresses {
						addr := strings.TrimSpace(addr)
						if len(addr) > 0 {
							cfg.P2P.BootstrapAddresses = append(cfg.P2P.BootstrapAddresses, addr)
						}
					}

					encConf := encoding.MakeConfig(codec.ModuleEncodingRegisters...)
					ring, err := keystore.OpenIndexedDB(encConf.Codec, keyringPassword)
					if err != nil {
						log(fmt.Sprintf("Failed to open keyring: %s", err), "error")
						reject.Invoke(err)
						return
					}
					if err := nodebuilder.InitWasm(ring, *cfg, basePath); err != nil {
						log(fmt.Sprintf("Failed to init: %s", err), "error")
						reject.Invoke(err)
						return
					}
					resolve.Invoke(js.Null())
				}()

				return nil
			}))
		}))
	*/

	js.Global().Set("startNode", js.FuncOf(func(this js.Value, args []js.Value) any {
		bootstrapAddressesStr := args[0].String()
		cfg := nodebuilder.DefaultConfig(node.Light)

		bootstrapAddresses := strings.Split(bootstrapAddressesStr, "\n")
		for _, addr := range bootstrapAddresses {
			addr := strings.TrimSpace(addr)
			if len(addr) > 0 {
				cfg.P2P.BootstrapAddresses = append(cfg.P2P.BootstrapAddresses, addr)
			}
		}

		go start(ctx, cfg, log)
		return nil
	}))

	js.Global().Set("stopNode", js.FuncOf(func(this js.Value, args []js.Value) any {
		go stop(ctx, log)
		return nil
	}))

	/* 	go func() {
		log("Starting up P2P connectivity tester...", "info")
		if err := startPeer(ctx, log); err != nil {
			log(fmt.Sprintf("Failed to start peer: %s", err), "error")
			return
		}
	}() */

	select {
	case <-ctx.Done():
		log("Node exited", "warn")
		return
	}
}

func start(ctx context.Context, cfg *nodebuilder.Config, log func(msg string, level string)) {
	encConf := encoding.MakeConfig(codec.ModuleEncodingRegisters...)
	ring, err := keystore.OpenIndexedDB(encConf.Codec, keyringPassword)
	if err != nil {
		log(fmt.Sprintf("Failed to open indexedDB: %s", err), "error")
	}

	log("Starting node", "info")

	store, err := nodebuilder.OpenStore(basePath, ring)
	if err != nil {
		log(fmt.Sprintf("Failed to open store: %s", err), "error")
		return
	}
	defer store.Close()

	log("Store opened successfully!", "debug")

	nd, err = nodebuilder.NewWithConfig(node.Light, p2p.Mainnet, store, cfg)
	if err != nil {
		log(fmt.Sprintf("Failed to create new node: %s", err), "error")
		return
	}

	log("New node created successfully!", "debug")

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

	fmt.Println("The p2p host is listening on AAAAA:")
	for _, addr := range addrs {
		fmt.Println("* ", addr.String())
		// Call a JavaScript function and pass the Peer ID
		js.Global().Call("setPeerID", addr.String())
	}
	fmt.Println()

	js.Global().Call("startedNode")

	<-ctx.Done()
	return
}

func stop(ctx context.Context, log func(msg string, level string)) error {
	if nd == nil {
		log("Node is not running", "warn")
		return nil
	}

	if err := nd.Stop(ctx); err != nil {
		log(fmt.Sprintf("Failed to stop node: %s", err), "error")
		return err
	}

	log("Node stopped successfully", "info")
	if cancel != nil {
		cancel()
	}
	return nil
}
