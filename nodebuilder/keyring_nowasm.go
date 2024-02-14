//go:build !wasm

package nodebuilder

import (
	"os"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	apptypes "github.com/celestiaorg/celestia-app/x/blob/types"
	kr "github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/celestiaorg/celestia-node/libs/keystore"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
)

const DefaultAccountName = "my_celes_key"

func InitKeyring(cfg *Config, path string) (kr.Keyring, error) {
	ksPath := keysPath(path)
	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	if cfg.State.KeyringBackend == kr.BackendTest {
		log.Warn("Detected plaintext keyring backend. For elevated security properties, consider using" +
			" the `file` keyring backend.")
	}
	ring, err := kr.New(app.Name, cfg.State.KeyringBackend, ksPath, os.Stdin, encConf.Codec)
	if err != nil {
		return nil, err
	}
	return ring, err
}

// KeyringSigner constructs a new keyring signer.
// NOTE: we construct keyring signer before constructing node for easier UX
// as having keyring-backend set to `file` prompts user for password.
func KeyringSigner(accName, backend string, ks keystore.Keystore, net p2p.Network) (*apptypes.KeyringSigner, error) {
	ring := ks.Keyring()
	var info *kr.Record
	// if custom accName provided, find key for that name
	if accName != "" {
		keyInfo, err := ring.Key(accName)
		if err != nil {
			log.Errorw("failed to find key by given name", "keyring.accname", accName)
			return nil, err
		}
		info = keyInfo
	} else {
		// use default key
		keyInfo, err := ring.Key(DefaultAccountName)
		if err != nil {
			log.Errorw("could not access key in keyring", "name", DefaultAccountName, "err", err)
			return nil, err
		}
		info = keyInfo
	}
	// construct signer using the default key found / generated above
	signer := apptypes.NewKeyringSigner(ring, info.Name, string(net))
	signerInfo := signer.GetSignerInfo()
	log.Infow("constructed keyring signer", "backend", backend, "path", ks.Path(),
		"key name", signerInfo.Name, "chain-id", string(net))

	return signer, nil
}
