package keeper_test

import (
	"fmt"
	"testing"
	"time"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/EpixZone/epix/app"
	"github.com/EpixZone/epix/x/inflation/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app         *app.Epix
	ctx         sdk.Context
	consAddress sdk.ConsAddress
	privKey     *ed25519.PrivKey
}

func TestKeeperIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Create test accounts
	suite.privKey = ed25519.GenPrivKey()
	priv2 := ed25519.GenPrivKey()
	acc1 := authtypes.NewBaseAccount(
		sdk.AccAddress(suite.privKey.PubKey().Address()),
		suite.privKey.PubKey(),
		0,
		0,
	)
	acc2 := authtypes.NewBaseAccount(
		sdk.AccAddress(priv2.PubKey().Address()),
		priv2.PubKey(),
		1,
		0,
	)

	// Create module accounts
	stakingAcc := authtypes.NewEmptyModuleAccount(stakingtypes.ModuleName, authtypes.Minter, authtypes.Burner)
	bondPool := authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName, authtypes.Burner, authtypes.Staking)
	notBondedPool := authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName, authtypes.Burner, authtypes.Staking)
	feeCollector := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName, authtypes.Minter, authtypes.Burner)
	inflationAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter)

	// Add all accounts to genesis
	genAccs := []authtypes.GenesisAccount{
		acc1, acc2, stakingAcc, bondPool, notBondedPool, feeCollector, inflationAcc,
	}

	// Set up initial balances
	bondedAmount := sdkmath.NewInt(1000000000000)
	balances := []banktypes.Balance{
		{
			Address: acc1.GetAddress().String(),
			Coins: sdk.NewCoins(
				sdk.NewCoin("aepix", sdkmath.NewInt(1000000000000000)),
			),
		},
	}

	// Calculate total supply
	totalSupply := sdk.NewCoins()
	for _, balance := range balances {
		totalSupply = totalSupply.Add(balance.Coins...)
	}

	// Initialize test app with genesis accounts
	suite.app = app.SetupWithGenesisAccounts(genAccs, balances...)

	header := tmtypes.Header{
		Height:  4,
		ChainID: "epix_1917-1",
		Time:    time.Now().UTC(),
	}
	suite.ctx = suite.app.BaseApp.NewContext(false)
	suite.ctx = suite.ctx.WithBlockHeader(header)

	// Fund the bonded pool account after context initialization
	var err error
	err = suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin("aepix", bondedAmount)))
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin("aepix", bondedAmount)))
	suite.Require().NoError(err)

	// Set validator consensus address
	suite.consAddress = sdk.ConsAddress(suite.privKey.PubKey().Address())

	// Setup test validator
	valAddr := sdk.ValAddress(suite.consAddress)
	validator, err := stakingtypes.NewValidator(valAddr.String(), suite.privKey.PubKey(), stakingtypes.Description{})
	suite.Require().NoError(err)

	// Set validator with tokens
	validator.Tokens = bondedAmount
	validator.DelegatorShares = sdkmath.LegacyNewDec(bondedAmount.Int64())
	validator.Status = stakingtypes.Bonded

	// Log initial state
	fmt.Printf("\n=== Initial State ===\n")
	fmt.Printf("Bonded Amount: %s\n", bondedAmount)
	fmt.Printf("Validator Tokens: %s\n", validator.Tokens)
	fmt.Printf("Validator Shares: %s\n", validator.DelegatorShares)
	fmt.Printf("Validator Status: %s\n", validator.Status)

	// Set staking module state
	stakingGenesis := stakingtypes.NewGenesisState(
		stakingtypes.DefaultParams(),
		[]stakingtypes.Validator{validator},
		[]stakingtypes.Delegation{},
	)
	stakingGenesis.Params.BondDenom = "aepix"

	// Create delegation
	delAddr := sdk.AccAddress(suite.privKey.PubKey().Address())
	delegation := stakingtypes.NewDelegation(delAddr.String(), valAddr.String(), sdkmath.LegacyNewDec(bondedAmount.Int64()))
	stakingGenesis.Delegations = []stakingtypes.Delegation{delegation}

	// Set last validator powers
	stakingGenesis.LastTotalPower = bondedAmount
	stakingGenesis.LastValidatorPowers = []stakingtypes.LastValidatorPower{
		{
			Address: valAddr.String(),
			Power:   bondedAmount.Int64(),
		},
	}

	// Initialize staking module
	suite.app.StakingKeeper.InitGenesis(suite.ctx, stakingGenesis)

	// Log state after staking initialization
	fmt.Printf("\n=== After Staking Init ===\n")
	bondedPool := suite.app.StakingKeeper.GetBondedPool(suite.ctx)
	bondedBalance := suite.app.BankKeeper.GetBalance(suite.ctx, bondedPool.GetAddress(), "aepix")
	fmt.Printf("Bonded Pool Balance: %s\n", bondedBalance.Amount)

	// Log validator state
	validatorAfter, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	fmt.Printf("Validator Tokens After: %s\n", validatorAfter.Tokens)
	fmt.Printf("Validator Shares After: %s\n", validatorAfter.DelegatorShares)
	fmt.Printf("Validator Status After: %s\n", validatorAfter.Status)

	// Log delegation state
	delegationAfter, _ := suite.app.StakingKeeper.GetDelegation(suite.ctx, delAddr, valAddr)
	fmt.Printf("Delegation Shares: %s\n", delegationAfter.Shares)

	// Set validator and power index after genesis
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
	suite.app.StakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	// Set up distribution module
	err = suite.app.DistrKeeper.SetValidatorHistoricalRewards(suite.ctx, valAddr, 0, distrtypes.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))
	suite.Require().NoError(err)
	err = suite.app.DistrKeeper.SetValidatorCurrentRewards(suite.ctx, valAddr, distrtypes.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))
	suite.Require().NoError(err)
	err = suite.app.DistrKeeper.SetValidatorAccumulatedCommission(suite.ctx, valAddr, distrtypes.ValidatorAccumulatedCommission{})
	suite.Require().NoError(err)

	// Set up delegator rewards
	err = suite.app.DistrKeeper.SetDelegatorStartingInfo(suite.ctx, valAddr, delAddr, distrtypes.NewDelegatorStartingInfo(1, sdkmath.LegacyNewDec(bondedAmount.Int64()), 1))
	suite.Require().NoError(err)

	// Log final state
	fmt.Printf("\n=== Final State ===\n")
	bondedBalanceFinal := suite.app.BankKeeper.GetBalance(suite.ctx, bondedPool.GetAddress(), "aepix")
	fmt.Printf("Final Bonded Pool Balance: %s\n", bondedBalanceFinal.Amount)

	validatorFinal, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	fmt.Printf("Final Validator Tokens: %s\n", validatorFinal.Tokens)
	fmt.Printf("Final Validator Shares: %s\n", validatorFinal.DelegatorShares)
	fmt.Printf("Final Validator Status: %s\n", validatorFinal.Status)

	// Set inflation parameters
	params := suite.app.InflationKeeper.GetParams(suite.ctx)
	params.EnableInflation = true
	params.InflationDistribution.StakingRewards = sdkmath.LegacyNewDecWithPrec(10, 2) // 10%
	params.InflationDistribution.CommunityPool = sdkmath.LegacyNewDecWithPrec(90, 2)  // 90%
	suite.app.InflationKeeper.SetParams(suite.ctx, params)

	// Set epoch identifier and epochs per period
	suite.app.InflationKeeper.SetEpochIdentifier(suite.ctx, "day")
	suite.app.InflationKeeper.SetEpochsPerPeriod(suite.ctx, 30) // 30 epochs per period

	// Calculate and set epoch mint provision
	bondedRatio, err := suite.app.StakingKeeper.BondedRatio(suite.ctx)
	suite.Require().NoError(err)
	epochMintProvision := types.CalculateEpochMintProvision(
		params,
		uint64(0), // period 0
		365,       // epochs per period
		bondedRatio,
	)
	suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, epochMintProvision)

	// Mint and allocate the coins
	mintedCoin := sdk.NewCoin(params.MintDenom, epochMintProvision.TruncateInt())
	staking, communityPool, err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, mintedCoin)
	suite.Require().NoError(err)

	// Log minting and allocation
	fmt.Printf("\n=== Minting and Allocation ===\n")
	fmt.Printf("Epoch Mint Provision: %s\n", epochMintProvision)
	fmt.Printf("Minted Coin: %s\n", mintedCoin)
	fmt.Printf("Staking Rewards: %s\n", staking)
	fmt.Printf("Community Pool: %s\n", communityPool)

	// Calculate validator rewards (10% of minted coins)
	validatorRewards := sdk.NewDecCoinFromCoin(mintedCoin).Amount.Mul(params.InflationDistribution.StakingRewards)
	validatorRewardCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(params.MintDenom, validatorRewards)}

	// Trigger reward distribution by allocating tokens to validator
	err = suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validatorFinal, validatorRewardCoins)
	suite.Require().NoError(err)

	// Log rewards state before verification
	fmt.Printf("\n=== Rewards State ===\n")
	rewards, err := suite.app.DistrKeeper.GetValidatorOutstandingRewards(suite.ctx, valAddr)
	suite.Require().NoError(err)
	fmt.Printf("Validator Outstanding Rewards: %s\n", rewards.Rewards)

	feePool, err := suite.app.DistrKeeper.FeePool.Get(suite.ctx)
	suite.Require().NoError(err)
	fmt.Printf("Community Pool Balance: %s\n", feePool.CommunityPool)

	// Verify reward distribution
	// Get validator rewards
	rewards, err = suite.app.DistrKeeper.GetValidatorOutstandingRewards(suite.ctx, valAddr)
	suite.Require().NoError(err)
	suite.Require().False(rewards.Rewards.IsZero(), "Validator rewards should not be zero")

	// Compare rewards with tolerance
	rewardsDiff := rewards.Rewards.AmountOf(params.MintDenom).Sub(staking[0].Amount.ToLegacyDec()).Abs()
	rewardsTolerance := sdkmath.LegacyNewDecWithPrec(1, 0) // Allow difference of 1 unit
	suite.Require().True(rewardsDiff.LTE(rewardsTolerance),
		"Validator rewards difference %s exceeds tolerance %s (expected: %s, got: %s)",
		rewardsDiff, rewardsTolerance, staking[0].Amount.ToLegacyDec(), rewards.Rewards.AmountOf(params.MintDenom))

	// Get community pool balance
	feePool, err = suite.app.DistrKeeper.FeePool.Get(suite.ctx)
	suite.Require().NoError(err)
	communityPoolBalance := feePool.CommunityPool

	// Calculate expected rewards
	expectedValidatorRewards := sdk.NewDecCoinFromCoin(mintedCoin).Amount.Mul(params.InflationDistribution.StakingRewards)
	expectedCommunityPool := sdk.NewDecCoinFromCoin(mintedCoin).Amount.Mul(params.InflationDistribution.CommunityPool)

	// Verify validator rewards (10%)
	suite.Require().Equal(expectedValidatorRewards, rewards.Rewards.AmountOf(params.MintDenom))

	// Verify community pool balance (90%) with tolerance for rounding
	actualCommunityPool := communityPoolBalance.AmountOf(params.MintDenom)
	difference := actualCommunityPool.Sub(expectedCommunityPool).Abs()
	tolerance := sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(1000)) // Allow small rounding difference
	suite.Require().True(difference.LTE(tolerance),
		"Community pool balance difference %s exceeds tolerance %s", difference, tolerance)

	// Verify distribution ratio (90/10) with tolerance
	totalRewards := rewards.Rewards.Add(communityPoolBalance...)
	validatorShare := rewards.Rewards.AmountOf(params.MintDenom).Quo(totalRewards.AmountOf(params.MintDenom))
	communityShare := communityPoolBalance.AmountOf(params.MintDenom).Quo(totalRewards.AmountOf(params.MintDenom))

	// Allow small rounding differences in ratios
	ratioTolerance := sdkmath.LegacyNewDecWithPrec(1, 6) // 0.0001% tolerance
	validatorDiff := validatorShare.Sub(params.InflationDistribution.StakingRewards).Abs()
	communityDiff := communityShare.Sub(params.InflationDistribution.CommunityPool).Abs()

	suite.Require().True(validatorDiff.LTE(ratioTolerance),
		"Validator share difference %s exceeds tolerance %s", validatorDiff, ratioTolerance)
	suite.Require().True(communityDiff.LTE(ratioTolerance),
		"Community share difference %s exceeds tolerance %s", communityDiff, ratioTolerance)

	// Verify staking rewards (10%)
	expectedStaking := sdk.NewCoin(params.MintDenom, mintedCoin.Amount.ToLegacyDec().Mul(params.InflationDistribution.StakingRewards).TruncateInt())
	stakingDiff := expectedStaking.Amount.Sub(staking[0].Amount).Abs()
	stakingTolerance := sdkmath.NewInt(1) // Allow difference of 1 due to rounding
	suite.Require().True(stakingDiff.LTE(stakingTolerance),
		"Staking rewards difference %s exceeds tolerance %s (expected: %s, got: %s)",
		stakingDiff, stakingTolerance, expectedStaking, staking[0])

	// Verify community pool (90%)
	expectedCommunityPoolCoin := sdk.NewCoin(params.MintDenom, mintedCoin.Amount.ToLegacyDec().Mul(params.InflationDistribution.CommunityPool).TruncateInt())
	communityDiffPool := expectedCommunityPoolCoin.Amount.Sub(communityPool[0].Amount).Abs()
	communityTolerance := sdkmath.NewInt(1) // Allow difference of 1 due to rounding
	suite.Require().True(communityDiffPool.LTE(communityTolerance),
		"Community pool difference %s exceeds tolerance %s (expected: %s, got: %s)",
		communityDiffPool, communityTolerance, expectedCommunityPoolCoin, communityPool[0])
}

