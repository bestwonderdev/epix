package ibc

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("epix", "epixpub")
}

func TestGetTransferSenderRecipient(t *testing.T) {
	testCases := []struct {
		name         string
		packet       channeltypes.Packet
		expSender    string
		expRecipient string
		expError     bool
	}{
		{
			"empty packet",
			channeltypes.Packet{},
			"", "",
			true,
		},
		{
			"invalid packet data",
			channeltypes.Packet{
				Data: ibctesting.MockFailPacketData,
			},
			"", "",
			true,
		},
		{
			"empty FungibleTokenPacketData",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{},
				),
			},
			"", "",
			true,
		},
		{
			"invalid sender",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1",
						Receiver: "epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
						Amount:   "123456",
					},
				),
			},
			"", "",
			true,
		},
		{
			"invalid recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1x2lktda892nmlpjfrp0l7n6mzhe8mvfu3gd9v8",
						Receiver: "epix1",
						Amount:   "123456",
					},
				),
			},
			"", "",
			true,
		},
		{
			"valid - cosmos sender, epix recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1x2lktda892nmlpjfrp0l7n6mzhe8mvfu3gd9v8",
						Receiver: "epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
						Amount:   "123456",
					},
				),
			},
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			false,
		},
		{
			"valid - epix sender, cosmos recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
						Receiver: "cosmos1x2lktda892nmlpjfrp0l7n6mzhe8mvfu3gd9v8",
						Amount:   "123456",
					},
				),
			},
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			false,
		},
		{
			"valid - osmosis sender, epix recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "osmo1x2lktda892nmlpjfrp0l7n6mzhe8mvfuen7464",
						Receiver: "epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
						Amount:   "123456",
					},
				),
			},
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			false,
		},
	}

	for _, tc := range testCases {
		sender, recipient, _, _, err := GetTransferSenderRecipient(tc.packet)

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expSender, sender.String())
			require.Equal(t, tc.expRecipient, recipient.String())
		}
	}
}

func TestGetTransferAmount(t *testing.T) {
	testCases := []struct {
		name      string
		packet    channeltypes.Packet
		expAmount string
		expError  bool
	}{
		{
			"empty packet",
			channeltypes.Packet{},
			"",
			true,
		},
		{
			"invalid packet data",
			channeltypes.Packet{
				Data: ibctesting.MockFailPacketData,
			},
			"",
			true,
		},
		{
			"invalid amount - empty",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1x2lktda892nmlpjfrp0l7n6mzhe8mvfu3gd9v8",
						Receiver: "epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
						Amount:   "",
					},
				),
			},
			"",
			true,
		},
		{
			"invalid amount - non-int",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1x2lktda892nmlpjfrp0l7n6mzhe8mvfu3gd9v8",
						Receiver: "epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
						Amount:   "test",
					},
				),
			},
			"test",
			true,
		},
		{
			"valid",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1x2lktda892nmlpjfrp0l7n6mzhe8mvfu3gd9v8",
						Receiver: "epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
						Amount:   "10000",
					},
				),
			},
			"10000",
			false,
		},
	}

	for _, tc := range testCases {
		amt, err := GetTransferAmount(tc.packet)
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expAmount, amt)
		}
	}
}
