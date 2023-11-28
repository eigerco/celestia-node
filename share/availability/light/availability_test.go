package light_test

import (
	"context"
	_ "embed"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share/availability/light"
	"github.com/celestiaorg/celestia-node/share/getters"
	"github.com/ipfs/boxo/blockservice"
	"github.com/ipfs/go-datastore"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/header/headertest"
	"github.com/celestiaorg/celestia-node/share"
	availability_test "github.com/celestiaorg/celestia-node/share/availability/test"
	"github.com/celestiaorg/celestia-node/share/ipld"
	"github.com/celestiaorg/celestia-node/share/sharetest"
)

func TestSharesAvailableCaches(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getter, eh := GetterWithRandSquare(t, 16)
	dah := eh.DAH
	avail := AvailabilityTest(getter)

	// cache doesn't have dah yet
	has, err := avail.DS.Has(ctx, light.RootKey(dah))
	assert.NoError(t, err)
	assert.False(t, has)

	err = avail.SharesAvailable(ctx, eh)
	assert.NoError(t, err)

	// is now cached
	has, err = avail.DS.Has(ctx, light.RootKey(dah))
	assert.NoError(t, err)
	assert.True(t, has)
}

func TestSharesAvailableHitsCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getter, _ := GetterWithRandSquare(t, 16)
	avail := AvailabilityTest(getter)

	bServ := ipld.NewMemBlockservice()
	dah := availability_test.RandFillBS(t, 16, bServ)
	eh := headertest.RandExtendedHeaderWithRoot(t, dah)

	// blockstore doesn't actually have the dah
	err := avail.SharesAvailable(ctx, eh)
	require.Error(t, err)

	// cache doesn't have dah yet, since it errored
	has, err := avail.DS.Has(ctx, light.RootKey(dah))
	assert.NoError(t, err)
	assert.False(t, has)

	err = avail.DS.Put(ctx, light.RootKey(dah), []byte{})
	require.NoError(t, err)

	// should hit cache after putting
	err = avail.SharesAvailable(ctx, eh)
	require.NoError(t, err)
}

func TestSharesAvailableEmptyRoot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getter, _ := GetterWithRandSquare(t, 16)
	avail := AvailabilityTest(getter)

	eh := headertest.RandExtendedHeaderWithRoot(t, share.EmptyRoot())
	err := avail.SharesAvailable(ctx, eh)
	assert.NoError(t, err)
}

func TestSharesAvailable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getter, dah := GetterWithRandSquare(t, 16)
	avail := AvailabilityTest(getter)
	err := avail.SharesAvailable(ctx, dah)
	assert.NoError(t, err)
}

func TestSharesAvailableFailed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bServ := ipld.NewMemBlockservice()
	dah := availability_test.RandFillBS(t, 16, bServ)
	eh := headertest.RandExtendedHeaderWithRoot(t, dah)

	getter, _ := GetterWithRandSquare(t, 16)
	avail := AvailabilityTest(getter)
	err := avail.SharesAvailable(ctx, eh)
	assert.Error(t, err)
}

func TestShareAvailableOverMocknet_Light(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	net := availability_test.NewTestDAGNet(ctx, t)
	_, root := RandNode(net, 16)
	eh := headertest.RandExtendedHeader(t)
	eh.DAH = root
	nd := Node(net)
	net.ConnectAll()

	err := nd.SharesAvailable(ctx, eh)
	assert.NoError(t, err)
}

func TestGetShare(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 16
	getter, eh := GetterWithRandSquare(t, n)

	for i := range make([]bool, n) {
		for j := range make([]bool, n) {
			sh, err := getter.GetShare(ctx, eh, i, j)
			assert.NotNil(t, sh)
			assert.NoError(t, err)
		}
	}
}

