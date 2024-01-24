package keystore_test

import (
	"testing"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/libs/keystore"
)

func TestMapKeystore(t *testing.T) {
	kstore := keystore.NewMapKeystore(app.ModuleEncodingRegisters...)
	err := kstore.Put("test", keystore.PrivKey{Body: []byte("test_private_key")})
	require.NoError(t, err)

	key, err := kstore.Get("test")
	require.NoError(t, err)
	assert.Equal(t, []byte("test_private_key"), key.Body)

	keys, err := kstore.List()
	require.NoError(t, err)
	assert.Len(t, keys, 1)

	err = kstore.Delete("test")
	require.NoError(t, err)

	keys, err = kstore.List()
	require.NoError(t, err)
	assert.Len(t, keys, 0)
}
