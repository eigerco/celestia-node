//go:build !wasm

package nodebuilder

import (
	"os"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-node/libs/fslock"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// Init initializes the Node FileSystem Store for the given Node Type 'tp' in the directory under
// 'path'.
func Init(cfg Config, path string, tp node.Type) error {
	path, err := storePath(path)
	if err != nil {
		return err
	}
	log.Infof("Initializing %s Node Store over '%s'", tp, path)

	err = initRoot(path)
	if err != nil {
		return err
	}

	flock, err := fslock.Lock(lockPath(path))
	if err != nil {
		if err == fslock.ErrLocked {
			return ErrOpened
		}
		return err
	}
	defer flock.Unlock() //nolint: errcheck

	ksPath := keysPath(path)
	err = initDir(ksPath)
	if err != nil {
		return err
	}

	err = initDir(dataPath(path))
	if err != nil {
		return err
	}

	cfgPath := configPath(path)
	err = SaveConfig(cfgPath, &cfg)
	if err != nil {
		return err
	}
	log.Infow("Saved config", "path", cfgPath)

	log.Infow("Accessing keyring @ nodebuilder.Init() ...")
	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	if cfg.State.KeyringBackend == keyring.BackendTest {
		log.Warn("Detected plaintext keyring backend. For elevated security properties, consider using" +
			" the `file` keyring backend.")
	}
	ring, err := keyring.New(app.Name, cfg.State.KeyringBackend, ksPath, os.Stdin, encConf.Codec)
	if err != nil {
		return err
	}
	err = GenerateKeys(ring)
	if err != nil {
		log.Errorw("generating account keys", "err", err)
		return err
	}

	log.Info("Node Store initialized")
	return nil
}
