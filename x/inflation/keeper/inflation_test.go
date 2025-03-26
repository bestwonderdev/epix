package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/EpixZone/epix/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ethermint "github.com/evmos/ethermint/types"
)

func (suite *KeeperTestSuite) TestMintAndAllocateInflation() {
	testCases := []struct {
		name                string
		mintCoin            sdk.Coin
		malleate            func()
		expStakingRewardAmt sdk.Coin
		expCommunityPoolAmt sdk.DecCoins
		expPass             bool
	}{
		{
			"pass",
			sdk.NewCoin(denomMint, sdkmath.NewInt(1_000_000)),
			func() {},
			sdk.NewCoin(denomMint, sdkmath.NewInt(100_000)),                     // 10% to staking rewards
			sdk.NewDecCoins(sdk.NewDecCoin(denomMint, sdkmath.NewInt(900_000))), // 90% to community pool
			true,
		},
		{
			"pass - no coins minted ",
			sdk.NewCoin(denomMint, sdkmath.ZeroInt()),
			func() {},
			sdk.NewCoin(denomMint, sdkmath.ZeroInt()),
			sdk.DecCoins(nil),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			_, _, err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, tc.mintCoin)

			// Get balances
			balanceModule := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				suite.app.AccountKeeper.GetModuleAddress(types.ModuleName),
				denomMint,
			)

			feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			balanceStakingRewards := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				feeCollector,
				denomMint,
			)

			feePool, err := s.app.DistrKeeper.FeePool.Get(s.ctx)
			s.Require().NoError(err)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().True(balanceModule.IsZero())
				suite.Require().Equal(tc.expStakingRewardAmt, balanceStakingRewards)
				suite.Require().Equal(tc.expCommunityPoolAmt, feePool.CommunityPool)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetCirculatingSupplyAndInflationRate() {
	testCases := []struct {
		name             string
		bankSupply       int64
		malleate         func()
		expInflationRate sdkmath.LegacyDec
	}{
		{
			"no mint provision",
			23_689_538,
			func() {
				suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, sdkmath.LegacyZeroDec())
			},
			sdkmath.LegacyZeroDec(),
		},
		{
			"genesis supply",
			23_689_538,
			func() {},
			sdkmath.LegacyMustNewDecFromStr("12.241690825713865800"), // 12.24% initial inflation rate (2.9M/23.69M)
		},
		{
			"year 1 supply",
			26_589_538, // genesis + 2.9M
			func() {},
			sdkmath.LegacyMustNewDecFromStr("10.906545273558344600"), // 10.9% inflation rate after 1 year
		},
		{
			"year 2 supply",
			29_028_438, //  + 2.44M
			func() {},
			sdkmath.LegacyMustNewDecFromStr("9.990203399852241400"), // 9.99% inflation rate after 2 year
		},
		{
			"year 3 supply",
			30_907_710, // + 2.05M
			func() {},
			sdkmath.LegacyMustNewDecFromStr("9.382772130319586900"), // 9.38% inflation rate after 3 year
		},
		{
			"year 4 supply (after first halving)",
			32_420_524, // + 1.725M (first 4 years)
			func() {},
			sdkmath.LegacyMustNewDecFromStr("8.944951043974489700"), // 8.945% inflation rate after halving
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Team allocation is only set on mainnet
			tc.malleate()
			// Mint coins to increase supply
			coin := sdk.NewCoin(types.DefaultInflationDenom, sdk.TokensFromConsensusPower(tc.bankSupply, ethermint.PowerReduction))
			decCoin := sdk.NewDecCoinFromCoin(coin)
			err := suite.app.InflationKeeper.MintCoins(suite.ctx, coin)
			suite.Require().NoError(err)

			circulatingSupply := s.app.InflationKeeper.GetCirculatingSupply(suite.ctx)
			suite.Require().Equal(decCoin.Amount, circulatingSupply)

			inflationRate := s.app.InflationKeeper.GetInflationRate(suite.ctx)
			suite.Require().Equal(tc.expInflationRate, inflationRate)
		})
	}
}
