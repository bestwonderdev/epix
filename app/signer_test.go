package app

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	coinswapv1 "github.com/EpixZone/epix/api/epix/coinswap/v1"
	erc20v1 "github.com/EpixZone/epix/api/epix/erc20/v1"
	coinswaptypes "github.com/EpixZone/epix/x/coinswap/types"
	erc20types "github.com/EpixZone/epix/x/erc20/types"
)

func TestDefineCustomGetSigners(t *testing.T) {
	addr := "epix17xpfvakm2amg962yls6f84z3kell8c5l9jd4vx"

	// Debug: Print Bech32 prefix and address details
	hrp, bz, err := bech32.DecodeAndConvert(addr)
	require.NoError(t, err)
	fmt.Printf("Debug - Bech32 details:\n")
	fmt.Printf("  HRP (prefix): %s\n", hrp)
	fmt.Printf("  Address bytes: %x\n", bz)

	// Debug: Print SDK config details
	config := sdk.GetConfig()
	fmt.Printf("Debug - SDK Config:\n")
	fmt.Printf("  Expected prefix: %s\n", config.GetBech32AccountAddrPrefix())

	accAddr, err := sdk.AccAddressFromBech32(addr)
	require.NoError(t, err)
	fmt.Printf("Debug - Converted address:\n")
	fmt.Printf("  Bytes: %x\n", accAddr.Bytes())
	fmt.Printf("  String: %s\n", accAddr.String())

	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
	}
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&erc20v1.MsgConvertERC20{}), erc20types.GetSignersFromMsgConvertERC20V2)
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&coinswapv1.MsgSwapOrder{}), coinswaptypes.CreateGetSignersFromMsgSwapOrderV2(&signingOptions))

	ctx, err := signing.NewContext(signingOptions)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     protov2.Message
		want    [][]byte
		wantErr bool
	}{
		{
			name: "MsgConvertERC20",
			msg: &erc20v1.MsgConvertERC20{
				Sender: common.BytesToAddress(accAddr.Bytes()).String(),
			},
			want: [][]byte{accAddr.Bytes()},
		},
		{
			name: "MsgSwapOrder",
			msg: &coinswapv1.MsgSwapOrder{
				Input: &coinswapv1.Input{Address: addr},
			},
			want: [][]byte{accAddr.Bytes()},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signers, err := ctx.GetSigners(test.msg)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.want, signers)
		})
	}
}