func TestService_GetSharesByNamespace(t *testing.T) {
	var tests = []struct {
		squareSize         int
		expectedShareCount int
	}{
		{squareSize: 4, expectedShareCount: 2},
		{squareSize: 16, expectedShareCount: 2},
		{squareSize: 128, expectedShareCount: 2},
	}

	for _, tt := range tests {
		t.Run("size: "+strconv.Itoa(tt.squareSize), func(t *testing.T) {
			getter, bServ := EmptyGetter()
			totalShares := tt.squareSize * tt.squareSize
			randShares := sharetest.RandShares(t, totalShares)
			idx1 := (totalShares - 1) / 2
			idx2 := totalShares / 2
			if tt.expectedShareCount > 1 {
				// make it so that two rows have the same namespace
				copy(share.GetNamespace(randShares[idx2]), share.GetNamespace(randShares[idx1]))
			}
			root := availability_test.FillBS(t, bServ, randShares)
			eh := headertest.RandExtendedHeader(t)
			eh.DAH = root
			randNamespace := share.GetNamespace(randShares[idx1])

			shares, err := getter.GetSharesByNamespace(context.Background(), eh, randNamespace)
			require.NoError(t, err)
			require.NoError(t, shares.Verify(root, randNamespace))
			flattened := shares.Flatten()
			assert.Len(t, flattened, tt.expectedShareCount)
			for _, value := range flattened {
				assert.Equal(t, randNamespace, share.GetNamespace(value))
			}
			if tt.expectedShareCount > 1 {
				// idx1 is always smaller than idx2
				assert.Equal(t, randShares[idx1], flattened[0])
				assert.Equal(t, randShares[idx2], flattened[1])
			}
		})
		t.Run("last two rows of a 4x4 square that have the same namespace have valid NMT proofs", func(t *testing.T) {
			squareSize := 4
			totalShares := squareSize * squareSize
			getter, bServ := EmptyGetter()
			randShares := sharetest.RandShares(t, totalShares)
			lastNID := share.GetNamespace(randShares[totalShares-1])
			for i := totalShares / 2; i < totalShares; i++ {
				copy(share.GetNamespace(randShares[i]), lastNID)
			}
			root := availability_test.FillBS(t, bServ, randShares)
			eh := headertest.RandExtendedHeader(t)
			eh.DAH = root

			shares, err := getter.GetSharesByNamespace(context.Background(), eh, lastNID)
			require.NoError(t, err)
			require.NoError(t, shares.Verify(root, lastNID))
		})
	}
}

func TestGetShares(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 16
	getter, eh := GetterWithRandSquare(t, n)

	eds, err := getter.GetEDS(ctx, eh)
	require.NoError(t, err)
	gotDAH, err := share.NewRoot(eds)
	require.NoError(t, err)

	require.True(t, eh.DAH.Equals(gotDAH))
}

func TestService_GetSharesByNamespaceNotFound(t *testing.T) {
	getter, eh := GetterWithRandSquare(t, 1)
	eh.DAH.RowRoots = nil

	emptyShares, err := getter.GetSharesByNamespace(context.Background(), eh, sharetest.RandV0Namespace())
	require.NoError(t, err)
	require.Empty(t, emptyShares.Flatten())
}

func BenchmarkService_GetSharesByNamespace(b *testing.B) {
	var tests = []struct {
		amountShares int
	}{
		{amountShares: 4},
		{amountShares: 16},
		{amountShares: 128},
	}

	for _, tt := range tests {
		b.Run(strconv.Itoa(tt.amountShares), func(b *testing.B) {
			t := &testing.T{}
			getter, eh := GetterWithRandSquare(t, tt.amountShares)
			root := eh.DAH
			randNamespace := root.RowRoots[(len(root.RowRoots)-1)/2][:share.NamespaceSize]
			root.RowRoots[(len(root.RowRoots) / 2)] = root.RowRoots[(len(root.RowRoots)-1)/2]
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := getter.GetSharesByNamespace(context.Background(), eh, randNamespace)
				require.NoError(t, err)
			}
		})
	}
}

// GetterWithRandSquare provides a share.Getter filled with 'n' NMT trees of 'n' random shares,
// essentially storing a whole square.
func GetterWithRandSquare(t *testing.T, n int) (share.Getter, *header.ExtendedHeader) {
	bServ := ipld.NewMemBlockservice()
	getter := getters.NewIPLDGetter(bServ)
	root := availability_test.RandFillBS(t, n, bServ)
	eh := headertest.RandExtendedHeader(t)
	eh.DAH = root

	return getter, eh
}

// EmptyGetter provides an unfilled share.Getter with corresponding blockservice.BlockService than
// can be filled by the test.
func EmptyGetter() (share.Getter, blockservice.BlockService) {
	bServ := ipld.NewMemBlockservice()
	getter := getters.NewIPLDGetter(bServ)
	return getter, bServ
}

// RandNode creates a Light Node filled with a random block of the given size.
func RandNode(dn *availability_test.TestDagNet, squareSize int) (*availability_test.TestNode, *share.Root) {
	nd := Node(dn)
	return nd, availability_test.RandFillBS(dn.T, squareSize, nd.BlockService)
}

// Node creates a new empty Light Node.
func Node(dn *availability_test.TestDagNet) *availability_test.TestNode {
	nd := dn.NewTestNode()
	nd.Getter = getters.NewIPLDGetter(nd.BlockService)
	nd.Availability = AvailabilityTest(nd.Getter)
	return nd
}

func AvailabilityTest(getter share.Getter) *light.ShareAvailability {
	ds := datastore.NewMapDatastore()
	return light.NewShareAvailability(getter, ds)
}

func SubNetNode(sn *availability_test.SubNet) *availability_test.TestNode {
	nd := Node(sn.TestDagNet)
	sn.AddNode(nd)
	return nd
}
