package lighteds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/celestiaorg/celestia-app/pkg/wrapper"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/celestia-node/share/ipld"
	"github.com/celestiaorg/nmt"
	"github.com/celestiaorg/rsmt2d"
	"github.com/ipld/go-car"
)

// ReadEDS reads the first EDS quadrant (1/4) from an io.Reader CAR file.
// Only the first quadrant will be read, which represents the original data.
// The returned EDS is guaranteed to be full and valid against the DataRoot, otherwise ReadEDS
// errors.
func ReadEDS(ctx context.Context, r io.Reader, root share.DataHash) (eds *rsmt2d.ExtendedDataSquare, err error) {
	//_, span := tracer.Start(ctx, "read-eds")
	//defer func() {
	//	utils.SetStatusAndEnd(span, err)
	//}()

	carReader, err := car.NewCarReader(r)
	if err != nil {
		return nil, fmt.Errorf("share: reading car file: %w", err)
	}

	// car header includes both row and col roots in header
	odsWidth := len(carReader.Header.Roots) / 4
	odsSquareSize := odsWidth * odsWidth
	shares := make([][]byte, odsSquareSize)
	// the first quadrant is stored directly after the header,
	// so we can just read the first odsSquareSize blocks
	for i := 0; i < odsSquareSize; i++ {
		block, err := carReader.Next()
		if err != nil {
			return nil, fmt.Errorf("share: reading next car entry: %w", err)
		}
		// the stored first quadrant shares are wrapped with the namespace twice.
		// we cut it off here, because it is added again while importing to the tree below
		shares[i] = share.GetData(block.RawData())
	}

	// use proofs adder if provided, to cache collected proofs while recomputing the eds
	var opts []nmt.Option
	visitor := ipld.ProofsAdderFromCtx(ctx).VisitFn()
	if visitor != nil {
		opts = append(opts, nmt.NodeVisitor(visitor))
	}

	eds, err = rsmt2d.ComputeExtendedDataSquare(
		shares,
		share.DefaultRSMT2DCodec(),
		wrapper.NewConstructor(uint64(odsWidth), opts...),
	)
	if err != nil {
		return nil, fmt.Errorf("share: computing eds: %w", err)
	}

	newDah, err := share.NewRoot(eds)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(newDah.Hash(), root) {
		return nil, fmt.Errorf(
			"share: content integrity mismatch: imported root %s doesn't match expected root %s",
			newDah.Hash(),
			root,
		)
	}
	return eds, nil
}

type Parameters struct {
	// GC performs DAG store garbage collection by reclaiming transient files of
	// shards that are currently available but inactive, or errored.
	// We don't use transient files right now, so GC is turned off by default.
	GCInterval time.Duration

	// RecentBlocksCacheSize is the size of the cache for recent blocks.
	RecentBlocksCacheSize int

	// BlockstoreCacheSize is the size of the cache for blockstore requested accessors.
	BlockstoreCacheSize int
}

// DefaultParameters returns the default configuration values for the EDS store parameters.
func DefaultParameters() *Parameters {
	return &Parameters{
		GCInterval:            0,
		RecentBlocksCacheSize: 10,
		BlockstoreCacheSize:   128,
	}
}

func (p *Parameters) Validate() error {
	if p.GCInterval < 0 {
		return fmt.Errorf("eds: GC interval cannot be negative")
	}

	if p.RecentBlocksCacheSize < 1 {
		return fmt.Errorf("eds: recent blocks cache size must be positive")
	}

	if p.BlockstoreCacheSize < 1 {
		return fmt.Errorf("eds: blockstore cache size must be positive")
	}
	return nil
}
