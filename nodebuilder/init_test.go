package nodebuilder_test

import (
	"github.com/celestiaorg/celestia-node/libs/codec"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"

	"github.com/celestiaorg/celestia-node/libs/fslock"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

func TestInit(t *testing.T) {
	dir := t.TempDir()
	nodes := []node.Type{node.Light, node.Bridge}
	kr := keyring.NewInMemory(encoding.MakeConfig(codec.ModuleEncodingRegisters...).Codec)
	for _, node := range nodes {
		cfg := nodebuilder.DefaultConfig(node)
		require.NoError(t, nodebuilder.Init(kr, *cfg, dir, node))
		assert.True(t, nodebuilder.IsInit(dir))
	}
}

func TestInitErrForInvalidPath(t *testing.T) {
	path := "/invalid_path"
	nodes := []node.Type{node.Light, node.Bridge}

	for _, node := range nodes {
		cfg := nodebuilder.DefaultConfig(node)
		require.Error(t, nodebuilder.Init(*cfg, path, node))
	}
}

func TestIsInitWithBrokenConfig(t *testing.T) {
	dir := t.TempDir()
	f, err := os.Create(filepath.Join(dir, "config.toml"))
	require.NoError(t, err)
	defer f.Close()
	//nolint:errcheck
	f.Write([]byte(`
		[P2P]
		  ListenAddresses = [/ip4/0.0.0.0/tcp/2121]
    `))
	assert.False(t, nodebuilder.IsInit(dir))
}

func TestIsInitForNonExistDir(t *testing.T) {
	path := "/invalid_path"
	assert.False(t, nodebuilder.IsInit(path))
}

func TestInitErrForLockedDir(t *testing.T) {
	dir := t.TempDir()
	flock, err := fslock.Lock(lockPath(dir))
	require.NoError(t, err)
	defer flock.Unlock() //nolint:errcheck
	nodes := []node.Type{node.Light, node.Bridge}
	kr := keyring.NewInMemory(encoding.MakeConfig(codec.ModuleEncodingRegisters...).Codec)
	for _, node := range nodes {
		cfg := nodebuilder.DefaultConfig(node)
		require.Error(t, nodebuilder.Init(kr, *cfg, dir, node))
	}
}

// TestInit_generateNewKey tests to ensure new account is generated
// correctly.
func TestInit_generateNewKey(t *testing.T) {
	cfg := nodebuilder.DefaultConfig(node.Bridge)

	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	ring, err := keyring.New(app.Name, cfg.State.KeyringBackend, t.TempDir(), os.Stdin, encConf.Codec)
	require.NoError(t, err)

	originalKey, mn, err := generateNewKey(ring)
	require.NoError(t, err)

	// check ring and make sure it generated + stored key
	keys, err := ring.List()
	require.NoError(t, err)
	assert.Equal(t, originalKey, keys[0])

	// ensure the generated account is actually a celestia account
	addr, err := originalKey.GetAddress()
	require.NoError(t, err)
	assert.Contains(t, addr.String(), "celestia")

	// ensure account is recoverable from mnemonic
	ring2, err := keyring.New(app.Name, cfg.State.KeyringBackend, t.TempDir(), os.Stdin, encConf.Codec)
	require.NoError(t, err)
	duplicateKey, err := ring2.NewAccount("test", mn, keyring.DefaultBIP39Passphrase, sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1)
	require.NoError(t, err)
	got, err := duplicateKey.GetAddress()
	require.NoError(t, err)
	assert.Equal(t, addr.String(), got.String())
}
