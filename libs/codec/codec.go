package codec

import (
	"github.com/celestiaorg/celestia-app/app/encoding"
	blobtypes "github.com/celestiaorg/celestia-app/x/blob/types"
	qgbtypes "github.com/celestiaorg/celestia-app/x/qgb/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/core/types"
)

var ModuleEncodingRegisters = []encoding.ModuleRegister{
	newGenericRegister(authtypes.RegisterLegacyAminoCodec, authtypes.RegisterInterfaces),
	newGenericRegister(banktypes.RegisterLegacyAminoCodec, banktypes.RegisterInterfaces),
	newGenericRegister(stakingtypes.RegisterLegacyAminoCodec, stakingtypes.RegisterInterfaces),
	genericRegister{
		registerLegacyAminoCodec: func(amino *codec.LegacyAmino) {
			v1.RegisterLegacyAminoCodec(amino)
			v1beta1.RegisterLegacyAminoCodec(amino)
		},
		registerInterfaces: func(registry types.InterfaceRegistry) {
			v1.RegisterInterfaces(registry)
			v1beta1.RegisterInterfaces(registry)
		},
	},
	newGenericRegister(proposal.RegisterLegacyAminoCodec, proposal.RegisterInterfaces), //params
	newGenericRegister(crisistypes.RegisterLegacyAminoCodec, crisistypes.RegisterInterfaces),
	newGenericRegister(slashingtypes.RegisterLegacyAminoCodec, slashingtypes.RegisterInterfaces),
	newGenericRegister(authz.RegisterLegacyAminoCodec, authz.RegisterInterfaces),
	newGenericRegister(feegrant.RegisterLegacyAminoCodec, feegrant.RegisterInterfaces),
	newGenericRegister(func(amino *codec.LegacyAmino) {}, ibctypes.RegisterInterfaces),
	newGenericRegister(evidencetypes.RegisterLegacyAminoCodec, evidencetypes.RegisterInterfaces),
	newGenericRegister(transfertypes.RegisterLegacyAminoCodec, transfertypes.RegisterInterfaces),
	newGenericRegister(vestingtypes.RegisterLegacyAminoCodec, vestingtypes.RegisterInterfaces),
	newGenericRegister(blobtypes.RegisterLegacyAminoCodec, blobtypes.RegisterInterfaces),
	newGenericRegister(qgbtypes.RegisterLegacyAminoCodec, qgbtypes.RegisterInterfaces),
}

func newGenericRegister(registerLegacyAminoCodec func(amino *codec.LegacyAmino), registerInterfaces func(registry types.InterfaceRegistry)) genericRegister {
	return genericRegister{
		registerLegacyAminoCodec: registerLegacyAminoCodec,
		registerInterfaces:       registerInterfaces,
	}
}

type genericRegister struct {
	registerLegacyAminoCodec func(amino *codec.LegacyAmino)
	registerInterfaces       func(registry types.InterfaceRegistry)
}

func (g genericRegister) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	g.registerLegacyAminoCodec(amino)
}

func (g genericRegister) RegisterInterfaces(registry types.InterfaceRegistry) {
	g.registerInterfaces(registry)
}
