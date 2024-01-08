package nodebuilder

import (
	"fmt"
	"path/filepath"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celestiaorg/celestia-node/libs/fslock"
	"github.com/celestiaorg/celestia-node/libs/utils"
)

// PrintKeyringInfo whether to print keyring information during init.
var PrintKeyringInfo = true

// Init initializes the Node FileSystem Store for the given Node Type 'tp' in the directory under
// 'path'.
func Init(ring keyring.Keyring, cfg Config, path string, tp node.Type) error {
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
	err = generateKeys(ring)
	if err != nil {
		log.Errorw("generating account keys", "err", err)
		return err
	}

	log.Info("Node Store initialized")
	return nil
}

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

// Reset removes all data from the datastore and dagstore directories. It leaves the keystore and
// config intact.
func Reset(path string, tp node.Type) error {
	path, err := storePath(path)
	if err != nil {
		return err
	}
	log.Infof("Resetting %s Node Store over '%s'", tp, path)

	flock, err := fslock.Lock(lockPath(path))
	if err != nil {
		if err == fslock.ErrLocked {
			return ErrOpened
		}
		return err
	}
	defer flock.Unlock() //nolint: errcheck

	err = resetDir(dataPath(path))
	if err != nil {
		return err
	}

	// light nodes don't have dagstore paths
	if tp == node.Light {
		log.Info("Node Store reset")
		return nil
	}

	err = resetDir(blocksPath(path))
	if err != nil {
		return err
	}

	err = resetDir(transientsPath(path))
	if err != nil {
		return err
	}

	err = resetDir(indexPath(path))
	if err != nil {
		return err
	}

	log.Info("Node Store reset")
	return nil
}

// IsInit checks whether FileSystem Store was setup under given 'path'.
// If any required file/subdirectory does not exist, then false is reported.
func IsInit(path string) bool {
	path, err := storePath(path)
	if err != nil {
		log.Errorw("parsing store path", "path", path, "err", err)
		return false
	}
	log.Errorw("store path ", "path", path)
	_, err = LoadConfig(configPath(path)) // load the Config and implicitly check for its existence
	if err != nil {
		log.Errorw("loading config", "path", path, "err", err)
		return false
	}

	if utils.Exists(fs, keysPath(path)) &&
		utils.Exists(fs, dataPath(path)) {
		return true
	}

	return false
}

const perms = 0755

// initRoot initializes(creates) directory if not created and check if it is writable
func initRoot(path string) error {
	err := initDir(path)
	if err != nil {
		return err
	}

	// check for writing permissions
	f, err := fs.Create(filepath.Join(path, ".check"))
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return fs.Remove(f.Name())
}

// resetDir removes all files from the given directory and reinitializes it
func resetDir(path string) error {
	err := fs.RemoveAll(path)
	if err != nil {
		return err
	}
	return initDir(path)
}

// initDir creates a dir if not exist
func initDir(path string) error {
	if utils.Exists(fs, path) {
		return nil
	}
	return fs.Mkdir(path, perms)
}

// generateKeys will construct a keyring from the given keystore path and check
// if account keys already exist. If not, it will generate a new account key and
// store it.
func generateKeys(ring keyring.Keyring) error {
	keys, err := ring.List()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		// at least one key is already present
		return nil
	}
	log.Infow("NO KEY FOUND IN STORE, GENERATING NEW KEY...")
	keyInfo, mn, err := generateNewKey(ring)
	if err != nil {
		return err
	}
	log.Info("NEW KEY GENERATED...")
	addr, err := keyInfo.GetAddress()
	if err != nil {
		return err
	}
	if PrintKeyringInfo {
		fmt.Printf("\nNAME: %s\nADDRESS: %s\nMNEMONIC (save this somewhere safe!!!): \n%s\n\n",
			keyInfo.Name, addr.String(), mn)
	}
	return nil
}

// generateNewKey generates and returns a new key on the given keyring called
// "my_celes_key".
func generateNewKey(ring keyring.Keyring) (*keyring.Record, string, error) {
	return ring.NewMnemonic(DefaultAccountName, keyring.English, sdk.GetConfig().GetFullBIP44Path(),
		keyring.DefaultBIP39Passphrase, hd.Secp256k1)
}
