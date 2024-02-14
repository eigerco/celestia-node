//go:build wasm

package nodebuilder

import "github.com/cosmos/cosmos-sdk/crypto/keyring"

func InitWasm(ring keyring.Keyring, cfg Config, path string) error {
	path, err := storePath(path)
	if err != nil {
		return err
	}
	log.Infof("Initializing light Node Store over %q", path)

	err = initRoot(path)
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
	err = generateKeys(ring)
	if err != nil {
		log.Errorw("generating account keys", "err", err)
		return err
	}

	log.Info("Node Store initialized")
	return nil
}
