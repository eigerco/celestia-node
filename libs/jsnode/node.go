//go:build js

package jsnode

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"
	"time"

	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
)

func RegisterJSFunctions() {
	global := js.Global()
	h := &nodeHandler{}
	global.Set("startNode", js.FuncOf(h.startNode))
	global.Set("stopNode", js.FuncOf(h.stopNode))
}

type nodeHandler struct {
	node   *nodebuilder.Node
	ctx    context.Context
	cancel func()
}

func (h *nodeHandler) startNode(this js.Value, args []js.Value) any {
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

	go h.start(cfg)
	return nil
}

func (h *nodeHandler) stopNode(this js.Value, args []js.Value) any {
	go h.stop()
	return nil
}

func (h *nodeHandler) start(cfg *nodebuilder.Config) {
	h.ctx, h.cancel = context.WithCancel(context.Background())

	store, err := nodebuilder.NewIndexedDBStore(h.ctx, cfg)
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

	h.node, err = nodebuilder.NewWithConfig(node.Light, p2p.Mainnet, store, cfg, nodebuilder.WithMetrics())
	if err != nil {
		log(fmt.Sprintf("Failed to create new node: %s", err), "error")
		return
	}

	log("New node created successfully!", "debug")

	log("Starting node", "info")
	if err := h.node.Start(h.ctx); err != nil {
		log(fmt.Sprintf("Failed to start node: %s", err), "error")
		return
	}

	log("Node started successfully!", "info")

	// Call a JavaScript function and pass the Peer ID
	// We use this peer ids to display which peer current running node is using.
	js.Global().Call("setPeerID", h.node.Host.ID().String())

	go func() {
		// update the nod info every second
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				for _, topic := range h.node.PubSub.GetTopics() {
					peerIDs := h.node.PubSub.ListPeers(topic)

					jsPeerValues := js.Global().Get("Array").New(len(peerIDs))
					for i, s := range peerIDs {
						jsPeerValues.SetIndex(i, s.String())
					}
					js.Global().Call("setConnectedPeers", jsPeerValues)
				}
				syncState, err := h.node.HeaderServ.SyncState(h.ctx)
				if err != nil {
					log(fmt.Sprintf("Failed to get sync syncState: %s", err), "error")
					return
				}
				syncStateBytes, err := json.Marshal(syncState)
				if err != nil {
					log(fmt.Sprintf("Failed to marshal sync sync state: %s", err), "error")
					return
				}
				js.Global().Call("syncerInfo", string(syncStateBytes))

				networkHead, err := h.node.HeaderServ.NetworkHead(h.ctx)
				if err != nil {
					log(fmt.Sprintf("Failed to get sync syncState: %s", err), "error")
				}
				networkHeadBytes, err := json.Marshal(networkHead)
				if err != nil {
					log(fmt.Sprintf("Failed to get network head: %s", err), "error")
				}
				js.Global().Call("setNetworkHead", string(networkHeadBytes))
			case <-h.ctx.Done():
				return
			}
		}
	}()

	js.Global().Call("startedNode")

	<-h.ctx.Done()
	return
}

func (h *nodeHandler) stop() {
	if h.node == nil {
		log("Node is not running", "warn")
		return
	}

	ctx, cl := context.WithTimeout(context.Background(), 2*time.Second)
	defer cl()
	if err := h.node.Stop(ctx); err != nil {
		log(fmt.Sprintf("Failed to stop node: %s", err), "error")
		return
	}

	log("Node stopped successfully!", "info")
	if h.cancel != nil {
		h.cancel()
	}
	return
}

func log(msg string, level string) {
	js.Global().Call("appendLog", msg, level)
}
