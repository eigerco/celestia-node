package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WebtransportBootstrappers holds a map of peer addresses.
type WebtransportBootstrappers struct {
	Addrs map[string]string `json:"addrs"`
}

// Map to store custom ports for specific IP addresses
var customPorts = map[string]int{
	"40.85.94.176":   6060, // Eiger custom node exposing docker port under different port
	"40.127.100.171": 6060, // ...
	"10.0.2.100":     6060, // ...
}

var customIPs = map[string]string{
	"10.0.2.100": "40.127.100.171",
}

func main() {
	// Initialize logger with development configurations.
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()
	zap.ReplaceGlobals(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get environment variables for Celestia node configuration.
	addr := os.Getenv("CELESTIA_NODE_IP_ADDR")
	token := os.Getenv("CELESTIA_NODE_AUTH_TOKEN")

	// Initialize Celestia client.
	nodeCli, err := client.NewClient(ctx, addr, token)
	if err != nil {
		logger.Error("Failure to resolve new celestia client", zap.Error(err))
		return
	}

	// Set up HTTP handler for the "/peers" endpoint.
	http.HandleFunc("/bootstrap-peers", BootstrapPeersHandler(nodeCli))

	// Set up HTTP handler for the "/peers" endpoint.
	http.HandleFunc("/peers", PeersHandler(nodeCli))

	// Start HTTP server.
	logger.Info("Started bootstrapper server at http://localhost:8096")
	if err := http.ListenAndServe(":8096", nil); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

func BootstrapPeersHandler(nodeCli *client.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		peerInfo, err := nodeCli.P2P.Info(ctx)
		if err != nil {
			zap.L().Error("failed to retrieve node peer information", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Map to store unique certhashes
		uniqueCerthashes := make(map[string]bool)
		uniqueAddrs := make(map[string]string)

		for _, addr := range peerInfo.Addrs {
			if isWebtransportPeer(addr) {
				certhashes := extractCerthashes(addr)
				for _, certhash := range certhashes {
					// Check if certhash already exists, if so, skip adding it
					if _, ok := uniqueCerthashes[certhash]; ok {
						continue
					}
					uniqueCerthashes[certhash] = true

					// Replace the port if there is a custom port defined for this IP
					addrStr := addr.String()

					// Get the original IP address from the multiaddress
					ip, err := addr.ValueForProtocol(multiaddr.P_IP4)
					if err != nil {
						zap.L().Error(
							"failure to extract IPV4 value of an address",
							zap.String("addr", addr.String()),
							zap.Error(err),
						)
						continue
					}

					if customPort, ok := customPorts[ip]; ok {
						addrStr = replacePort(addrStr, customPort)
					}

					if customIP, ok := customIPs[ip]; ok {
						addrStr = replaceIP(addrStr, customIP)
					}

					uniqueAddrs[peerInfo.ID.String()] = fmt.Sprintf("%s/p2p/%s", addrStr, peerInfo.ID.String())
				}
			}
		}

		bootstrappers := WebtransportBootstrappers{
			Addrs: uniqueAddrs,
		}

		responseJSON, err := json.Marshal(bootstrappers)
		if err != nil {
			zap.L().Error("failed to marshal response to JSON", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}
}

// replacePort replaces the existing QUIC UDP port in the multiaddress string with the given custom port
func replacePort(addrStr string, customPort int) string {
	// Define the regular expression pattern to match the port number
	pattern := regexp.MustCompile(`udp/\d+/quic-v1`)

	// Replace the port number with the custom port
	replaced := pattern.ReplaceAllString(addrStr, "udp/"+strconv.Itoa(customPort)+"/quic-v1")

	return replaced
}

// replaceIP replaces the existing IP address in the multiaddress string with the given custom IP
func replaceIP(addrStr string, customIP string) string {
	// This pattern matches:
	// - IPv4 addresses (e.g., 192.168.1.1)
	// - IPv6 addresses (e.g., [fe80::1])
	pattern := regexp.MustCompile(`(?:\[/[0-9a-fA-F:]+\]|/[0-9\.]+)`)

	// Replace the IP address with the custom IP
	replaced := pattern.ReplaceAllString(addrStr, customIP)

	return replaced
}

// Assuming the certhash follows the format "/webtransport/certhash/<certhash>"
// Extract the certhash value from the protocol string
func extractCerthashes(addr multiaddr.Multiaddr) []string {
	var certhashes []string

	// Iterate over each protocol code to extract certhashes
	for _, protocol := range addr.Protocols() {
		// Check if the protocol represents a webtransport with certhash
		if protocol.Name == "webtransport" {
			// Get the certhash value for the webtransport protocol
			certhash, err := addr.ValueForProtocol(protocol.Code)
			if err == nil {
				certhashes = append(certhashes, certhash)
			}
		}
	}

	return certhashes
}

// PeerHandler returns an HTTP handler function that retrieves and serves peer information.
func PeersHandler(nodeCli *client.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		peers, err := nodeCli.P2P.Peers(ctx)
		if err != nil {
			zap.L().Error("failed to retrieve peers", zap.Error(err))
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
					zap.L().Error("failed to get peer info", zap.Error(err))
					return
				}

				for _, addr := range peerInfo.Addrs {
					maddr, err := multiaddr.NewMultiaddrBytes(addr.Bytes())
					if err != nil {
						zap.L().Error("failed to parse multiaddr", zap.Error(err))
						continue
					}
					if isWebtransportPeer(maddr) {
						mutex.Lock()
						// Currently have no idea how to apply /p2p/ better. /p2p/{peer_id} is necessary
						// to pass. Otherwise, it will start complaining that about incorrect multi-addr.
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
			zap.L().Error("failed to marshal response to JSON", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}
}

// Function to check if a peer is a secure IPv4 web transport, containing certhash. Has to be UDP too.
// WASM supports connectivity over go-libp2p only with these peers. DNS or IPv6 is not supported.
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

		if strings.Contains(addr.String(), "127.0.0.1") ||
			strings.Contains(addr.String(), "172.") {
			return false
		}

		zap.L().Info("Discovered WebTransport peer addr info", zap.Any("peer", addr.String()), zap.Any("protocol", addr.Protocols()))
		return true
	}

	return false
}