func (suite *IntegrationTestSuite) TestRewardDistribution() {
	// Set up initial state
	suite.SetupTest()

	// Get initial state
	params := suite.app.InflationKeeper.GetParams(suite.ctx)

	// Calculate epoch mint provision
	bondedRatio, err := suite.app.StakingKeeper.BondedRatio(suite.ctx)
	suite.Require().NoError(err)
	epochMintProvision := types.CalculateEpochMintProvision(
		params,
		uint64(0), // period 0
		30,        // epochs per period
		bondedRatio,
	)

	// Mint and allocate coins
	mintedCoin := sdk.NewCoin(params.MintDenom, epochMintProvision.TruncateInt())
	staking, communityPool, err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, mintedCoin)
	suite.Require().NoError(err)
	suite.Require().Len(staking, 1, "Expected one staking coin")
	suite.Require().Len(communityPool, 1, "Expected one community pool coin")

	// Verify total allocation equals minted amount
	totalAllocated := staking[0].Amount.Add(communityPool[0].Amount)
	suite.Require().Equal(mintedCoin.Amount.String(), totalAllocated.String(),
		"Total allocated (%s) does not match minted amount (%s)",
		totalAllocated, mintedCoin.Amount)

	// Verify distribution matches parameters
	stakingRatio := staking[0].Amount.ToLegacyDec().Quo(mintedCoin.Amount.ToLegacyDec())
	communityRatio := communityPool[0].Amount.ToLegacyDec().Quo(mintedCoin.Amount.ToLegacyDec())

	suite.Require().Equal(params.InflationDistribution.StakingRewards.String(), stakingRatio.String(),
		"Staking ratio (%s) does not match parameter (%s)",
		stakingRatio, params.InflationDistribution.StakingRewards)

	suite.Require().Equal(params.InflationDistribution.CommunityPool.String(), communityRatio.String(),
		"Community pool ratio (%s) does not match parameter (%s)",
		communityRatio, params.InflationDistribution.CommunityPool)

	// Get validator rewards
	valAddr := sdk.ValAddress(suite.consAddress)
	rewards, err := suite.app.DistrKeeper.GetValidatorOutstandingRewards(suite.ctx, valAddr)
	suite.Require().NoError(err)
	suite.Require().False(rewards.Rewards.IsZero(), "Validator rewards should not be zero")

	// Get community pool balance
	feePool, err := suite.app.DistrKeeper.FeePool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().False(feePool.CommunityPool.IsZero(), "Community pool should not be zero")
}
