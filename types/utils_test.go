package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("epix", "epixpub")
}

func TestIsSupportedKeys(t *testing.T) {
	testCases := []struct {
		name        string
		pk          cryptotypes.PubKey
		isSupported bool
	}{
		{
			"nil key",
			nil,
			false,
		},
		{
			"ethsecp256k1 key",
			&ethsecp256k1.PubKey{},
			true,
		},
		{
			"ed25519 key",
			&ed25519.PubKey{},
			true,
		},
		{
			"multisig key - no pubkeys",
			&multisig.LegacyAminoPubKey{},
			false,
		},
		{
			"multisig key - valid pubkeys",
			multisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{&ed25519.PubKey{}, &ed25519.PubKey{}, &ed25519.PubKey{}}),
			true,
		},
		{
			"multisig key - nested multisig",
			multisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{&ed25519.PubKey{}, &ed25519.PubKey{}, &multisig.LegacyAminoPubKey{}}),
			false,
		},
		{
			"multisig key - invalid pubkey",
			multisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{&ed25519.PubKey{}, &ed25519.PubKey{}, &secp256k1.PubKey{}}),
			false,
		},
		{
			"cosmos secp256k1",
			&secp256k1.PubKey{},
			false,
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.isSupported, IsSupportedKey(tc.pk), tc.name)
	}
}

func TestGetepixAddressFromBech32(t *testing.T) {
	testCases := []struct {
		name       string
		address    string
		expAddress string
		expError   bool
	}{
		{
			"blank bech32 address",
			" ",
			"",
			true,
		},
		{
			"invalid bech32 address",
			"epix",
			"",
			true,
		},
		{
			"invalid address bytes",
			"epix1123",
			"",
			true,
		},
		{
			"epix address",
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			false,
		},
		{
			"cosmos address",
			"cosmos1x2lktda892nmlpjfrp0l7n6mzhe8mvfu3gd9v8",
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			false,
		},
		{
			"osmosis address",
			"osmo1x2lktda892nmlpjfrp0l7n6mzhe8mvfuen7464",
			"epix1x2lktda892nmlpjfrp0l7n6mzhe8mvfuyrrstu",
			false,
		},
	}

	for _, tc := range testCases {
		addr, err := GetepixAddressFromBech32(tc.address)
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expAddress, addr.String(), tc.name)
		}
	}
}

// Helper test to generate new addresses
func TestGenerateAddresses(t *testing.T) {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("epix", "epixpub")
	cfg.SetBech32PrefixForAccount("cosmos", "cosmospub")
	cfg.SetBech32PrefixForAccount("osmo", "osmopub")

	// Create a random address
	privKey := ed25519.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address())

	epixAddr := sdk.MustBech32ifyAddressBytes("epix", address)
	cosmosAddr := sdk.MustBech32ifyAddressBytes("cosmos", address)
	osmoAddr := sdk.MustBech32ifyAddressBytes("osmo", address)

	t.Logf("Generated addresses for the same key bytes:")
	t.Logf("epix address: %s", epixAddr)
	t.Logf("cosmos address: %s", cosmosAddr)
	t.Logf("osmo address: %s", osmoAddr)
}
