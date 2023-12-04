//go:build light && wasm
// +build light,wasm

package main

import (
	"context"
	"fmt"
	"os"
	"syscall/js"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

var (
	nd     *nodebuilder.Node
	cancel context.CancelFunc
)

func main() {
	appendLog := js.Global().Get("appendLog")

	log := func(msg string, level string) {
		appendLog.Invoke(msg, level)
	}

	ctx, stop := context.WithCancel(context.Background())
	cancel = stop

	js.Global().Set("startNode", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go Start(ctx, log)
		return nil
	}))

	js.Global().Set("stopNode", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go Stop(ctx, log)
		return nil
	}))

	select {
	case <-ctx.Done():
		log("Node exited", "warn")
		return
	}
}

func Start(ctx context.Context, log func(msg string, level string)) error {
	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	ring, err := keyring.New(app.Name, keyring.BackendMemory, "", os.Stdin, encConf.Codec)
	if err != nil {
		log(fmt.Sprintf("Failed to create keyring: %s", err), "error")
		return err
	}

	store, err := nodebuilder.OpenStore(".celestia-light-arabica-10", ring)
	if err != nil {
		log(fmt.Sprintf("Failed to open store: %s", err), "error")
		return err
	}
	defer store.Close()

	nd, err = nodebuilder.New(node.Light, p2p.Arabica, store)
	if err != nil {
		log(fmt.Sprintf("Failed to create new node: %s", err), "error")
		return err
	}

	if err := nd.Start(ctx); err != nil {
		log(fmt.Sprintf("Failed to start node: %s", err), "error")
		return err
	}

	log("Node started successfully", "info")
	<-ctx.Done()
	return nil
}

func Stop(ctx context.Context, log func(msg string, level string)) error {
	if nd != nil {
		if err := nd.Stop(ctx); err != nil {
			log(fmt.Sprintf("Failed to stop node: %s", err), "error")
			return err
		}
		log("Node stopped successfully", "info")
	} else {
		log("No node instance found to stop", "warn")
	}
	if cancel != nil {
		cancel()
	}
	return nil
}
