package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global erc20 module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to modules/erc20 and
	// defined at the application level.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	// AminoCdc is a amino codec created to support amino JSON compatible msgs.
	AminoCdc = codec.NewAminoCodec(amino)
)

// NOTE: This is required for the GetSignBytes function
func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgConvertCoin{},
		&MsgConvertERC20{},
		&MsgUpdateParams{},
		&MsgRegisterCoin{},
		&MsgRegisterERC20{},
		&MsgToggleTokenConversion{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&RegisterCoinProposal{},
		&RegisterERC20Proposal{},
		&ToggleTokenConversionProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/erc20 interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization and EIP-712 compatibility.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterCoin{}, "epix/MsgRegisterCoin", nil)
	cdc.RegisterConcrete(&MsgRegisterERC20{}, "epix/MsgRegisterERC20", nil)
	cdc.RegisterConcrete(&MsgConvertCoin{}, "epix/MsgConvertCoin", nil)
	cdc.RegisterConcrete(&MsgConvertERC20{}, "epix/MsgConvertERC20", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "epix/x/erc20/MsgUpdateParams", nil)
	cdc.RegisterConcrete(&Params{}, "epix/x/erc20/Params", nil)
}
