package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/peer"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type WebtransportBootstrappers struct {
	Addrs map[string]string `json:"addrs"`
}

func main() {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()
	zap.ReplaceGlobals(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := os.Getenv("CELESTIA_NODE_IP_ADDR")
	token := os.Getenv("CELESTIA_NODE_AUTH_TOKEN")

	nodeCli, err := client.NewClient(ctx, addr, token)
	if err != nil {
		logger.Error("failure to resolve new celestia client", zap.Error(err))
		return
	}

	http.HandleFunc("/peers", func(w http.ResponseWriter, r *http.Request) {
		// Set the Access-Control-Allow-Origin header to allow requests from localhost:8080
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		peers, err := nodeCli.P2P.Peers(ctx)
		if err != nil {
			logger.Error("failed to retrieve peers", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		wtPeers := make(map[string]string)

		var mutex sync.Mutex
		var wg sync.WaitGroup

		for _, prr := range peers {
			wg.Add(1)
			go func(pID peer.ID) {
				defer wg.Done()

				peerInfo, err := nodeCli.P2P.PeerInfo(ctx, pID)
				if err != nil {
					logger.Error("failed to get peer info", zap.Error(err))
					return
				}

				for _, addr := range peerInfo.Addrs {
					maddr, err := multiaddr.NewMultiaddrBytes(addr.Bytes())
					if err != nil {
						logger.Error("failed to parse multiaddr", zap.Error(err))
						continue
					}
					if isWebtransportPeer(maddr) {
						mutex.Lock()
						wtPeers[pID.String()] = fmt.Sprintf("%s/p2p/%s", maddr, pID.String())
						mutex.Unlock()
						break
					}
				}
			}(prr)
		}

		wg.Wait()

		webtransportBootstrappers := WebtransportBootstrappers{
			Addrs: wtPeers,
		}

		responseJSON, err := json.Marshal(webtransportBootstrappers)
		if err != nil {
			logger.Error("failed to marshal response to JSON", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	logger.Info("Starting bootstrapper server on http://localhost:8096")
	if err := http.ListenAndServe(":8096", nil); err != nil {
		logger.Fatal("failed to start HTTP server", zap.Error(err))
	}
}

// Function to check if a peer is a secure webtransport containing certhash and p2p
func isWebtransportPeer(addr multiaddr.Multiaddr) bool {
	var hasIPv4, hasUDP, hasWebtransport, hasCerthash bool

	for _, protocol := range addr.Protocols() {
		switch protocol.Code {
		case multiaddr.P_IP4:
			hasIPv4 = true
		case multiaddr.P_UDP:
			hasUDP = true
		case multiaddr.P_WEBTRANSPORT:
			hasWebtransport = true
		case multiaddr.P_CERTHASH:
			hasCerthash = true
		}
	}

	if hasIPv4 && hasUDP && hasWebtransport && hasCerthash {
		zap.L().Info("Discovered WebTransport peer addr info", zap.Any("peer", addr.String()), zap.Any("protocol", addr.Protocols()))
		return true
	}

	return false
}
