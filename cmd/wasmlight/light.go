//go:build light && wasm
// +build light,wasm

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall/js"

	"github.com/BurntSushi/toml"
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

	js.Global().Set("startNode", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		configStr := args[0].String()

		fmt.Println("attempting to start node")
		go start(ctx, log, configStr)
		return nil
	}))

	js.Global().Set("stopNode", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("attempting to stop node")
		go stop(ctx, log)
		return nil
	}))

	select {
	case <-ctx.Done():
		log("Node exited", "warn")
		return
	}
}

func start(ctx context.Context, log func(msg string, level string), configStr string) {
	cfg := &nodebuilder.Config{}
	toml.Decode(configStr, cfg)

	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	ring, err := keyring.New(app.Name, keyring.BackendMemory, "", os.Stdin, encConf.Codec)
	if err != nil {
		log(fmt.Sprintf("Failed to create keyring: %s", err), "error")
		return
	}

	basePath := ".celestia-light-arabica-10"
	if err := nodebuilder.SaveConfig(strings.Join([]string{basePath, "config.toml"}, "/"), cfg); err != nil {
		log(fmt.Sprintf("unable to save config %s", err), "error")
		return
	}

	store, err := nodebuilder.OpenStore(basePath, ring)
	if err != nil {
		log(fmt.Sprintf("Failed to open store: %s", err), "error")
		return
	}
	defer store.Close()

	nd, err = nodebuilder.New(node.Light, p2p.Arabica, store)
	if err != nil {
		log(fmt.Sprintf("Failed to create new node: %s", err), "error")
		return
	}
	if err := nd.Start(ctx); err != nil {
		log(fmt.Sprintf("Failed to start node: %s", err), "error")
		return
	}

	log("Node started successfully", "info")
	<-ctx.Done()
	return
}

func stop(ctx context.Context, log func(msg string, level string)) error {
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
