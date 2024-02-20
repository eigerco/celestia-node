package nodebuilder

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DefaultAccountName = "my_celes_key"

// PrintKeyringInfo whether to print keyring information during init.
var PrintKeyringInfo = true

// GenerateKeys will construct a keyring from the given keystore path and check
// if account keys already exist. If not, it will generate a new account key and
// store it.
func GenerateKeys(ring keyring.Keyring) error {
	keys, err := ring.List()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		// at least one key is already present
		return nil
	}
	keyInfo, mn, err := generateNewKey(ring)
	if err != nil {
		return err
	}
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
