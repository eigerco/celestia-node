package p2p_test

import (
	"context"
	"testing"

	"github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/celestiaorg/celestia-node/libs/codec"
	"github.com/celestiaorg/celestia-node/libs/keystore"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
)

func testModule(tp node.Type) fx.Option {
	cfg := p2p.DefaultConfig(tp)
	constructMod, err := p2p.ConstructModule(tp, &cfg)
	if err != nil {
		panic(err)
	}
	// TODO(@Wondertan): Most of these can be deduplicated
	//  by moving Store into the modnode and introducing there a TestModNode module
	//  that testers would import
	return fx.Options(
		fx.NopLogger,
		constructMod,
		fx.Provide(context.Background),
		fx.Supply(p2p.Private),
		fx.Supply(p2p.Bootstrappers{}),
		fx.Supply(tp),
		fx.Provide(keystore.NewMapKeystore(codec.ModuleEncodingRegisters...)),
		fx.Supply(fx.Annotate(ds_sync.MutexWrap(datastore.NewMapDatastore()), fx.As(new(datastore.Batching)))),
	)
}

func TestModuleBuild(t *testing.T) {
	var test = []struct {
		tp node.Type
	}{
		{tp: node.Bridge},
		{tp: node.Full},
		{tp: node.Light},
	}

	for _, tt := range test {
		t.Run(tt.tp.String(), func(t *testing.T) {
			app := fxtest.New(t, testModule(tt.tp))
			app.RequireStart()
			app.RequireStop()
		})
	}
}

func TestModuleBuild_WithMetrics(t *testing.T) {
	var test = []struct {
		tp node.Type
	}{
		{tp: node.Full},
		{tp: node.Bridge},
		{tp: node.Light},
	}

	for _, tt := range test {
		t.Run(tt.tp.String(), func(t *testing.T) {
			app := fxtest.New(t, testModule(tt.tp), p2p.WithMetrics())
			app.RequireStart()
			app.RequireStop()
		})
	}
}
