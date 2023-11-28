///go:build light && wasm

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// NOTE: We should always ensure that the added Flags below are parsed somewhere, like in the
// PersistentPreRun func on parent command.

func main() {
	if err := Start(); err != nil {
		panic(err)
	}
}

// Start constructs a CLI command to start Celestia Node daemon of any type with the given flags.
func Start() error {
	ctx := context.Background()

	// override config with all modifiers passed on start
	//cfg := NodeConfig(ctx)

	//storePath := StorePath(ctx)
	//keysPath := filepath.Join(".celestia-light-arabica-10", "keys")

	// construct ring
	// TODO @renaynay: Include option for setting custom `userInput` parameter with
	//  implementation of https://github.com/celestiaorg/celestia-node/issues/415.
	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	ring, err := keyring.New(app.Name, keyring.BackendMemory, "", os.Stdin, encConf.Codec)
	if err != nil {
		return err
	}

	store, err := nodebuilder.OpenStore(".celestia-light-arabica-10", ring) // TODO What shall we do with the store thing???
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, store.Close())
	}()

	nd, err := nodebuilder.New(node.Light, p2p.Arabica, store) // TODO, NodeOptions(ctx)...)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	err = nd.Start(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	cancel() // ensure we stop reading more signals for start context

	ctx, cancel = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	return nd.Stop(ctx)
}
